// Package mock provides AI-related utilities including mock support for testing.
package mock

import (
	"os"
	"path/filepath"
	"strings"
)

// Environment variable names for mock configuration.
const (
	// EnvMockAIDir is the base directory containing mock response files.
	// Structure: $CLIE_MOCK_AI_DIR/<command>/mock-response.txt.
	EnvMockAIDir = "CLIE_MOCK_AI_DIR"

	// EnvMockAIPrefix is the prefix for command-specific mock file overrides.
	// Example: CLIE_MOCK_AI_SPECS=mock-response-conflict.txt.
	EnvMockAIPrefix = "CLIE_MOCK_AI_"
)

// GetMockResponse returns a mock AI response for the given command if configured.
//
// It checks in order:
// 1. CLIE_MOCK_AI_<COMMAND> env var for a specific mock file name
// 2. CLIE_MOCK_AI_DIR/<command>/mock-response.{txt,md,json} as defaults
//
// Returns the mock content and true if found, empty string and false otherwise.
//
// Example usage in a command:
//
//	if mock, ok := ai.GetMockResponse("specs"); ok {
//	    return mock, nil  // Use mock instead of real AI
//	}
//	// ... real AI call
func GetMockResponse(command string) (string, bool) {
	mockDir := os.Getenv(EnvMockAIDir)
	if mockDir == "" {
		return "", false
	}

	commandDir := filepath.Join(mockDir, command)

	// Check for command-specific override via env var
	// e.g., CLIE_MOCK_AI_SPECS=mock-response-conflict.txt
	envKey := EnvMockAIPrefix + strings.ToUpper(strings.ReplaceAll(command, "-", "_"))
	if specificFile := os.Getenv(envKey); specificFile != "" {
		mockFile := filepath.Join(commandDir, specificFile)
		if content, err := os.ReadFile(mockFile); err == nil {
			return string(content), true
		}
	}

	// Try default mock file patterns
	defaults := []string{
		"mock-response.txt",
		"mock-response.md",
		"mock-response.json",
	}
	for _, name := range defaults {
		mockFile := filepath.Join(commandDir, name)
		if content, err := os.ReadFile(mockFile); err == nil {
			return string(content), true
		}
	}

	return "", false
}

// GetMockResponseWithSubcommand returns a mock for a specific subcommand.
//
// It checks in order:
// 1. CLIE_MOCK_AI_<COMMAND>_<SUBCOMMAND> env var
// 2. CLIE_MOCK_AI_DIR/<command>/mock-<subcommand>-response.{txt,md,json}
// 3. Falls back to GetMockResponse(command) for general command mock
//
// Example: GetMockResponseWithSubcommand("risks", "assessment")
// checks CLIE_MOCK_AI_RISKS_ASSESSMENT, then risks/mock-assessment-response.*.
func GetMockResponseWithSubcommand(command, subcommand string) (string, bool) {
	mockDir := os.Getenv(EnvMockAIDir)
	if mockDir == "" {
		return "", false
	}

	commandDir := filepath.Join(mockDir, command)

	// Check for subcommand-specific override via env var
	// e.g., CLIE_MOCK_AI_RISKS_ASSESSMENT=custom-assessment.md
	envKey := EnvMockAIPrefix + strings.ToUpper(strings.ReplaceAll(command, "-", "_")) +
		"_" + strings.ToUpper(strings.ReplaceAll(subcommand, "-", "_"))
	if specificFile := os.Getenv(envKey); specificFile != "" {
		mockFile := filepath.Join(commandDir, specificFile)
		if content, err := os.ReadFile(mockFile); err == nil {
			return string(content), true
		}
	}

	// Try subcommand-specific patterns
	// e.g., mock-assessment-response.md for subcommand "assessment"
	patterns := []string{
		"mock-" + subcommand + "-response.txt",
		"mock-" + subcommand + "-response.md",
		"mock-" + subcommand + "-response.json",
	}
	for _, name := range patterns {
		mockFile := filepath.Join(commandDir, name)
		if content, err := os.ReadFile(mockFile); err == nil {
			return string(content), true
		}
	}

	// Fall back to general command mock
	return GetMockResponse(command)
}

// IsMockEnabled returns true if mock AI responses are configured.
func IsMockEnabled() bool {
	return os.Getenv(EnvMockAIDir) != ""
}
