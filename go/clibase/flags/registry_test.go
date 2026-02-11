package flags

import (
	"strings"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "abcd", 1},
		{"abc", "adc", 1},
		{"api-token", "ai-token", 1},
		{"module", "modules", 1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			result := levenshteinDistance(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestSuggestSimilarFlags(t *testing.T) {
	validFlags := []core.FlagSpec{
		{Name: "ai-token"},
		{Name: "ai-provider"},
		{Name: "module"},
		{Name: "force"},
	}

	tests := []struct {
		unknownFlag string
		expected    []string
	}{
		{"api-token", []string{"ai-token"}}, // 1 char difference
		{"ai-tokn", []string{"ai-token"}},   // 1 char difference
		{"modules", []string{"module"}},     // 1 char difference
		{"xyz", nil},                        // Too different
	}

	for _, tt := range tests {
		t.Run(tt.unknownFlag, func(t *testing.T) {
			result := suggestSimilarFlags(tt.unknownFlag, validFlags)
			if len(result) != len(tt.expected) {
				t.Fatalf("suggestSimilarFlags(%q) returned %d suggestions; want %d", tt.unknownFlag, len(result), len(tt.expected))
			}
			for i, suggestion := range result {
				if suggestion != tt.expected[i] {
					t.Errorf("suggestion[%d] = %q; want %q", i, suggestion, tt.expected[i])
				}
			}
		})
	}
}

func TestFindFlag(t *testing.T) {
	flags := []core.FlagSpec{
		{Name: "ai-provider", Shorthand: "a"},
		{Name: "debug", Shorthand: "d"},
		{Name: "module"},
	}

	tests := []struct {
		name     string
		flagName string
		found    bool
		expected string
	}{
		{"full name", "--ai-provider", true, "ai-provider"},
		{"full name no dashes", "ai-provider", true, "ai-provider"},
		{"shorthand", "-a", true, "ai-provider"},
		{"shorthand no dash", "a", true, "ai-provider"},
		{"double dash shorthand", "--a", true, "ai-provider"},
		{"no shorthand", "--module", true, "module"},
		{"unknown", "--unknown", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findFlag(flags, tt.flagName)
			if tt.found {
				if result == nil {
					t.Fatalf("findFlag(%q) returned nil; want %q", tt.flagName, tt.expected)
				}
				if result.Name != tt.expected {
					t.Errorf("findFlag(%q).Name = %q; want %q", tt.flagName, result.Name, tt.expected)
				}
			} else {
				if result != nil {
					t.Errorf("findFlag(%q) returned %v; want nil", tt.flagName, result)
				}
			}
		})
	}
}

func TestValidateParsedFlags(t *testing.T) {
	flagSpecs := []core.FlagSpec{
		{Name: "ai-provider", Type: "string", Required: true},
		{Name: "debug", Type: "bool", Required: false},
		{Name: "module", Type: "string", Required: false},
	}

	tests := []struct {
		name        string
		parsedFlags map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid flags with required",
			parsedFlags: map[string]string{"ai-provider": "claude-api", "debug": ""},
			expectError: false,
		},
		{
			name:        "missing required flag",
			parsedFlags: map[string]string{"debug": ""},
			expectError: true,
			errorMsg:    "required flag missing: --ai-provider",
		},
		{
			name:        "unknown flag",
			parsedFlags: map[string]string{"ai-provider": "claude-api", "unknown": "value"},
			expectError: true,
			errorMsg:    "Unknown flag: --unknown",
		},
		{
			name:        "only required flags",
			parsedFlags: map[string]string{"ai-provider": "claude-api"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParsedFlags(tt.parsedFlags, flagSpecs)
			if tt.expectError {
				if err == nil {
					t.Fatalf("validateParsedFlags() expected error; got nil")
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateParsedFlags() error = %q; want to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateParsedFlags() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBuildUnknownFlagError(t *testing.T) {
	validFlags := []core.FlagSpec{
		{Name: "ai-token", Shorthand: "a", Type: "string", Usage: "AI API token"},
		{Name: "debug", Shorthand: "d", Type: "bool", Usage: "Enable debug mode"},
	}

	err := buildUnknownFlagError("api-token", validFlags)
	if err == nil {
		t.Fatal("buildUnknownFlagError() returned nil; want error")
	}

	errMsg := err.Error()

	// Check that error contains expected sections
	if !strings.Contains(errMsg, "Unknown flag: --api-token") {
		t.Errorf("error missing 'Unknown flag' header")
	}
	if !strings.Contains(errMsg, "Valid flags:") {
		t.Errorf("error missing 'Valid flags' section")
	}
	if !strings.Contains(errMsg, "--ai-token") {
		t.Errorf("error missing valid flag name")
	}
	if !strings.Contains(errMsg, "Did you mean:") {
		t.Errorf("error missing 'Did you mean' section")
	}
	if !strings.Contains(errMsg, "ai-token") {
		t.Errorf("error missing suggestion")
	}
}

func TestValidateFlags(t *testing.T) {
	specs := []core.FlagSpec{
		{Name: "output", Shorthand: "o", Type: "string"},
		{Name: "verbose", Shorthand: "v", Type: "bool"},
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"valid long flag", []string{"--output", "json"}, false},
		{"valid short flag", []string{"-o", "json"}, false},
		{"valid bool flag", []string{"--verbose"}, false},
		{"valid equals syntax", []string{"--output=json"}, false},
		{"unknown flag", []string{"--unknown"}, true},
		{"positional args ignored", []string{"foo", "bar"}, false},
		{"help always accepted", []string{"--help"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFlags(tt.args, specs)
			if tt.expectError && err == nil {
				t.Errorf("ValidateFlags() expected error; got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateFlags() unexpected error: %v", err)
			}
		})
	}
}
