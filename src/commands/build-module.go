// Command: build module
// Description: Build a module by its moniker using type-based dispatch
// Usage: build module <moniker>
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ready-to-release/eac/src/internal/contracts/modules"
	"github.com/ready-to-release/eac/src/internal/contracts/reports"
)

func init() {
	Register("build module", BuildModule)
}

// BuildFunc is the signature for module type build functions
// Parameters: module contract, workspace root, output directory, log writer
// Returns: exit code
type BuildFunc func(*modules.ModuleContract, string, string, io.Writer) int

// buildFunctions maps module types to their build functions
var buildFunctions = map[string]BuildFunc{
	"go-cli":      buildGoCLI,
	"go-commands": buildGoCommands,
	"go-mcp":      buildGoMCP,
	"go-library":  buildGoLibrary,
	"go-tests":    buildGoTests,
}

// BuildModule builds a module by its moniker
func BuildModule() int {
	// Parse arguments
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: missing module moniker\n")
		fmt.Fprintf(os.Stderr, "Usage: build module <moniker>\n")
		return 1
	}

	moniker := os.Args[3]

	// Get repository root (two levels up from src/commands)
	workspaceRoot, err := filepath.Abs("../..")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to determine workspace root: %v\n", err)
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

	// Get build function for module type
	buildFunc, hasBuilder := buildFunctions[module.Type]
	if !hasBuilder {
		fmt.Fprintf(os.Stderr, "Error: no build function for type: %s\n", module.Type)
		fmt.Fprintf(os.Stderr, "Module: %s\n", moniker)
		fmt.Fprintf(os.Stderr, "Type: %s\n", module.Type)
		fmt.Fprintf(os.Stderr, "\nAvailable build functions:\n")
		if len(buildFunctions) == 0 {
			fmt.Fprintf(os.Stderr, "  (none - infrastructure only)\n")
		} else {
			for moduleType := range buildFunctions {
				fmt.Fprintf(os.Stderr, "  - %s\n", moduleType)
			}
		}
		return 1
	}

	// Purge existing output directory for this module
	outputDir := filepath.Join(workspaceRoot, "out", moniker)
	if err := os.RemoveAll(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to purge output directory: %v\n", err)
		return 1
	}

	// Create fresh output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create output directory: %v\n", err)
		return 1
	}

	// Create build log file
	logPath := filepath.Join(outputDir, fmt.Sprintf("build-%s.log", time.Now().Format("2006-01-02-150405")))
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create log file: %v\n", err)
		return 1
	}
	defer logFile.Close()

	// Create multi-writer to log to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Print header to both console and log
	fmt.Fprintf(multiWriter, "Building module: %s (type: %s)\n", moniker, module.Type)
	fmt.Fprintf(multiWriter, "Module root: %s\n", module.Source.Root)
	fmt.Fprintf(multiWriter, "Output directory: %s\n", outputDir)
	fmt.Fprintf(multiWriter, "Build log: %s\n", logPath)

	// Execute the build function with output directory and log writer
	return buildFunc(module, workspaceRoot, outputDir, multiWriter)
}

// buildGoCLI builds a Cobra CLI binary (Pattern A)
// Requires: go generate && go build
func buildGoCLI(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)
	binaryName := "r2r-cli"
	if runtime.GOOS == "windows" {
		binaryName = "r2r-cli.exe"
	}
	binaryPath := filepath.Join(outputDir, binaryName)

	fmt.Fprintf(logWriter, "\n=== Building go-cli: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go generate ./...\n")

	// Step 1: go generate
	if exitCode := runCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	fmt.Fprintf(logWriter, "Running: go build -o %s\n", binaryPath)

	// Step 2: go build with output to out/<moniker>/
	return runCommandWithLog(moduleRoot, logWriter, "go", "build", "-o", binaryPath)
}

// buildGoCommands builds the runtime command dispatcher (Pattern B)
// Note: This is a development tool that's always run with "go run ."
func buildGoCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	fmt.Fprintf(logWriter, "\n=== go-commands: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "ℹ️  This module uses 'go run .' and is never compiled to a binary\n")
	fmt.Fprintf(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)\n")
	return 0
}

// buildGoMCP builds an MCP JSON-RPC server binary (Pattern C)
// Requires: go build -o mcp-server-<name>
func buildGoMCP(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	// Extract server name from moniker (e.g., "src-mcp-docs" -> "docs")
	serverName := module.Moniker
	if len(serverName) > 8 && serverName[:8] == "src-mcp-" {
		serverName = serverName[8:]
	}

	binaryName := fmt.Sprintf("mcp-server-%s", serverName)
	binaryPath := filepath.Join(outputDir, binaryName)

	fmt.Fprintf(logWriter, "\n=== Building go-mcp: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go build -o %s\n", binaryPath)

	return runCommandWithLog(moduleRoot, logWriter, "go", "build", "-o", binaryPath)
}

// buildGoLibrary builds a Go library module (Pattern D)
// Note: Libraries are imported as dependencies, no binary output
func buildGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	fmt.Fprintf(logWriter, "\n=== go-library: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "ℹ️  This is a library module (no binary to build)\n")
	fmt.Fprintf(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)\n")
	return 0
}

// buildGoTests builds a Godog BDD test module (Pattern D variant)
// Note: Tests are run with "go test", not built separately
func buildGoTests(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	fmt.Fprintf(logWriter, "\n=== go-tests: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "ℹ️  This is a test module (use 'test module' command to run tests)\n")
	fmt.Fprintf(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)\n")
	return 0
}

// runCommandWithLog executes a command in the specified directory
// Output is written to both console and log file via the provided writer
// Returns exit code (0 = success, non-zero = failure)
func runCommandWithLog(dir string, logWriter io.Writer, name string, args ...string) int {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	// Create multi-writer for stderr to capture errors in log
	stderrWriter := io.MultiWriter(os.Stderr, logWriter)

	cmd.Stdout = logWriter
	cmd.Stderr = stderrWriter

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(stderrWriter, "\nError: failed to execute command: %v\n", err)
		return 1
	}

	return 0
}
