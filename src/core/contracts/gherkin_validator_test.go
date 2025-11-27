package contracts

import (
	"strings"
	"testing"
)

// Intent: Comprehensive tests for GherkinValidator with TagContract integration
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Table-driven tests with clear test case names
//   - Each test covers a specific validation rule
//   - Expected outcomes are explicit
//
// Easy to change:
//   - Tests organized by validation category
//   - Shared test helpers reduce duplication
//   - Contract fixtures are reusable
//
// Hard to break:
//   - Tests cover all validation rules from contracts
//   - Edge cases and error conditions tested
//   - Both positive and negative test cases

// createTestTagContract creates a TagContract for testing
func createTestTagContract() *TagContract {
	tc := &TagContract{
		Metadata: TagMetadata{
			Version:     "0.1.0",
			Description: "Test tag contract",
			Scope:       "Tests",
		},
		Tags: []TagDefinition{
			// Verification tags
			{Tag: "@ov", Name: "Operational Verification", Type: "verification"},
			{Tag: "@iv", Name: "Installation Verification", Type: "verification"},
			{Tag: "@pv", Name: "Performance Verification", Type: "verification"},
			{Tag: "@piv", Name: "Production Installation Verification", Type: "verification"},
			{Tag: "@ppv", Name: "Production Performance Verification", Type: "verification"},
			// Taxonomy levels
			{Tag: "@L0", Name: "Very Fast Unit Tests", Type: "taxonomy-level"},
			{Tag: "@L1", Name: "Fast Unit Tests", Type: "taxonomy-level"},
			{Tag: "@L2", Name: "Emulated System Tests", Type: "taxonomy-level"},
			{Tag: "@L3", Name: "Production-like System Tests", Type: "taxonomy-level"},
			{Tag: "@L4", Name: "Production System Tests", Type: "taxonomy-level"},
			{Tag: "@HE2E", Name: "Horizontal E2E", Type: "taxonomy-level"},
			// System dependencies
			{Tag: "@deps:docker", Description: "Requires Docker", Type: "system_dependency"},
			{Tag: "@deps:git", Description: "Requires Git", Type: "system_dependency"},
			{Tag: "@deps:go", Description: "Requires Go", Type: "system_dependency"},
			// Execution control
			{Tag: "@skip:<reason>", Description: "Skip test", Type: "execution_control", Pattern: "^@skip:[a-z]+$", Example: "@skip:wip"},
			{Tag: "@Manual", Description: "Manual test", Type: "execution_control", Constraint: "mutually_exclusive_with_taxonomy_levels"},
			// Risk control
			{Tag: "@risk-control:<name>-<id>", Description: "Risk control", Type: "risk_control", Pattern: "^@risk-control:[a-z0-9-]+-[0-9]{2}$", Example: "@risk-control:auth-mfa-01"},
			{Tag: "@risk-control:gxp-<name>", Description: "GxP risk control", Type: "risk_control_gxp", Pattern: "^@risk-control:gxp-[a-z0-9-]+$", Example: "@risk-control:gxp-account-lockout"},
			// GxP
			{Tag: "@gxp", Description: "GxP requirement", Type: "gxp_regulatory"},
			{Tag: "@critical-aspect", Description: "GmP Critical Aspect", Type: "gxp_regulatory"},
		},
		SkipReasons: []SkipReason{
			{Code: "wip", Name: "Work In Progress"},
			{Code: "broken", Name: "Broken Test"},
			{Code: "flaky", Name: "Flaky Test"},
			{Code: "infra", Name: "Infrastructure Unavailable"},
			{Code: "platform", Name: "Platform Specific"},
			{Code: "deprecated", Name: "Deprecated Feature"},
			{Code: "blocked", Name: "Blocked"},
			{Code: "todo", Name: "To Do"},
		},
	}
	_ = tc.Initialize()
	return tc
}

// createTestContract creates a structure Contract for testing
func createTestContract() *Contract {
	return &Contract{
		Version: "0.1.0",
		Name:    "Gherkin Specification Structure",
		RawData: map[string]interface{}{
			"feature_naming_pattern": "^[a-z][a-z0-9-]*_[a-z][a-z0-9-]*$",
			"required_verification_tags": []interface{}{
				"@ov", "@iv", "@pv", "@piv", "@ppv",
			},
		},
	}
}

func TestGherkinValidator_BasicStructure(t *testing.T) {
	contract := createTestContract()
	tagContract := createTestTagContract()
	validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

	tests := []struct {
		name           string
		input          string
		expectErrors   []string
		expectNoErrors bool
	}{
		{
			name:         "empty input",
			input:        "",
			expectErrors: []string{"EMPTY_OUTPUT"},
		},
		{
			name:         "whitespace only",
			input:        "   \n\t\n  ",
			expectErrors: []string{"EMPTY_OUTPUT"},
		},
		{
			name: "valid minimal specification",
			input: `Feature: src-core_test-feature
  As a developer
  I want to test

  Rule: Basic rule

    @ov @L1
    Scenario: Test scenario
      Given something
      Then something happens`,
			expectNoErrors: true,
		},
		{
			name: "missing feature",
			input: `Rule: Some rule

  @ov
  Scenario: Test
    Given something`,
			expectErrors: []string{"MISSING_FEATURE", "RULE_BEFORE_FEATURE"},
		},
		{
			name: "missing rule",
			input: `Feature: src-core_test-feature

  @ov
  Scenario: Test
    Given something`,
			expectErrors: []string{"MISSING_RULE"},
		},
		{
			name: "missing scenario",
			input: `Feature: src-core_test-feature

  Rule: Some rule`,
			expectErrors: []string{"MISSING_SCENARIO"},
		},
		{
			name: "multiple features",
			input: `Feature: src-core_first
Feature: src-core_second

  Rule: Some rule

    @ov
    Scenario: Test
      Given something`,
			expectErrors: []string{"MULTIPLE_FEATURES"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.Validate(tt.input, nil)

			if tt.expectNoErrors {
				if len(errors) > 0 {
					t.Errorf("expected no errors, got %d: %v", len(errors), errors)
				}
				return
			}

			for _, expectedCode := range tt.expectErrors {
				found := false
				for _, err := range errors {
					if err.Code == expectedCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %s not found in errors: %v", expectedCode, errors)
				}
			}
		})
	}
}

func TestGherkinValidator_FeatureNaming(t *testing.T) {
	contract := createTestContract()
	validator := NewGherkinValidator(contract, nil)

	tests := []struct {
		name        string
		featureName string
		expectError bool
	}{
		{name: "valid: simple", featureName: "src-core_test-feature", expectError: false},
		{name: "valid: with numbers", featureName: "src-core2_test123", expectError: false},
		{name: "valid: all lowercase", featureName: "module_feature", expectError: false},
		{name: "invalid: no underscore", featureName: "testfeature", expectError: true},
		{name: "invalid: uppercase", featureName: "Src-Core_Test", expectError: true},
		{name: "invalid: starts with number", featureName: "1module_feature", expectError: true},
		{name: "invalid: special chars", featureName: "src@core_test!", expectError: true},
		{name: "invalid: spaces", featureName: "src core_test feature", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "Feature: " + tt.featureName + "\n\n  Rule: R\n\n    @ov\n    Scenario: S\n      Given g"
			errors := validator.Validate(input, nil)

			hasNamingError := false
			for _, err := range errors {
				if err.Code == "INVALID_FEATURE_NAMING" {
					hasNamingError = true
					break
				}
			}

			if tt.expectError && !hasNamingError {
				t.Errorf("expected INVALID_FEATURE_NAMING error for '%s'", tt.featureName)
			}
			if !tt.expectError && hasNamingError {
				t.Errorf("unexpected INVALID_FEATURE_NAMING error for '%s'", tt.featureName)
			}
		})
	}
}

func TestGherkinValidator_VerificationTags(t *testing.T) {
	contract := createTestContract()
	tagContract := createTestTagContract()
	validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

	tests := []struct {
		name        string
		tags        string
		expectError bool
	}{
		{name: "has @ov", tags: "@ov", expectError: false},
		{name: "has @iv", tags: "@iv", expectError: false},
		{name: "has @pv", tags: "@pv", expectError: false},
		{name: "has @piv", tags: "@piv", expectError: false},
		{name: "has @ppv", tags: "@ppv", expectError: false},
		{name: "has @ov with others", tags: "@L1 @ov @deps:docker", expectError: false},
		{name: "missing verification tag", tags: "@L1 @deps:docker", expectError: true},
		{name: "no tags", tags: "", expectError: true},
		{name: "only taxonomy", tags: "@L0", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagLine := ""
			if tt.tags != "" {
				tagLine = "    " + tt.tags + "\n"
			}
			input := `Feature: src-core_test

  Rule: Test rule

` + tagLine + `    Scenario: Test scenario
      Given something`

			errors := validator.Validate(input, nil)

			hasMissingTag := false
			for _, err := range errors {
				if err.Code == "MISSING_VERIFICATION_TAG" {
					hasMissingTag = true
					break
				}
			}

			if tt.expectError && !hasMissingTag {
				t.Errorf("expected MISSING_VERIFICATION_TAG error")
			}
			if !tt.expectError && hasMissingTag {
				t.Errorf("unexpected MISSING_VERIFICATION_TAG error")
			}
		})
	}
}

func TestGherkinValidator_SkipReasonValidation(t *testing.T) {
	contract := createTestContract()
	tagContract := createTestTagContract()
	validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

	tests := []struct {
		name        string
		skipTag     string
		expectError bool
		errorCode   string
	}{
		{name: "valid: wip", skipTag: "@skip:wip", expectError: false},
		{name: "valid: broken", skipTag: "@skip:broken", expectError: false},
		{name: "valid: flaky", skipTag: "@skip:flaky", expectError: false},
		{name: "valid: infra", skipTag: "@skip:infra", expectError: false},
		{name: "valid: platform", skipTag: "@skip:platform", expectError: false},
		{name: "valid: deprecated", skipTag: "@skip:deprecated", expectError: false},
		{name: "valid: blocked", skipTag: "@skip:blocked", expectError: false},
		{name: "valid: todo", skipTag: "@skip:todo", expectError: false},
		{name: "invalid: unknown reason", skipTag: "@skip:unknown", expectError: true, errorCode: "INVALID_SKIP_REASON"},
		{name: "invalid: uppercase", skipTag: "@skip:WIP", expectError: true, errorCode: "INVALID_TAG_FORMAT"},
		{name: "invalid: empty reason", skipTag: "@skip:", expectError: true, errorCode: "INVALID_TAG_FORMAT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `Feature: src-core_test

  Rule: Test rule

    @ov ` + tt.skipTag + `
    Scenario: Test scenario
      Given something`

			errors := validator.Validate(input, nil)

			hasExpectedError := false
			for _, err := range errors {
				if tt.errorCode != "" && err.Code == tt.errorCode {
					hasExpectedError = true
					break
				}
			}

			if tt.expectError && !hasExpectedError {
				t.Errorf("expected error %s for '%s', got: %v", tt.errorCode, tt.skipTag, errors)
			}
			if !tt.expectError {
				for _, err := range errors {
					if err.Code == "INVALID_SKIP_REASON" || err.Code == "INVALID_TAG_FORMAT" {
						t.Errorf("unexpected error %s for '%s'", err.Code, tt.skipTag)
					}
				}
			}
		})
	}
}

func TestGherkinValidator_MutualExclusion(t *testing.T) {
	contract := createTestContract()
	tagContract := createTestTagContract()
	validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

	tests := []struct {
		name        string
		tags        string
		expectError bool
	}{
		{name: "@Manual alone", tags: "@ov @Manual", expectError: false},
		{name: "@L0 alone", tags: "@ov @L0", expectError: false},
		{name: "@Manual with @L0", tags: "@ov @Manual @L0", expectError: true},
		{name: "@Manual with @L1", tags: "@ov @Manual @L1", expectError: true},
		{name: "@Manual with @L2", tags: "@ov @Manual @L2", expectError: true},
		{name: "@Manual with @L3", tags: "@ov @Manual @L3", expectError: true},
		{name: "@Manual with @L4", tags: "@ov @Manual @L4", expectError: true},
		{name: "@Manual with @HE2E", tags: "@ov @Manual @HE2E", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `Feature: src-core_test

  Rule: Test rule

    ` + tt.tags + `
    Scenario: Test scenario
      Given something`

			errors := validator.Validate(input, nil)

			hasError := false
			for _, err := range errors {
				if err.Code == "MUTUAL_EXCLUSION_VIOLATION" {
					hasError = true
					break
				}
			}

			if tt.expectError && !hasError {
				t.Errorf("expected MUTUAL_EXCLUSION_VIOLATION error for tags '%s'", tt.tags)
			}
			if !tt.expectError && hasError {
				t.Errorf("unexpected MUTUAL_EXCLUSION_VIOLATION error for tags '%s'", tt.tags)
			}
		})
	}
}

func TestGherkinValidator_GxPRequirements(t *testing.T) {
	contract := createTestContract()
	tagContract := createTestTagContract()
	validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

	tests := []struct {
		name        string
		tags        string
		expectError string
	}{
		{
			name:        "@gxp without risk control",
			tags:        "@ov @gxp",
			expectError: "GXP_MISSING_RISK_CONTROL",
		},
		{
			name:        "@gxp with valid risk control",
			tags:        "@ov @gxp @risk-control:gxp-account-lockout",
			expectError: "",
		},
		{
			name:        "@critical-aspect without @gxp",
			tags:        "@ov @critical-aspect",
			expectError: "CRITICAL_ASPECT_REQUIRES_GXP",
		},
		{
			name:        "@critical-aspect with @gxp and risk control",
			tags:        "@ov @gxp @critical-aspect @risk-control:gxp-data-integrity",
			expectError: "",
		},
		{
			name:        "@gxp with non-gxp risk control",
			tags:        "@ov @gxp @risk-control:auth-mfa-01",
			expectError: "GXP_MISSING_RISK_CONTROL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `Feature: src-core_test

  Rule: Test rule

    ` + tt.tags + `
    Scenario: Test scenario
      Given something`

			errors := validator.Validate(input, nil)

			if tt.expectError != "" {
				found := false
				for _, err := range errors {
					if err.Code == tt.expectError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error %s for tags '%s', got: %v", tt.expectError, tt.tags, errors)
				}
			} else {
				for _, err := range errors {
					if err.Code == "GXP_MISSING_RISK_CONTROL" || err.Code == "CRITICAL_ASPECT_REQUIRES_GXP" {
						t.Errorf("unexpected GxP error %s for tags '%s'", err.Code, tt.tags)
					}
				}
			}
		})
	}
}

func TestGherkinValidator_UnknownTagWarning(t *testing.T) {
	contract := createTestContract()
	tagContract := createTestTagContract()
	validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

	tests := []struct {
		name          string
		tags          string
		expectWarning bool
		warningFor    string
	}{
		{name: "known tags only", tags: "@ov @L1 @deps:docker", expectWarning: false},
		{name: "typo in @ov", tags: "@0v @L1", expectWarning: true, warningFor: "@0v"},
		{name: "unknown custom tag", tags: "@ov @custom-tag", expectWarning: true, warningFor: "@custom-tag"},
		{name: "typo in @Manual", tags: "@ov @manual", expectWarning: true, warningFor: "@manual"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `Feature: src-core_test

  Rule: Test rule

    ` + tt.tags + `
    Scenario: Test scenario
      Given something`

			errors := validator.Validate(input, nil)

			hasWarning := false
			for _, err := range errors {
				if err.Code == "UNKNOWN_TAG" && err.Severity == "warning" {
					if tt.warningFor != "" && strings.Contains(err.Message, tt.warningFor) {
						hasWarning = true
						break
					} else if tt.warningFor == "" {
						hasWarning = true
						break
					}
				}
			}

			if tt.expectWarning && !hasWarning {
				t.Errorf("expected UNKNOWN_TAG warning for '%s'", tt.warningFor)
			}
			if !tt.expectWarning && hasWarning {
				t.Errorf("unexpected UNKNOWN_TAG warning in tags '%s'", tt.tags)
			}
		})
	}
}

func TestGherkinValidator_RiskControlPatterns(t *testing.T) {
	contract := createTestContract()
	tagContract := createTestTagContract()
	validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

	tests := []struct {
		name        string
		tag         string
		expectError bool
	}{
		{name: "valid risk-control", tag: "@risk-control:auth-mfa-01", expectError: false},
		{name: "valid gxp risk-control", tag: "@risk-control:gxp-account-lockout", expectError: false},
		{name: "invalid: missing id", tag: "@risk-control:auth-mfa", expectError: true},
		{name: "invalid: uppercase", tag: "@risk-control:Auth-MFA-01", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `Feature: src-core_test

  Rule: Test rule

    @ov ` + tt.tag + `
    Scenario: Test scenario
      Given something`

			errors := validator.Validate(input, nil)

			hasFormatError := false
			for _, err := range errors {
				if err.Code == "INVALID_TAG_FORMAT" {
					hasFormatError = true
					break
				}
			}

			if tt.expectError && !hasFormatError {
				t.Errorf("expected INVALID_TAG_FORMAT error for '%s'", tt.tag)
			}
			if !tt.expectError && hasFormatError {
				t.Errorf("unexpected INVALID_TAG_FORMAT error for '%s'", tt.tag)
			}
		})
	}
}

func TestGherkinValidator_WithoutTagContract(t *testing.T) {
	contract := createTestContract()
	// Create validator WITHOUT tag contract
	validator := NewGherkinValidator(contract, nil)

	input := `Feature: src-core_test

  Rule: Test rule

    @ov @unknown-tag @Manual @L1
    Scenario: Test scenario
      Given something`

	errors := validator.Validate(input, nil)

	// Should still validate verification tags but not advanced tag rules
	for _, err := range errors {
		if err.Code == "UNKNOWN_TAG" || err.Code == "MUTUAL_EXCLUSION_VIOLATION" {
			t.Errorf("unexpected advanced validation error without tag contract: %v", err)
		}
	}
}

func TestGherkinValidator_MultipleScenarios(t *testing.T) {
	contract := createTestContract()
	tagContract := createTestTagContract()
	validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

	input := `Feature: src-core_test

  Rule: Test rule

    @ov @L1
    Scenario: First scenario
      Given something

    @iv @L2
    Scenario: Second scenario
      Given something else

    @L0
    Scenario: Third scenario without verification
      Given another thing`

	errors := validator.Validate(input, nil)

	// Should have error for third scenario only
	missingTagErrors := 0
	for _, err := range errors {
		if err.Code == "MISSING_VERIFICATION_TAG" {
			missingTagErrors++
		}
	}

	if missingTagErrors != 1 {
		t.Errorf("expected 1 MISSING_VERIFICATION_TAG error, got %d", missingTagErrors)
	}
}

func TestGherkinValidator_VerifyImplementation(t *testing.T) {
	t.Run("with both contracts", func(t *testing.T) {
		contract := createTestContract()
		tagContract := createTestTagContract()
		validator := NewGherkinValidatorWithTags(contract, tagContract, nil)

		errors := validator.VerifyImplementation()

		for _, err := range errors {
			if err.Severity == "error" {
				t.Errorf("unexpected error in VerifyImplementation: %v", err)
			}
		}
	})

	t.Run("without tag contract", func(t *testing.T) {
		contract := createTestContract()
		validator := NewGherkinValidator(contract, nil)

		errors := validator.VerifyImplementation()

		hasNoTagWarning := false
		for _, err := range errors {
			if err.Code == "NO_TAG_CONTRACT" {
				hasNoTagWarning = true
				break
			}
		}

		if !hasNoTagWarning {
			t.Error("expected NO_TAG_CONTRACT warning")
		}
	})

	t.Run("without any contract", func(t *testing.T) {
		validator := NewGherkinValidator(nil, nil)

		errors := validator.VerifyImplementation()

		hasNoContractError := false
		for _, err := range errors {
			if err.Code == "NO_CONTRACT" {
				hasNoContractError = true
				break
			}
		}

		if !hasNoContractError {
			t.Error("expected NO_CONTRACT error")
		}
	})
}

func TestTagContract_Initialize(t *testing.T) {
	tc := createTestTagContract()

	t.Run("verification tags populated", func(t *testing.T) {
		tags := tc.GetVerificationTags()
		if len(tags) != 5 {
			t.Errorf("expected 5 verification tags, got %d", len(tags))
		}
	})

	t.Run("taxonomy level tags populated", func(t *testing.T) {
		tags := tc.GetTaxonomyLevelTags()
		if len(tags) != 6 {
			t.Errorf("expected 6 taxonomy level tags, got %d", len(tags))
		}
	})

	t.Run("skip reasons populated", func(t *testing.T) {
		reasons := tc.GetValidSkipReasons()
		if len(reasons) != 8 {
			t.Errorf("expected 8 skip reasons, got %d", len(reasons))
		}
	})

	t.Run("pattern compiled", func(t *testing.T) {
		if len(tc.compiledPatterns) == 0 {
			t.Error("expected compiled patterns")
		}
	})
}

func TestTagContract_IsKnownTag(t *testing.T) {
	tc := createTestTagContract()

	tests := []struct {
		tag      string
		expected bool
	}{
		{"@ov", true},
		{"@L0", true},
		{"@deps:docker", true},
		{"@skip:wip", true},
		{"@skip:broken", true},
		{"@risk-control:auth-mfa-01", true},
		{"@risk-control:gxp-test", true},
		{"@Manual", true},
		{"@unknown", false},
		{"@typo", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := tc.IsKnownTag(tt.tag)
			if result != tt.expected {
				t.Errorf("IsKnownTag(%s) = %v, expected %v", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestTagContract_ValidateTag(t *testing.T) {
	tc := createTestTagContract()

	tests := []struct {
		tag         string
		expectError bool
		errorCode   string
	}{
		{"@ov", false, ""},
		{"@skip:wip", false, ""},
		{"@skip:invalid", true, "INVALID_SKIP_REASON"},
		{"@skip:WIP", true, "INVALID_TAG_FORMAT"},
		{"@risk-control:auth-mfa-01", false, ""},
		{"@risk-control:auth-mfa", true, "INVALID_TAG_FORMAT"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			err := tc.ValidateTag(tt.tag, 1)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for tag %s", tt.tag)
				} else if err.Code != tt.errorCode {
					t.Errorf("expected error code %s, got %s", tt.errorCode, err.Code)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for tag %s: %v", tt.tag, err)
				}
			}
		})
	}
}
