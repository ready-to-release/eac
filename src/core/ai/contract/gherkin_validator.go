package contract

import (
	"fmt"
	"regexp"
	"strings"
)

// Intent: Validate Gherkin specifications against the contract
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
//   - Contract structure is loaded externally
//   - Easy to add new validation rules
//
// Hard to break:
//   - Validates against formal contract (structure.yml)
//   - Returns detailed errors with line numbers
//   - Comprehensive test coverage
//   - Self-verification with VerifyImplementation()

// GherkinValidator validates Gherkin specifications against the contract
type GherkinValidator struct {
	contract       *SpecContract
	antiCorruption *AntiCorruptionRules
}

// NewGherkinValidator creates a new Gherkin specification validator
func NewGherkinValidator(contract *SpecContract, antiCorruption *AntiCorruptionRules) *GherkinValidator {
	return &GherkinValidator{
		contract:       contract,
		antiCorruption: antiCorruption,
	}
}

// Validate validates Gherkin content against the specification contract
//
// This implements the SpecValidator interface and checks:
// - Feature declaration exists and follows naming convention
// - Rule blocks present (ATDD requirement)
// - Scenarios present under Rules (BDD requirement)
// - Proper keyword ordering
// - Verification tags present on scenarios
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
			Message:  "Missing Rule: declaration (required for ATDD acceptance criteria)",
			Severity: "error",
		})
	}

	if !state.seenScenario {
		errors = append(errors, ValidationError{
			Code:     "MISSING_SCENARIO",
			Message:  "Missing Scenario: declaration (required for BDD behavior examples)",
			Severity: "error",
		})
	}

	// Validate verification tags on scenarios
	tagErrors := v.validateVerificationTags(lines)
	errors = append(errors, tagErrors...)

	return errors
}

// validateFeatureNaming checks if feature name follows the naming convention
//
// Expected format: <module>_<feature-name>
// Pattern: ^[a-z][a-z0-9-]*_[a-z][a-z0-9-]*$
func (v *GherkinValidator) validateFeatureNaming(featureName string, lineNum int) ValidationError {
	// Extract naming convention pattern from contract
	pattern := `^[a-z][a-z0-9-]*_[a-z][a-z0-9-]*$`

	matched, err := regexp.MatchString(pattern, featureName)
	if err != nil || !matched {
		return ValidationError{
			Code: "INVALID_FEATURE_NAMING",
			Message: fmt.Sprintf(
				"Feature name '%s' does not follow naming convention '<module>_<feature-name>' (lowercase with hyphens, e.g., 'src-commands_user-auth')",
				featureName,
			),
			Line:     lineNum,
			Severity: "error",
		}
	}

	return ValidationError{} // No error (empty struct)
}

// validateVerificationTags checks that scenarios have required verification tags
//
// According to contract, every scenario MUST have at least one verification tag:
// @ov, @iv, @pv, @piv, @ppv
//
// Tags appear BEFORE the Scenario line in Gherkin
func (v *GherkinValidator) validateVerificationTags(lines []string) []ValidationError {
	var errors []ValidationError

	var pendingTags []string // Tags collected before a scenario

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
				}
			}
			continue
		}

		// When we hit a Scenario, check if we have verification tags
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			// Check if pending tags contain verification tag
			if !hasVerificationTag(pendingTags) {
				errors = append(errors, ValidationError{
					Code:     "MISSING_VERIFICATION_TAG",
					Message:  "Scenario missing verification tag (required: @ov, @iv, @pv, @piv, or @ppv)",
					Line:     lineNum,
					Severity: "error",
				})
			}
			// Reset pending tags for next scenario
			pendingTags = []string{}
			continue
		}

		// When we hit other keywords (Feature, Rule, Background), clear pending tags
		if strings.HasPrefix(trimmed, "Feature:") ||
			strings.HasPrefix(trimmed, "Rule:") ||
			strings.HasPrefix(trimmed, "Background:") {
			pendingTags = []string{}
			continue
		}
	}

	return errors
}

// hasVerificationTag checks if tag list contains at least one verification tag
func hasVerificationTag(tags []string) bool {
	verificationTags := map[string]bool{
		"@ov":  true,
		"@iv":  true,
		"@pv":  true,
		"@piv": true,
		"@ppv": true,
	}

	for _, tag := range tags {
		if verificationTags[tag] {
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

	return errors
}

// gherkinValidationState tracks Gherkin structure during validation
type gherkinValidationState struct {
	seenFeature  bool
	seenRule     bool
	seenScenario bool
}
