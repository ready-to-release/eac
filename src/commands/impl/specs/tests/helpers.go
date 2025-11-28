// Package tests provides BDD step definitions for the specs commands.
//
// This file contains shared helper functions used across step definitions.
package tests

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

// ============================================================================
// Embedded Asset Files for Validation Tests
// ============================================================================

//go:embed assets/valid-spec.txt
var validSpecAsset string

//go:embed assets/invalid-spec-no-rule.txt
var invalidSpecNoRuleAsset string

//go:embed assets/invalid-spec-multiple-errors.txt
var invalidSpecMultipleErrorsAsset string

//go:embed assets/spec-with-warnings.txt
var specWithWarningsAsset string

//go:embed assets/spec-missing-verification-tags.txt
var specMissingVerificationTagsAsset string

//go:embed assets/spec-no-scenarios.txt
var specNoScenariosAsset string

//go:embed assets/spec-wrong-order.txt
var specWrongOrderAsset string

//go:embed assets/spec-fixable-tags.txt
var specFixableTagsAsset string

//go:embed assets/spec-bad-tags.txt
var specBadTagsAsset string

// ============================================================================
// Helper Functions
// ============================================================================

// createTestFile creates a file in the test directory
func createTestFile(relativePath, content string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	fullPath := filepath.Join(Ctx.TestDir, relativePath)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	Ctx.CreatedFiles = append(Ctx.CreatedFiles, fullPath)
	return nil
}

// cleanupTestFiles removes all files created during the test
func cleanupTestFiles() error {
	if Ctx == nil {
		return nil
	}

	for _, file := range Ctx.CreatedFiles {
		os.Remove(file)
		dir := filepath.Dir(file)
		os.Remove(dir) // Remove empty directories
	}

	Ctx.CreatedFiles = nil
	return nil
}

// validSpecContent returns a valid Gherkin specification from embedded asset
func validSpecContent() string {
	return validSpecAsset
}

// invalidSpecContent returns an invalid Gherkin specification (missing Rule) from embedded asset
func invalidSpecContent() string {
	return invalidSpecNoRuleAsset
}

// multipleErrorsSpecContent returns a specification with multiple validation errors from embedded asset
func multipleErrorsSpecContent() string {
	return invalidSpecMultipleErrorsAsset
}

// specWithWarningsContent returns a specification with warnings from embedded asset
func specWithWarningsContent() string {
	return specWithWarningsAsset
}

// specMissingVerificationTagsContent returns a specification with missing verification tags
func specMissingVerificationTagsContent() string {
	return specMissingVerificationTagsAsset
}

// specNoScenariosContent returns a specification with no scenarios
func specNoScenariosContent() string {
	return specNoScenariosAsset
}

// specWrongOrderContent returns a specification with wrong element order
func specWrongOrderContent() string {
	return specWrongOrderAsset
}

// specFixableTagsContent returns a specification with fixable tag issues
func specFixableTagsContent() string {
	return specFixableTagsAsset
}

// specBadTagsContent returns a specification with bad tags
func specBadTagsContent() string {
	return specBadTagsAsset
}

// getLastCreatedFileContent reads the content of the last file mentioned in stdout.
// The specs create command outputs "   File: <path>" when a file is created.
// This helper parses that output and reads the actual file content.
func getLastCreatedFileContent() (string, error) {
	if Ctx == nil {
		return "", fmt.Errorf("test context not initialized")
	}

	// Parse the file path from command output
	// Format: "   File: <path>"
	output := Ctx.CommandOutput
	lines := splitLines(output)
	var filePath string

	for _, line := range lines {
		if trimmed := trimPrefix(line, "   File: "); trimmed != line {
			filePath = trimmed
		}
	}

	if filePath == "" {
		// Fall back to checking CreatedFiles list
		if len(Ctx.CreatedFiles) > 0 {
			filePath = Ctx.CreatedFiles[len(Ctx.CreatedFiles)-1]
		}
	}

	if filePath == "" {
		return "", fmt.Errorf("no file path found in output:\n%s", output)
	}

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return string(content), nil
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// trimPrefix returns s without the provided leading prefix string.
// If s doesn't start with prefix, s is returned unchanged.
func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}
