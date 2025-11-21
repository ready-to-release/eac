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
// Long: Example:
// Long:   test debug
// HasSideEffects: false
package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/test/internal/cucumber"
	"github.com/ready-to-release/eac/src/commands/impl/test/internal/testjson"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

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
		fmt.Fprintf(os.Stderr, "Error: failed to get workspace root: %v\n", err)
		return 1
	}

	testOutputDir := filepath.Join(workspaceRoot, "out", "test")

	// Find all test result files
	goTestFiles, err := findGoTestJSONFiles(testOutputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding Go test JSON files: %v\n", err)
		return 1
	}

	cucumberFiles, err := findCucumberJSONFiles(testOutputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding Cucumber JSON files: %v\n", err)
		return 1
	}

	if len(goTestFiles) == 0 && len(cucumberFiles) == 0 {
		if _, err := os.Stat(testOutputDir); os.IsNotExist(err) {
			fmt.Println("No test output directory found. Run tests first to generate output.")
		} else {
			fmt.Println("No test results found. Run tests to generate output.")
		}
		return 0
	}

	// Collect failures from all sources
	var allFailures []Failure

	// Parse Go test JSON files
	for _, jsonFile := range goTestFiles {
		failures, err := collectFailuresFromGoTestJSON(jsonFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", jsonFile, err)
			continue
		}
		allFailures = append(allFailures, failures...)
	}

	// Parse Cucumber JSON files
	for _, jsonFile := range cucumberFiles {
		failures, err := collectFailuresFromCucumberJSON(jsonFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", jsonFile, err)
			continue
		}
		allFailures = append(allFailures, failures...)
	}

	if len(allFailures) == 0 {
		fmt.Println("No test failures found. All tests passed!")
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

		// Collect error output (look for error indicators)
		var errorLines []string
		for _, line := range outputs {
			// Look for error indicators
			if strings.Contains(line, "Error:") || strings.Contains(line, "error:") ||
				strings.Contains(line, "FAIL") || strings.Contains(line, "❌") ||
				strings.Contains(line, "expected") || strings.Contains(line, "actual") {
				errorLines = append(errorLines, line)
			}
		}

		errorOutput := "(no error details)"
		if len(errorLines) > 0 {
			errorOutput = strings.Join(errorLines, "; ")
			if len(errorOutput) > 200 {
				errorOutput = errorOutput[:197] + "..."
			}
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

// printFailureTable prints test failures in a formatted table
func printFailureTable(failures []Failure) {
	fmt.Println("\n=== Test Failures Found ===")
	fmt.Printf("Total failures: %d\n\n", len(failures))

	// Find max widths for table formatting
	maxTest := len("Test")
	maxPackage := len("Package")
	maxError := len("Error")

	for _, f := range failures {
		if len(f.TestName) > maxTest && len(f.TestName) <= 50 {
			maxTest = len(f.TestName)
		}
		if len(f.Package) > maxPackage && len(f.Package) <= 50 {
			maxPackage = len(f.Package)
		}
		if len(f.ErrorOutput) > maxError && len(f.ErrorOutput) <= 100 {
			maxError = len(f.ErrorOutput)
		}
	}

	// Cap widths
	if maxTest > 50 {
		maxTest = 50
	}
	if maxPackage > 50 {
		maxPackage = 50
	}
	if maxError > 100 {
		maxError = 100
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s\n",
		maxTest, "Test",
		maxPackage, "Package",
		maxError, "Error")
	fmt.Println(strings.Repeat("-", maxTest+2+maxPackage+2+maxError))

	// Print failures
	for _, f := range failures {
		test := f.TestName
		if len(test) > maxTest {
			test = test[:maxTest-3] + "..."
		}

		pkg := f.Package
		if len(pkg) > maxPackage {
			pkg = pkg[:maxPackage-3] + "..."
		}

		errorMsg := f.ErrorOutput
		if len(errorMsg) > maxError {
			errorMsg = errorMsg[:maxError-3] + "..."
		}

		fmt.Printf("%-*s  %-*s  %-*s\n",
			maxTest, test,
			maxPackage, pkg,
			maxError, errorMsg)
	}

	fmt.Println()
}
