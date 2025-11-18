package contract

import (
	"testing"
)

// Intent: Test Gherkin validator against all contract rules
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Table-driven tests with descriptive names
//   - Each test validates one specific rule
//   - Clear expected vs actual assertions
//
// Easy to change:
//   - Test data separated from test logic
//   - Helper functions for common assertions
//   - Easy to add new validation rules
//
// Hard to break:
//   - Tests cover all contract requirements
//   - Edge cases are explicitly tested
//   - Invalid and valid examples for each rule

func TestGherkinValidator_ValidGherkin(t *testing.T) {
	validator := createTestValidator()

	validGherkin := `Feature: src-commands_user-authentication
  As a user
  I want to authenticate
  So that I can access the system

  Rule: Valid credentials grant access

    @ov
    Scenario: User logs in successfully
      Given I have valid credentials
      When I log in
      Then I am authenticated
`

	errors := validator.Validate(validGherkin, nil)

	if len(errors) != 0 {
		t.Errorf("expected no errors for valid Gherkin, got %d error(s):\n%s",
			len(errors), FormatValidationErrors(errors))
	}
}

func TestGherkinValidator_MissingFeature(t *testing.T) {
	validator := createTestValidator()

	missingFeature := `Rule: Some rule
    Scenario: Some scenario
      Given something
`

	errors := validator.Validate(missingFeature, nil)

	assertHasError(t, errors, "MISSING_FEATURE")
}

func TestGherkinValidator_MissingRule(t *testing.T) {
	validator := createTestValidator()

	missingRule := `Feature: src-commands_test
  As a user
  I want something

  Scenario: Some scenario
    Given something
`

	errors := validator.Validate(missingRule, nil)

	assertHasError(t, errors, "MISSING_RULE")
}

func TestGherkinValidator_MissingScenario(t *testing.T) {
	validator := createTestValidator()

	missingScenario := `Feature: src-commands_test
  As a user
  I want something

  Rule: Some rule
`

	errors := validator.Validate(missingScenario, nil)

	assertHasError(t, errors, "MISSING_SCENARIO")
}

func TestGherkinValidator_MultipleFeatures(t *testing.T) {
	validator := createTestValidator()

	multipleFeatures := `Feature: src-commands_test-one
  As a user
  I want something

  Rule: Rule one
    @ov
    Scenario: Scenario one
      Given something

Feature: src-commands_test-two
  As a user
  I want something else

  Rule: Rule two
    @ov
    Scenario: Scenario two
      Given something
`

	errors := validator.Validate(multipleFeatures, nil)

	assertHasError(t, errors, "MULTIPLE_FEATURES")
}

func TestGherkinValidator_InvalidFeatureNaming(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		wantError   bool
	}{
		{
			name:        "valid naming",
			featureName: "src-commands_user-auth",
			wantError:   false,
		},
		{
			name:        "missing underscore separator",
			featureName: "src-commands-user-auth",
			wantError:   true,
		},
		{
			name:        "uppercase not allowed",
			featureName: "Src-Commands_User-Auth",
			wantError:   true,
		},
		{
			name:        "no module prefix",
			featureName: "user-auth",
			wantError:   true,
		},
		{
			name:        "space in name",
			featureName: "src-commands_user auth",
			wantError:   true,
		},
	}

	validator := createTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gherkin := `Feature: ` + tt.featureName + `
  As a user
  I want something

  Rule: Test rule
    @ov
    Scenario: Test scenario
      Given something
`

			errors := validator.Validate(gherkin, nil)

			hasError := hasErrorCode(errors, "INVALID_FEATURE_NAMING")

			if hasError != tt.wantError {
				t.Errorf("hasError = %v, wantError = %v for feature name '%s'",
					hasError, tt.wantError, tt.featureName)
				if len(errors) > 0 {
					t.Logf("Errors: %s", FormatValidationErrors(errors))
				}
			}
		})
	}
}

func TestGherkinValidator_MissingVerificationTag(t *testing.T) {
	validator := createTestValidator()

	missingTag := `Feature: src-commands_test
  As a user
  I want something

  Rule: Test rule

    Scenario: Test scenario without tag
      Given something
      When something happens
      Then something occurs
`

	errors := validator.Validate(missingTag, nil)

	assertHasError(t, errors, "MISSING_VERIFICATION_TAG")
}

func TestGherkinValidator_ValidVerificationTags(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{"operational verification", "@ov"},
		{"installation verification", "@iv"},
		{"performance verification", "@pv"},
		{"production installation", "@piv"},
		{"production performance", "@ppv"},
	}

	validator := createTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gherkin := `Feature: src-commands_test
  As a user
  I want something

  Rule: Test rule

    ` + tt.tag + `
    Scenario: Test scenario
      Given something
`

			errors := validator.Validate(gherkin, nil)

			if hasErrorCode(errors, "MISSING_VERIFICATION_TAG") {
				t.Errorf("tag %s should be recognized as valid verification tag", tt.tag)
			}
		})
	}
}

func TestGherkinValidator_WrongKeywordOrder(t *testing.T) {
	tests := []struct {
		name      string
		gherkin   string
		errorCode string
	}{
		{
			name: "background before feature",
			gherkin: `Background:
  Given something

Feature: src-commands_test
  Rule: Test
    @ov
    Scenario: Test
      Given something
`,
			errorCode: "BACKGROUND_BEFORE_FEATURE",
		},
		{
			name: "rule before feature",
			gherkin: `Rule: Test rule

Feature: src-commands_test
  Rule: Another rule
    @ov
    Scenario: Test
      Given something
`,
			errorCode: "RULE_BEFORE_FEATURE",
		},
		{
			name: "scenario before feature",
			gherkin: `Scenario: Test scenario
  Given something

Feature: src-commands_test
  Rule: Test
    Scenario: Another
      Given something
`,
			errorCode: "SCENARIO_BEFORE_FEATURE",
		},
		{
			name: "examples without scenario",
			gherkin: `Feature: src-commands_test
  Rule: Test

  Examples:
    | value |
    | 1     |
`,
			errorCode: "EXAMPLES_WITHOUT_SCENARIO",
		},
	}

	validator := createTestValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.Validate(tt.gherkin, nil)

			assertHasError(t, errors, tt.errorCode)
		})
	}
}

func TestGherkinValidator_EmptyOutput(t *testing.T) {
	validator := createTestValidator()

	errors := validator.Validate("", nil)

	assertHasError(t, errors, "EMPTY_OUTPUT")
}

func TestGherkinValidator_CommentsAndEmptyLines(t *testing.T) {
	validator := createTestValidator()

	withComments := `# This is a comment
Feature: src-commands_test
  # Feature description
  As a user
  I want something

  # Acceptance criteria
  Rule: Test rule

    # Test scenario
    @ov
    Scenario: Test scenario
      # Setup
      Given something
      # Action
      When something happens
      # Verification
      Then something occurs
`

	errors := validator.Validate(withComments, nil)

	// Should still be valid - comments don't affect validation
	if len(errors) != 0 {
		t.Errorf("expected no errors with comments, got %d error(s):\n%s",
			len(errors), FormatValidationErrors(errors))
	}
}

func TestGherkinValidator_MultipleScenarios(t *testing.T) {
	validator := createTestValidator()

	multipleScenarios := `Feature: src-commands_test
  As a user
  I want something

  Rule: Test rule

    @ov
    Scenario: First scenario
      Given something
      When something happens
      Then something occurs

    @ov
    Scenario: Second scenario
      Given something else
      When another thing happens
      Then another thing occurs

    @iv
    Scenario: Third scenario
      Given yet another thing
      When it happens
      Then it works
`

	errors := validator.Validate(multipleScenarios, nil)

	if len(errors) != 0 {
		t.Errorf("expected no errors for multiple scenarios, got %d error(s):\n%s",
			len(errors), FormatValidationErrors(errors))
	}
}

func TestGherkinValidator_MixedTagsOnScenario(t *testing.T) {
	validator := createTestValidator()

	mixedTags := `Feature: src-commands_test
  As a user
  I want something

  Rule: Test rule

    @ov @deps:git @depm:src-cli
    Scenario: Scenario with multiple tags
      Given something
      When something happens
      Then something occurs
`

	errors := validator.Validate(mixedTags, nil)

	// Should be valid - @ov is present
	if hasErrorCode(errors, "MISSING_VERIFICATION_TAG") {
		t.Errorf("should recognize @ov even with other tags present")
	}
}

func TestGherkinValidator_VerifyImplementation(t *testing.T) {
	// Test with valid contract
	contract := &SpecContract{
		Version: "0.1.0",
		Name:    "Gherkin Specification Structure",
		RawData: make(map[string]interface{}),
	}

	validator := NewGherkinValidator(contract, nil)

	errors := validator.VerifyImplementation()

	if len(errors) != 0 {
		t.Errorf("expected no implementation errors, got %d:\n%s",
			len(errors), FormatValidationErrors(errors))
	}
}

func TestGherkinValidator_VerifyImplementation_WrongVersion(t *testing.T) {
	// Test with wrong contract version
	contract := &SpecContract{
		Version: "0.2.0", // Wrong version
		Name:    "Gherkin Specification Structure",
		RawData: make(map[string]interface{}),
	}

	validator := NewGherkinValidator(contract, nil)

	errors := validator.VerifyImplementation()

	assertHasError(t, errors, "CONTRACT_VERSION_MISMATCH")
}

func TestGherkinValidator_VerifyImplementation_NoContract(t *testing.T) {
	validator := NewGherkinValidator(nil, nil)

	errors := validator.VerifyImplementation()

	assertHasError(t, errors, "NO_CONTRACT")
}

// Helper functions

func createTestValidator() *GherkinValidator {
	contract := &SpecContract{
		Version: "0.1.0",
		Name:    "Gherkin Specification Structure",
		RawData: make(map[string]interface{}),
	}

	return NewGherkinValidator(contract, nil)
}

func assertHasError(t *testing.T, errors []ValidationError, errorCode string) {
	t.Helper()

	if !hasErrorCode(errors, errorCode) {
		t.Errorf("expected error code %s, but not found in errors:\n%s",
			errorCode, FormatValidationErrors(errors))
	}
}

func hasErrorCode(errors []ValidationError, code string) bool {
	for _, err := range errors {
		if err.Code == code {
			return true
		}
	}
	return false
}

func TestHasVerificationTag(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want bool
	}{
		{
			name: "has @ov",
			tags: []string{"@ov"},
			want: true,
		},
		{
			name: "has @iv among others",
			tags: []string{"@deps:git", "@iv", "@depm:src-cli"},
			want: true,
		},
		{
			name: "no verification tags",
			tags: []string{"@deps:git", "@depm:src-cli"},
			want: false,
		},
		{
			name: "empty tags",
			tags: []string{},
			want: false,
		},
		{
			name: "all verification tags",
			tags: []string{"@ov", "@iv", "@pv", "@piv", "@ppv"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasVerificationTag(tt.tags)
			if got != tt.want {
				t.Errorf("hasVerificationTag(%v) = %v, want %v",
					tt.tags, got, tt.want)
			}
		})
	}
}

func TestGherkinValidator_ScenarioOutline(t *testing.T) {
	validator := createTestValidator()

	scenarioOutline := `Feature: src-commands_test
  As a user
  I want something

  Rule: Test rule

    @ov
    Scenario Outline: Test with examples
      Given I have <count> items
      When I process them
      Then I should see <result>

      Examples:
        | count | result |
        | 1     | one    |
        | 2     | two    |
`

	errors := validator.Validate(scenarioOutline, nil)

	if len(errors) != 0 {
		t.Errorf("expected no errors for Scenario Outline, got %d error(s):\n%s",
			len(errors), FormatValidationErrors(errors))
	}
}

func TestGherkinValidator_TagsOnSeparateLines(t *testing.T) {
	validator := createTestValidator()

	tagsOnSeparateLines := `Feature: src-commands_test
  As a user
  I want something

  Rule: Test rule

    @ov
    @deps:git
    @depm:src-cli
    Scenario: Scenario with tags on separate lines
      Given something
      When something happens
      Then something occurs
`

	errors := validator.Validate(tagsOnSeparateLines, nil)

	// Should be valid - @ov is present
	if hasErrorCode(errors, "MISSING_VERIFICATION_TAG") {
		t.Errorf("should recognize @ov on separate line")
	}
}

func TestGherkinValidator_Background(t *testing.T) {
	validator := createTestValidator()

	withBackground := `Feature: src-commands_test
  As a user
  I want something

  Background:
    Given I am logged in
    And the system is initialized

  Rule: Test rule

    @ov
    Scenario: Test scenario
      Given something specific
      When something happens
      Then something occurs
`

	errors := validator.Validate(withBackground, nil)

	if len(errors) != 0 {
		t.Errorf("expected no errors with Background, got %d error(s):\n%s",
			len(errors), FormatValidationErrors(errors))
	}
}

func TestGherkinValidator_MultipleRules(t *testing.T) {
	validator := createTestValidator()

	multipleRules := `Feature: src-commands_test
  As a user
  I want something

  Rule: First acceptance criterion

    @ov
    Scenario: First scenario
      Given something
      When something happens
      Then something occurs

  Rule: Second acceptance criterion

    @ov
    Scenario: Second scenario
      Given something else
      When another thing happens
      Then another thing occurs
`

	errors := validator.Validate(multipleRules, nil)

	if len(errors) != 0 {
		t.Errorf("expected no errors with multiple Rules, got %d error(s):\n%s",
			len(errors), FormatValidationErrors(errors))
	}
}

func TestGherkinValidator_OnlyOneScenarioNeedsTag(t *testing.T) {
	// Test that EACH scenario needs a tag (not just one per file)
	validator := createTestValidator()

	oneScenarioMissingTag := `Feature: src-commands_test
  As a user
  I want something

  Rule: Test rule

    @ov
    Scenario: First scenario with tag
      Given something

    Scenario: Second scenario without tag
      Given something else
`

	errors := validator.Validate(oneScenarioMissingTag, nil)

	// Should have exactly one MISSING_VERIFICATION_TAG error for the second scenario
	tagErrorCount := 0
	for _, err := range errors {
		if err.Code == "MISSING_VERIFICATION_TAG" {
			tagErrorCount++
		}
	}

	if tagErrorCount != 1 {
		t.Errorf("expected 1 MISSING_VERIFICATION_TAG error, got %d", tagErrorCount)
		if len(errors) > 0 {
			t.Logf("All errors:\n%s", FormatValidationErrors(errors))
		}
	}
}
