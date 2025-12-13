// Command: test debug
// Short: Parse test results and list all failures
// Long: Parse test results (Go test JSON and Cucumber JSON) in out/test directory
// Long: and list all failed tests with their locations.
// Long:
// Long: This command scans for test-results.json (Go test JSON) and .cucumber.json
// Long: files, parses test results, and extracts failure information.
// Long:
// Long: Results are presented in a clear table format showing test failures.
// Long:
// Long: Expected Output:
// Long:   - Table of failed tests with test name, package, and error details
// Long:   - File locations (for Cucumber tests) with line numbers
// Long:   - Parses out/test/**/*.json files (both Go test JSON and Cucumber JSON)
// Long:   - If no failures found, displays success message
// Long:
// Long: Example:
// Long:   test debug
package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/cucumber"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/testjson"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// ansiRegex matches ANSI escape sequences for color/formatting
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func init() {
	registry.Register(TestDebug)
}

// Failure represents a test failure found in test results
type Failure struct {
	TestName    string // Name of the failed test
	Package     string // Package name
	File        string // Source file (for Cucumber) or package (for Go tests)
	Line        int    // Line number (for Cucumber)
	ErrorOutput string // Error output
	Source      string // "go-test" or "cucumber"
}

// TestDebug parses test results and lists all failures
func TestDebug() int {
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		log.Errorf("failed to get workspace root: %v", err)
		return 1
	}

	testOutputDir := paths.TestOutputDir(workspaceRoot)

	// Find all test result files
	goTestFiles, err := findGoTestJSONFiles(testOutputDir)
	if err != nil {
		log.Errorf("Error finding Go test JSON files: %v", err)
		return 1
	}

	cucumberFiles, err := findCucumberJSONFiles(testOutputDir)
	if err != nil {
		log.Errorf("Error finding Cucumber JSON files: %v", err)
		return 1
	}

	if len(goTestFiles) == 0 && len(cucumberFiles) == 0 {
		if _, err := os.Stat(testOutputDir); os.IsNotExist(err) {
			log.Info("No test output directory found. Run tests first to generate output.")
		} else {
			log.Info("No test results found. Run tests to generate output.")
		}
		return 0
	}

	// Collect failures from all sources
	var allFailures []Failure

	// Parse Go test JSON files
	for _, jsonFile := range goTestFiles {
		failures, err := collectFailuresFromGoTestJSON(jsonFile)
		if err != nil {
			log.Errorf("Warning: failed to parse %s: %v", jsonFile, err)
			continue
		}
		allFailures = append(allFailures, failures...)
	}

	// Parse Cucumber JSON files
	for _, jsonFile := range cucumberFiles {
		failures, err := collectFailuresFromCucumberJSON(jsonFile)
		if err != nil {
			log.Errorf("Warning: failed to parse %s: %v", jsonFile, err)
			continue
		}
		allFailures = append(allFailures, failures...)
	}

	if len(allFailures) == 0 {
		log.Info("No test failures found. All tests passed!")
		return 0
	}

	// Print results in table format
	printFailureTable(allFailures)

	return 0
}

// findGoTestJSONFiles finds all Go test JSON files
// Looks for test-results.json (from test module) and *.json (from test suite)
// Excludes .cucumber.json files
func findGoTestJSONFiles(dir string) ([]string, error) {
	var jsonFiles []string

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return jsonFiles, nil
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			// Exclude Cucumber JSON files
			if !strings.HasSuffix(path, ".cucumber.json") {
				jsonFiles = append(jsonFiles, path)
			}
		}
		return nil
	})

	return jsonFiles, err
}

// findCucumberJSONFiles finds all .cucumber.json files
func findCucumberJSONFiles(dir string) ([]string, error) {
	var jsonFiles []string

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return jsonFiles, nil
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".cucumber.json") {
			jsonFiles = append(jsonFiles, path)
		}
		return nil
	})

	return jsonFiles, err
}

// collectFailuresFromGoTestJSON parses a Go test JSON file
func collectFailuresFromGoTestJSON(jsonPath string) ([]Failure, error) {
	events, err := testjson.ParseJSONFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var failures []Failure
	failedTestOutputs := testjson.ExtractFailedTests(events)

	for key, outputs := range failedTestOutputs {
		parts := strings.SplitN(key, "::", 2)
		if len(parts) != 2 {
			continue
		}

		pkg := parts[0]
		test := parts[1]

		// Collect all output lines except RUN/PASS/FAIL markers
		var errorLines []string
		for _, line := range outputs {
			// Strip ANSI escape codes for clean output
			line = ansiRegex.ReplaceAllString(line, "")
			trimmed := strings.TrimSpace(line)
			// Skip empty lines and test framework markers
			if trimmed == "" ||
				strings.HasPrefix(trimmed, "=== RUN") ||
				strings.HasPrefix(trimmed, "--- PASS") ||
				strings.HasPrefix(trimmed, "--- FAIL") {
				continue
			}
			// Replace problematic Unicode characters for Windows terminal compatibility
			line = strings.ReplaceAll(line, "❌", "[X]")
			line = strings.ReplaceAll(line, "✓", "[OK]")
			line = strings.ReplaceAll(line, "✗", "[X]")
			errorLines = append(errorLines, strings.TrimRight(line, "\n\r"))
		}

		errorOutput := "(no error details)"
		if len(errorLines) > 0 {
			errorOutput = strings.Join(errorLines, "\n")
		}

		failures = append(failures, Failure{
			TestName:    test,
			Package:     pkg,
			File:        pkg,
			Source:      "go-test",
			ErrorOutput: errorOutput,
		})
	}

	return failures, nil
}

// collectFailuresFromCucumberJSON parses a Cucumber JSON file
func collectFailuresFromCucumberJSON(jsonPath string) ([]Failure, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var report cucumber.CucumberReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	var failures []Failure

	for _, feature := range report {
		for _, scenario := range feature.Elements {
			for _, step := range scenario.Steps {
				if step.Result.Status == "failed" {
					failure := Failure{
						TestName:    scenario.Name,
						Package:     filepath.Base(filepath.Dir(feature.URI)),
						File:        feature.URI,
						Line:        step.Line,
						ErrorOutput: step.Result.Error,
						Source:      "cucumber",
					}
					if failure.ErrorOutput == "" {
						failure.ErrorOutput = "(no error message)"
					}
					failures = append(failures, failure)
				}
			}
		}
	}

	return failures, nil
}

// printFailureTable prints test failures in a readable format
func printFailureTable(failures []Failure) {
	log.Info("")
	log.Info("=== Test Failures Found ===")
	log.Infof("Total failures: %d", len(failures))

	for i, f := range failures {
		log.Info("")
		log.Infof("--- Failure %d ---", i+1)
		log.Infof("Test:    %s", f.TestName)

		// Show short package name for readability
		pkg := f.Package
		if idx := strings.LastIndex(pkg, "/"); idx != -1 {
			pkg = ".../" + pkg[idx+1:]
		}
		log.Infof("Package: %s", pkg)

		if f.Source == "cucumber" && f.File != "" {
			log.Infof("File:    %s:%d", f.File, f.Line)
		}

		log.Info("Output:")
		// Print each line of error output with indentation
		for _, line := range strings.Split(f.ErrorOutput, "\n") {
			if strings.TrimSpace(line) != "" {
				log.Infof("  %s", line)
			}
		}
	}

	log.Info("")
}
