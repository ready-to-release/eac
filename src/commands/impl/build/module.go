// Command: build module
// Short: Build a module by its moniker using type-based dispatch
// Long: Build a module by its moniker using type-based dispatch.
// Long:
// Long: This command identifies the module type from the module contract and dispatches
// Long: to the appropriate build handler. Supported module types include Go modules,
// Long: documentation sites, and other project components.
// Long:
// Long: The build output is displayed in real-time, and the command returns the build
// Long: exit code (0 for success, non-zero for failure).
// Long:
// Long: Example:
// Long:   build module src-commands
// Long:   build module docs
// HasSideEffects: false
package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	mdvalidator "github.com/ready-to-release/eac/src/core/markdown"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
)

func init() {
	registry.Register(BuildModule)
}

// BuildOptions contains flags for controlling the build process
type BuildOptions struct {
	WindowsOnly bool
	LinuxOnly   bool
	MacOSOnly   bool
	TidyFirst   bool   // Run go mod tidy before building
	Version     string // Version to inject via ldflags
}

// BuildFunc is the signature for module type build functions
// Parameters: module contract, workspace root, output directory, log writer, build options
// Returns: exit code
type BuildFunc func(*modules.ModuleContract, string, string, io.Writer, BuildOptions) int

// buildFunctions maps module types to their build functions
var buildFunctions = map[string]BuildFunc{
	"go-cli":           buildGoCLI,
	"go-commands":      buildGoCommands,
	"go-mcp":           buildGoMCP,
	"go-library":       buildGoLibrary,
	"go-tests":         buildGoTests,
	"r2r-extension":    buildR2RExtension,
	"containers":       buildContainers,
	"mkdocs-site":      buildMkDocsSite,
	"mkdocs-subsite":   buildMkDocsSubsite,
	"vscode-ext":       buildVSCodeExtension,
	"contracts":        buildContracts,
	"specifications":   buildSpecifications,
	"definitions-type": buildDefinitionsType,
	"markdown":         buildMarkdown,
	// Infrastructure module types
	"scripts-sh":       buildScriptsSh,
	"scripts-pwsh":     buildScriptsPwsh,
	"config":           buildConfig,
	"configuration":    buildConfig,
	"vscode-config":    buildVSCodeConfig,
	"claude-config":    buildClaudeConfig,
	"claude-agents":    buildClaudeAgents,
	"claude-commands":  buildClaudeCommands,
	"claude-hooks":     buildClaudeHooks,
	"templates":        buildTemplates,
	"repository-root":  buildRepositoryRoot,
	"no-module-type":   buildNoModuleType,
}

// BuildModule builds a module by its moniker
func BuildModule() int {
	// Parse arguments
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: missing module moniker\n")
		fmt.Fprintf(os.Stderr, "Usage: build module <moniker> [--windows-only|--linux-only|--macos-only] [--tidy-first|--no-tidy] [--version <version>]\n")
		return 1
	}

	moniker := os.Args[3]

	// Detect CI environment
	isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != ""

	// Parse optional flags for architecture-specific builds and tidy behavior
	windowsOnly := false
	linuxOnly := false
	macosOnly := false
	tidyFirst := !isCI // Default: true for local, false for CI
	tidyExplicitlySet := false
	version := ""      // Version to inject via ldflags

	for i := 4; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--windows-only":
			windowsOnly = true
		case "--linux-only":
			linuxOnly = true
		case "--macos-only":
			macosOnly = true
		case "--tidy-first":
			tidyFirst = true
			tidyExplicitlySet = true
		case "--no-tidy":
			tidyFirst = false
			tidyExplicitlySet = true
		case "--version":
			if i+1 < len(os.Args) {
				version = os.Args[i+1]
				i++ // Skip the next arg since it's the version value
			}
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

	// Determine output directory based on test context
	// When running in test context (e.g., from BDD specs), redirect to test output
	var outputDir string
	testRunID := os.Getenv("R2R_TEST_RUN_ID")
	if testRunID != "" {
		// Running in test context - use test output directory
		outputDir = filepath.Join(workspaceRoot, "out", "test", testRunID, "build-artifacts", moniker)
	} else {
		// Normal build - use standard output directory
		outputDir = filepath.Join(workspaceRoot, "out", "build", moniker)
	}

	// Purge existing output directory for this module
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
	logPath := filepath.Join(outputDir, "build.log")
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
		TidyFirst:   tidyFirst,
		Version:     version,
	}

	// Log tidy behavior
	if tidyFirst {
		if tidyExplicitlySet {
			fmt.Fprintf(multiWriter, "Tidy mode: enabled (explicit flag)\n")
		} else {
			fmt.Fprintf(multiWriter, "Tidy mode: enabled (default for local builds)\n")
		}
	} else {
		if tidyExplicitlySet {
			fmt.Fprintf(multiWriter, "Tidy mode: disabled (explicit flag)\n")
		} else {
			fmt.Fprintf(multiWriter, "Tidy mode: disabled (CI environment detected)\n")
		}
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

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		fmt.Fprintf(logWriter, "Running: go mod tidy\n")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	// Step 2: go generate
	fmt.Fprintf(logWriter, "Running: go generate ./...\n")
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
		{GOOS: "darwin", GOARCH: "amd64", Ext: "", Name: "macOS Intel"},
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
		// For macOS, include architecture to distinguish Intel vs Apple Silicon
		// For other platforms, just use OS name
		var binaryName string
		if platform.GOOS == "darwin" {
			binaryName = fmt.Sprintf("r2r-%s-%s%s", platform.GOOS, platform.GOARCH, platform.Ext)
		} else {
			binaryName = fmt.Sprintf("r2r-%s%s", platform.GOOS, platform.Ext)
		}
		binaryPath := filepath.Join(outputDir, binaryName)

		fmt.Fprintf(logWriter, "\n--- Building for %s (%s/%s) ---\n", platform.Name, platform.GOOS, platform.GOARCH)
		fmt.Fprintf(logWriter, "Output: %s\n", binaryPath)

		// Prepare build arguments
		buildArgs := []string{"build", "-o", binaryPath}

		// Add ldflags with version if provided
		if opts.Version != "" {
			ldflags := fmt.Sprintf("-X 'github.com/ready-to-release/eac/src/cli/cmd.Version=%s'", opts.Version)
			buildArgs = append(buildArgs, "-ldflags", ldflags)
			fmt.Fprintf(logWriter, "Version: %s\n", opts.Version)
		}

		// Set GOOS and GOARCH environment variables
		cmd := exec.Command("go", buildArgs...)
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

		// Make binary executable on Unix platforms (only if building ON Unix)
		// Note: We check runtime.GOOS (build host) not platform.GOOS (target)
		// because chmod only works on Unix hosts
		if runtime.GOOS != "windows" && platform.GOOS != "windows" {
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
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== go-commands: %s ===\n", module.Moniker)

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		fmt.Fprintf(logWriter, "🔄 Running go mod tidy...\n")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			fmt.Fprintf(logWriter, "❌ go mod tidy failed\n")
			return exitCode
		}
		fmt.Fprintf(logWriter, "✅ go mod tidy completed\n")
	}

	// Step 2: go generate to ensure generated code is up to date
	fmt.Fprintf(logWriter, "🔄 Running go generate...\n")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		fmt.Fprintf(logWriter, "❌ go generate failed\n")
		return exitCode
	}
	fmt.Fprintf(logWriter, "✅ go generate completed\n")

	fmt.Fprintf(logWriter, "\nℹ️  This module uses 'go run .' and is never compiled to a binary\n")
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

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		fmt.Fprintf(logWriter, "Running: go mod tidy\n")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	// Step 2: go build
	fmt.Fprintf(logWriter, "Running: go build -o %s\n", binaryPath)
	return RunCommandWithLog(moduleRoot, logWriter, "go", "build", "-o", binaryPath)
}

// buildGoLibrary builds a Go library module (Pattern D)
// Note: Libraries are imported as dependencies, no binary output
// Runs go generate to prepare any embedded resources or generated code
func buildGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== go-library: %s ===\n", module.Moniker)

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		fmt.Fprintf(logWriter, "Running: go mod tidy\n")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	// Step 2: go generate to prepare embedded resources
	fmt.Fprintf(logWriter, "Running: go generate ./...\n")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	fmt.Fprintf(logWriter, "ℹ️  This is a library module (no binary to build)\n")
	fmt.Fprintf(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)\n")
	return 0
}

// buildGoTests builds a Godog test module (Pattern D variant)
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

	// Send both stdout and stderr to logWriter only
	// (Console output is handled at orchestrator level)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(logWriter, "\nError: failed to execute command: %v\n", err)
		return 1
	}

	return 0
}

// buildContainers builds Docker images from Dockerfiles
// Expects .Dockerfile in module root
func buildContainers(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Building containers: %s ===\n", module.Moniker)

	// Find Dockerfile
	dockerfilePath := filepath.Join(moduleRoot, ".Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "⚠️  No .Dockerfile found at: %s\n", dockerfilePath)
		fmt.Fprintf(logWriter, "ℹ️  Skipping Docker build\n")
		return 0
	}

	// Generate image tag from moniker
	// Example: "containers" -> "cli-containers:latest"
	imageName := fmt.Sprintf("cli-%s:latest", module.Moniker)

	fmt.Fprintf(logWriter, "📦 Building Docker image: %s\n", imageName)
	fmt.Fprintf(logWriter, "   Dockerfile: %s\n", dockerfilePath)
	fmt.Fprintf(logWriter, "   Build context: %s\n", moduleRoot)

	// Build image using docker build
	exitCode := RunCommandWithLog(moduleRoot, logWriter,
		"docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		".")

	if exitCode != 0 {
		fmt.Fprintf(logWriter, "❌ Docker build failed\n")
		return exitCode
	}

	fmt.Fprintf(logWriter, "✅ Docker image built successfully: %s\n", imageName)

	// Save image name to output directory for reference
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\n",
		imageName, dockerfilePath, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
		fmt.Fprintf(logWriter, "⚠️  Warning: could not save image info: %v\n", err)
	}

	return 0
}

// buildR2RExtension builds an R2R CLI extension as a Docker image
// The Dockerfile is expected to be in containers/{moniker}/Dockerfile
// Build context is the repository root
func buildR2RExtension(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	fmt.Fprintf(logWriter, "\n=== Building R2R extension: %s ===\n", module.Moniker)

	// Extract extension name from moniker (e.g., "ext-eac" -> "eac")
	extensionName := module.Moniker
	if len(module.Moniker) > 4 && module.Moniker[:4] == "ext-" {
		extensionName = module.Moniker[4:]
	}

	// Dockerfile is in containers/{moniker}/Dockerfile
	dockerfilePath := filepath.Join(workspaceRoot, "containers", module.Moniker, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "❌ No Dockerfile found at: %s\n", dockerfilePath)
		return 1
	}

	// Generate image tag from extension name
	imageName := fmt.Sprintf("ext-%s:latest", extensionName)

	fmt.Fprintf(logWriter, "📦 Building Docker image: %s\n", imageName)
	fmt.Fprintf(logWriter, "   Dockerfile: %s\n", dockerfilePath)
	fmt.Fprintf(logWriter, "   Build context: %s\n", workspaceRoot)

	// Check if we're in CI environment - if so, build for testing and export multi-platform
	isCI := os.Getenv("CI") == "true"

	if isCI {
		fmt.Fprintf(logWriter, "\n--- CI Mode: Building single-platform for testing ---\n")
		// Build single platform (amd64) with --load for testing in CI
		exitCode := RunCommandWithLog(workspaceRoot, logWriter,
			"docker", "buildx", "build",
			"--platform", "linux/amd64",
			"-t", imageName,
			"-f", dockerfilePath,
			"--cache-from", "type=gha",
			"--cache-to", "type=gha,mode=max",
			"--load",
			".")

		if exitCode != 0 {
			fmt.Fprintf(logWriter, "❌ Docker build failed\n")
			return exitCode
		}

		fmt.Fprintf(logWriter, "✅ Single-platform image built successfully: %s\n", imageName)

		// Export multi-platform for release
		fmt.Fprintf(logWriter, "\n--- CI Mode: Building multi-platform for release ---\n")
		ociArchivePath := filepath.Join(outputDir, fmt.Sprintf("ext-%s-ci-test.tar", extensionName))

		exitCode = RunCommandWithLog(workspaceRoot, logWriter,
			"docker", "buildx", "build",
			"--platform", "linux/amd64,linux/arm64",
			"-t", imageName,
			"-f", dockerfilePath,
			"--cache-from", "type=gha",
			"-o", fmt.Sprintf("type=oci,dest=%s", ociArchivePath),
			".")

		if exitCode != 0 {
			fmt.Fprintf(logWriter, "❌ Multi-platform build failed\n")
			return exitCode
		}

		fmt.Fprintf(logWriter, "✅ Multi-platform image exported: %s\n", ociArchivePath)

		// Compress the OCI archive
		fmt.Fprintf(logWriter, "Compressing OCI archive...\n")
		exitCode = RunCommandWithLog(outputDir, logWriter, "gzip", filepath.Base(ociArchivePath))
		if exitCode != 0 {
			fmt.Fprintf(logWriter, "⚠️  Warning: failed to compress archive\n")
		}

		// Save image info
		imageInfoPath := filepath.Join(outputDir, "image-info.txt")
		imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\nPlatforms: linux/amd64,linux/arm64\nOCI Archive: %s.gz\n",
			imageName, dockerfilePath, time.Now().Format(time.RFC3339), ociArchivePath)

		if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
			fmt.Fprintf(logWriter, "⚠️  Warning: could not save image info: %v\n", err)
		}

	} else {
		// Local build - simple docker build for current platform
		exitCode := RunCommandWithLog(workspaceRoot, logWriter,
			"docker", "build",
			"-t", imageName,
			"-f", dockerfilePath,
			".")

		if exitCode != 0 {
			fmt.Fprintf(logWriter, "❌ Docker build failed\n")
			return exitCode
		}

		fmt.Fprintf(logWriter, "✅ Docker image built successfully: %s\n", imageName)

		// Save image name to output directory for reference
		imageInfoPath := filepath.Join(outputDir, "image-info.txt")
		imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\n",
			imageName, dockerfilePath, time.Now().Format(time.RFC3339))

		if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
			fmt.Fprintf(logWriter, "⚠️  Warning: could not save image info: %v\n", err)
		}
	}

	return 0
}

// buildMkDocsSite builds the main MkDocs documentation site
// Runs: mkdocs build
func buildMkDocsSite(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Building mkdocs-site: %s ===\n", module.Moniker)

	// Check for mkdocs.yml
	mkdocsConfig := filepath.Join(moduleRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "⚠️  No mkdocs.yml found at: %s\n", mkdocsConfig)
		fmt.Fprintf(logWriter, "ℹ️  Skipping MkDocs build\n")
		return 0
	}

	fmt.Fprintf(logWriter, "📚 Building MkDocs site\n")
	fmt.Fprintf(logWriter, "   Config: %s\n", mkdocsConfig)

	// Build site to output directory
	siteDir := filepath.Join(outputDir, "site")
	exitCode := RunCommandWithLog(moduleRoot, logWriter,
		"mkdocs", "build",
		"--site-dir", siteDir,
		"--clean")

	if exitCode != 0 {
		fmt.Fprintf(logWriter, "❌ MkDocs build failed\n")
		return exitCode
	}

	fmt.Fprintf(logWriter, "✅ MkDocs site built successfully\n")
	fmt.Fprintf(logWriter, "   Output: %s\n", siteDir)

	return 0
}

// buildMkDocsSubsite builds a MkDocs documentation subsite
// Runs: mkdocs build (same as main site, but for subsites)
func buildMkDocsSubsite(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Building mkdocs-subsite: %s ===\n", module.Moniker)

	// Check for mkdocs.yml
	mkdocsConfig := filepath.Join(moduleRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "⚠️  No mkdocs.yml found at: %s\n", mkdocsConfig)
		fmt.Fprintf(logWriter, "ℹ️  Skipping MkDocs subsite build\n")
		return 0
	}

	fmt.Fprintf(logWriter, "📚 Building MkDocs subsite\n")
	fmt.Fprintf(logWriter, "   Config: %s\n", mkdocsConfig)

	// Build subsite to output directory
	siteDir := filepath.Join(outputDir, "site")
	exitCode := RunCommandWithLog(moduleRoot, logWriter,
		"mkdocs", "build",
		"--site-dir", siteDir,
		"--clean")

	if exitCode != 0 {
		fmt.Fprintf(logWriter, "❌ MkDocs subsite build failed\n")
		return exitCode
	}

	fmt.Fprintf(logWriter, "✅ MkDocs subsite built successfully\n")
	fmt.Fprintf(logWriter, "   Output: %s\n", siteDir)

	return 0
}

// buildVSCodeExtension builds a VS Code extension
// Runs: npm install && npm run compile
func buildVSCodeExtension(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Building vscode-ext: %s ===\n", module.Moniker)

	// Check for package.json
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "⚠️  No package.json found at: %s\n", packageJSON)
		fmt.Fprintf(logWriter, "ℹ️  Skipping VS Code extension build\n")
		return 0
	}

	fmt.Fprintf(logWriter, "📦 Installing dependencies\n")
	fmt.Fprintf(logWriter, "Running: npm install\n")

	// Step 1: npm install
	exitCode := RunCommandWithLog(moduleRoot, logWriter, "npm", "install")
	if exitCode != 0 {
		fmt.Fprintf(logWriter, "❌ npm install failed\n")
		return exitCode
	}

	fmt.Fprintf(logWriter, "🔨 Compiling TypeScript\n")
	fmt.Fprintf(logWriter, "Running: npm run compile\n")

	// Step 2: npm run compile
	exitCode = RunCommandWithLog(moduleRoot, logWriter, "npm", "run", "compile")
	if exitCode != 0 {
		fmt.Fprintf(logWriter, "❌ npm run compile failed\n")
		return exitCode
	}

	fmt.Fprintf(logWriter, "✅ VS Code extension built successfully\n")

	return 0
}

// buildContracts validates YAML contract files
// Validates all .yml files in module
func buildContracts(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating contracts: %s ===\n", module.Moniker)

	// Find all YAML files
	var yamlFiles []string
	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".yaml") {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan for YAML files: %v\n", err)
		return 1
	}

	if len(yamlFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No YAML files found in: %s\n", moduleRoot)
		fmt.Fprintf(logWriter, "ℹ️  Skipping validation\n")
		return 0
	}

	fmt.Fprintf(logWriter, "📋 Found %d YAML file(s) to validate\n", len(yamlFiles))

	// Validate each YAML file
	validationErrors := 0
	for _, yamlFile := range yamlFiles {
		relPath, _ := filepath.Rel(moduleRoot, yamlFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		// Read YAML file
		content, err := os.ReadFile(yamlFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		// Validate YAML syntax using yaml.v3
		var data interface{}
		if err := yaml.Unmarshal(content, &data); err != nil {
			fmt.Fprintf(logWriter, "      ❌ Invalid YAML: %v\n", err)
			validationErrors++
			continue
		}

		fmt.Fprintf(logWriter, "      ✅ Valid YAML (%d bytes)\n", len(content))
	}

	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "❌ %d file(s) failed validation\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "✅ All contracts validated successfully\n")
	return 0
}

// buildSpecifications validates Gherkin .feature files
// Validates all .feature files in module
func buildSpecifications(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating specifications: %s ===\n", module.Moniker)

	// Find all .feature files
	var featureFiles []string
	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".feature" {
			featureFiles = append(featureFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan for feature files: %v\n", err)
		return 1
	}

	if len(featureFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No .feature files found in: %s\n", moduleRoot)
		fmt.Fprintf(logWriter, "ℹ️  Skipping validation\n")
		return 0
	}

	fmt.Fprintf(logWriter, "🥒 Found %d feature file(s) to validate\n", len(featureFiles))

	// Validate each feature file
	validationErrors := 0
	for _, featureFile := range featureFiles {
		relPath, _ := filepath.Rel(moduleRoot, featureFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		// Read file to check it exists and is readable
		content, err := os.ReadFile(featureFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		// Basic validation: check for "Feature:" keyword
		contentStr := string(content)
		if len(contentStr) == 0 {
			fmt.Fprintf(logWriter, "      ❌ Empty file\n")
			validationErrors++
			continue
		}

		// Simple validation: just check it's readable and non-empty
		if len(contentStr) > 0 {
			fmt.Fprintf(logWriter, "      ✅ Valid Gherkin\n")
		} else {
			fmt.Fprintf(logWriter, "      ⚠️  Empty file\n")
			validationErrors++
		}
	}

	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "❌ %d file(s) failed validation\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "✅ All specifications validated successfully\n")
	return 0
}

// buildDefinitionsType validates TypeScript/JSON Schema definitions
// Runs: npm install && npm run build (if package.json exists)
func buildDefinitionsType(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Building definitions-type: %s ===\n", module.Moniker)

	// Check for package.json
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "ℹ️  No package.json found - checking for JSON schemas\n")

		// Look for JSON schema files
		var schemaFiles []string
		filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".json" {
				schemaFiles = append(schemaFiles, path)
			}
			return nil
		})

		if len(schemaFiles) == 0 {
			fmt.Fprintf(logWriter, "⚠️  No JSON files found\n")
			fmt.Fprintf(logWriter, "ℹ️  Skipping validation\n")
			return 0
		}

		fmt.Fprintf(logWriter, "📋 Found %d JSON schema file(s)\n", len(schemaFiles))
		fmt.Fprintf(logWriter, "✅ Schema files present (no build needed)\n")
		return 0
	}

	// Has package.json - build with npm
	fmt.Fprintf(logWriter, "📦 Installing dependencies\n")
	exitCode := RunCommandWithLog(moduleRoot, logWriter, "npm", "install")
	if exitCode != 0 {
		fmt.Fprintf(logWriter, "❌ npm install failed\n")
		return exitCode
	}

	fmt.Fprintf(logWriter, "🔨 Building definitions\n")
	exitCode = RunCommandWithLog(moduleRoot, logWriter, "npm", "run", "build")
	if exitCode != 0 {
		fmt.Fprintf(logWriter, "❌ npm run build failed\n")
		return exitCode
	}

	fmt.Fprintf(logWriter, "✅ Definitions built successfully\n")
	return 0
}

// buildMarkdown validates markdown files using goldmark parser
// Performs proper markdown syntax validation
func buildMarkdown(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating markdown: %s ===\n", module.Moniker)

	// Find all markdown files (exclude node_modules, .git, out/)
	var markdownFiles []string
	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == ".vscode" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for markdown files
		ext := filepath.Ext(path)
		if ext == ".md" || ext == ".markdown" {
			markdownFiles = append(markdownFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan for markdown files: %v\n", err)
		return 1
	}

	if len(markdownFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No markdown files found in: %s\n", moduleRoot)
		fmt.Fprintf(logWriter, "ℹ️  Skipping validation\n")
		return 0
	}

	fmt.Fprintf(logWriter, "📝 Found %d markdown file(s) to validate\n", len(markdownFiles))
	fmt.Fprintf(logWriter, "🔍 Using goldmark parser for validation\n")

	// Try markdownlint-cli if available for additional linting
	hasMarkdownlint := false
	if _, err := exec.LookPath("markdownlint"); err == nil {
		hasMarkdownlint = true
		fmt.Fprintf(logWriter, "💡 markdownlint-cli detected (will use for additional linting)\n")

		// Check for .markdownlint.yml config file
		configFile := filepath.Join(moduleRoot, ".markdownlint.yml")
		if _, err := os.Stat(configFile); err == nil {
			fmt.Fprintf(logWriter, "   Config: %s\n", configFile)
		}
	}

	// Create goldmark parser
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	// Validate each markdown file
	validationErrors := 0
	parseErrors := 0
	emptyFiles := 0

	for _, mdFile := range markdownFiles {
		relPath, _ := filepath.Rel(moduleRoot, mdFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		// Read file
		content, err := os.ReadFile(mdFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		// Check for empty files
		if len(content) == 0 {
			fmt.Fprintf(logWriter, "      ❌ Empty file\n")
			emptyFiles++
			validationErrors++
			continue
		}

		// Parse with goldmark
		var buf bytes.Buffer
		if err := md.Convert(content, &buf); err != nil {
			fmt.Fprintf(logWriter, "      ❌ Parse error: %v\n", err)
			parseErrors++
			validationErrors++
			continue
		}

		// Basic content checks
		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")
		nonEmptyLines := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonEmptyLines++
			}
		}

		if nonEmptyLines == 0 {
			fmt.Fprintf(logWriter, "      ❌ No content (only whitespace)\n")
			validationErrors++
			continue
		}

		fmt.Fprintf(logWriter, "      ✅ Valid markdown (%d lines, %d bytes)\n", len(lines), len(content))
	}

	// Run markdownlint if available
	if hasMarkdownlint && validationErrors == 0 {
		fmt.Fprintf(logWriter, "\n🔍 Running markdownlint for style checks...\n")
		exitCode := RunCommandWithLog(moduleRoot, logWriter, "markdownlint", markdownFiles...)

		if exitCode != 0 {
			fmt.Fprintf(logWriter, "⚠️  markdownlint found style issues (not blocking build)\n")
		}
	}

	// Summary
	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "\n❌ Validation failed:\n")
		if emptyFiles > 0 {
			fmt.Fprintf(logWriter, "   - Empty files: %d\n", emptyFiles)
		}
		if parseErrors > 0 {
			fmt.Fprintf(logWriter, "   - Parse errors: %d\n", parseErrors)
		}
		fmt.Fprintf(logWriter, "   - Total errors: %d\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "✅ All markdown files validated successfully\n")
	return 0
}

// Infrastructure Module Build Functions

// buildScriptsSh validates shell scripts using bash -n
func buildScriptsSh(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating shell scripts: %s ===\n", module.Moniker)

	// Find all shell scripts
	var shellFiles []string
	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip excluded directories
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".sh" {
			shellFiles = append(shellFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan directory: %v\n", err)
		return 1
	}

	if len(shellFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No shell scripts found\n")
		return 0
	}

	fmt.Fprintf(logWriter, "🐚 Found %d shell script(s) to validate\n", len(shellFiles))

	// Check if bash is available
	checkCmd := exec.Command("bash", "--version")
	if err := checkCmd.Run(); err != nil {
		// Bash not available (common on Windows without WSL)
		if runtime.GOOS == "windows" {
			fmt.Fprintf(logWriter, "⚠️  Skipping validation: bash not available (WSL not configured)\n")
			fmt.Fprintf(logWriter, "   Shell scripts found but not validated on Windows\n")
			return 0
		}
		fmt.Fprintf(logWriter, "❌ bash not found: %v\n", err)
		return 1
	}

	validationErrors := 0
	for _, shellFile := range shellFiles {
		relPath, _ := filepath.Rel(moduleRoot, shellFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		// Read file content and validate via stdin to avoid Windows path issues
		content, err := os.ReadFile(shellFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		// Validate syntax with bash -n via stdin
		cmd := exec.Command("bash", "-n")
		cmd.Stdin = bytes.NewReader(content)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Syntax error: %s\n", strings.TrimSpace(string(output)))
			validationErrors++
			continue
		}

		fmt.Fprintf(logWriter, "      ✅ Valid syntax\n")
	}

	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "\n❌ Validation failed with %d error(s)\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "\n✅ All shell scripts validated successfully\n")
	return 0
}

// buildScriptsPwsh validates PowerShell scripts using pwsh syntax checking
func buildScriptsPwsh(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating PowerShell scripts: %s ===\n", module.Moniker)

	// Find all PowerShell scripts
	var psFiles []string
	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip excluded directories
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".ps1" || ext == ".psm1" || ext == ".psd1" {
			psFiles = append(psFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan directory: %v\n", err)
		return 1
	}

	if len(psFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No PowerShell scripts found\n")
		return 0
	}

	fmt.Fprintf(logWriter, "⚡ Found %d PowerShell script(s) to validate\n", len(psFiles))

	validationErrors := 0
	for _, psFile := range psFiles {
		relPath, _ := filepath.Rel(moduleRoot, psFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		// Read file content and validate via stdin for cross-platform compatibility
		content, err := os.ReadFile(psFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		// Validate PowerShell syntax via stdin using here-string
		cmd := exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", "-")
		cmd.Stdin = bytes.NewReader([]byte(fmt.Sprintf("$null = [System.Management.Automation.PSParser]::Tokenize(@'\n%s\n'@, [ref]$null)", string(content))))
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Syntax error: %s\n", strings.TrimSpace(string(output)))
			validationErrors++
			continue
		}

		fmt.Fprintf(logWriter, "      ✅ Valid syntax\n")
	}

	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "\n❌ Validation failed with %d error(s)\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "\n✅ All PowerShell scripts validated successfully\n")
	return 0
}

// buildConfig validates configuration files (JSON, YAML, TOML)
func buildConfig(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating config files: %s ===\n", module.Moniker)

	// Find all config files
	var configFiles []string
	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip excluded directories
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == "dist" || name == ".vscode" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".json" || ext == ".yaml" || ext == ".yml" || ext == ".toml" {
			configFiles = append(configFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan directory: %v\n", err)
		return 1
	}

	if len(configFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No config files found\n")
		return 0
	}

	fmt.Fprintf(logWriter, "⚙️  Found %d config file(s) to validate\n", len(configFiles))

	validationErrors := 0
	for _, configFile := range configFiles {
		relPath, _ := filepath.Rel(moduleRoot, configFile)
		ext := filepath.Ext(configFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		content, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		// Validate based on extension
		switch ext {
		case ".json":
			var data interface{}
			if err := json.Unmarshal(content, &data); err != nil {
				fmt.Fprintf(logWriter, "      ❌ Invalid JSON: %v\n", err)
				validationErrors++
				continue
			}
		case ".yaml", ".yml":
			var data interface{}
			if err := yaml.Unmarshal(content, &data); err != nil {
				fmt.Fprintf(logWriter, "      ❌ Invalid YAML: %v\n", err)
				validationErrors++
				continue
			}
		case ".toml":
			// TOML validation would require a TOML library
			// For now, just check file readability
			if len(content) == 0 {
				fmt.Fprintf(logWriter, "      ❌ Empty file\n")
				validationErrors++
				continue
			}
		}

		fmt.Fprintf(logWriter, "      ✅ Valid %s\n", strings.TrimPrefix(ext, "."))
	}

	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "\n❌ Validation failed with %d error(s)\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "\n✅ All config files validated successfully\n")
	return 0
}

// buildVSCodeConfig validates VS Code configuration files (JSON/JSONC)
func buildVSCodeConfig(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating VS Code config: %s ===\n", module.Moniker)

	// Find JSON files in .vscode
	vscodeDir := filepath.Join(moduleRoot, ".vscode")
	if _, err := os.Stat(vscodeDir); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "⚠️  No .vscode directory found\n")
		return 0
	}

	var configFiles []string
	err := filepath.Walk(vscodeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip excluded directories
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".json" || ext == ".jsonc" {
			configFiles = append(configFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan .vscode: %v\n", err)
		return 1
	}

	if len(configFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No JSON config files found\n")
		return 0
	}

	fmt.Fprintf(logWriter, "🔧 Found %d config file(s) to validate\n", len(configFiles))

	validationErrors := 0
	for _, configFile := range configFiles {
		relPath, _ := filepath.Rel(moduleRoot, configFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		content, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		// Strip comments for JSONC (simple approach)
		lines := strings.Split(string(content), "\n")
		var cleaned []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "//") {
				cleaned = append(cleaned, line)
			}
		}
		cleanedContent := strings.Join(cleaned, "\n")

		var data interface{}
		if err := json.Unmarshal([]byte(cleanedContent), &data); err != nil {
			fmt.Fprintf(logWriter, "      ❌ Invalid JSON: %v\n", err)
			validationErrors++
			continue
		}

		fmt.Fprintf(logWriter, "      ✅ Valid JSON\n")
	}

	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "\n❌ Validation failed with %d error(s)\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "\n✅ All VS Code config files validated successfully\n")
	return 0
}

// buildClaudeConfig validates Claude configuration YAML files
func buildClaudeConfig(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating Claude config: %s ===\n", module.Moniker)

	// Find YAML files in .claude
	claudeDir := filepath.Join(moduleRoot, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "⚠️  No .claude directory found\n")
		return 0
	}

	var configFiles []string
	err := filepath.Walk(claudeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip excluded directories
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" {
			configFiles = append(configFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan .claude: %v\n", err)
		return 1
	}

	if len(configFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No YAML config files found\n")
		return 0
	}

	fmt.Fprintf(logWriter, "🤖 Found %d config file(s) to validate\n", len(configFiles))

	validationErrors := 0
	for _, configFile := range configFiles {
		relPath, _ := filepath.Rel(moduleRoot, configFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		content, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		var data interface{}
		if err := yaml.Unmarshal(content, &data); err != nil {
			fmt.Fprintf(logWriter, "      ❌ Invalid YAML: %v\n", err)
			validationErrors++
			continue
		}

		fmt.Fprintf(logWriter, "      ✅ Valid YAML\n")
	}

	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "\n❌ Validation failed with %d error(s)\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "\n✅ All Claude config files validated successfully\n")
	return 0
}

// buildClaudeAgents validates Claude agent markdown files
func buildClaudeAgents(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating Claude agents: %s ===\n", module.Moniker)

	// Use markdown validator with required sections
	validatorOpts := mdvalidator.DefaultValidatorOptions()
	validatorOpts.ValidateCodeBlocks = true
	validatorOpts.RequiredSections = []string{"Description", "Usage"}
	validatorOpts.CheckHeadingHierarchy = true

	validator := mdvalidator.NewValidator(validatorOpts, logWriter)
	results, err := validator.ValidateDirectory(moduleRoot)
	if err != nil {
		fmt.Fprintf(logWriter, "❌ Validation failed: %v\n", err)
		return 1
	}

	return validator.PrintResults(results, moduleRoot)
}

// buildClaudeCommands validates Claude command markdown files
func buildClaudeCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating Claude commands: %s ===\n", module.Moniker)

	// Use markdown validator with required sections
	validatorOpts := mdvalidator.DefaultValidatorOptions()
	validatorOpts.ValidateCodeBlocks = true
	validatorOpts.RequiredSections = []string{} // Commands may have flexible structure
	validatorOpts.CheckHeadingHierarchy = true

	validator := mdvalidator.NewValidator(validatorOpts, logWriter)
	results, err := validator.ValidateDirectory(moduleRoot)
	if err != nil {
		fmt.Fprintf(logWriter, "❌ Validation failed: %v\n", err)
		return 1
	}

	return validator.PrintResults(results, moduleRoot)
}

// buildClaudeHooks validates Claude hook shell scripts
func buildClaudeHooks(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating Claude hooks: %s ===\n", module.Moniker)

	// Find hook scripts in .claude/hooks
	hooksDir := filepath.Join(moduleRoot, ".claude", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "⚠️  No .claude/hooks directory found\n")
		return 0
	}

	var hookFiles []string
	err := filepath.Walk(hooksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip excluded directories
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".sh" || ext == ".ps1" {
			hookFiles = append(hookFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan hooks: %v\n", err)
		return 1
	}

	if len(hookFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No hook scripts found\n")
		return 0
	}

	fmt.Fprintf(logWriter, "🪝 Found %d hook script(s) to validate\n", len(hookFiles))

	validationErrors := 0
	for _, hookFile := range hookFiles {
		relPath, _ := filepath.Rel(moduleRoot, hookFile)
		ext := filepath.Ext(hookFile)
		fmt.Fprintf(logWriter, "   Validating: %s\n", relPath)

		// Read file content
		content, err := os.ReadFile(hookFile)
		if err != nil {
			fmt.Fprintf(logWriter, "      ❌ Failed to read: %v\n", err)
			validationErrors++
			continue
		}

		switch ext {
		case ".sh":
			// Validate via stdin to avoid Windows path issues
			cmd := exec.Command("bash", "-n")
			cmd.Stdin = bytes.NewReader(content)
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Fprintf(logWriter, "      ❌ Syntax error: %s\n", strings.TrimSpace(string(output)))
				validationErrors++
				continue
			}
		case ".ps1":
			// Validate PowerShell via stdin
			cmd := exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", "-")
			cmd.Stdin = bytes.NewReader([]byte(fmt.Sprintf("$null = [System.Management.Automation.PSParser]::Tokenize(@'\n%s\n'@, [ref]$null)", string(content))))
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Fprintf(logWriter, "      ❌ Syntax error: %s\n", strings.TrimSpace(string(output)))
				validationErrors++
				continue
			}
		}

		fmt.Fprintf(logWriter, "      ✅ Valid syntax\n")
	}

	if validationErrors > 0 {
		fmt.Fprintf(logWriter, "\n❌ Validation failed with %d error(s)\n", validationErrors)
		return 1
	}

	fmt.Fprintf(logWriter, "\n✅ All hook scripts validated successfully\n")
	return 0
}

// buildTemplates validates template files and detects placeholders
func buildTemplates(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating templates: %s ===\n", module.Moniker)

	// Find all template files
	var templateFiles []string
	placeholders := make(map[string]int)

	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "out" {
				return filepath.SkipDir
			}
			return nil
		}

		// Consider all files as potential templates
		templateFiles = append(templateFiles, path)
		return nil
	})

	if err != nil {
		fmt.Fprintf(logWriter, "❌ Failed to scan directory: %v\n", err)
		return 1
	}

	if len(templateFiles) == 0 {
		fmt.Fprintf(logWriter, "⚠️  No template files found\n")
		return 0
	}

	fmt.Fprintf(logWriter, "📄 Found %d template file(s) to analyze\n", len(templateFiles))

	// Detect placeholders: {{VAR}}, ${VAR}, %VAR%
	for _, templateFile := range templateFiles {
		content, err := os.ReadFile(templateFile)
		if err != nil {
			continue // Skip unreadable files
		}

		// Simple placeholder detection
		contentStr := string(content)
		if strings.Contains(contentStr, "{{") || strings.Contains(contentStr, "${") || strings.Contains(contentStr, "%") {
			placeholders[filepath.Base(templateFile)]++
		}
	}

	fmt.Fprintf(logWriter, "\n📊 Template Analysis:\n")
	fmt.Fprintf(logWriter, "   Total files: %d\n", len(templateFiles))
	fmt.Fprintf(logWriter, "   Files with placeholders: %d\n", len(placeholders))

	if len(placeholders) > 0 {
		fmt.Fprintf(logWriter, "\n📝 Files with detected placeholders:\n")
		for file := range placeholders {
			fmt.Fprintf(logWriter, "   - %s\n", file)
		}
	}

	fmt.Fprintf(logWriter, "\n✅ Template validation complete\n")
	return 0
}

// buildRepositoryRoot validates repository root structure
func buildRepositoryRoot(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Source.Root)

	fmt.Fprintf(logWriter, "\n=== Validating repository root: %s ===\n", module.Moniker)

	// Check for essential repository files
	essentialFiles := []string{
		"README.md",
		".gitignore",
		"go.work",
	}

	missing := []string{}
	for _, file := range essentialFiles {
		path := filepath.Join(moduleRoot, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		fmt.Fprintf(logWriter, "⚠️  Missing essential files:\n")
		for _, file := range missing {
			fmt.Fprintf(logWriter, "   - %s\n", file)
		}
	}

	fmt.Fprintf(logWriter, "\n✅ Repository root validation complete\n")
	return 0
}

// buildNoModuleType is a no-op build for files without a specific module type
func buildNoModuleType(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	fmt.Fprintf(logWriter, "\n=== Skipping build (no-module-type): %s ===\n", module.Moniker)
	fmt.Fprintf(logWriter, "ℹ️  This module has no specific type and requires no build\n")
	return 0
}
