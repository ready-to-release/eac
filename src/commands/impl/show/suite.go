// Command: show suite
// Description: Display detailed information about a test suite
// HasSideEffects: false
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/repository"
	testing "github.com/ready-to-release/eac/src/core/testing"
)

func init() {
	registry.Register(ShowSuite)
}

// ShowSuite displays detailed information about a test suite in markdown table format
//
// Command: show suite <suite-moniker>
// Example: show suite commit
//
// Output format:
// - Suite header with metadata
// - Markdown table with one row per test
// - Columns: Test Name, Type, Module, Tags
func ShowSuite() int {
	// Parse arguments - expect suite moniker after "show suite"
	args := os.Args[1:]

	// Find where "show suite" ends
	suiteIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "show" && args[i+1] == "suite" {
			suiteIdx = i + 2
			break
		}
	}

	if suiteIdx == -1 || suiteIdx >= len(args) {
		fmt.Fprintf(os.Stderr, "Error: suite moniker required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: show suite <suite-moniker>\n\n")
		fmt.Fprintf(os.Stderr, "Available suites:\n")
		for _, moniker := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", moniker)
		}
		return 1
	}

	suiteMoniker := args[suiteIdx]

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not in a git repository: %v\n", err)
		return 1
	}

	// Get suite
	suite, err := testing.GetSuite(suiteMoniker)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Available suites:\n")
		for _, moniker := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", moniker)
		}
		return 1
	}

	// Phase 1: Discover all tests
	allTests, err := testing.DiscoverAllTests(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering tests: %v\n", err)
		return 1
	}

	// Phase 2: Apply inferences
	allTests = testing.ApplyInferences(allTests, suite.Inferences)

	// Phase 3: Select tests for this suite
	selectedTests := suite.SelectTests(allTests)

	// Phase 3.5: Filter out framework tests (tests about the testing framework itself)
	productionTests := []testing.TestReference{}
	frameworkTests := []testing.TestReference{}
	for _, test := range selectedTests {
		if testing.ShouldSkipValidation(test) {
			frameworkTests = append(frameworkTests, test)
		} else {
			productionTests = append(productionTests, test)
		}
	}

	// Phase 4: Validate post-inference tags
	validationErrors := testing.ValidateAllPostInference(productionTests, repoRoot)

	if len(validationErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: %d tests have validation errors:\n", len(validationErrors))
		if len(frameworkTests) > 0 {
			fmt.Fprintf(os.Stderr, "          (%d framework tests excluded from validation)\n", len(frameworkTests))
		}
		fmt.Fprintf(os.Stderr, "\n")
		for testName, errors := range validationErrors {
			fmt.Fprintf(os.Stderr, "  - %s:\n", testName)
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "    • %s\n", err)
			}
		}
		fmt.Fprintf(os.Stderr, "\n")
	} else if len(frameworkTests) > 0 {
		fmt.Fprintf(os.Stderr, "\n✓ All tests pass validation (%d framework tests excluded from display)\n\n", len(frameworkTests))
	}

	// Display suite information
	fmt.Printf("# Test Suite: %s\n\n", suite.Name)
	fmt.Printf("**Moniker**: `%s`  \n", suite.Moniker)
	fmt.Printf("**Description**: %s  \n", suite.Description)
	fmt.Printf("**Production Tests**: %d  \n", len(productionTests))
	fmt.Printf("**Framework Tests**: %d (excluded from display)  \n", len(frameworkTests))
	fmt.Printf("**Total Discovered**: %d  \n", len(allTests))
	fmt.Printf("\n")

	// Display selection criteria
	fmt.Printf("## Selection Criteria\n\n")
	for i, selector := range suite.Selectors {
		fmt.Printf("**Selector %d**:\n", i+1)
		if len(selector.AnyOfTags) > 0 {
			fmt.Printf("  - **AnyOf**: %s\n", strings.Join(selector.AnyOfTags, ", "))
		}
		if len(selector.RequireTags) > 0 {
			fmt.Printf("  - **RequireAll**: %s\n", strings.Join(selector.RequireTags, ", "))
		}
		if len(selector.ExcludeTags) > 0 {
			fmt.Printf("  - **Exclude**: %s\n", strings.Join(selector.ExcludeTags, ", "))
		}
		fmt.Printf("\n")
	}

	// Build file-to-module mapping once for efficiency
	fileModuleMap, err := buildFileModuleMap(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load file-module mapping: %v\n", err)
		fileModuleMap = make(map[string]string) // Use empty map
	}

	// Display tests in markdown table using TableBuilder
	fmt.Printf("## Production Tests\n\n")

	tb := render.NewTableBuilder().
		WithHeaders("#", "Test Name", "Type", "Module", "Level", "Verification", "System Deps", "Module Deps")

	for i, test := range productionTests {
		// Extract module from file path
		module := extractModuleFromPath(test.FilePath, fileModuleMap, repoRoot)

		// Separate tags by type
		levelTags := filterTagsByPrefix(test.Tags, "@L")
		verificationTags := filterTagsByPatterns(test.Tags, []string{"@ov", "@iv", "@pv", "@piv", "@ppv"})
		systemDeps := filterTagsByPrefix(test.Tags, "@deps:")
		moduleDeps := filterTagsByPrefix(test.Tags, "@depm:")

		// Format tag columns
		levelStr := strings.Join(levelTags, ", ")
		verificationStr := strings.Join(verificationTags, ", ")
		systemDepsStr := strings.Join(systemDeps, ", ")
		moduleDepsStr := strings.Join(moduleDeps, ", ")

		tb.AddRow(
			fmt.Sprintf("%d", i+1),
			test.TestName,
			test.Type,
			module,
			levelStr,
			verificationStr,
			systemDepsStr,
			moduleDepsStr,
		)
	}

	fmt.Println(tb.Build())
	fmt.Printf("\n")

	// Display summary statistics
	fmt.Printf("## Statistics\n\n")

	// Count by type
	typeCounts := make(map[string]int)
	for _, test := range productionTests {
		typeCounts[test.Type]++
	}

	fmt.Printf("**By Type**:\n")
	for testType, count := range typeCounts {
		fmt.Printf("  - %s: %d\n", testType, count)
	}
	fmt.Printf("\n")

	// Count by module
	moduleCounts := make(map[string]int)
	for _, test := range productionTests {
		module := extractModuleFromPath(test.FilePath, fileModuleMap, repoRoot)
		moduleCounts[module]++
	}

	fmt.Printf("**By Module**:\n")
	for module, count := range moduleCounts {
		fmt.Printf("  - %s: %d\n", module, count)
	}
	fmt.Printf("\n")

	// Extract and display dependencies
	systemDeps := testing.GetSystemDependencies(productionTests)
	moduleDeps := testing.GetModuleDependencies(productionTests)

	if len(systemDeps) > 0 || len(moduleDeps) > 0 {
		fmt.Printf("**Dependencies**:\n")
		if len(systemDeps) > 0 {
			fmt.Printf("  - System: %s\n", strings.Join(systemDeps, ", "))
		}
		if len(moduleDeps) > 0 {
			fmt.Printf("  - Module: %s\n", strings.Join(moduleDeps, ", "))
		}
		fmt.Printf("\n")
	}

	return 0
}

// buildFileModuleMap creates a map of file paths to module names
func buildFileModuleMap(repoRoot string) (map[string]string, error) {
	fileMap := make(map[string]string)

	// Get all files with module ownership
	files, err := repository.GetRepositoryFilesWithModules(true, false, false, repoRoot, "0.1.0")
	if err != nil {
		return nil, err
	}

	// Build the mapping
	for _, file := range files {
		if len(file.Modules) > 0 {
			fileMap[file.Name] = file.Modules[0] // Use first module
		}
	}

	return fileMap, nil
}

// filterTagsByPrefix filters tags that start with the given prefix (deduplicated)
func filterTagsByPrefix(tags []string, prefix string) []string {
	seen := make(map[string]bool)
	filtered := []string{}
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) && !seen[tag] {
			filtered = append(filtered, tag)
			seen[tag] = true
		}
	}
	return filtered
}

// filterTagsByPatterns filters tags that match any of the given exact patterns (deduplicated)
func filterTagsByPatterns(tags []string, patterns []string) []string {
	patternMap := make(map[string]bool)
	for _, p := range patterns {
		patternMap[p] = true
	}

	seen := make(map[string]bool)
	filtered := []string{}
	for _, tag := range tags {
		if patternMap[tag] && !seen[tag] {
			filtered = append(filtered, tag)
			seen[tag] = true
		}
	}
	return filtered
}

// extractModuleFromPath looks up the module for a file path using the file-module mapping
func extractModuleFromPath(filePath string, fileModuleMap map[string]string, repoRoot string) string {
	// Normalize separators
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	repoRoot = strings.ReplaceAll(repoRoot, "\\", "/")

	// Convert absolute path to relative path from repo root
	relativePath := filePath
	if strings.HasPrefix(filePath, repoRoot) {
		relativePath = strings.TrimPrefix(filePath, repoRoot)
		relativePath = strings.TrimPrefix(relativePath, "/")
	}

	// For specs/ files, extract module from path structure (specs/MODULE/...)
	parts := strings.Split(relativePath, "/")
	if len(parts) >= 2 && parts[0] == "specs" {
		return parts[1]
	}

	// For src/ files, look up in the file-module map
	if module, found := fileModuleMap[relativePath]; found {
		return module
	}

	// Try direct lookup as fallback
	if module, found := fileModuleMap[filePath]; found {
		return module
	}

	return "unknown"
}
