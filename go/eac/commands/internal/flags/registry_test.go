package flags

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
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
	validFlags := []registry.FlagMetadata{
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

func TestFindFlagMetadata(t *testing.T) {
	flags := []registry.FlagMetadata{
		{Name: "ai-provider", Shorthand: "-a"},
		{Name: "debug", Shorthand: "-d"},
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
			result := findFlagMetadata(flags, tt.flagName)
			if tt.found {
				if result == nil {
					t.Fatalf("findFlagMetadata(%q) returned nil; want %q", tt.flagName, tt.expected)
				}
				if result.Name != tt.expected {
					t.Errorf("findFlagMetadata(%q).Name = %q; want %q", tt.flagName, result.Name, tt.expected)
				}
			} else {
				if result != nil {
					t.Errorf("findFlagMetadata(%q) returned %v; want nil", tt.flagName, result)
				}
			}
		})
	}
}

func TestValidateParsedFlags(t *testing.T) {
	flagMetadata := []registry.FlagMetadata{
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
			err := validateParsedFlags(tt.parsedFlags, flagMetadata)
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
	validFlags := []registry.FlagMetadata{
		{Name: "ai-token", Shorthand: "-a", Type: "string", Usage: "AI API token"},
		{Name: "debug", Shorthand: "-d", Type: "bool", Usage: "Enable debug mode"},
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

func TestExtractCommandName(t *testing.T) {
	// Create a temporary test file with // Command: comment
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package test

// Command: test-command
// This is a test command

func TestCommand() int {
	return 0
}
`
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	commandName, err := extractCommandName(testFile)
	if err != nil {
		t.Fatalf("extractCommandName() error = %v; want nil", err)
	}

	if commandName != "test-command" {
		t.Errorf("extractCommandName() = %q; want %q", commandName, "test-command")
	}
}

func TestExtractCommandName_MissingComment(t *testing.T) {
	// Create a temporary test file without // Command: comment
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package test

// This is a test command without Command: comment

func TestCommand() int {
	return 0
}
`
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, err := extractCommandName(testFile)
	if err == nil {
		t.Fatal("extractCommandName() expected error for missing comment; got nil")
	}

	if !strings.Contains(err.Error(), "no // Command: comment found") {
		t.Errorf("extractCommandName() error = %q; want 'no // Command: comment found'", err.Error())
	}
}

func TestTranslateCrossCompilePath(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	if err := os.WriteFile(testFile, []byte("package test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test with the file that exists
	translated, err := translateCrossCompilePath(testFile)
	if err != nil {
		t.Fatalf("translateCrossCompilePath() error = %v; want nil", err)
	}

	// Verify file exists at translated path
	if _, err := os.Stat(translated); err != nil {
		t.Errorf("translated path %q does not exist: %v", translated, err)
	}
}

// TestValidateFlagsFromRegistry_Integration tests the full validation flow
// Note: This requires actual registry initialization, so we'll mock it.
func TestValidateFlagsFromRegistry_MockScenario(t *testing.T) {
	// This test demonstrates the expected behavior but requires:
	// 1. Registry to be initialized with command
	// 2. Caller to be a registered command file
	// Since this is complex to set up in unit tests, we'll test components individually above
	t.Skip("Integration test - requires full registry setup")
}
