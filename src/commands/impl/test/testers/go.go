// go.go - Test functions for Go module types
package testers

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// TestGoDefault is the default test handler for Go modules without specific handlers.
// Delegates to test suite with module filter.
func TestGoDefault(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing Go module: %s (type: %s) ===", module.Moniker, module.Type)
	Writeln(logWriter, "Suite: %s", suiteName)
	return RunTestSuiteForModule(module.Moniker, suiteName)
}

// TestGoCLI tests a Cobra CLI binary (Pattern A).
// Delegates to test suite with module filter for proper tag-based filtering.
func TestGoCLI(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing go-cli: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	return RunTestSuiteForModule(module.Moniker, suiteName)
}

// TestGoCommands tests the runtime command dispatcher (Pattern B).
// Delegates to test suite with module filter for proper tag-based filtering.
func TestGoCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing go-commands: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	return RunTestSuiteForModule(module.Moniker, suiteName)
}

// TestGoMCP tests an MCP JSON-RPC server (Pattern C).
// Delegates to test suite with module filter for proper tag-based filtering.
func TestGoMCP(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing go-mcp: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	return RunTestSuiteForModule(module.Moniker, suiteName)
}

// TestGoLibrary tests a Go library module (Pattern D).
// Delegates to test suite with module filter for proper tag-based filtering.
func TestGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing go-library: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	return RunTestSuiteForModule(module.Moniker, suiteName)
}

// TestGoTests tests a Godog test module (Pattern D variant).
// Runs: go test with Godog formatters for reports.
func TestGoTests(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Writeln(logWriter, "\n=== Testing go-tests: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	Writeln(logWriter, "Running: go test (Godog tests)")

	// Generate report file path based on format
	var reportPath string

	if reportFormat == "junit" {
		reportPath = filepath.Join(outputDir, "junit.xml")
		Writeln(logWriter, "Report: JUnit XML - %s", reportPath)
	} else {
		// Default: cucumber
		reportPath = filepath.Join(outputDir, "cucumber.json")
		Writeln(logWriter, "Report: Cucumber JSON - %s", reportPath)
	}

	env := map[string]string{
		"GODOG_OUTPUT_DIR":    outputDir,
		"GODOG_REPORT_FORMAT": reportFormat,
	}

	// Run go test - Godog will read format from test code via environment
	exitCode := RunTestCommandWithEnv(moduleRoot, logWriter, env, "go", "test", "-v")

	// Generate summary_acceptance.md if cucumber.json was created
	if reportFormat == "cucumber" && exitCode == 0 {
		Writeln(logWriter, "\n=== Generating summary_acceptance.md ===")
		GenerateGherkinSummaryMarkdown(module.Moniker, workspaceRoot, outputDir, logWriter)
	}

	return exitCode
}

// TestRepositoryRoot tests the repository-root module (runs repository validation tests).
func TestRepositoryRoot(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	// The repository-root module contains repository-level validation tests
	// Location is defined in the module contract's test_impl field or derived from type
	testImplPath := module.GetTestImplementationPath()
	if testImplPath == "" {
		Writeln(logWriter, "\n=== Testing repository-root: %s ===", module.Moniker)
		Writeln(logWriter, "⚠️ No test_impl path defined in module contract")
		return 0
	}
	testDir := filepath.Join(workspaceRoot, testImplPath)

	Writeln(logWriter, "\n=== Testing repository-root: %s ===", module.Moniker)
	Writeln(logWriter, "Running repository validation tests from: %s", testDir)

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		Writeln(logWriter, "✅ No repository validation tests found at %s (this is OK)", testDir)
		return 0
	}

	// Run go test with JSON output for test debug command compatibility
	exitCode, output := RunTestCommandWithCapture(testDir, logWriter, "go", "test", "-json", "-v")

	// Save JSON output to file so test debug can find failures
	jsonFile := filepath.Join(outputDir, "test-results.json")
	if err := os.WriteFile(jsonFile, []byte(output), 0644); err != nil {
		Writeln(logWriter, "Warning: failed to save JSON results: %v", err)
	} else {
		Writeln(logWriter, "✅ Saved JSON results: %s", jsonFile)
	}

	if exitCode != 0 {
		Writeln(logWriter, "\n❌ Repository validation tests failed with exit code %d", exitCode)
	} else {
		Writeln(logWriter, "\n✅ Repository validation tests passed")
	}

	return exitCode
}

// TestScriptsPackage tests script package modules by running Godog tests if a tests directory exists.
func TestScriptsPackage(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing scripts-package: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)

	// Check if module has test patterns
	if len(module.Files.Tests) == 0 {
		Writeln(logWriter, "No tests defined for this module")
		Writeln(logWriter, "✅ Scripts module - no tests to run")
		return 0
	}

	// Use the module root for tests (tests are patterns relative to root)
	testDir := filepath.Join(workspaceRoot, module.Files.Root)

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		Writeln(logWriter, "Test directory not found: %s", testDir)
		Writeln(logWriter, "✅ Scripts module - no tests to run")
		return 0
	}

	Writeln(logWriter, "Running tests from: %s", testDir)

	// Run go test with JSON output
	exitCode, output := RunTestCommandWithCapture(testDir, logWriter, "go", "test", "-json", "-v")

	// Save JSON output to file
	jsonFile := filepath.Join(outputDir, "test-results.json")
	if err := os.WriteFile(jsonFile, []byte(output), 0644); err != nil {
		Writeln(logWriter, "Warning: failed to save JSON results: %v", err)
	} else {
		Writeln(logWriter, "✅ Saved JSON results: %s", jsonFile)
	}

	if exitCode != 0 {
		Writeln(logWriter, "\n❌ Script tests failed with exit code %d", exitCode)
	} else {
		Writeln(logWriter, "\n✅ Script tests passed")
	}

	return exitCode
}

// RunTestSuiteForModule is a placeholder that will be set by the dispatch package
// to avoid circular imports. The actual implementation calls TestSuite().
var RunTestSuiteForModule func(moniker string, suiteName string) int
