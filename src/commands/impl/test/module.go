// Command: test module
// Short: Test a module by its moniker using type-based dispatch
// Long: Test a module by its moniker using type-based dispatch.
// Long:
// Long: This command identifies the module type from the module contract and dispatches
// Long: to the appropriate test handler. Supported module types include Go modules with
// Long: unit tests, specifications, and other testable components.
// Long:
// Long: Test output formats can be controlled with flags. The default output shows test
// Long: results in real-time. Use --as-cucumber for cucumber-style JSON output or --as-junit
// Long: for JUnit XML format suitable for CI/CD systems.
// Long:
// Long: Use --suite to filter tests by suite. The default suite is "commit".
// Long:
// Long: Example:
// Long:   test module src-commands
// Long:   test module src-core --as-junit
// Long:   test module src-commands --suite integration
// Flag.as-cucumber: type=bool, usage=Output test results in Cucumber JSON format
// Flag.as-junit: type=bool, usage=Output test results in JUnit XML format
// Flag.suite: type=string, usage=Filter tests by suite (default: "commit")
// HasSideEffects: false
package test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/build"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/commands/impl/test/internal/cucumber"
)

func init() {
	registry.Register(TestModule)
}

// TestFunc is the signature for module type test functions
// Parameters: module contract, workspace root, output directory, log writer, report format, suite name
// Returns: exit code
type TestFunc func(*modules.ModuleContract, string, string, io.Writer, string, string) int

// testFunctions maps module types to their test functions
var testFunctions = map[string]TestFunc{
	"go-cli":      testGoCLI,
	"go-commands": testGoCommands,
	"go-mcp":      testGoMCP,
	"go-library":  testGoLibrary,
	"go-tests":    testGoTests,
}

// TestModule tests a module by its moniker
// This is a convenience wrapper that redirects to TestSuite with module filter
func TestModule() int {
	// Parse arguments - expect: test module <moniker> [flags]
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: missing module moniker\n")
		fmt.Fprintf(os.Stderr, "Usage: test module <moniker> [--as-cucumber|--as-junit] [--suite <suite-name>]\n")
		return 1
	}

	moniker := os.Args[3]

	// Parse flags to extract suite name (default: commit)
	suiteName := "commit"
	otherFlags := []string{}

	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--suite" {
			if i+1 >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "Error: --suite requires a suite name\n")
				return 1
			}
			i++
			suiteName = os.Args[i]
		} else {
			otherFlags = append(otherFlags, arg)
		}
	}

	// Redirect to test suite by reconstructing args
	// Change: [binary, "test", "module", moniker, --suite <suite>, ...other flags]
	// To:     [binary, "test", "suite", <suite>, --module moniker, ...other flags]

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }() // Restore original args when done

	// Build new args: binary, "test", "suite", suiteName, --module, moniker, ...otherFlags
	newArgs := []string{
		os.Args[0],  // binary
		"test",
		"suite",
		suiteName,
		"--module",
		moniker,
	}
	newArgs = append(newArgs, otherFlags...)

	os.Args = newArgs

	// Call TestSuite which implements the full 5-phase process
	return TestSuite()
}

// testGoCLI tests a Cobra CLI binary (Pattern A)
// Runs: go test ./...
func testGoCLI(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-cli: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Suite: %s\n", suiteName)
	fmt.Fprintf(logWriter, "Running: go generate ./...\n")

	// Step 1: go generate (required for embedded files from contracts)
	if exitCode := build.RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	fmt.Fprintf(logWriter, "Running: go test ./...\n")

	exitCode, output := runTestCommandWithCapture(moduleRoot, logWriter, "go", "test", "./...")

	// Generate summary_unit.md
	fmt.Fprintf(logWriter, "\n=== Generating summary_unit.md ===\n")
	generateUnitTestSummaryMarkdown(module.Moniker, module.Type, outputDir, logWriter, output, exitCode)

	return exitCode
}

// testGoCommands tests the runtime command dispatcher (Pattern B)
// Runs: go test ./...
func testGoCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-commands: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Suite: %s\n", suiteName)
	fmt.Fprintf(logWriter, "Running: go test ./...\n")

	exitCode, output := runTestCommandWithCapture(moduleRoot, logWriter, "go", "test", "./...")

	// Generate summary_unit.md
	fmt.Fprintf(logWriter, "\n=== Generating summary_unit.md ===\n")
	generateUnitTestSummaryMarkdown(module.Moniker, module.Type, outputDir, logWriter, output, exitCode)

	return exitCode
}

// testGoMCP tests an MCP JSON-RPC server (Pattern C)
// Runs: go test ./...
func testGoMCP(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-mcp: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Suite: %s\n", suiteName)
	fmt.Fprintf(logWriter, "Running: go test ./...\n")

	exitCode, output := runTestCommandWithCapture(moduleRoot, logWriter, "go", "test", "./...")

	// Generate summary_unit.md
	fmt.Fprintf(logWriter, "\n=== Generating summary_unit.md ===\n")
	generateUnitTestSummaryMarkdown(module.Moniker, module.Type, outputDir, logWriter, output, exitCode)

	return exitCode
}

// testGoLibrary tests a Go library module (Pattern D)
// Runs: go test ./...
func testGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-library: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Suite: %s\n", suiteName)
	fmt.Fprintf(logWriter, "Running: go test ./...\n")

	exitCode, output := runTestCommandWithCapture(moduleRoot, logWriter, "go", "test", "./...")

	// Generate summary_unit.md
	fmt.Fprintf(logWriter, "\n=== Generating summary_unit.md ===\n")
	generateUnitTestSummaryMarkdown(module.Moniker, module.Type, outputDir, logWriter, output, exitCode)

	return exitCode
}

// testGoTests tests a Godog test module (Pattern D variant)
// Runs: go test with Godog formatters for reports
func testGoTests(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-tests: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Suite: %s\n", suiteName)
	fmt.Fprintf(logWriter, "Running: go test (Godog tests)\n")

	// Generate report file path based on format
	var reportPath string

	if reportFormat == "junit" {
		reportPath = filepath.Join(outputDir, "junit.xml")
		fmt.Fprintf(logWriter, "Report: JUnit XML - %s\n", reportPath)
	} else {
		// Default: cucumber
		reportPath = filepath.Join(outputDir, "cucumber.json")
		fmt.Fprintf(logWriter, "Report: Cucumber JSON - %s\n", reportPath)
	}

	env := map[string]string{
		"GODOG_OUTPUT_DIR":    outputDir,
		"GODOG_REPORT_FORMAT": reportFormat,
	}

	// Run go test - Godog will read format from test code via environment
	exitCode := runTestCommandWithEnv(moduleRoot, logWriter, env, "go", "test", "-v")

	// Generate summary_acceptance.md if cucumber.json was created
	if reportFormat == "cucumber" && exitCode == 0 {
		fmt.Fprintf(logWriter, "\n=== Generating summary_acceptance.md ===\n")
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
		} else {
			fmt.Fprintf(stderrWriter, "\nError: failed to execute test command: %v\n", err)
			exitCode = 1
		}
	} else {
		fmt.Fprintf(logWriter, "\n✅ Tests passed\n")
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
		fmt.Fprintf(stderrWriter, "\nError: failed to execute test command: %v\n", err)
		return 1
	}

	fmt.Fprintf(logWriter, "\n✅ Tests passed\n")
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
		fmt.Fprintf(logWriter, "Warning: failed to parse cucumber.json: %v\n", err)
		return
	}

	fmt.Fprintf(logWriter, "Found %d features\n", len(report))

	// Generate summary markdown without Appendix A (fragment starting at level 2)
	var summary string
	summary += "## Acceptance Test Summary\n\n"
	summary += cucumber.RenderAllFeatures(report, nil)

	// Write summary_acceptance.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		fmt.Fprintf(logWriter, "Warning: failed to write summary_acceptance.md: %v\n", err)
		return
	}

	fmt.Fprintf(logWriter, "✅ Generated: %s\n", summaryPath)

	// Generate Appendix A as separate file (fragment starting at level 2)
	var appendix string
	appendix += "## Appendix A: Specifications and Test Results\n\n"
	appendix += cucumber.RenderAppendixA(report, workspaceRoot)

	// Write appendix_a.md
	if err := os.WriteFile(appendixPath, []byte(appendix), 0644); err != nil {
		fmt.Fprintf(logWriter, "Warning: failed to write appendix_a.md: %v\n", err)
		return
	}

	fmt.Fprintf(logWriter, "✅ Generated: %s\n", appendixPath)
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
		fmt.Fprintf(logWriter, "Warning: failed to write summary_unit.md: %v\n", err)
		return
	}

	fmt.Fprintf(logWriter, "✅ Generated: %s\n", summaryPath)
}
