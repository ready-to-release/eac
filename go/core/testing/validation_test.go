//go:build L1

package testing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFeatureLevelTags(t *testing.T) {
	tests := []struct {
		name           string
		featureContent string
		expectErrors   int
		errorContains  string
	}{
		{
			name: "Feature with L2, Scenario with L3 - should error",
			featureContent: `@env:isolated-test-project
@L2 @ov @depm:clie-installer
Feature: clie-installer_cli-installation

  As a user
  I want to install the CLI

  Rule: Installer downloads and installs CLI binary

    @L3 @ov @deps:windows @depm:clie-installer
    Scenario: Install latest version on Windows
      Given I am on Windows
      When I run the PowerShell installer
      Then the latest version is fetched
`,
			expectErrors:  1,
			errorContains: "Feature has L-tag @L2 and Scenario",
		},
		{
			name: "Feature with L2, Scenario without L-tag - should not error",
			featureContent: `@L2 @ov @depm:clie
Feature: clie_cli-invocation

  As a user
  I want to run the CLI

  Rule: CLI can be invoked

    @ov @depm:clie
    Scenario: Basic invocation
      Given I have the CLI installed
      When I run "clie --help"
      Then I see help output
`,
			expectErrors:  0,
			errorContains: "",
		},
		{
			name: "Feature without L-tag, Scenario with L3 - should not error",
			featureContent: `@ov @depm:clie-installer
Feature: clie-installer_cli-installation

  As a user
  I want to install the CLI

  Rule: Installer downloads and installs CLI binary

    @L3 @ov @deps:windows @depm:clie-installer
    Scenario: Install latest version on Windows
      Given I am on Windows
      When I run the PowerShell installer
      Then the latest version is fetched
`,
			expectErrors:  0,
			errorContains: "",
		},
		{
			name: "Multiple scenarios, some with conflicting L-tags",
			featureContent: `@L2 @ov @depm:eac
Feature: eac-cli_commit

  As a developer
  I want to generate commit messages

  Rule: Commit message generation

    @L2 @ov @depm:eac
    Scenario: Generate basic commit message
      Given I have staged changes
      When I run "clie commit message"
      Then a commit message is generated

    @L3 @ov @depm:eac
    Scenario: Generate commit with AI
      Given I have staged changes
      When I run "clie commit message --ai"
      Then an AI-generated message is returned
`,
			expectErrors:  1,
			errorContains: "Feature has L-tag @L2 and Scenario 'Generate commit with AI'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			featurePath := filepath.Join(tmpDir, "test.feature")

			err := os.WriteFile(featurePath, []byte(tt.featureContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write test feature file: %v", err)
			}

			// Run validation
			errors, err := ValidateFeatureLevelTags(featurePath)
			if err != nil {
				t.Fatalf("ValidateFeatureLevelTags returned error: %v", err)
			}

			// Check error count
			if len(errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d. Errors: %v", tt.expectErrors, len(errors), errors)
			}

			// Check error content
			if tt.expectErrors > 0 && tt.errorContains != "" {
				found := false
				for _, errMsg := range errors {
					if containsSubstring(errMsg, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, got errors: %v", tt.errorContains, errors)
				}
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	if len(s) > 0 && len(substr) > 0 {
		// Use simple substring check
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
	}
	return false
}
