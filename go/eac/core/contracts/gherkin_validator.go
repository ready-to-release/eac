package contracts

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// Intent: Validate Gherkin specifications against structure and tag contracts
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear validation rules matching contract requirements
//   - Each check has a specific error code
//   - Error messages are descriptive and actionable
//
// Easy to change:
//   - Validation rules are separate functions
//   - Structure contract (contract.yml) and tag contract (tags.yml) loaded externally
//   - Easy to add new validation rules
//
// Hard to break:
//   - Validates against formal contracts
//   - Returns detailed errors with line numbers
//   - Comprehensive test coverage
//   - Self-verification with VerifyImplementation()

// GherkinValidator validates Gherkin specifications against structure and tag contracts
type GherkinValidator struct {
	contract       *Contract
	tagsConfig     *config.TestingTagsConfig
	antiCorruption *AntiCorruptionRules
}

// NewGherkinValidator creates a new Gherkin specification validator
func NewGherkinValidator(contract *Contract, antiCorruption *AntiCorruptionRules) *GherkinValidator {
	return &GherkinValidator{
		contract:       contract,
		antiCorruption: antiCorruption,
	}
}

// NewGherkinValidatorWithTags creates a validator with both structure and tag configs
func NewGherkinValidatorWithTags(contract *Contract, tagsConfig *config.TestingTagsConfig, antiCorruption *AntiCorruptionRules) *GherkinValidator {
	return &GherkinValidator{
		contract:       contract,
		tagsConfig:     tagsConfig,
		antiCorruption: antiCorruption,
	}
}

// NewGherkinValidatorFromConfig creates a validator from the unified AI config
func NewGherkinValidatorFromConfig(loader *AIConfigLoader, typeName string, antiCorruption *AntiCorruptionRules) *GherkinValidator {
	// Get validation rules from the type config
	validation, err := loader.GetValidation(typeName)
	if err != nil {
		// Return validator without contract data
		return &GherkinValidator{
			antiCorruption: antiCorruption,
		}
	}

	// Convert validation map to Contract for backward compatibility
	contract := &Contract{
		Version: "0.1.0",
		Name:    "Gherkin Specification Structure",
		RawData: make(map[string]interface{}),
	}

	// Extract feature_naming_pattern from patterns.feature_naming
	if patterns := ExtractMap(validation, "patterns"); patterns != nil {
		if featureNaming := ExtractString(patterns, "feature_naming"); featureNaming != "" {
			contract.RawData["feature_naming_pattern"] = featureNaming
		}
	}

	// Extract required_verification_tags from required_tags
	if requiredTags := ExtractStringList(validation, "required_tags"); len(requiredTags) > 0 {
		// Convert []string to []interface{} for RawData compatibility
		tags := make([]interface{}, len(requiredTags))
		for i, t := range requiredTags {
			tags[i] = t
		}
		contract.RawData["required_verification_tags"] = tags
	}

	return &GherkinValidator{
		contract:       contract,
		antiCorruption: antiCorruption,
	}
}

// SetTagsConfig sets the tags config for tag validation
func (v *GherkinValidator) SetTagsConfig(tagsConfig *config.TestingTagsConfig) {
	v.tagsConfig = tagsConfig
}

// Validate validates Gherkin content against the specification contract
//
// This implements the Validator interface and checks:
// - Feature declaration exists and follows naming convention
// - Rule blocks present
// - Scenarios present under Rules
// - Proper keyword ordering
// - Verification tags present on scenarios
// - Tag format validation (patterns from tags.yml)
// - Skip reason validation
// - Mutual exclusion constraints (@Manual vs @L0-L4)
// - GxP tag requirements
// - Unknown tag warnings
//
// Returns a list of validation errors (empty if valid)
func (v *GherkinValidator) Validate(output string, context map[string]interface{}) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(output) == "" {
		errors = append(errors, ValidationError{
			Code:     "EMPTY_OUTPUT",
			Message:  "Generated output is empty",
			Severity: "error",
		})
		return errors
	}

	lines := strings.Split(output, "\n")
	state := &gherkinValidationState{
		seenFeature:  false,
		seenRule:     false,
		seenScenario: false,
	}

	// Track feature name for naming convention validation
	var featureName string

	// Validate line by line
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for Feature declaration
		if strings.HasPrefix(trimmed, "Feature:") {
			if state.seenFeature {
				errors = append(errors, ValidationError{
					Code:     "MULTIPLE_FEATURES",
					Message:  "Multiple Feature declarations found (only one allowed per file)",
					Line:     lineNum,
					Severity: "error",
				})
			} else {
				state.seenFeature = true
				featureName = strings.TrimSpace(strings.TrimPrefix(trimmed, "Feature:"))

				// Validate naming convention
				namingError := v.validateFeatureNaming(featureName, lineNum)
				if namingError.Code != "" { // Non-empty error
					errors = append(errors, namingError)
				}
			}
			continue
		}

		// Check for Background declaration
		if strings.HasPrefix(trimmed, "Background:") {
			if !state.seenFeature {
				errors = append(errors, ValidationError{
					Code:     "BACKGROUND_BEFORE_FEATURE",
					Message:  "Background must come after Feature declaration",
					Line:     lineNum,
					Severity: "error",
				})
			}
			continue
		}

		// Check for Rule declaration
		if strings.HasPrefix(trimmed, "Rule:") {
			if !state.seenFeature {
				errors = append(errors, ValidationError{
					Code:     "RULE_BEFORE_FEATURE",
					Message:  "Rule must come after Feature declaration",
					Line:     lineNum,
					Severity: "error",
				})
			} else {
				state.seenRule = true
			}
			continue
		}

		// Check for Scenario declaration
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			if !state.seenFeature {
				errors = append(errors, ValidationError{
					Code:     "SCENARIO_BEFORE_FEATURE",
					Message:  "Scenario must come after Feature declaration",
					Line:     lineNum,
					Severity: "error",
				})
			} else {
				state.seenScenario = true
			}
			continue
		}

		// Check for Examples (must follow Scenario Outline)
		if strings.HasPrefix(trimmed, "Examples:") {
			if !state.seenScenario {
				errors = append(errors, ValidationError{
					Code:     "EXAMPLES_WITHOUT_SCENARIO",
					Message:  "Examples must come after Scenario Outline",
					Line:     lineNum,
					Severity: "error",
				})
			}
			continue
		}
	}

	// Final state validation
	if !state.seenFeature {
		errors = append(errors, ValidationError{
			Code:     "MISSING_FEATURE",
			Message:  "Missing Feature: declaration",
			Severity: "error",
		})
	}

	if !state.seenRule {
		errors = append(errors, ValidationError{
			Code:     "MISSING_RULE",
			Message:  "Missing Rule: declaration (required for acceptance criteria)",
			Severity: "error",
		})
	}

	if !state.seenScenario {
		errors = append(errors, ValidationError{
			Code:     "MISSING_SCENARIO",
			Message:  "Missing Scenario: declaration (required for behavior examples)",
			Severity: "error",
		})
	}

	// Validate tags on scenarios (verification, format, constraints)
	tagErrors := v.validateScenarioTags(lines)
	errors = append(errors, tagErrors...)

	return errors
}

// validateFeatureNaming checks if feature name follows the naming convention
//
// Expected format: <module>_<feature-name>
// Pattern is read from contract.feature_naming_pattern
func (v *GherkinValidator) validateFeatureNaming(featureName string, lineNum int) ValidationError {
	// Get naming pattern from contract
	pattern := `^[a-z][a-z0-9-]*_[a-z][a-z0-9-]*$` // Default fallback
	if v.contract != nil && v.contract.RawData != nil {
		if patternVal, ok := v.contract.RawData["feature_naming_pattern"].(string); ok && patternVal != "" {
			pattern = patternVal
		}
	}

	matched, err := regexp.MatchString(pattern, featureName)
	if err != nil || !matched {
		return ValidationError{
			Code: "INVALID_FEATURE_NAMING",
			Message: fmt.Sprintf(
				"Feature name '%s' does not follow naming convention '<module>_<feature-name>' (lowercase with hyphens, e.g., 'eac-commands_user-auth')",
				featureName,
			),
			Line:     lineNum,
			Severity: "error",
		}
	}

	return ValidationError{} // No error (empty struct)
}

// validateScenarioTags validates all tag-related rules on scenarios:
// - Verification tags present (with inheritance from Feature and Rule)
// - Tag format validation
// - Skip reason validation
// - Mutual exclusion constraints
// - GxP tag requirements
// - Unknown tag warnings
//
// Tag inheritance follows Gherkin semantics:
// - Feature tags are inherited by all Rules and Scenarios
// - Rule tags are inherited by all Scenarios within that Rule
func (v *GherkinValidator) validateScenarioTags(lines []string) []ValidationError {
	var errors []ValidationError

	// Inherited tags from Feature and Rule levels
	var featureTags []string     // Tags on the Feature (inherited by all)
	var currentRuleTags []string // Tags on current Rule (inherited by scenarios in that rule)
	var pendingTags []string     // Tags collected immediately before a scenario
	var pendingTagLines []int    // Line numbers for pending tags

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Collect tags (lines starting with @)
		if strings.HasPrefix(trimmed, "@") {
			// Extract all tags from the line (space-separated)
			tags := strings.Fields(trimmed)
			for _, tag := range tags {
				if strings.HasPrefix(tag, "@") {
					pendingTags = append(pendingTags, tag)
					pendingTagLines = append(pendingTagLines, lineNum)
				}
			}
			continue
		}

		// When we hit Feature, capture pending tags as feature-level tags
		if strings.HasPrefix(trimmed, "Feature:") {
			featureTags = append([]string{}, pendingTags...)
			pendingTags = []string{}
			pendingTagLines = []int{}
			continue
		}

		// When we hit Rule, capture pending tags as rule-level tags
		if strings.HasPrefix(trimmed, "Rule:") {
			currentRuleTags = append([]string{}, pendingTags...)
			pendingTags = []string{}
			pendingTagLines = []int{}
			continue
		}

		// When we hit a Scenario, validate with inherited tags
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			// Combine inherited tags: Feature + Rule + Scenario
			allTags := v.combineInheritedTags(featureTags, currentRuleTags, pendingTags)
			scenarioErrors := v.validateTagsForScenario(allTags, pendingTagLines, lineNum)
			errors = append(errors, scenarioErrors...)

			// Reset only scenario-level pending tags (keep inherited)
			pendingTags = []string{}
			pendingTagLines = []int{}
			continue
		}

		// Background clears pending tags but doesn't affect inheritance
		if strings.HasPrefix(trimmed, "Background:") {
			pendingTags = []string{}
			pendingTagLines = []int{}
			continue
		}
	}

	return errors
}

// combineInheritedTags merges tags from Feature, Rule, and Scenario levels
// Returns a deduplicated list of all applicable tags
func (v *GherkinValidator) combineInheritedTags(featureTags, ruleTags, scenarioTags []string) []string {
	seen := make(map[string]bool)
	var result []string

	// Add in order: Feature, Rule, Scenario (scenario tags take precedence visually)
	for _, tag := range featureTags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	for _, tag := range ruleTags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	for _, tag := range scenarioTags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}

// validateTagsForScenario validates all tags for a single scenario
func (v *GherkinValidator) validateTagsForScenario(tags []string, tagLines []int, scenarioLine int) []ValidationError {
	var errors []ValidationError

	// 1. Check verification tags (required)
	if !v.hasVerificationTag(tags) {
		verificationTags := v.getVerificationTags()
		errors = append(errors, ValidationError{
			Code:     "MISSING_VERIFICATION_TAG",
			Message:  fmt.Sprintf("Scenario missing verification tag (required: %s)", strings.Join(verificationTags, ", ")),
			Line:     scenarioLine,
			Severity: "error",
		})
	}

	// Only perform advanced tag validation if tagContract is available
	if v.tagsConfig == nil {
		return errors
	}

	// 2. Validate tag formats and collect metadata
	hasManual := false
	hasTaxonomyLevel := false
	hasGxP := false
	hasGxPRiskControl := false
	taxonomyLevelTags := v.tagsConfig.GetTaxonomyLevelTags()

	for i, tag := range tags {
		lineNum := scenarioLine
		if i < len(tagLines) {
			lineNum = tagLines[i]
		}

		// Validate tag format
		if formatErr := v.tagsConfig.ValidateTag(tag); formatErr != nil {
			errors = append(errors, ValidationError{
				Code:     "INVALID_TAG_FORMAT",
				Message:  formatErr.Error(),
				Line:     lineNum,
				Severity: "error",
			})
		}

		// Track special tags for constraint validation
		if tag == "@Manual" {
			hasManual = true
		}
		if v.isInTagList(tag, taxonomyLevelTags) {
			hasTaxonomyLevel = true
		}
		if tag == "@gxp" {
			hasGxP = true
		}
		if strings.HasPrefix(tag, "@risk-control:gxp-") {
			hasGxPRiskControl = true
		}

		// 7. Check for unknown tags (warning only)
		if !v.tagsConfig.IsKnownTag(tag) {
			errors = append(errors, ValidationError{
				Code:     "UNKNOWN_TAG",
				Message:  fmt.Sprintf("Unknown tag '%s' - possible typo? Check %s/testing-tags.yml for valid tags", tag, EACConfigRelPath),
				Line:     lineNum,
				Severity: "warning",
			})
		}
	}

	// 5. Enforce mutual exclusion: @Manual cannot be used with @L0-@L4
	if hasManual && hasTaxonomyLevel {
		errors = append(errors, ValidationError{
			Code:     "MUTUAL_EXCLUSION_VIOLATION",
			Message:  "@Manual tag cannot be used with taxonomy level tags (@L0-@L4, @HE2E) - manual tests are not automated",
			Line:     scenarioLine,
			Severity: "error",
		})
	}

	// 6. Validate GxP requirements: @gxp requires @risk-control:gxp-*
	if hasGxP && !hasGxPRiskControl {
		errors = append(errors, ValidationError{
			Code:     "GXP_MISSING_RISK_CONTROL",
			Message:  "@gxp tag requires a corresponding @risk-control:gxp-<name> tag",
			Line:     scenarioLine,
			Severity: "error",
		})
	}

	// Check @gmp-critical-aspect requires @gxp
	for i, tag := range tags {
		if tag == "@gmp-critical-aspect" && !hasGxP {
			lineNum := scenarioLine
			if i < len(tagLines) {
				lineNum = tagLines[i]
			}
			errors = append(errors, ValidationError{
				Code:     "CRITICAL_ASPECT_REQUIRES_GXP",
				Message:  "@gmp-critical-aspect tag requires @gxp tag to be present",
				Line:     lineNum,
				Severity: "error",
			})
			break
		}
	}

	return errors
}

// hasVerificationTag checks if tag list contains at least one verification tag
func (v *GherkinValidator) hasVerificationTag(tags []string) bool {
	verificationTags := v.getVerificationTags()
	tagMap := make(map[string]bool)
	for _, vTag := range verificationTags {
		tagMap[vTag] = true
	}

	for _, tag := range tags {
		if tagMap[tag] {
			return true
		}
	}
	return false
}

// getVerificationTags returns verification tags from contract or defaults
func (v *GherkinValidator) getVerificationTags() []string {
	// Try tag contract first
	if v.tagsConfig != nil {
		tags := v.tagsConfig.GetVerificationTags()
		if len(tags) > 0 {
			return tags
		}
	}

	// Try structure contract
	if v.contract != nil && v.contract.RawData != nil {
		// Try new unified config key first (ai-config.yml format)
		if tagsVal, ok := v.contract.RawData["required_tags"].([]interface{}); ok {
			var tags []string
			for _, tag := range tagsVal {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			if len(tags) > 0 {
				return tags
			}
		}

		// Fall back to legacy key name for backward compatibility
		if tagsVal, ok := v.contract.RawData["required_verification_tags"].([]interface{}); ok {
			var tags []string
			for _, tag := range tagsVal {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			if len(tags) > 0 {
				return tags
			}
		}
	}

	// Try global config as last resort
	if cfg := config.Global(); cfg != nil && cfg.TestingTags != nil {
		tags := cfg.TestingTags.GetVerificationTags()
		if len(tags) > 0 {
			return tags
		}
	}

	// Fallback to hardcoded defaults (should rarely be reached)
	return []string{"@ov", "@iv", "@pv", "@piv", "@ppv"}
}

// isInTagList checks if a tag is in the given list
func (v *GherkinValidator) isInTagList(tag string, list []string) bool {
	for _, t := range list {
		if t == tag {
			return true
		}
	}
	return false
}

// VerifyImplementation verifies that the validator implements all contract rules
//
// This is a self-check to ensure the validator stays in sync with the contract
func (v *GherkinValidator) VerifyImplementation() []ValidationError {
	var errors []ValidationError

	// Check that contract is loaded
	if v.contract == nil {
		errors = append(errors, ValidationError{
			Code:     "NO_CONTRACT",
			Message:  "Contract not loaded",
			Severity: "error",
		})
		return errors
	}

	// Verify contract version matches
	expectedVersion := "0.1.0"
	if v.contract.Version != expectedVersion {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_VERSION_MISMATCH",
			Message:  fmt.Sprintf("Expected contract version %s, got %s", expectedVersion, v.contract.Version),
			Severity: "error",
		})
	}

	// Verify contract name
	if v.contract.Name != "Gherkin Specification Structure" {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_NAME_MISMATCH",
			Message:  fmt.Sprintf("Expected contract name 'Gherkin Specification Structure', got '%s'", v.contract.Name),
			Severity: "warning",
		})
	}

	// Verify tag contract is loaded (warning only, not required)
	if v.tagsConfig == nil {
		errors = append(errors, ValidationError{
			Code:     "NO_TAG_CONTRACT",
			Message:  "Tag contract not loaded - advanced tag validation disabled",
			Severity: "warning",
		})
	}

	return errors
}

// gherkinValidationState tracks Gherkin structure during validation
type gherkinValidationState struct {
	seenFeature  bool
	seenRule     bool
	seenScenario bool
}
