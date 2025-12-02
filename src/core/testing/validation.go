package testing

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/contracts"
)

// NOTE: ValidTags map has been removed - validation now uses tag contract
// See LoadTagContract() and IsValidTag() for contract-based validation

// ValidTestTypes defines the valid test types.
// Unit tests: gotest (Go), mocha (npm)
// Spec tests: godog (Go), tscucumber (npm)
var ValidTestTypes = []string{"gotest", "godog", "mocha", "tscucumber"}

// GetLevelTags returns taxonomy level tags from config.
// Returns nil if config is unavailable - callers must handle this.
func GetLevelTags() []string {
	cfg := config.Global()
	if cfg == nil || cfg.TestingTags == nil {
		return nil
	}
	return cfg.TestingTags.GetTaxonomyLevelTags()
}

// GetLevelTagsFromConfig returns taxonomy level tags from a pre-loaded config.
func GetLevelTagsFromConfig(cfg *config.EACConfig) []string {
	if cfg == nil || cfg.TestingTags == nil {
		return nil
	}
	return cfg.TestingTags.GetTaxonomyLevelTags()
}

// GetVerificationTags returns verification type tags from config.
// Returns nil if config is unavailable - callers must handle this.
func GetVerificationTags() []string {
	cfg := config.Global()
	if cfg == nil || cfg.TestingTags == nil {
		return nil
	}
	return cfg.TestingTags.GetVerificationTags()
}

// GetVerificationTagsFromConfig returns verification type tags from a pre-loaded config.
func GetVerificationTagsFromConfig(cfg *config.EACConfig) []string {
	if cfg == nil || cfg.TestingTags == nil {
		return nil
	}
	return cfg.TestingTags.GetVerificationTags()
}

// ValidateTags checks if tags are valid and don't conflict
func ValidateTags(tags []string) []string {
	errors := []string{}

	// Load config once for all tag validation
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		errors = append(errors, fmt.Sprintf("cannot validate tags: failed to load config (%v)", err))
		return errors
	}

	// Check for invalid tags
	for _, tag := range tags {
		if !IsValidTagWithConfig(tag, cfg) {
			errors = append(errors, fmt.Sprintf("tag %s is not defined in %s/testing-tags.yml", tag, contracts.EACConfigRelPath))
		}
	}

	// Check for multiple level tags
	levelCount := 0
	levelTags := GetLevelTagsFromConfig(cfg)
	for _, tag := range tags {
		if contains(levelTags, tag) {
			levelCount++
		}
	}
	if levelCount > 1 {
		errors = append(errors, "test has multiple level tags (only one allowed)")
	}

	// Check for multiple verification tags
	verificationCount := 0
	verificationTags := GetVerificationTagsFromConfig(cfg)
	for _, tag := range tags {
		if contains(verificationTags, tag) {
			verificationCount++
		}
	}
	if verificationCount > 1 {
		errors = append(errors, "test has multiple verification tags (only one allowed)")
	}

	return errors
}

// ValidatePostInference validates test tags after inference has been applied
// This enforces strict rules that must be true after tag enrichment
func ValidatePostInference(test TestReference, validSkipReasons map[string]config.SkipReason) []string {
	errors := []string{}
	warnings := []string{}

	// Load config once for tag validation
	cfg, cfgErr := config.Load(config.DefaultLoadOptions())
	if cfgErr != nil {
		errors = append(errors, fmt.Sprintf("cannot validate test '%s': failed to load config (%v)", test.TestName, cfgErr))
		return errors
	}

	// CRITICAL: Check for undefined tags first
	for _, tag := range test.Tags {
		if !IsValidTagWithConfig(tag, cfg) {
			errors = append(errors, fmt.Sprintf("test '%s' has undefined tag '%s' (not in %s/testing-tags.yml)", test.TestName, tag, contracts.EACConfigRelPath))
		}

		// Validate skip reason codes if contract is loaded
		if strings.HasPrefix(tag, "@skip:") && validSkipReasons != nil {
			reasonCode := strings.TrimPrefix(tag, "@skip:")
			if _, ok := validSkipReasons[reasonCode]; !ok {
				validCodes := make([]string, 0, len(validSkipReasons))
				for code := range validSkipReasons {
					validCodes = append(validCodes, code)
				}
				errors = append(errors, fmt.Sprintf("test '%s' has invalid skip reason '%s', valid codes: %v", test.TestName, reasonCode, validCodes))
			}
		}
	}

	// CRITICAL: Must have exactly ONE level tag (except @Manual tests)
	foundLevelTags := []string{}
	allLevelTags := GetLevelTagsFromConfig(cfg)
	for _, tag := range test.Tags {
		if contains(allLevelTags, tag) {
			foundLevelTags = append(foundLevelTags, tag)
		}
	}

	// @Manual tests are exempt from L-tag requirements
	if !test.IsManual {
		if len(foundLevelTags) == 0 {
			errors = append(errors, fmt.Sprintf("test '%s' has NO level tag (must have exactly one)", test.TestName))
		} else if len(foundLevelTags) > 1 {
			errors = append(errors, fmt.Sprintf("test '%s' has MULTIPLE level tags %v (must have exactly one)", test.TestName, foundLevelTags))
		}
	}

	// Must have at least ONE verification tag
	foundVerificationTags := []string{}
	allVerificationTags := GetVerificationTagsFromConfig(cfg)
	for _, tag := range test.Tags {
		if contains(allVerificationTags, tag) {
			foundVerificationTags = append(foundVerificationTags, tag)
		}
	}

	if len(foundVerificationTags) == 0 {
		errors = append(errors, fmt.Sprintf("test '%s' has NO verification tag (must have one)", test.TestName))
	}

	// Consistency: @piv or @ppv MUST have @L4
	hasPIV := contains(test.Tags, "@piv")
	hasPPV := contains(test.Tags, "@ppv")
	hasL4 := contains(test.Tags, "@L4")

	if (hasPIV || hasPPV) && !hasL4 {
		errors = append(errors, fmt.Sprintf("test '%s' has production verification (@piv/@ppv) but is not @L4", test.TestName))
	}

	// Consistency: @iv or @pv MUST have @L3
	// Installation and Performance verification are PLTE-level tests (L3)
	hasIV := contains(test.Tags, "@iv")
	hasPV := contains(test.Tags, "@pv")
	hasL3 := contains(test.Tags, "@L3")

	if (hasIV || hasPV) && !hasL3 {
		errors = append(errors, fmt.Sprintf("test '%s' has PLTE verification (@iv/@pv) but is not @L3", test.TestName))
	}

	// Validate: @L0 should not have external runtime dependencies
	// Note: @deps:go is allowed since it's the build-time toolchain, not a runtime dependency
	hasL0 := contains(test.Tags, "@L0")
	if hasL0 {
		for _, tag := range test.Tags {
			if strings.HasPrefix(tag, "@deps:") && tag != "@deps:go" {
				errors = append(errors, fmt.Sprintf("test '%s' is @L0 (very fast unit test) but has runtime dependency %s", test.TestName, tag))
			}
		}
	}

	// Validate: @Manual tests MUST NOT have any taxonomy level tags (L0-L4)
	// Manual tests are outside the automated taxonomy - they require human execution
	if test.IsManual && len(foundLevelTags) > 0 {
		errors = append(errors, fmt.Sprintf("test '%s' is @Manual but has taxonomy level tag %v (@Manual is mutually exclusive with L0-L4)", test.TestName, foundLevelTags))
	}

	// Validate: @gxp tests must have risk controls
	if test.IsGxP && len(test.RiskControls) == 0 {
		errors = append(errors, fmt.Sprintf("test '%s' has @gxp tag but no @risk-control:* tags", test.TestName))
	}

	// Validate: @gmp-critical-aspect must be used with @gxp
	if test.IsCriticalAspect && !test.IsGxP {
		errors = append(errors, fmt.Sprintf("test '%s' has @gmp-critical-aspect but no @gxp tag", test.TestName))
	}

	// Validate: Production tags (@piv/@ppv) should not be in commit suite (these are @L4)
	// This is covered by the L4 check above

	// Validate: @skip tests should still have proper tags (for documentation)
	// No special validation needed - skipped tests still need valid tags for when @skip tag is removed

	// Combine errors and warnings
	allIssues := append(errors, warnings...)
	return allIssues
}

// ShouldSkipValidation determines if a test should be excluded from validation
// Returns true for test framework's own tests that may intentionally have invalid tags
func ShouldSkipValidation(test TestReference) bool {
	// All meta tests (testing framework tests) have been moved to .go.txt files
	// and are tested via specs/src-core/testing-framework/specification.feature
	// No tests should be skipped from validation anymore
	return false
}

// ValidateAllPostInference validates all tests after inference
// Returns map of test name to validation errors
// Skips test framework's own tests which may have intentionally invalid tags
func ValidateAllPostInference(tests []TestReference, repoRoot string) map[string][]string {
	validationErrors := make(map[string][]string)

	// Load config for skip reason validation
	cfg, err := config.Load(config.DefaultLoadOptions())
	var validSkipReasons map[string]config.SkipReason
	if err != nil {
		// If config loading fails, proceed without skip reason validation
		// This allows validation to work even if config is missing
		validSkipReasons = nil
	} else {
		validSkipReasons = cfg.TestingTags.GetSkipReasons()
	}

	for _, test := range tests {
		// Skip validation for framework tests
		if ShouldSkipValidation(test) {
			continue
		}

		issues := ValidatePostInference(test, validSkipReasons)
		if len(issues) > 0 {
			validationErrors[test.TestName] = issues
		}
	}

	return validationErrors
}

// ValidateTestReference validates a complete test reference
func ValidateTestReference(test TestReference) []string {
	errors := []string{}

	// Validate test type
	if !contains(ValidTestTypes, test.Type) {
		errors = append(errors, fmt.Sprintf("invalid test type: %s (valid types: %v)", test.Type, ValidTestTypes))
	}

	// Validate tags
	tagErrors := ValidateTags(test.Tags)
	errors = append(errors, tagErrors...)

	return errors
}

// IsValidTag checks if a tag is valid according to contracts.
// Returns false if config cannot be loaded (fail-closed behavior).
// Note: This does lightweight validation - for full validation use validate test-tags command
func IsValidTag(tag string) bool {
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		// Fail closed: if we can't load config, the tag is invalid
		// This ensures contract-based validation is enforced
		return false
	}

	// Use the config's IsKnownTag method which handles both exact and pattern matches
	return cfg.TestingTags.IsKnownTag(tag)
}

// IsValidTagWithConfig checks if a tag is valid using a pre-loaded config.
// Use this in loops to avoid repeated config loading.
func IsValidTagWithConfig(tag string, cfg *config.EACConfig) bool {
	if cfg == nil || cfg.TestingTags == nil {
		return false
	}
	return cfg.TestingTags.IsKnownTag(tag)
}

// validateRiskControlTag validates @risk-control:<name>-<id> format
func validateRiskControlTag(tag string) bool {
	// Format: @risk-control:<name>-<id>
	// Example: @risk-control:auth-mfa-01
	// GxP Format: @risk-control:gxp-<name>

	parts := strings.TrimPrefix(tag, "@risk-control:")
	if len(parts) == 0 {
		return false
	}

	// GxP format: @risk-control:gxp-<name>
	if strings.HasPrefix(parts, "gxp-") {
		return len(parts) > 4 // At least "gxp-x"
	}

	// Standard format: must have dash and at least 2-digit numeric ID
	dashIndex := strings.LastIndex(parts, "-")
	if dashIndex == -1 {
		return false
	}

	controlName := parts[:dashIndex]
	scenarioID := parts[dashIndex+1:]

	// Scenario ID must be at least 2 characters and all digits
	if len(scenarioID) < 2 {
		return false
	}

	for _, ch := range scenarioID {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return len(controlName) > 0
}

// GetKnownTags returns all known tags from the config
func GetKnownTags() []string {
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return []string{} // Return empty if config can't be loaded
	}

	tags := []string{}
	for _, tag := range cfg.TestingTags.Tags {
		tags = append(tags, tag.Tag)
	}
	return tags
}

// ValidateGxPRequirements validates GxP-specific requirements
func ValidateGxPRequirements(test TestReference) []string {
	errors := []string{}

	// GxP requirements must have risk control
	if test.IsGxP {
		hasGxPRiskControl := false
		for _, rc := range test.RiskControls {
			if strings.HasPrefix(rc, "@risk-control:gxp-") {
				hasGxPRiskControl = true
				break
			}
		}

		if !hasGxPRiskControl {
			errors = append(errors, "GxP requirement must have @risk-control:gxp-<name> tag")
		}
	}

	// @gmp-critical-aspect must be used with @gxp
	if test.IsCriticalAspect && !test.IsGxP {
		errors = append(errors, "@gmp-critical-aspect must be used with @gxp tag")
	}

	return errors
}

// ValidateRiskControls validates risk control tags
func ValidateRiskControls(test TestReference) []string {
	errors := []string{}

	for _, rc := range test.RiskControls {
		if !validateRiskControlTag(rc) {
			errors = append(errors, fmt.Sprintf("Invalid risk control tag format: %s", rc))
		}
	}

	return errors
}

// ParseRiskControlTag parses a risk control tag into components
func ParseRiskControlTag(tag string) (*RiskControlRef, error) {
	if !strings.HasPrefix(tag, "@risk-control:") {
		return nil, fmt.Errorf("not a risk control tag: %s", tag)
	}

	parts := strings.TrimPrefix(tag, "@risk-control:")

	ref := &RiskControlRef{
		FullTag: tag,
	}

	// GxP format: @risk-control:gxp-<name>
	if strings.HasPrefix(parts, "gxp-") {
		ref.ControlName = parts
		ref.ScenarioID = ""
		ref.IsGxP = true
		return ref, nil
	}

	// Standard format: @risk-control:<name>-<id>
	dashIndex := strings.LastIndex(parts, "-")
	if dashIndex == -1 {
		return nil, fmt.Errorf("invalid risk control tag format: %s", tag)
	}

	ref.ControlName = parts[:dashIndex]
	ref.ScenarioID = parts[dashIndex+1:]
	ref.IsGxP = false

	return ref, nil
}

// ValidateFeatureLevelTags validates that Feature-level tags don't conflict with Scenario tags.
// In Gherkin, scenarios INHERIT all tags from the Feature level. This causes problems when:
//   - Feature has @L2 tag and Scenario has @L3 tag → Scenario gets BOTH @L2 AND @L3 (invalid)
//   - Feature has @ov tag and Scenario has @iv tag → Scenario gets BOTH @ov AND @iv (invalid)
//
// This function detects conflicts for:
//   - L-level tags (@L0-@L4) - must have exactly one
//   - Verification tags (@ov/@iv/@pv/@piv/@ppv) - validation checks for multiple
//
// This function detects these anti-patterns and returns validation errors.
func ValidateFeatureLevelTags(featureFilePath string) ([]string, error) {
	errors := []string{}

	content, err := os.ReadFile(featureFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feature file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	var featureLevelTags []string
	var inFeature bool
	lineNum := 0

	for _, line := range lines {
		lineNum++
		trimmed := strings.TrimSpace(line)

		// Detect Feature: keyword first
		if strings.HasPrefix(trimmed, "Feature:") {
			inFeature = true
			continue
		}

		// Extract feature-level tags (before Feature:)
		if strings.HasPrefix(trimmed, "@") && !inFeature {
			tags := extractTagsFromLine(trimmed)
			featureLevelTags = append(featureLevelTags, tags...)
			continue
		}

		// Check for Scenario or Scenario Outline with tags
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			// Look back at previous lines for scenario-level tags
			scenarioTags := []string{}
			checkLine := lineNum - 2 // Start checking from line before scenario

			for checkLine >= 0 && checkLine < len(lines) {
				checkTrimmed := strings.TrimSpace(lines[checkLine])
				if strings.HasPrefix(checkTrimmed, "@") {
					tags := extractTagsFromLine(checkTrimmed)
					// Prepend tags (we're walking backwards)
					scenarioTags = append(tags, scenarioTags...)
					checkLine--
				} else if checkTrimmed == "" {
					// Empty line, keep looking
					checkLine--
				} else {
					// Non-tag, non-empty line - stop looking
					break
				}
			}

			// Check if Feature has L-level tag AND Scenario has L-level tag
			featureLevelTag := ""
			for _, tag := range featureLevelTags {
				if isLevelTag(tag) {
					featureLevelTag = tag
					break
				}
			}

			scenarioLevelTag := ""
			for _, tag := range scenarioTags {
				if isLevelTag(tag) {
					scenarioLevelTag = tag
					break
				}
			}

			// ANTI-PATTERN: Feature AND Scenario have DIFFERENT L-level tags
			// (Same L-tag is redundant but not invalid - it just results in duplicate tags)
			if featureLevelTag != "" && scenarioLevelTag != "" && featureLevelTag != scenarioLevelTag {
				scenarioName := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(trimmed, "Scenario Outline:"), "Scenario:"))
				errors = append(errors, fmt.Sprintf(
					"Feature has L-tag %s and Scenario '%s' (line %d) has L-tag %s. "+
						"Scenarios inherit Feature tags, resulting in BOTH tags on the scenario. "+
						"Remove the L-tag from either the Feature or all Scenarios.",
					featureLevelTag, scenarioName, lineNum, scenarioLevelTag))
			}

			// Check if Feature has verification tag AND Scenario has DIFFERENT verification tag
			featureVerificationTag := ""
			for _, tag := range featureLevelTags {
				if isVerificationTag(tag) {
					featureVerificationTag = tag
					break
				}
			}

			scenarioVerificationTag := ""
			for _, tag := range scenarioTags {
				if isVerificationTag(tag) {
					scenarioVerificationTag = tag
					break
				}
			}

			// ANTI-PATTERN: Feature AND Scenario have DIFFERENT verification tags
			if featureVerificationTag != "" && scenarioVerificationTag != "" && featureVerificationTag != scenarioVerificationTag {
				scenarioName := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(trimmed, "Scenario Outline:"), "Scenario:"))
				errors = append(errors, fmt.Sprintf(
					"Feature has verification tag %s and Scenario '%s' (line %d) has verification tag %s. "+
						"Scenarios inherit Feature tags, resulting in BOTH tags on the scenario. "+
						"Remove the verification tag from either the Feature or all Scenarios.",
					featureVerificationTag, scenarioName, lineNum, scenarioVerificationTag))
			}
		}
	}

	return errors, nil
}
