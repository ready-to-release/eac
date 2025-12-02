//go:build L1 && ov
// +build L1,ov

// File: go/eac/commands/impl/specs/validation_test.go
package specs

import (
	"testing"
)

// Intent: Test Gherkin validation functionality
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Table-driven tests with clear test case names
//   - Each test focuses on a single validation aspect
//
// Easy to change:
//   - Test data is clearly separated from test logic
//
// Hard to break:
//   - Tests cover happy path and error cases
//   - Edge cases are explicitly tested

func TestValidateGherkinContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid Gherkin with Feature, Rule, Scenario",
			content: `Feature: test_feature
  As a user
  I want something

  Rule: Acceptance Criterion

    Scenario: Happy path
      Given a precondition
      When an action
      Then an outcome`,
			wantErr: false,
		},
		{
			name:    "missing Feature declaration",
			content: "  As a user\n  I want something",
			wantErr: true,
		},
		{
			name:    "missing Rule declaration",
			content: "Feature: test\n  Scenario: test\n    Given something",
			wantErr: true,
		},
		{
			name:    "missing Scenario declaration",
			content: "Feature: test\n  Rule: test\n    Given something",
			wantErr: true,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGherkinContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGherkinContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGherkinContent_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid complete Gherkin",
			content: `Feature: test
  Rule: test
    Scenario: test
      Given test`,
			wantErr: false,
		},
		{
			name:    "empty string",
			content: "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			content: "   \t\n   ",
			wantErr: true,
		},
		{
			name:    "missing Feature",
			content: "Rule: test\nScenario: test",
			wantErr: true,
		},
		{
			name:    "missing Rule",
			content: "Feature: test\nScenario: test",
			wantErr: true,
		},
		{
			name:    "missing Scenario",
			content: "Feature: test\nRule: test",
			wantErr: true,
		},
		{
			name: "Feature and Rule but scenario is misspelled",
			content: `Feature: test
  Rule: test
    Scenarion: test`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGherkinContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGherkinContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGherkinContent_StructuralIssues(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "Feature after Scenario (wrong order)",
			content: `Scenario: Test
  Given test
Feature: Test
  Rule: Test`,
			wantErr: true, // ✅ NEW: Improved validation detects wrong order
		},
		{
			name: "Rule before Feature (wrong order)",
			content: `Rule: Test
Feature: Test
  Scenario: Test`,
			wantErr: true, // ✅ NEW: Improved validation detects wrong order
		},
		{
			name: "Feature in comment",
			content: `# This comment mentions Feature: and Rule: and Scenario:`,
			wantErr: true, // ✅ NEW: Improved validation correctly fails (no actual Feature)
		},
		{
			name: "Keywords in strings",
			content: `Some text with "Feature:" in quotes
And "Rule:" here
Plus "Scenario:" there`,
			wantErr: true, // ✅ NEW: Improved validation correctly fails (no actual keywords)
		},
		{
			name: "Multiple Features",
			content: `Feature: First
  Rule: Test
    Scenario: Test

Feature: Second
  Rule: Test
    Scenario: Test`,
			wantErr: true, // ✅ NEW: Improved validation detects multiple Features
		},
		{
			name: "Feature with Background but no Scenario",
			content: `Feature: Test
  Background:
    Given setup
  Rule: Test
    # Missing Scenario`,
			wantErr: true, // Missing Scenario keyword
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGherkinContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGherkinContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
