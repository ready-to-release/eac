// Command: test module
// Description: Test a module by its moniker using type-based dispatch
// Usage: test module <moniker> [--as-cucumber|--as-junit]
package test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/impl/build"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/commands/impl/test/internal/cucumber"
)

func init() {
	registry.Register(TestModule)
}

// TestFunc is the signature for module type test functions
// Parameters: module contract, workspace root, output directory, log writer, report format
// Returns: exit code
type TestFunc func(*modules.ModuleContract, string, string, io.Writer, string) int

// testFunctions maps module types to their test functions
var testFunctions = map[string]TestFunc{
	"go-cli":      testGoCLI,
	"go-commands": testGoCommands,
	"go-mcp":      testGoMCP,
	"go-library":  testGoLibrary,
	"go-tests":    testGoTests,
}

// TestModule tests a module by its moniker
func TestModule() int {
	// Parse arguments and flags
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: missing module moniker\n")
		fmt.Fprintf(os.Stderr, "Usage: test module <moniker> [--as-cucumber|--as-junit]\n")
		return 1
	}

	moniker := os.Args[3]

	// Parse flags (default: cucumber format, generate summary enabled)
	reportFormat := "cucumber"
	generateSummaryEnabled := true
	generateOnly := false

	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--as-cucumber" {
			reportFormat = "cucumber"
		} else if arg == "--as-junit" {
			reportFormat = "junit"
		} else if arg == "--no-generate" {
			generateSummaryEnabled = false
		} else if arg == "--generate-only" {
			generateOnly = true
		} else if strings.HasPrefix(arg, "--as-") {
			fmt.Fprintf(os.Stderr, "Error: unknown format flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid formats: --as-cucumber (default), --as-junit\n")
			return 1
		} else if strings.HasPrefix(arg, "--") {
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid flags: --as-cucumber, --as-junit, --no-generate, --generate-only\n")
			return 1
		}
	}

	// Handle --generate-only flag (skip tests, just generate summary)
	if generateOnly {
		fmt.Printf("📊 Generating summary for module: %s (skipping tests)\n", moniker)
		if err := generateSummary(moniker); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Load module contracts
	report, err := reports.GetModuleContracts(workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// Get the module from registry
	module, exists := report.Registry.Get(moniker)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moniker)
		return 1
	}

	// Get test function for module type
	testFunc, hasTester := testFunctions[module.Type]
	if !hasTester {
		fmt.Fprintf(os.Stderr, "Error: no test function for type: %s\n", module.Type)
		fmt.Fprintf(os.Stderr, "Module: %s\n", moniker)
		fmt.Fprintf(os.Stderr, "Type: %s\n", module.Type)
		fmt.Fprintf(os.Stderr, "\nAvailable test functions:\n")
		if len(testFunctions) == 0 {
			fmt.Fprintf(os.Stderr, "  (none)\n")
		} else {
			for moduleType := range testFunctions {
				fmt.Fprintf(os.Stderr, "  - %s\n", moduleType)
			}
		}
		return 1
	}

	// Create test-run-id directory (timestamp-based)
	testRunID := time.Now().Format("2006-01-02-150405")
	testRunDir := filepath.Join(workspaceRoot, "out", "test-results", testRunID)
	if err := os.MkdirAll(testRunDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create test run directory: %v\n", err)
		return 1
	}

	// Create module output directory within test run
	outputDir := filepath.Join(testRunDir, moniker)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create output directory: %v\n", err)
		return 1
	}

	// Create test log file
	logPath := filepath.Join(outputDir, "test.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create log file: %v\n", err)
		return 1
	}
	defer logFile.Close()

	// Create multi-writer to log to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Print header to both console and log
	fmt.Fprintf(multiWriter, "Test Run ID: %s\n", testRunID)
	fmt.Fprintf(multiWriter, "Testing module: %s (type: %s)\n", moniker, module.Type)
	fmt.Fprintf(multiWriter, "Module root: %s\n", module.Source.Root)
	fmt.Fprintf(multiWriter, "Output directory: %s\n", outputDir)
	fmt.Fprintf(multiWriter, "Test log: %s\n", logPath)
	fmt.Fprintf(multiWriter, "Report format: %s\n", reportFormat)

	// Execute the test function with output directory, log writer, and report format
	exitCode := testFunc(module, workspaceRoot, outputDir, multiWriter, reportFormat)

	// Print summary
	fmt.Println("\n===========================================")
	fmt.Printf("Test Run Summary (ID: %s)\n", testRunID)
	fmt.Println("===========================================")
	if exitCode == 0 {
		fmt.Printf("✅ Module %s passed\n", moniker)
	} else {
		fmt.Printf("❌ Module %s failed with exit code %d\n", moniker, exitCode)
	}
	fmt.Printf("Results directory: %s\n", outputDir)

	// Generate summary if enabled and using cucumber format
	if generateSummaryEnabled && reportFormat == "cucumber" && exitCode == 0 {
		fmt.Println("\n📊 Generating test summary...")
		if err := generateSummaryForOutputDir(outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to generate summary: %v\n", err)
			// Don't fail the test run, just warn
		}
	}

	return exitCode
}

// testGoCLI tests a Cobra CLI binary (Pattern A)
// Runs: go test ./...
func testGoCLI(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-cli: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go generate ./...\n")

	// Step 1: go generate (required for embedded files from contracts)
	if exitCode := build.RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	fmt.Fprintf(logWriter, "Running: go test ./...\n")

	exitCode, output := runTestCommandWithCapture(moduleRoot, logWriter, "go", "test", "./...")

	// Generate summary_unit.md
	fmt.Fprintf(logWriter, "\n=== Generating summary_unit.md ===\n")
	generateTDDSummaryMarkdown(module.Moniker, module.Type, outputDir, logWriter, output, exitCode)

	return exitCode
}

// testGoCommands tests the runtime command dispatcher (Pattern B)
// Runs: go test ./...
func testGoCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-commands: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go test ./...\n")

	exitCode, output := runTestCommandWithCapture(moduleRoot, logWriter, "go", "test", "./...")

	// Generate summary_unit.md
	fmt.Fprintf(logWriter, "\n=== Generating summary_unit.md ===\n")
	generateTDDSummaryMarkdown(module.Moniker, module.Type, outputDir, logWriter, output, exitCode)

	return exitCode
}

// testGoMCP tests an MCP JSON-RPC server (Pattern C)
// Runs: go test ./...
func testGoMCP(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-mcp: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go test ./...\n")

	exitCode, output := runTestCommandWithCapture(moduleRoot, logWriter, "go", "test", "./...")

	// Generate summary_unit.md
	fmt.Fprintf(logWriter, "\n=== Generating summary_unit.md ===\n")
	generateTDDSummaryMarkdown(module.Moniker, module.Type, outputDir, logWriter, output, exitCode)

	return exitCode
}

// testGoLibrary tests a Go library module (Pattern D)
// Runs: go test ./...
func testGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-library: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go test ./...\n")

	exitCode, output := runTestCommandWithCapture(moduleRoot, logWriter, "go", "test", "./...")

	// Generate summary_unit.md
	fmt.Fprintf(logWriter, "\n=== Generating summary_unit.md ===\n")
	generateTDDSummaryMarkdown(module.Moniker, module.Type, outputDir, logWriter, output, exitCode)

	return exitCode
}

// testGoTests tests a Godog BDD test module (Pattern D variant)
// Runs: go test with Godog formatters for reports
func testGoTests(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Testing go-tests: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go test (Godog BDD tests)\n")

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
		generateBDDSummaryMarkdown(module.Moniker, workspaceRoot, outputDir, logWriter)
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

// generateBDDSummaryMarkdown generates summary_acceptance.md from cucumber.json
func generateBDDSummaryMarkdown(moniker string, workspaceRoot string, outputDir string, logWriter io.Writer) {
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

// generateTDDSummaryMarkdown generates summary_unit.md from go test output
func generateTDDSummaryMarkdown(moniker string, moduleType string, outputDir string, logWriter io.Writer, testOutput string, exitCode int) {
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
