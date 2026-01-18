//go:build L1 && ov
// +build L1,ov

package commitmessage

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsModuleName_EdgeCases tests module name validation edge cases
func TestIsModuleName_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "single character module name",
			input: "a",
			want:  true,
		},
		{
			name:  "max length module name (50 chars)",
			input: "abcdefghij-klmnopqrst-uvwxyz0123456789-abcdefgh",
			want:  true,
		},
		{
			name:  "exceeds max length (51 chars)",
			input: "abcdefghij-klmnopqrst-uvwxyz0123456789-abcdefghijkl",
			want:  false,
		},
		{
			name:  "consecutive dashes",
			input: "module--name",
			want:  true,
		},
		{
			name:  "starting with dash",
			input: "-module",
			want:  true,
		},
		{
			name:  "ending with dash",
			input: "module-",
			want:  true,
		},
		{
			name:  "only dashes",
			input: "---",
			want:  true,
		},
		{
			name:  "with underscore",
			input: "module_name",
			want:  true,
		},
		{
			name:  "with numbers",
			input: "module123",
			want:  true,
		},
		{
			name:  "with uppercase (invalid)",
			input: "Module",
			want:  false,
		},
		{
			name:  "with space (invalid)",
			input: "module name",
			want:  false,
		},
		{
			name:  "with special char (invalid)",
			input: "module@name",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "with colon (invalid)",
			input: "module:name",
			want:  false,
		},
		{
			name:  "with period (invalid)",
			input: "module.name",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isModuleName(tt.input)
			if got != tt.want {
				t.Errorf("isModuleName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsDashesLine_EdgeCases tests dashes line validation edge cases
func TestIsDashesLine_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "exactly 3 dashes (minimum)",
			input: "---",
			want:  true,
		},
		{
			name:  "exactly 2 dashes (below minimum)",
			input: "--",
			want:  false,
		},
		{
			name:  "many dashes",
			input: "--------------------",
			want:  true,
		},
		{
			name:  "single dash",
			input: "-",
			want:  false,
		},
		{
			name:  "dashes with space",
			input: "- - -",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "underscores (invalid)",
			input: "___",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDashesLine(tt.input)
			if got != tt.want {
				t.Errorf("isDashesLine(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestLoadContractFromConfig_EdgeCases tests contract loading edge cases
func TestLoadContractFromConfig_EdgeCases(t *testing.T) {
	t.Run("missing ai-config.yml", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := LoadContractFromConfig(tmpDir)
		if err == nil {
			t.Error("LoadContractFromConfig() expected error for missing config, got nil")
		}
		if !contains(err.Error(), "failed to load commit-message config") {
			t.Errorf("LoadContractFromConfig() error = %q, want error containing 'failed to load commit-message config'", err.Error())
		}
	})
}

// TestLoadAntiCorruptionRules_EdgeCases tests anti-corruption rules loading edge cases
func TestLoadAntiCorruptionRules_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string // Returns path to test rules file
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing rules file",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/rules.yml"
			},
			wantErr: true,
			errMsg:  "failed to read anti-corruption rules file",
		},
		{
			name: "malformed YAML",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				rulesPath := filepath.Join(tmpDir, "rules.yml")
				content := `version: 0.1.0
forbidden_prefixes:
  - invalid: yaml: syntax{{{
`
				if err := os.WriteFile(rulesPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return rulesPath
			},
			wantErr: true,
			errMsg:  "failed to parse anti-corruption rules YAML",
		},
		{
			name: "valid rules file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				rulesPath := filepath.Join(tmpDir, "rules.yml")
				content := `version: 0.1.0
name: anti-corruption-rules
description: Test rules
forbidden_prefixes:
  - "# "
  - "**"
forbidden_contains:
  - "noise"
forbidden_emojis:
  - "🎉"
agent_signatures:
  - "Agent:"
`
				if err := os.WriteFile(rulesPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return rulesPath
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesPath := tt.setup(t)
			rules, err := LoadAntiCorruptionRules(rulesPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadAntiCorruptionRules() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("LoadAntiCorruptionRules() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("LoadAntiCorruptionRules() unexpected error: %v", err)
					return
				}
				if rules == nil {
					t.Error("LoadAntiCorruptionRules() returned nil rules")
				}
			}
		})
	}
}

// TestVerifyCommitMessageContract_Unicode tests Unicode handling
func TestVerifyCommitMessageContract_Unicode(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name: "Unicode in body text",
			message: `feat(cli): add feature

Auditor-Summary: Added feature with émojis and spëcial chars.

Body text with Unicode: café, naïve, 日本語, 中文, עברית, العربية.`,
			wantErr: false, // Should handle Unicode gracefully
		},
		{
			name: "Emoji in body (not in header)",
			message: `feat(cli): add feature

Auditor-Summary: Added feature.

Body text with emoji: 🎉 🚀 ✨`,
			wantErr: false,
		},
		{
			name: "Very long line with Unicode",
			message: `feat(cli): add feature

Auditor-Summary: Added feature.

This is a very long line with Unicode characters: café naïve résumé fiancé über schön日本語`,
			wantErr: false, // Will have LINE_TOO_LONG warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := VerifyCommitMessageContract(tt.message, []string{"cli"})
			hasErrors := false
			for _, err := range errors {
				if !err.IsWarning() {
					hasErrors = true
					break
				}
			}

			if hasErrors != tt.wantErr {
				t.Errorf("VerifyCommitMessageContract() hasErrors = %v, want %v\nErrors: %v", hasErrors, tt.wantErr, errors)
			}
		})
	}
}

// TestVerifyCommitMessageContract_EmptyAndInvalid tests empty and invalid inputs
func TestVerifyCommitMessageContract_EmptyAndInvalid(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr string
	}{
		{
			name:    "empty message",
			message: "",
			wantErr: "INVALID_HEADER_FORMAT", // Empty message still triggers header validation
		},
		{
			name:    "only whitespace",
			message: "   \n\n\t  ",
			wantErr: "INVALID_HEADER_FORMAT",
		},
		{
			name:    "invalid header format",
			message: "This is not a valid commit header",
			wantErr: "INVALID_HEADER_FORMAT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := VerifyCommitMessageContract(tt.message, nil)
			if len(errors) == 0 {
				t.Error("VerifyCommitMessageContract() expected errors, got none")
				return
			}

			found := false
			for _, err := range errors {
				if err.Code == tt.wantErr {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("VerifyCommitMessageContract() expected error code %q, got errors: %v", tt.wantErr, errors)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
