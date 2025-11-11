// Feature: src-commands_design-command
// Unit tests for Structurizr validator with Docker
package design

import (
	"strings"
	"testing"
)

// TestParseValidationOutput_Success tests parsing successful validation output
func TestParseValidationOutput_Success(t *testing.T) {
	validator := &StructurizrValidatorImpl{}

	output := `Structurizr CLI v1.30.0
Validating workspace...
Workspace is valid`

	result := validator.parseValidationOutput(output)

	if !result.Valid {
		t.Errorf("Expected Valid=true, got Valid=false")
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}

	if len(result.Warnings) != 0 {
		t.Errorf("Expected 0 warnings, got %d", len(result.Warnings))
	}
}

// TestParseValidationOutput_WithErrors tests parsing output with errors
func TestParseValidationOutput_WithErrors(t *testing.T) {
	validator := &StructurizrValidatorImpl{}

	output := `Structurizr CLI v1.30.0
Validating workspace...
ERROR: Line 15: Unknown element type 'Database'
- api is not a valid identifier (expected: lowercase letters, numbers, and hyphens)`

	result := validator.parseValidationOutput(output)

	if result.Valid {
		t.Errorf("Expected Valid=false, got Valid=true")
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	// Check first error has line number
	if result.Errors[0].Line != 15 {
		t.Errorf("Expected first error at line 15, got line %d", result.Errors[0].Line)
	}

	// Check error messages
	if !strings.Contains(result.Errors[0].Message, "Unknown element type") {
		t.Errorf("Expected error message to contain 'Unknown element type', got: %s", result.Errors[0].Message)
	}

	if !strings.Contains(result.Errors[1].Message, "is not a valid identifier") {
		t.Errorf("Expected error message to contain 'is not a valid identifier', got: %s", result.Errors[1].Message)
	}
}

// TestParseValidationOutput_WithWarnings tests parsing output with warnings
func TestParseValidationOutput_WithWarnings(t *testing.T) {
	validator := &StructurizrValidatorImpl{}

	output := `Structurizr CLI v1.30.0
Validating workspace...
WARNING: Line 10: Element has no description
Workspace is valid`

	result := validator.parseValidationOutput(output)

	if !result.Valid {
		t.Errorf("Expected Valid=true (warnings don't invalidate), got Valid=false")
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	// Check warning has line number
	if result.Warnings[0].Line != 10 {
		t.Errorf("Expected warning at line 10, got line %d", result.Warnings[0].Line)
	}
}

// TestParseValidationOutput_EmptyOutput tests parsing empty output
func TestParseValidationOutput_EmptyOutput(t *testing.T) {
	validator := &StructurizrValidatorImpl{}

	output := ""

	result := validator.parseValidationOutput(output)

	if !result.Valid {
		t.Errorf("Expected Valid=true for empty output, got Valid=false")
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

// TestParseValidationOutput_MixedErrorsAndWarnings tests parsing output with both
func TestParseValidationOutput_MixedErrorsAndWarnings(t *testing.T) {
	validator := &StructurizrValidatorImpl{}

	output := `Structurizr CLI v1.30.0
Validating workspace...
ERROR: Invalid syntax at line 5
WARNING: Line 10: Missing documentation
- element123 is not a valid identifier
WARNING: Unused relationship`

	result := validator.parseValidationOutput(output)

	if result.Valid {
		t.Errorf("Expected Valid=false (has errors), got Valid=true")
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	if len(result.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(result.Warnings))
	}
}

// TestValidationResultJSON tests JSON marshaling of ValidationResult
func TestValidationResultJSON(t *testing.T) {
	result := &ValidationResult{
		Module:        "test-module",
		WorkspacePath: "/path/to/workspace.dsl",
		Valid:         false,
		Errors: []ValidationMessage{
			{
				Severity: "error",
				Message:  "Test error",
				Line:     15,
			},
		},
		Warnings: []ValidationMessage{},
	}

	// Test that we can marshal to JSON
	_, err := result.MarshalJSON()
	if err != nil {
		t.Errorf("Failed to marshal ValidationResult to JSON: %v", err)
	}
}

// TestValidationSummaryJSON tests JSON marshaling of ValidationSummary
func TestValidationSummaryJSON(t *testing.T) {
	summary := &ValidationSummary{
		TotalModules:  3,
		PassedModules: 2,
		FailedModules: 1,
		TotalErrors:   2,
		TotalWarnings: 1,
		Results:       []ValidationResult{},
	}

	// Test that we can marshal to JSON
	_, err := summary.MarshalJSON()
	if err != nil {
		t.Errorf("Failed to marshal ValidationSummary to JSON: %v", err)
	}
}

// TestFormatDockerVolume tests Windows path conversion for Docker volumes
func TestFormatDockerVolume(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows absolute path",
			input:    `C:\source\simply-cli\cli\specs\src-cli\design`,
			expected: "/c/source/simply-cli/cli/specs/src-cli/design",
		},
		{
			name:     "Windows path with different drive",
			input:    `D:\projects\myapp\workspace`,
			expected: "/d/projects/myapp/workspace",
		},
		{
			name:     "Unix path (no conversion needed)",
			input:    "/home/user/project/workspace",
			expected: "/home/user/project/workspace",
		},
		{
			name:     "Relative path (no conversion)",
			input:    "specs/src-cli/design",
			expected: "specs/src-cli/design",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDockerVolume(tt.input)
			if result != tt.expected {
				t.Errorf("formatDockerVolume(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
