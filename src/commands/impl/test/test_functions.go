// test_functions.go - Type-based test dispatch functions
// These functions are used by the unified test command in test.go
package test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/impl/test/internal/cucumber"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// TestFunc is the signature for module type test functions
// Parameters: module contract, workspace root, output directory, log writer, report format, suite name
// Returns: exit code
type TestFunc func(*modules.ModuleContract, string, string, io.Writer, string, string) int

// testFunctions maps module types to their test functions
var testFunctions = map[string]TestFunc{
	// Go modules - run go test
	"go-cli":      testGoCLI,
	"go-commands": testGoCommands,
	"go-mcp":      testGoMCP,
	"go-library":  testGoLibrary,
	"go-tests":    testGoTests,

	// Documentation - verify build output
	"mkdocs-site":    testMkDocsSite,
	"mkdocs-subsite": testStaticModule, // Subsites are validated via parent site build

	// Repository-level tests
	"repository-root": testRepositoryRoot,

	// Script modules - run syntax validation
	"scripts":      testScripts,
	"scripts-sh":   testScriptsSh,
	"scripts-pwsh": testScriptsPwsh,

	// Configuration/static modules - no tests needed, verify build output exists
	"claude-agents":    testStaticModule,
	"claude-commands":  testStaticModule,
	"claude-config":    testStaticModule,
	"claude-hooks":     testStaticModule,
	"config":           testStaticModule,
	"configuration":    testStaticModule,
	"containers":       testStaticModule,
	"contracts":        testStaticModule,
	"definitions-type": testStaticModule,
	"markdown":         testStaticModule,
	"no-module-type":   testStaticModule,
	"go-r2r-extension": testStaticModule,
	"specifications":   testStaticModule,
	"templates":        testStaticModule,
	"vscode-config":    testStaticModule,
	"vscode-ext":       testStaticModule,
}

// testGoCLI tests a Cobra CLI binary (Pattern A)
// Delegates to test suite with module filter for proper tag-based filtering
func testGoCLI(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	writeln(logWriter, "\n=== Testing go-cli: %s ===", module.Moniker)
	writeln(logWriter, "Suite: %s", suiteName)
	return runTestSuiteForModule(module.Moniker, suiteName)
}

// testGoCommands tests the runtime command dispatcher (Pattern B)
// Delegates to test suite with module filter for proper tag-based filtering
func testGoCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	writeln(logWriter, "\n=== Testing go-commands: %s ===", module.Moniker)
	writeln(logWriter, "Suite: %s", suiteName)
	return runTestSuiteForModule(module.Moniker, suiteName)
}

// testGoMCP tests an MCP JSON-RPC server (Pattern C)
// Delegates to test suite with module filter for proper tag-based filtering
func testGoMCP(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	writeln(logWriter, "\n=== Testing go-mcp: %s ===", module.Moniker)
	writeln(logWriter, "Suite: %s", suiteName)
	return runTestSuiteForModule(module.Moniker, suiteName)
}

// testGoLibrary tests a Go library module (Pattern D)
// Delegates to test suite with module filter for proper tag-based filtering
func testGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	writeln(logWriter, "\n=== Testing go-library: %s ===", module.Moniker)
	writeln(logWriter, "Suite: %s", suiteName)
	return runTestSuiteForModule(module.Moniker, suiteName)
}

// testGoTests tests a Godog test module (Pattern D variant)
// Runs: go test with Godog formatters for reports
func testGoTests(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	writeln(logWriter, "\n=== Testing go-tests: %s ===", module.Moniker)
	writeln(logWriter, "Suite: %s", suiteName)
	writeln(logWriter, "Running: go test (Godog tests)")

	// Generate report file path based on format
	var reportPath string

	if reportFormat == "junit" {
		reportPath = filepath.Join(outputDir, "junit.xml")
		writeln(logWriter, "Report: JUnit XML - %s", reportPath)
	} else {
		// Default: cucumber
		reportPath = filepath.Join(outputDir, "cucumber.json")
		writeln(logWriter, "Report: Cucumber JSON - %s", reportPath)
	}

	env := map[string]string{
		"GODOG_OUTPUT_DIR":    outputDir,
		"GODOG_REPORT_FORMAT": reportFormat,
	}

	// Run go test - Godog will read format from test code via environment
	exitCode := runTestCommandWithEnv(moduleRoot, logWriter, env, "go", "test", "-v")

	// Generate summary_acceptance.md if cucumber.json was created
	if reportFormat == "cucumber" && exitCode == 0 {
		writeln(logWriter, "\n=== Generating summary_acceptance.md ===")
		generateGherkinSummaryMarkdown(module.Moniker, workspaceRoot, outputDir, logWriter)
	}

	return exitCode
}

// runTestCommand executes a test command in the specified directory
// Output is written to both console and log file via the provided writer
// Returns exit code (0 = success, non-zero = failure)
func runTestCommand(dir string, logWriter io.Writer, name string, args ...string) int {
	return runTestCommandWithEnv(dir, logWriter, nil, name, args...)
}

// runTestCommandWithCapture executes a test command and captures output
// Output is written to both console and log file, and also captured for summary generation
// Returns exit code and captured output
func runTestCommandWithCapture(dir string, logWriter io.Writer, name string, args ...string) (int, string) {
	var outputBuffer strings.Builder

	// Create multi-writer to capture output while also writing to log
	captureWriter := io.MultiWriter(logWriter, &outputBuffer)

	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	// Create multi-writer for stderr to capture errors in log
	stderrWriter := io.MultiWriter(os.Stderr, captureWriter)

	cmd.Stdout = captureWriter
	cmd.Stderr = stderrWriter

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			writeln(logWriter, "\n❌ Tests exited with code %d (error: %v)", exitCode, err)
		} else {
			writeln(stderrWriter, "\nError: failed to execute test command: %v", err)
			exitCode = 1
		}
	} else {
		writeln(logWriter, "\n✅ Tests passed")
	}

	return exitCode, outputBuffer.String()
}

// runTestCommandWithEnv executes a test command with custom environment variables
// Output is written to both console and log file via the provided writer
// Returns exit code (0 = success, non-zero = failure)
func runTestCommandWithEnv(dir string, logWriter io.Writer, env map[string]string, name string, args ...string) int {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	// Set custom environment variables
	if env != nil {
		cmd.Env = os.Environ()
		for key, value := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Create multi-writer for stderr to capture errors in log
	stderrWriter := io.MultiWriter(os.Stderr, logWriter)

	cmd.Stdout = logWriter
	cmd.Stderr = stderrWriter

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		writeln(stderrWriter, "\nError: failed to execute test command: %v", err)
		return 1
	}

	writeln(logWriter, "\n✅ Tests passed")
	return 0
}

// generateGherkinSummaryMarkdown generates summary_acceptance.md from cucumber.json
func generateGherkinSummaryMarkdown(moniker string, workspaceRoot string, outputDir string, logWriter io.Writer) {
	cucumberPath := filepath.Join(outputDir, "cucumber.json")
	summaryPath := filepath.Join(outputDir, "summary_acceptance.md")
	appendixPath := filepath.Join(outputDir, "appendix_a.md")

	// Parse cucumber.json
	report, err := cucumber.ParseFile(cucumberPath)
	if err != nil {
		writeln(logWriter, "Warning: failed to parse cucumber.json: %v", err)
		return
	}

	writeln(logWriter, "Found %d features", len(report))

	// Generate summary markdown without Appendix A (fragment starting at level 2)
	var summary string
	summary += "## Acceptance Test Summary\n\n"
	summary += cucumber.RenderAllFeatures(report, nil)

	// Write summary_acceptance.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		writeln(logWriter, "Warning: failed to write summary_acceptance.md: %v", err)
		return
	}

	writeln(logWriter, "✅ Generated: %s", summaryPath)

	// Generate Appendix A as separate file (fragment starting at level 2)
	var appendix string
	appendix += "## Appendix A: Specifications and Test Results\n\n"
	appendix += cucumber.RenderAppendixA(report, workspaceRoot)

	// Write appendix_a.md
	if err := os.WriteFile(appendixPath, []byte(appendix), 0644); err != nil {
		writeln(logWriter, "Warning: failed to write appendix_a.md: %v", err)
		return
	}

	writeln(logWriter, "✅ Generated: %s", appendixPath)
}

// generateUnitTestSummaryMarkdown generates summary_unit.md from go test output
func generateUnitTestSummaryMarkdown(moniker string, moduleType string, outputDir string, logWriter io.Writer, testOutput string, exitCode int) {
	summaryPath := filepath.Join(outputDir, "summary_unit.md")

	var summary string
	summary += "## Unit Test Summary\n\n"
	summary += fmt.Sprintf("**Module**: %s\n", moniker)
	summary += fmt.Sprintf("**Type**: %s\n", moduleType)

	if exitCode == 0 {
		summary += fmt.Sprintf("**Status**: ✅ Passed\n\n")
	} else {
		summary += fmt.Sprintf("**Status**: ❌ Failed\n\n")
	}

	summary += "### Test Output\n\n"
	summary += "```\n"
	summary += testOutput
	summary += "\n```\n"

	// Write summary_unit.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		writeln(logWriter, "Warning: failed to write summary_unit.md: %v", err)
		return
	}

	writeln(logWriter, "✅ Generated: %s", summaryPath)
}

// testStaticModule is a passthrough test for static/configuration modules
// These modules don't have runtime tests - they are validated by the build process
func testStaticModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	writeln(logWriter, "\n=== Testing %s: %s ===", module.Type, module.Moniker)
	writeln(logWriter, "Suite: %s", suiteName)
	writeln(logWriter, "Module type '%s' has no runtime tests", module.Type)
	writeln(logWriter, "✅ Static module - validation done at build time")
	return 0
}

// testMkDocsSite tests the MkDocs documentation site by verifying the build output exists
// The build is done separately with strict mode - this test just verifies the output
func testMkDocsSite(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	writeln(logWriter, "\n=== Testing mkdocs-site: %s ===", module.Moniker)
	writeln(logWriter, "Suite: %s", suiteName)

	// Check that the build output exists
	buildOutputDir := filepath.Join(workspaceRoot, "out", "build", module.Moniker, "site")
	indexFile := filepath.Join(buildOutputDir, "index.html")

	writeln(logWriter, "Checking build output: %s", buildOutputDir)

	// Verify site directory exists
	if _, err := os.Stat(buildOutputDir); os.IsNotExist(err) {
		writeln(logWriter, "\n❌ Build output not found: %s", buildOutputDir)
		writeln(logWriter, "   Run 'build module %s' first", module.Moniker)
		return 1
	}

	// Verify index.html exists (indicates successful build)
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		writeln(logWriter, "\n❌ index.html not found in build output")
		writeln(logWriter, "   Build may have failed - run 'build module %s'", module.Moniker)
		return 1
	}

	writeln(logWriter, "✅ Build output verified: %s", indexFile)
	writeln(logWriter, "\n✅ MkDocs site validation passed")

	return 0
}

// testRepositoryRoot tests the repository-root module (runs repository validation tests)
func testRepositoryRoot(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	// The repository-root module contains repository-level validation tests
	// These are typically in src/core/repository/tests
	testDir := filepath.Join(workspaceRoot, "src", "core", "repository", "tests")

	writeln(logWriter, "\n=== Testing repository-root: %s ===", module.Moniker)
	writeln(logWriter, "Running repository validation tests from: %s", testDir)

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		writeln(logWriter, "✅ No repository validation tests found at %s (this is OK)", testDir)
		return 0
	}

	// Run go test with JSON output for test debug command compatibility
	exitCode, output := runTestCommandWithCapture(testDir, logWriter, "go", "test", "-json", "-v")

	// Save JSON output to file so test debug can find failures
	jsonFile := filepath.Join(outputDir, "test-results.json")
	if err := os.WriteFile(jsonFile, []byte(output), 0644); err != nil {
		writeln(logWriter, "Warning: failed to save JSON results: %v", err)
	} else {
		writeln(logWriter, "✅ Saved JSON results: %s", jsonFile)
	}

	if exitCode != 0 {
		writeln(logWriter, "\n❌ Repository validation tests failed with exit code %d", exitCode)
	} else {
		writeln(logWriter, "\n✅ Repository validation tests passed")
	}

	return exitCode
}

// testScripts tests script modules by running Godog tests if a tests directory exists
func testScripts(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	writeln(logWriter, "\n=== Testing scripts: %s ===", module.Moniker)
	writeln(logWriter, "Suite: %s", suiteName)

	// Check if module has test patterns
	if len(module.Files.Tests) == 0 {
		writeln(logWriter, "No tests defined for this module")
		writeln(logWriter, "✅ Scripts module - no tests to run")
		return 0
	}

	// Use the module root for tests (tests are patterns relative to root)
	testDir := filepath.Join(workspaceRoot, module.Files.Root)

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		writeln(logWriter, "Test directory not found: %s", testDir)
		writeln(logWriter, "✅ Scripts module - no tests to run")
		return 0
	}

	writeln(logWriter, "Running tests from: %s", testDir)

	// Run go test with JSON output
	exitCode, output := runTestCommandWithCapture(testDir, logWriter, "go", "test", "-json", "-v")

	// Save JSON output to file
	jsonFile := filepath.Join(outputDir, "test-results.json")
	if err := os.WriteFile(jsonFile, []byte(output), 0644); err != nil {
		writeln(logWriter, "Warning: failed to save JSON results: %v", err)
	} else {
		writeln(logWriter, "✅ Saved JSON results: %s", jsonFile)
	}

	if exitCode != 0 {
		writeln(logWriter, "\n❌ Script tests failed with exit code %d", exitCode)
	} else {
		writeln(logWriter, "\n✅ Script tests passed")
	}

	return exitCode
}

// testScriptsSh tests shell script modules - runs Godog tests if defined
func testScriptsSh(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	// Delegate to testScripts - same logic applies
	return testScripts(module, workspaceRoot, outputDir, logWriter, reportFormat, suiteName)
}

// testScriptsPwsh tests PowerShell script modules - runs Godog tests if defined
func testScriptsPwsh(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	// Delegate to testScripts - same logic applies
	return testScripts(module, workspaceRoot, outputDir, logWriter, reportFormat, suiteName)
}

// runModuleTest runs tests for a single module
func runModuleTest(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	// Get test function for module type
	testFunc, hasTester := testFunctions[module.Type]
	if !hasTester {
		writeln(logWriter, "Error: no test function for type: %s", module.Type)
		return 1
	}

	// Execute the test function
	return testFunc(module, workspaceRoot, outputDir, logWriter, reportFormat, suiteName)
}

// runTestSuiteForModule runs the test suite command with a module filter
// This ensures proper test discovery, inference, and suite-based filtering
func runTestSuiteForModule(moniker string, suiteName string) int {
	// Redirect to test suite with module filter
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		oldArgs[0],
		"test",
		"suite",
		suiteName,
		"--module",
		moniker,
	}

	return TestSuite()
}

// FindModulesWithResults finds all subdirectories containing cucumber.json
func FindModulesWithResults(testRunDir string) ([]string, error) {
	entries, err := os.ReadDir(testRunDir)
	if err != nil {
		return nil, err
	}

	var foundModules []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this directory has cucumber.json
		cucumberPath := filepath.Join(testRunDir, entry.Name(), "cucumber.json")
		if _, err := os.Stat(cucumberPath); err == nil {
			foundModules = append(foundModules, entry.Name())
		}
	}

	return foundModules, nil
}

// formatDuration formats a duration as "1m 23s" or "45s"
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
