package testing

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
)

// NOTE: ValidTags map has been removed - validation now uses tag contract
// See LoadTagContract() and IsValidTag() for contract-based validation

// ValidTestTypes returns the valid test types, queried from registered adapters.
// Falls back to hardcoded defaults if no adapters are registered.
func ValidTestTypes() []string {
	return getSupportedTypes()
}

// IsValidTestType checks if a string is a valid test type.
func IsValidTestType(s string) bool {
	for _, t := range getSupportedTypes() {
		if t == s {
			return true
		}
	}
	return false
}

// GetLevelTags returns taxonomy level tags from config.
// Returns nil if config is unavailable - callers must handle this.
func GetLevelTags() []string {
	cfg := config.Global()
	if cfg == nil || cfg.Testing == nil {
		return nil
	}
	return cfg.Testing.GetTaxonomyLevelTags()
}

// GetLevelTagsFromConfig returns taxonomy level tags from a pre-loaded config.
func GetLevelTagsFromConfig(cfg *config.EACConfig) []string {
	if cfg == nil || cfg.Testing == nil {
		return nil
	}
	return cfg.Testing.GetTaxonomyLevelTags()
}

// GetVerificationTags returns verification type tags from config.
// Returns nil if config is unavailable - callers must handle this.
func GetVerificationTags() []string {
	cfg := config.Global()
	if cfg == nil || cfg.Testing == nil {
		return nil
	}
	return cfg.Testing.GetVerificationTags()
}

// GetVerificationTagsFromConfig returns verification type tags from a pre-loaded config.
func GetVerificationTagsFromConfig(cfg *config.EACConfig) []string {
	if cfg == nil || cfg.Testing == nil {
		return nil
	}
	return cfg.Testing.GetVerificationTags()
}

// ValidateTags checks if tags are valid and don't conflict.
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
			errors = append(errors, fmt.Sprintf("tag %s is not defined in %s/testing-tags.yml", tag, domain.EACConfigRelPath))
		}
	}

	// Check for multiple level tags
	levelCount := 0
	levelTags := GetLevelTagsFromConfig(cfg)
	for _, tag := range tags {
		if slices.Contains(levelTags, tag) {
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
		if slices.Contains(verificationTags, tag) {
			verificationCount++
		}
	}
	if verificationCount > 1 {
		errors = append(errors, "test has multiple verification tags (only one allowed)")
	}

	return errors
}

// ValidatePostInference validates test tags after inference has been applied
// This enforces strict rules that must be true after tag enrichment.
func ValidatePostInference(test TestReference, cfg *config.EACConfig) []string {
	errors := []string{}
	warnings := []string{}

	// Use provided config or load fresh if nil
	if cfg == nil {
		var cfgErr error
		cfg, cfgErr = config.Load(config.DefaultLoadOptions())
		if cfgErr != nil {
			errors = append(errors, fmt.Sprintf("cannot validate test '%s': failed to load config (%v)", test.TestName, cfgErr))
			return errors
		}
	}

	// CRITICAL: Check for undefined tags first
	for _, tag := range test.Tags {
		if !IsValidTagWithConfig(tag, cfg) {
			errors = append(errors, fmt.Sprintf("test '%s' has undefined tag '%s' (not in %s/testing-tags.yml)", test.TestName, tag, domain.EACConfigRelPath))
		}

		// Validate skip reason codes using the new Testing API
		if strings.HasPrefix(tag, "@skip:") && cfg.Testing != nil {
			reasonCode := strings.TrimPrefix(tag, "@skip:")
			if _, ok := cfg.Testing.ValidateSkipReason(reasonCode); !ok {
				validCodes := cfg.Testing.GetValidSkipReasons()
				errors = append(errors, fmt.Sprintf("test '%s' has invalid skip reason '%s', valid codes: %v", test.TestName, reasonCode, validCodes))
			}
		}
	}

	// CRITICAL: Must have exactly ONE level tag (except @Manual tests)
	foundLevelTags := []string{}
	allLevelTags := GetLevelTagsFromConfig(cfg)
	for _, tag := range test.Tags {
		if slices.Contains(allLevelTags, tag) {
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
		if slices.Contains(allVerificationTags, tag) {
			foundVerificationTags = append(foundVerificationTags, tag)
		}
	}

	if len(foundVerificationTags) == 0 {
		errors = append(errors, fmt.Sprintf("test '%s' has NO verification tag (must have one)", test.TestName))
	}

	// Consistency: @piv or @ppv MUST have @L4
	hasPIV := slices.Contains(test.Tags, "@piv")
	hasPPV := slices.Contains(test.Tags, "@ppv")
	hasL4 := slices.Contains(test.Tags, "@L4")

	if (hasPIV || hasPPV) && !hasL4 {
		errors = append(errors, fmt.Sprintf("test '%s' has production verification (@piv/@ppv) but is not @L4", test.TestName))
	}

	// Consistency: @iv or @pv MUST have @L3
	// Installation and Performance verification are PLTE-level tests (L3)
	hasIV := slices.Contains(test.Tags, "@iv")
	hasPV := slices.Contains(test.Tags, "@pv")
	hasL3 := slices.Contains(test.Tags, "@L3")

	if (hasIV || hasPV) && !hasL3 {
		errors = append(errors, fmt.Sprintf("test '%s' has PLTE verification (@iv/@pv) but is not @L3", test.TestName))
	}

	// Validate: @L0 should not have external runtime dependencies
	// Note: @deps:go is allowed since it's the build-time toolchain, not a runtime dependency
	hasL0 := slices.Contains(test.Tags, "@L0")
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

	// Validate: Production tags (@piv/@ppv) should not be in unit suite (these are @L4)
	// This is covered by the L4 check above

	// Validate: @skip tests should still have proper tags (for documentation)
	// No special validation needed - skipped tests still need valid tags for when @skip tag is removed

	// Combine errors and warnings
	allIssues := append(errors, warnings...)
	return allIssues
}

// ShouldSkipValidation determines if a test should be excluded from validation
// Returns true for test framework's own tests that may intentionally have invalid tags.
func ShouldSkipValidation(test TestReference) bool {
	// All meta tests (testing framework tests) have been moved to .go.txt files
	// and are tested via specs/core/testing-framework/specification.feature
	// No tests should be skipped from validation anymore
	return false
}

// ValidateAllPostInference validates all tests after inference
// Returns map of test name to validation errors
// Skips test framework's own tests which may have intentionally invalid tags.
func ValidateAllPostInference(tests []TestReference, repoRoot string) map[string][]string {
	validationErrors := make(map[string][]string)

	// Load config once for all validations (uses new Testing API)
	cfg, _ := config.Load(config.DefaultLoadOptions())
	// cfg may be nil if loading fails - ValidatePostInference handles this

	for _, test := range tests {
		// Skip validation for framework tests
		if ShouldSkipValidation(test) {
			continue
		}

		issues := ValidatePostInference(test, cfg)
		if len(issues) > 0 {
			validationErrors[test.TestName] = issues
		}
	}

	return validationErrors
}

// ValidateTestReference validates a complete test reference.
func ValidateTestReference(test TestReference) []string {
	errors := []string{}

	// Validate test type
	validTypes := ValidTestTypes()
	if !slices.Contains(validTypes, test.Type) {
		errors = append(errors, fmt.Sprintf("invalid test type: %s (valid types: %v)", test.Type, validTypes))
	}

	// Validate tags
	tagErrors := ValidateTags(test.Tags)
	errors = append(errors, tagErrors...)

	return errors
}

// IsValidTag checks if a tag is valid according to domain.
// Returns false if config cannot be loaded (fail-closed behavior).
// Note: This does lightweight validation - for full validation use validate test-tags command.
func IsValidTag(tag string) bool {
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		// Fail closed: if we can't load config, the tag is invalid
		// This ensures contract-based validation is enforced
		return false
	}

	// Use the config's IsKnownTag method which handles both exact and pattern matches
	return cfg.Testing.IsKnownTag(tag)
}

// IsValidTagWithConfig checks if a tag is valid using a pre-loaded config.
// Use this in loops to avoid repeated config loading.
func IsValidTagWithConfig(tag string, cfg *config.EACConfig) bool {
	if cfg == nil || cfg.Testing == nil {
		return false
	}
	return cfg.Testing.IsKnownTag(tag)
}

// GetKnownTags returns all known tags from the config.
func GetKnownTags() []string {
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil || cfg.Testing == nil {
		return []string{} // Return empty if config can't be loaded
	}

	return cfg.Testing.ListTags()
}
