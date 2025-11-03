package validator

import (
	"strings"
	"testing"
)

func TestCommandValidator_ValidateCommand(t *testing.T) {
	cv := NewCommandValidator()

	tests := []struct {
		name          string
		args          []string
		wantValid     bool
		wantErrors    int
		wantWarnings  int
		errorContains string
	}{
		{
			name:          "Empty command",
			args:          []string{},
			wantValid:     false,
			wantErrors:    1,
			errorContains: "No command provided",
		},
		{
			name:         "Valid binary only",
			args:         []string{"r2r"},
			wantValid:    true,
			wantWarnings: 1, // No subcommand warning
		},
		{
			name:      "Valid simple command",
			args:      []string{"r2r", "version"},
			wantValid: true,
		},
		{
			name:      "Valid with global flags",
			args:      []string{"r2r", "--r2r-debug", "version"},
			wantValid: true,
		},
		{
			name:          "Invalid binary name",
			args:          []string{"invalid", "version"},
			wantValid:     false,
			wantErrors:    1,
			errorContains: "Invalid binary name",
		},
		{
			name:          "Invalid subcommand",
			args:          []string{"r2r", "invalid"},
			wantValid:     false,
			wantErrors:    1,
			errorContains: "Invalid subcommand",
		},
		{
			name:          "Conflicting global flags",
			args:          []string{"r2r", "--r2r-debug", "--r2r-quiet", "version"},
			wantValid:     false,
			wantErrors:    1,
			errorContains: "Cannot use both --r2r-debug and --r2r-quiet",
		},
		{
			name:          "Run without extension",
			args:          []string{"r2r", "run"},
			wantValid:     false,
			wantErrors:    1,
			errorContains: "requires an extension name",
		},
		{
			name:          "Run with invalid extension",
			args:          []string{"r2r", "run", "-invalid"},
			wantValid:     false,
			wantErrors:    1,
			errorContains: "requires an extension name",
		},
		{
			name:      "Valid run command",
			args:      []string{"r2r", "run", "pwsh"},
			wantValid: true,
		},
		{
			name:      "Run with container args",
			args:      []string{"r2r", "run", "python", "script.py", "--verbose"},
			wantValid: true,
		},
		{
			name:         "Run with global flag in container args",
			args:         []string{"r2r", "run", "python", "script.py", "--r2r-debug"},
			wantValid:    true,
			wantWarnings: 1, // Warning about global flag in container args
		},
		{
			name:          "Metadata without extension",
			args:          []string{"r2r", "metadata"},
			wantValid:     false,
			wantErrors:    1,
			errorContains: "requires an extension name",
		},
		{
			name:      "Valid metadata command",
			args:      []string{"r2r", "metadata", "pwsh"},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cv.ValidateCommand(tt.args)

			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
				if !result.Valid {
					t.Errorf("Errors: %v", result.Errors)
				}
			}

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("Error count = %v, want %v", len(result.Errors), tt.wantErrors)
				t.Errorf("Errors: %v", result.Errors)
			}

			if tt.errorContains != "" && len(result.Errors) > 0 {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, got %v", tt.errorContains, result.Errors)
				}
			}

			if len(result.Warnings) != tt.wantWarnings {
				t.Errorf("Warning count = %v, want %v", len(result.Warnings), tt.wantWarnings)
				t.Errorf("Warnings: %v", result.Warnings)
			}
		})
	}
}

func TestCommandValidator_GetViperArguments(t *testing.T) {
	cv := NewCommandValidator()

	tests := []struct {
		name      string
		args      []string
		wantViper []string
	}{
		{
			name:      "Simple command",
			args:      []string{"r2r", "version"},
			wantViper: []string{"r2r", "version"},
		},
		{
			name:      "Run without container args",
			args:      []string{"r2r", "run", "pwsh"},
			wantViper: []string{"r2r", "run", "pwsh"},
		},
		{
			name:      "Run with container args",
			args:      []string{"r2r", "run", "python", "script.py", "--verbose"},
			wantViper: []string{"r2r", "run", "python"},
		},
		{
			name:      "Complex with global flags",
			args:      []string{"r2r", "--r2r-debug", "run", "pwsh", "Write-Host", "Hello"},
			wantViper: []string{"r2r", "--r2r-debug", "run", "pwsh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper := cv.GetViperArguments(tt.args)

			if len(viper) != len(tt.wantViper) {
				t.Errorf("GetViperArguments() = %v, want %v", viper, tt.wantViper)
				return
			}

			for i, arg := range viper {
				if arg != tt.wantViper[i] {
					t.Errorf("GetViperArguments()[%d] = %v, want %v", i, arg, tt.wantViper[i])
				}
			}
		})
	}
}

func TestCommandValidator_GetContainerArguments(t *testing.T) {
	cv := NewCommandValidator()

	tests := []struct {
		name          string
		args          []string
		wantContainer []string
	}{
		{
			name:          "No container args",
			args:          []string{"r2r", "version"},
			wantContainer: []string{},
		},
		{
			name:          "Run without container args",
			args:          []string{"r2r", "run", "pwsh"},
			wantContainer: []string{},
		},
		{
			name:          "Run with container args",
			args:          []string{"r2r", "run", "python", "script.py", "--verbose"},
			wantContainer: []string{"script.py", "--verbose"},
		},
		{
			name:          "Complex container args",
			args:          []string{"r2r", "run", "pwsh", "-Command", "Get-Date | Format-Table"},
			wantContainer: []string{"-Command", "Get-Date | Format-Table"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := cv.GetContainerArguments(tt.args)

			if len(container) != len(tt.wantContainer) {
				t.Errorf("GetContainerArguments() = %v, want %v", container, tt.wantContainer)
				return
			}

			for i, arg := range container {
				if arg != tt.wantContainer[i] {
					t.Errorf("GetContainerArguments()[%d] = %v, want %v", i, arg, tt.wantContainer[i])
				}
			}
		})
	}
}

func TestCommandValidator_IsViperArgument(t *testing.T) {
	cv := NewCommandValidator()

	args := []string{"r2r", "--r2r-debug", "run", "python", "script.py", "--verbose"}

	tests := []struct {
		position  int
		wantViper bool
	}{
		{0, true},  // r2r
		{1, true},  // --r2r-debug
		{2, true},  // run
		{3, true},  // python
		{4, false}, // script.py (container arg)
		{5, false}, // --verbose (container arg)
	}

	for _, tt := range tests {
		isViper := cv.IsViperArgument(args, tt.position)
		if isViper != tt.wantViper {
			t.Errorf("IsViperArgument(args, %d) = %v, want %v", tt.position, isViper, tt.wantViper)
		}
	}
}

func TestCommandValidator_ValidateExtensionName(t *testing.T) {
	cv := NewCommandValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid simple", "pwsh", false},
		{"Valid complex", "my-extension_123", false},
		{"Invalid empty", "", true},
		{"Invalid start", "123-ext", true},
		{"Invalid char", "ext@name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cv.ValidateExtensionName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExtensionName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCommandValidator_ValidateForRun(t *testing.T) {
	cv := NewCommandValidator()

	tests := []struct {
		name      string
		args      []string
		wantValid bool
	}{
		{
			name:      "Valid run command",
			args:      []string{"r2r", "run", "pwsh"},
			wantValid: true,
		},
		{
			name:      "Run without extension",
			args:      []string{"r2r", "run"},
			wantValid: false,
		},
		{
			name:      "Run with container args",
			args:      []string{"r2r", "run", "python", "script.py"},
			wantValid: true,
		},
		{
			name:      "Not a run command",
			args:      []string{"r2r", "version"},
			wantValid: true, // Should still be valid, just not a run command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cv.ValidateForRun(tt.args)
			if result.Valid != tt.wantValid {
				t.Errorf("ValidateForRun() valid = %v, want %v", result.Valid, tt.wantValid)
				if !result.Valid {
					t.Errorf("Errors: %v", result.Errors)
				}
			}
		})
	}
}

func TestCommandValidationResult_Summary(t *testing.T) {
	tests := []struct {
		name     string
		result   CommandValidationResult
		wantText string
	}{
		{
			name: "Valid no warnings",
			result: CommandValidationResult{
				Valid: true,
			},
			wantText: "Command valid",
		},
		{
			name: "Valid with warnings",
			result: CommandValidationResult{
				Valid:    true,
				Warnings: []string{"warning1", "warning2"},
			},
			wantText: "Command valid with 2 warning(s)",
		},
		{
			name: "Invalid",
			result: CommandValidationResult{
				Valid:    false,
				Errors:   []string{"error1"},
				Warnings: []string{"warning1"},
			},
			wantText: "Command invalid: 1 error(s), 1 warning(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.result.Summary()
			if summary != tt.wantText {
				t.Errorf("Summary() = %q, want %q", summary, tt.wantText)
			}
		})
	}
}
