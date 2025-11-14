// Command: build module
// Description: Build a module by its moniker using type-based dispatch
// Usage: build module <moniker>
// HasSideEffects: false
package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(BuildModule)
}

// BuildOptions contains flags for controlling the build process
type BuildOptions struct {
	WindowsOnly bool
	LinuxOnly   bool
	MacOSOnly   bool
}

// BuildFunc is the signature for module type build functions
// Parameters: module contract, workspace root, output directory, log writer, build options
// Returns: exit code
type BuildFunc func(*modules.ModuleContract, string, string, io.Writer, BuildOptions) int

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
		fmt.Fprintf(os.Stderr, "Usage: build module <moniker> [--windows-only|--linux-only|--macos-only]\n")
		return 1
	}

	moniker := os.Args[3]

	// Parse optional flags for architecture-specific builds
	windowsOnly := false
	linuxOnly := false
	macosOnly := false
	for i := 4; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--windows-only":
			windowsOnly = true
		case "--linux-only":
			linuxOnly = true
		case "--macos-only":
			macosOnly = true
		}
	}

	// Get repository root using repository package
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

	// Create build options
	buildOpts := BuildOptions{
		WindowsOnly: windowsOnly,
		LinuxOnly:   linuxOnly,
		MacOSOnly:   macosOnly,
	}

	// Execute the build function with output directory, log writer, and options
	return buildFunc(module, workspaceRoot, outputDir, multiWriter, buildOpts)
}

// buildGoCLI builds a Cobra CLI binary (Pattern A)
// Requires: go generate && go build
// By default, builds for Windows, Linux, and macOS/ARM
func buildGoCLI(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Building go-cli: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go generate ./...\n")

	// Step 1: go generate
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	// Define target platforms
	type Platform struct {
		GOOS   string
		GOARCH string
		Ext    string
		Name   string
	}

	platforms := []Platform{
		{GOOS: "windows", GOARCH: "amd64", Ext: ".exe", Name: "Windows x64"},
		{GOOS: "linux", GOARCH: "amd64", Ext: "", Name: "Linux x64"},
		{GOOS: "darwin", GOARCH: "arm64", Ext: "", Name: "macOS ARM64"},
	}

	// Filter platforms based on flags
	var targetPlatforms []Platform
	if opts.WindowsOnly {
		targetPlatforms = []Platform{platforms[0]}
	} else if opts.LinuxOnly {
		targetPlatforms = []Platform{platforms[1]}
	} else if opts.MacOSOnly {
		targetPlatforms = []Platform{platforms[2]}
	} else {
		// Default: build for all platforms
		targetPlatforms = platforms
	}

	// Build for each target platform
	for _, platform := range targetPlatforms {
		binaryName := "r2r-cli" + platform.Ext
		binaryPath := filepath.Join(outputDir, fmt.Sprintf("%s-%s", platform.GOOS, binaryName))

		fmt.Fprintf(logWriter, "\n--- Building for %s (%s/%s) ---\n", platform.Name, platform.GOOS, platform.GOARCH)
		fmt.Fprintf(logWriter, "Output: %s\n", binaryPath)

		// Set GOOS and GOARCH environment variables
		cmd := exec.Command("go", "build", "-o", binaryPath)
		cmd.Dir = moduleRoot
		cmd.Stdout = logWriter
		cmd.Stderr = logWriter
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOOS=%s", platform.GOOS),
			fmt.Sprintf("GOARCH=%s", platform.GOARCH),
		)

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(logWriter, "❌ Build failed for %s: %v\n", platform.Name, err)
			return 1
		}

		fmt.Fprintf(logWriter, "✅ Built successfully: %s\n", binaryPath)

		// Make binary executable on Unix platforms
		if platform.GOOS != "windows" {
			if exitCode := RunCommandWithLog(moduleRoot, logWriter, "chmod", "+x", binaryPath); exitCode != 0 {
				fmt.Fprintf(logWriter, "⚠️  Warning: could not set executable permissions\n")
			}
		}
	}

	fmt.Fprintf(logWriter, "\n✅ All builds completed successfully\n")
	return 0
}

// buildGoCommands builds the runtime command dispatcher (Pattern B)
// Note: This is a development tool that's always run with "go run ."
func buildGoCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	fmt.Fprintf(logWriter, "\n=== go-commands: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "ℹ️  This module uses 'go run .' and is never compiled to a binary\n")
	fmt.Fprintf(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)\n")
	return 0
}

// buildGoMCP builds an MCP JSON-RPC server binary (Pattern C)
// Requires: go build -o mcp-server-<name>
func buildGoMCP(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
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

	return RunCommandWithLog(moduleRoot, logWriter, "go", "build", "-o", binaryPath)
}

// buildGoLibrary builds a Go library module (Pattern D)
// Note: Libraries are imported as dependencies, no binary output
// Runs go generate to prepare any embedded resources or generated code
func buildGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== go-library: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "Running: go generate ./...\n")

	// Step 1: go generate to prepare embedded resources
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	fmt.Fprintf(logWriter, "ℹ️  This is a library module (no binary to build)\n")
	fmt.Fprintf(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)\n")
	return 0
}

// buildGoTests builds a Godog BDD test module (Pattern D variant)
// Note: Tests are run with "go test", not built separately
func buildGoTests(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	fmt.Fprintf(logWriter, "\n=== go-tests: %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "ℹ️  This is a test module (use 'test module' command to run tests)\n")
	fmt.Fprintf(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)\n")
	return 0
}

// runCommandWithLog executes a command in the specified directory
// Output is written to both console and log file via the provided writer
// Returns exit code (0 = success, non-zero = failure)
func RunCommandWithLog(dir string, logWriter io.Writer, name string, args ...string) int {
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
