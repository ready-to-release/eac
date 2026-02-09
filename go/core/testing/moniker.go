package testing

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reCamelCaseBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	reMultipleHyphens   = regexp.MustCompile(`-+`)
)

// GenerateTestMoniker generates a unique moniker for a test based on its type.
// Uses the adapter registry to determine moniker style:
// - "feature" style: module_feature-name_scenario-name (for BDD/Gherkin tests)
// - "file" style: module_test-file_TestName (for unit tests)
func GenerateTestMoniker(testRef TestReference, module string) string {
	style := getMonikerStyle(testRef.Type)
	if style == "feature" {
		return generateBDDMoniker(testRef, module)
	}
	return generateGoTestMoniker(testRef, module)
}

// generateBDDMoniker creates moniker for BDD/Gherkin tests (godog, tscucumber).
// Format: module_feature-name_scenario-name
// Example: clie_cli-invocation_version-flag-displays-version.
func generateBDDMoniker(testRef TestReference, module string) string {
	// Extract feature name from file path
	// Path: specs/clie/cli-invocation/specification.feature
	// Feature: cli-invocation
	featureName := extractFeatureName(testRef.FilePath)

	// Convert scenario name to kebab-case
	scenarioName := toKebabCase(testRef.TestName)

	// Build moniker: module_feature_scenario
	if module == "" {
		module = "unknown"
	}

	return module + "_" + featureName + "_" + scenarioName
}

// generateGoTestMoniker creates moniker for Go unit tests
// Format: module_test-file_TestName
// Example: clie_install-test_TestInstallCommand-CreateConfigFile.
func generateGoTestMoniker(testRef TestReference, module string) string {
	// Extract test file name without extension
	// Path: C:\projects\eac\go\clie\cli\cmd\install_test.go
	// File: install_test
	fileName := extractTestFileName(testRef.FilePath)

	// Convert test name to kebab-case
	testName := toKebabCase(testRef.TestName)

	// Build moniker: module_file_testname
	if module == "" {
		module = "unknown"
	}

	return module + "_" + fileName + "_" + testName
}

// extractFeatureName extracts the feature directory name from a feature file path
// specs/clie/cli-invocation/specification.feature -> cli-invocation.
func extractFeatureName(filePath string) string {
	normalized := filepath.ToSlash(filePath)
	parts := strings.Split(normalized, "/")

	// Find "specs" in path and get the feature directory
	for i, part := range parts {
		if part == "specs" && i+2 < len(parts) {
			// parts[i+1] is module, parts[i+2] is feature directory
			return parts[i+2]
		}
	}

	// Fallback: use last directory before filename
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}

	return "unknown-feature"
}

// extractTestFileName extracts the test file name without _test.go suffix
// C:\projects\eac\go\clie\cli\cmd\install_test.go -> install-test.
func extractTestFileName(filePath string) string {
	base := filepath.Base(filePath)

	// Remove .go extension
	name := strings.TrimSuffix(base, ".go")

	// Convert to kebab-case (install_test -> install-test)
	return toKebabCase(name)
}

// toKebabCase converts a string to kebab-case
// Handles: PascalCase, snake_case, spaces, etc.
func toKebabCase(s string) string {
	// Replace underscores with hyphens
	s = strings.ReplaceAll(s, "_", "-")

	// Insert hyphens before uppercase letters (for PascalCase/camelCase)
	// TestInstallCommand -> Test-Install-Command
	s = reCamelCaseBoundary.ReplaceAllString(s, "${1}-${2}")

	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove multiple consecutive hyphens
	s = reMultipleHyphens.ReplaceAllString(s, "-")

	// Trim leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s
}
