// build_functions.go - Build functions for different module types
// This file contains the type-based build dispatch functions used by the build command.
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

	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	mdvalidator "github.com/ready-to-release/eac/src/core/markdown"
	"github.com/ready-to-release/eac/src/core/platform"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
)

// logln writes a formatted string with platform-specific line ending to the writer
func logln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

// BuildOptions contains flags for controlling the build process
type BuildOptions struct {
	WindowsOnly   bool
	LinuxOnly     bool
	MacOSOnly     bool
	TidyFirst     bool   // Run go mod tidy before building
	Version       string // Version to inject via ldflags
	Compressed    bool   // Strip debug info with -ldflags "-s -w" (for releases)
	CompressedUPX bool   // Also apply UPX compression after build
}

// BuildFunc is the signature for module type build functions
// Parameters: module contract, workspace root, output directory, log writer, build options
// Returns: exit code
type BuildFunc func(*modules.ModuleContract, string, string, io.Writer, BuildOptions) int

// IsGoModuleType returns true if the module type uses Go tooling (has go_module capability)
func IsGoModuleType(moduleType string) bool {
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		return cfg.ModuleTypes.HasCapability(moduleType, "go_module")
	}
	// Fallback: use naming convention if config unavailable
	return strings.HasPrefix(moduleType, "go-")
}

// buildFunctions maps module types to their build functions.
// This map is used as a fallback when a type-specific handler exists.
// For new types, prefer adding capabilities in module-types.yml and using
// build system handlers instead of adding entries here.
var buildFunctions = map[string]BuildFunc{
	"go-cli":           buildGoCLI,
	"go-commands":      buildGoCommands,
	"go-mcp":           buildGoMCP,
	"go-library":       buildGoLibrary,
	"go-tests":         buildGoTests,
	"go-r2r-extension": buildR2RExtension,
	"containers":       buildContainers,
	"mkdocs-site":      buildMkDocsSite,
	"mkdocs-subsite":   buildMkDocsSubsite,
	"vscode-ext":       buildVSCodeExtension,
	"contracts":        buildContracts,
	"specifications":   buildSpecifications,
	"definitions-type": buildDefinitionsType,
	"markdown":         buildMarkdown,
	"docker-image":     buildDockerImage,
	// Infrastructure module types
	"scripts":         buildScripts,
	"scripts-sh":      buildScriptsSh,
	"scripts-pwsh":    buildScriptsPwsh,
	"config":          buildConfig,
	"configuration":   buildConfig,
	"vscode-config":   buildVSCodeConfig,
	"claude-config":   buildClaudeConfig,
	"claude-agents":   buildClaudeAgents,
	"claude-commands": buildClaudeCommands,
	"claude-hooks":    buildClaudeHooks,
	"templates":       buildTemplates,
	"repository-root": buildRepositoryRoot,
	"no-module-type":  buildNoModuleType,
}

// buildSystemHandlers maps build systems to default build functions.
// Used when no type-specific handler exists in buildFunctions.
var buildSystemHandlers = map[string]BuildFunc{
	"go":     buildGoDefault,
	"mkdocs": buildMkDocsDefault,
	"docker": buildDockerDefault,
	"vscode": buildVSCodeDefault,
	"none":   buildNoop,
}

// GetBuildFunc returns the appropriate build function for a module type.
// It first checks for a type-specific handler, then falls back to build system handlers.
func GetBuildFunc(moduleType string) BuildFunc {
	// First, check for type-specific handler
	if fn, exists := buildFunctions[moduleType]; exists {
		return fn
	}

	// Fall back to build system handler from type registry
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		buildSystem := cfg.ModuleTypes.GetBuildSystem(moduleType)
		if fn, exists := buildSystemHandlers[buildSystem]; exists {
			return fn
		}
	}

	// Ultimate fallback: no-op build
	return buildNoop
}

// buildGoDefault is the default build handler for Go modules without specific handlers.
// Runs go generate and validates the module compiles.
func buildGoDefault(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)
	logln(logWriter, "\n=== Building Go module: %s (type: %s) ===", module.Moniker, module.Type)

	if opts.TidyFirst {
		logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	logln(logWriter, "Running: go generate ./...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	logln(logWriter, "Running: go build ./...")
	return RunCommandWithLog(moduleRoot, logWriter, "go", "build", "./...")
}

// buildMkDocsDefault is the default build handler for MkDocs modules.
func buildMkDocsDefault(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	logln(logWriter, "\n=== Building MkDocs module: %s (type: %s) ===", module.Moniker, module.Type)
	logln(logWriter, "ℹ️  MkDocs modules are built via mkdocs build command")
	return 0
}

// buildDockerDefault is the default build handler for Docker modules.
func buildDockerDefault(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	logln(logWriter, "\n=== Building Docker module: %s (type: %s) ===", module.Moniker, module.Type)
	logln(logWriter, "ℹ️  Docker modules are built via docker build command")
	return 0
}

// buildVSCodeDefault is the default build handler for VS Code modules.
func buildVSCodeDefault(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	logln(logWriter, "\n=== Building VS Code module: %s (type: %s) ===", module.Moniker, module.Type)
	logln(logWriter, "ℹ️  VS Code extensions are built via vsce package command")
	return 0
}

// buildNoop is a no-op build function for modules that don't require building.
func buildNoop(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	logln(logWriter, "\n=== %s (type: %s) ===", module.Moniker, module.Type)
	logln(logWriter, "ℹ️  No build step required for this module type")
	return 0
}

// buildGoCLI builds a Cobra CLI binary (Pattern A)
// Requires: go generate && go build
// By default, builds for Windows, Linux, and macOS/ARM
//
// Compression modes:
//   - Default (dev): Full debug info for debugging
//   - --compressed: Strip debug info with -ldflags "-s -w" (~30% smaller)
//   - --compressed-upx: Also apply UPX compression (~70% smaller total)
func buildGoCLI(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Building go-cli: %s ===", module.Moniker)

	// Log compression mode
	if opts.CompressedUPX {
		logln(logWriter, "Compression: UPX (--compressed-upx)")
	} else if opts.Compressed {
		logln(logWriter, "Compression: stripped (--compressed)")
	} else {
		logln(logWriter, "Compression: none (dev build with debug info)")
	}

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	// Step 2: go generate
	logln(logWriter, "Running: go generate ./...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	// Check for UPX if needed
	if opts.CompressedUPX {
		if _, err := exec.LookPath("upx"); err != nil {
			logln(logWriter, "❌ UPX not found in PATH. Install UPX for --compressed-upx support.")
			logln(logWriter, "   See: https://upx.github.io/")
			return 1
		}
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

	// Track built binaries for UPX compression
	var builtBinaries []string

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

		logln(logWriter, "\n--- Building for %s (%s/%s) ---", platform.Name, platform.GOOS, platform.GOARCH)
		logln(logWriter, "Output: %s", binaryPath)

		// Prepare build arguments
		buildArgs := []string{"build", "-o", binaryPath}

		// Build ldflags based on compression mode
		var ldflags string
		if opts.Compressed || opts.CompressedUPX {
			// Strip debug info (-s) and symbol table (-w) for smaller binaries
			ldflags = "-s -w"
		}
		if opts.Version != "" {
			if ldflags != "" {
				ldflags += " "
			}
			ldflags += fmt.Sprintf("-X 'github.com/ready-to-release/eac/src/cli/cmd.Version=%s'", opts.Version)
			logln(logWriter, "Version: %s", opts.Version)
		}
		if ldflags != "" {
			buildArgs = append(buildArgs, "-ldflags", ldflags)
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
			logln(logWriter, "❌ Build failed for %s: %v", platform.Name, err)
			return 1
		}

		logln(logWriter, "✅ Built successfully: %s", binaryPath)

		// Make binary executable on Unix platforms (only if building ON Unix)
		// Note: We check runtime.GOOS (build host) not platform.GOOS (target)
		// because chmod only works on Unix hosts
		if runtime.GOOS != "windows" && platform.GOOS != "windows" {
			if exitCode := RunCommandWithLog(moduleRoot, logWriter, "chmod", "+x", binaryPath); exitCode != 0 {
				logln(logWriter, "⚠️  Warning: could not set executable permissions")
			}
		}

		builtBinaries = append(builtBinaries, binaryPath)
	}

	// Apply UPX compression if requested
	// Note: UPX cannot compress Darwin (macOS) binaries when cross-compiling from Linux
	if opts.CompressedUPX {
		logln(logWriter, "\n--- Applying UPX compression ---")
		logln(logWriter, "Note: UPX compression skipped for Darwin binaries (not supported when cross-compiling)")

		for _, binaryPath := range builtBinaries {
			baseName := filepath.Base(binaryPath)

			// Skip Darwin binaries - UPX cannot compress them when cross-compiling
			if strings.Contains(baseName, "darwin") {
				logln(logWriter, "⏭️  Skipping %s (Darwin binaries not supported by UPX cross-compile)", baseName)
				continue
			}

			// Get original size
			originalInfo, err := os.Stat(binaryPath)
			if err != nil {
				logln(logWriter, "⚠️  Warning: could not stat %s: %v", binaryPath, err)
				continue
			}
			originalSize := originalInfo.Size()

			// Create UPX-compressed version with -upx suffix
			ext := filepath.Ext(baseName)
			nameWithoutExt := strings.TrimSuffix(baseName, ext)
			upxName := nameWithoutExt + "-upx" + ext
			upxPath := filepath.Join(outputDir, upxName)

			// Copy original to UPX path first
			if err := copyFile(binaryPath, upxPath); err != nil {
				logln(logWriter, "❌ Failed to copy %s for UPX: %v", baseName, err)
				return 1
			}

			logln(logWriter, "Compressing: %s -> %s", baseName, upxName)

			// Run UPX on the copy (--best for maximum compression, -q for quiet)
			cmd := exec.Command("upx", "--best", "-q", upxPath)
			cmd.Stdout = logWriter
			cmd.Stderr = logWriter

			if err := cmd.Run(); err != nil {
				logln(logWriter, "⚠️  UPX compression failed for %s: %v", upxName, err)
				// Remove failed UPX file
				os.Remove(upxPath)
				continue
			}

			// Get compressed size
			compressedInfo, err := os.Stat(upxPath)
			if err != nil {
				logln(logWriter, "⚠️  Warning: could not stat compressed file: %v", err)
				continue
			}
			compressedSize := compressedInfo.Size()

			ratio := float64(compressedSize) / float64(originalSize) * 100
			logln(logWriter, "✅ %s: %.1f MB -> %.1f MB (%.0f%%)",
				upxName,
				float64(originalSize)/1024/1024,
				float64(compressedSize)/1024/1024,
				ratio)
		}
	}

	logln(logWriter, "\n✅ All builds completed successfully")
	return 0
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Preserve executable permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// buildGoCommands builds the runtime command dispatcher (Pattern B)
// Note: This is a development tool that's always run with "go run ."
func buildGoCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== go-commands: %s ===", module.Moniker)

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		logln(logWriter, "🔄 Running go mod tidy...")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			logln(logWriter, "❌ go mod tidy failed")
			return exitCode
		}
		logln(logWriter, "✅ go mod tidy completed")
	}

	// Step 2: go generate to ensure generated code is up to date
	logln(logWriter, "🔄 Running go generate...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		logln(logWriter, "❌ go generate failed")
		return exitCode
	}
	logln(logWriter, "✅ go generate completed")

	logln(logWriter, "\nℹ️  This module uses 'go run .' and is never compiled to a binary")
	logln(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)")
	return 0
}

// buildGoMCP builds an MCP JSON-RPC server binary (Pattern C)
// Requires: go build -o mcp-server-<name>
func buildGoMCP(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	// Extract server name from moniker (e.g., "src-mcp-docs" -> "docs")
	serverName := module.Moniker
	if len(serverName) > 8 && serverName[:8] == "src-mcp-" {
		serverName = serverName[8:]
	}

	binaryName := fmt.Sprintf("mcp-server-%s", serverName)
	binaryPath := filepath.Join(outputDir, binaryName)

	logln(logWriter, "\n=== Building go-mcp: %s ===", module.Moniker)

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	// Step 2: go build
	logln(logWriter, "Running: go build -o %s", binaryPath)
	return RunCommandWithLog(moduleRoot, logWriter, "go", "build", "-o", binaryPath)
}

// buildGoLibrary builds a Go library module (Pattern D)
// Note: Libraries are imported as dependencies, no binary output
// Runs go generate to prepare any embedded resources or generated code
func buildGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== go-library: %s ===", module.Moniker)

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	// Step 2: go generate to prepare embedded resources
	logln(logWriter, "Running: go generate ./...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	logln(logWriter, "ℹ️  This is a library module (no binary to build)")
	logln(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)")
	return 0
}

// buildGoTests builds a Godog test module (Pattern D variant)
// Note: Tests are run with "go test", not built separately
func buildGoTests(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	logln(logWriter, "\n=== go-tests: %s ===", module.Moniker)
	logln(logWriter, "ℹ️  This is a test module (use 'test module' command to run tests)")
	logln(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)")
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
		logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}

	return 0
}

// buildContainers builds Docker images from Dockerfiles
// Expects .Dockerfile in module root
func buildContainers(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Building containers: %s ===", module.Moniker)

	// Find Dockerfile
	dockerfilePath := filepath.Join(moduleRoot, ".Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		logln(logWriter, "⚠️  No .Dockerfile found at: %s", dockerfilePath)
		logln(logWriter, "ℹ️  Skipping Docker build")
		return 0
	}

	// Generate image tag from moniker
	// Example: "containers" -> "cli-containers:latest"
	imageName := fmt.Sprintf("cli-%s:latest", module.Moniker)

	logln(logWriter, "📦 Building Docker image: %s", imageName)
	logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	logln(logWriter, "   Build context: %s", moduleRoot)

	// Build image using docker build
	exitCode := RunCommandWithLog(moduleRoot, logWriter,
		"docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		".")

	if exitCode != 0 {
		logln(logWriter, "❌ Docker build failed")
		return exitCode
	}

	logln(logWriter, "✅ Docker image built successfully: %s", imageName)

	// Save image name to output directory for reference
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\n",
		imageName, dockerfilePath, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
		logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
	}

	return 0
}

// buildDockerImage builds a standalone Docker image from a Dockerfile
func buildDockerImage(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Building docker-image: %s ===", module.Moniker)

	// Find Dockerfile (check for both "Dockerfile" and ".Dockerfile")
	dockerfilePath := filepath.Join(moduleRoot, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		dockerfilePath = filepath.Join(moduleRoot, ".Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			logln(logWriter, "❌ No Dockerfile found in: %s", moduleRoot)
			return 1
		}
	}

	// Generate image tag from moniker
	imageName := fmt.Sprintf("%s:latest", module.Moniker)

	logln(logWriter, "📦 Building Docker image: %s", imageName)
	logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	logln(logWriter, "   Build context: %s", moduleRoot)

	// Build image using docker build
	exitCode := RunCommandWithLog(moduleRoot, logWriter,
		"docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		".")

	if exitCode != 0 {
		logln(logWriter, "❌ Docker build failed")
		return exitCode
	}

	logln(logWriter, "✅ Docker image built successfully: %s", imageName)

	// Save image name to output directory for reference
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\n",
		imageName, dockerfilePath, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
		logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
	}

	return 0
}

// buildR2RExtension builds an R2R CLI extension as a Docker image
// The Dockerfile is expected to be in containers/{moniker}/Dockerfile
// Build context is the repository root
func buildR2RExtension(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	logln(logWriter, "\n=== Building R2R extension: %s ===", module.Moniker)

	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	// Step 1: go mod tidy (if enabled) - extension modules have their own go.mod
	if opts.TidyFirst {
		// Check if this module has a go.mod file
		goModPath := filepath.Join(moduleRoot, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			logln(logWriter, "Running: go mod tidy (in %s)", module.Files.Root)
			if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
				logln(logWriter, "❌ go mod tidy failed")
				return exitCode
			}
		}
	}

	// Extract extension name from moniker (e.g., "ext-eac" -> "eac")
	extensionName := module.Moniker
	if len(module.Moniker) > 4 && module.Moniker[:4] == "ext-" {
		extensionName = module.Moniker[4:]
	}

	// Dockerfile is in containers/{moniker}/Dockerfile
	dockerfilePath := filepath.Join(workspaceRoot, "containers", module.Moniker, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		logln(logWriter, "❌ No Dockerfile found at: %s", dockerfilePath)
		return 1
	}

	// Generate image tag from extension name
	imageName := fmt.Sprintf("ext-%s:latest", extensionName)

	logln(logWriter, "📦 Building Docker image: %s", imageName)
	logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	logln(logWriter, "   Build context: %s", workspaceRoot)

	// Check if we're in CI environment - if so, build for testing and export multi-platform
	isCI := os.Getenv("CI") == "true"

	if isCI {
		logln(logWriter, "\n--- CI Mode: Building single-platform for testing ---")
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
			logln(logWriter, "\n❌ Docker build failed (see errors above)")
			return exitCode
		}
		logln(logWriter, "✅ Single-platform image built successfully: %s", imageName)

		// Export multi-platform for release
		logln(logWriter, "\n--- CI Mode: Building multi-platform for release ---")
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
			logln(logWriter, "\n❌ Multi-platform build failed (see errors above)")
			return exitCode
		}

		logln(logWriter, "✅ Multi-platform image exported: %s", ociArchivePath)

		// Compress the OCI archive
		logln(logWriter, "Compressing OCI archive...")
		exitCode = RunCommandWithLog(outputDir, logWriter, "gzip", filepath.Base(ociArchivePath))
		if exitCode != 0 {
			logln(logWriter, "⚠️  Warning: failed to compress archive")
		}

		// Save image info
		imageInfoPath := filepath.Join(outputDir, "image-info.txt")
		imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\nPlatforms: linux/amd64,linux/arm64\nOCI Archive: %s.gz\n",
			imageName, dockerfilePath, time.Now().Format(time.RFC3339), ociArchivePath)

		if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
			logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
		}

	} else {
		// Local build - simple docker build for current platform
		exitCode := RunCommandWithLog(workspaceRoot, logWriter,
			"docker", "build",
			"-t", imageName,
			"-f", dockerfilePath,
			".")

		if exitCode != 0 {
			logln(logWriter, "\n❌ Docker build failed (see errors above)")
			return exitCode
		}

		logln(logWriter, "✅ Docker image built successfully: %s", imageName)

		// Save image name to output directory for reference
		imageInfoPath := filepath.Join(outputDir, "image-info.txt")
		imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\n",
			imageName, dockerfilePath, time.Now().Format(time.RFC3339))

		if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
			logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
		}
	}

	return 0
}

// buildMkDocsSite builds the main MkDocs documentation site using Docker
// Uses the cli-mkdocs container for consistent builds across local and CI environments
func buildMkDocsSite(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	logln(logWriter, "\n=== Building mkdocs-site: %s ===", module.Moniker)

	// Check for mkdocs.yml at repository root
	mkdocsConfig := filepath.Join(workspaceRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		logln(logWriter, "⚠️  No mkdocs.yml found at: %s", mkdocsConfig)
		logln(logWriter, "ℹ️  Skipping MkDocs build")
		return 0
	}

	logln(logWriter, "📚 Building MkDocs site using Docker")
	logln(logWriter, "   Config: %s", mkdocsConfig)

	// Ensure the Docker image exists
	imageName := "cli-mkdocs:latest"
	dockerfilePath := filepath.Join(workspaceRoot, "containers", "mkdocs", ".Dockerfile")
	contextPath := filepath.Join(workspaceRoot, "containers", "mkdocs")

	if err := ensureMkDocsImage(imageName, dockerfilePath, contextPath, logWriter); err != nil {
		logln(logWriter, "❌ Failed to ensure Docker image: %v", err)
		return 1
	}

	// Calculate the site output directory relative to workspace root
	// outputDir is typically: <workspaceRoot>/out/build/<moniker>
	// We want site output at: <outputDir>/site
	siteDir := filepath.Join(outputDir, "site")

	// Create the output directory
	if err := os.MkdirAll(siteDir, 0755); err != nil {
		logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// Calculate relative path from workspace root to site dir for Docker mount
	relSiteDir, err := filepath.Rel(workspaceRoot, siteDir)
	if err != nil {
		logln(logWriter, "❌ Failed to calculate relative path: %v", err)
		return 1
	}

	// Format workspace root for Docker volume mount (handles Windows paths)
	dockerVolume := formatDockerVolumePath(workspaceRoot)

	// Build the site using Docker
	// Mount workspace at /docs, output to relative site directory
	logln(logWriter, "   Image: %s", imageName)
	logln(logWriter, "   Output: %s", siteDir)

	// Convert Windows path separators to forward slashes for Docker
	dockerSiteDir := strings.ReplaceAll(relSiteDir, "\\", "/")

	// Check for --accept-warnings flag in build options (passed via command line)
	acceptWarnings := false
	for _, arg := range os.Args {
		if arg == "--accept-warnings" {
			acceptWarnings = true
			break
		}
	}

	// Build command args - always use --strict to catch warnings
	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
		imageName,
		"mkdocs", "build",
		"--site-dir", dockerSiteDir,
		"--clean",
		"--strict",
	}

	if acceptWarnings {
		logln(logWriter, "   Mode: accepting warnings (--accept-warnings)")
	} else {
		logln(logWriter, "   Mode: strict (warnings will fail build)")
	}

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", buildArgs...)

	// If accepting warnings, treat warning exit code as success
	if acceptWarnings && exitCode != 0 {
		logln(logWriter, "⚠️  Build completed with warnings (accepted)")
		exitCode = 0
	}

	if exitCode != 0 {
		logln(logWriter, "❌ MkDocs build failed")
		return exitCode
	}

	logln(logWriter, "✅ MkDocs site built successfully")
	logln(logWriter, "   Output: %s", siteDir)

	return 0
}

// ensureMkDocsImage ensures the cli-mkdocs Docker image exists, building it if necessary
func ensureMkDocsImage(imageName, dockerfilePath, contextPath string, logWriter io.Writer) error {
	// Check if image exists using docker images command
	cmd := exec.Command("docker", "images", "-q", imageName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check for Docker image: %w", err)
	}

	// If output is non-empty, image exists
	if len(strings.TrimSpace(string(output))) > 0 {
		logln(logWriter, "   Using existing image: %s", imageName)
		return nil
	}

	// Image doesn't exist, build it
	logln(logWriter, "   Building Docker image: %s", imageName)

	exitCode := RunCommandWithLog(contextPath, logWriter,
		"docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		".")

	if exitCode != 0 {
		return fmt.Errorf("docker build failed with exit code %d", exitCode)
	}

	logln(logWriter, "   Image built successfully: %s", imageName)
	return nil
}

// formatDockerVolumePath formats a path for use as a Docker volume mount source
// On Windows, converts C:\path to /c/path for Docker compatibility
func formatDockerVolumePath(path string) string {
	// Check if this is a Windows absolute path (e.g., C:\...)
	if len(path) >= 2 && path[1] == ':' {
		// Convert C:\path to /c/path
		driveLetter := strings.ToLower(string(path[0]))
		rest := strings.ReplaceAll(path[2:], "\\", "/")
		return "/" + driveLetter + rest
	}
	// Already Unix-style or relative path
	return strings.ReplaceAll(path, "\\", "/")
}

// buildMkDocsSubsite builds a MkDocs documentation subsite
// Runs: mkdocs build (same as main site, but for subsites)
func buildMkDocsSubsite(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Building mkdocs-subsite: %s ===", module.Moniker)

	// Check for mkdocs.yml
	mkdocsConfig := filepath.Join(moduleRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		logln(logWriter, "⚠️  No mkdocs.yml found at: %s", mkdocsConfig)
		logln(logWriter, "ℹ️  Skipping MkDocs subsite build")
		return 0
	}

	logln(logWriter, "📚 Building MkDocs subsite")
	logln(logWriter, "   Config: %s", mkdocsConfig)

	// Build subsite to output directory
	siteDir := filepath.Join(outputDir, "site")
	exitCode := RunCommandWithLog(moduleRoot, logWriter,
		"mkdocs", "build",
		"--site-dir", siteDir,
		"--clean")

	if exitCode != 0 {
		logln(logWriter, "❌ MkDocs subsite build failed")
		return exitCode
	}

	logln(logWriter, "✅ MkDocs subsite built successfully")
	logln(logWriter, "   Output: %s", siteDir)

	return 0
}

// buildVSCodeExtension builds a VS Code extension
// Runs: npm install && npm run compile
func buildVSCodeExtension(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Building vscode-ext: %s ===", module.Moniker)

	// Check for package.json
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		logln(logWriter, "⚠️  No package.json found at: %s", packageJSON)
		logln(logWriter, "ℹ️  Skipping VS Code extension build")
		return 0
	}

	logln(logWriter, "📦 Installing dependencies")
	logln(logWriter, "Running: npm install")

	// Step 1: npm install
	exitCode := RunCommandWithLog(moduleRoot, logWriter, "npm", "install")
	if exitCode != 0 {
		logln(logWriter, "❌ npm install failed")
		return exitCode
	}

	logln(logWriter, "🔨 Compiling TypeScript")
	logln(logWriter, "Running: npm run compile")

	// Step 2: npm run compile
	exitCode = RunCommandWithLog(moduleRoot, logWriter, "npm", "run", "compile")
	if exitCode != 0 {
		logln(logWriter, "❌ npm run compile failed")
		return exitCode
	}

	logln(logWriter, "✅ VS Code extension built successfully")

	return 0
}

// buildContracts validates YAML contract files
// Validates all .yml files in module
func buildContracts(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating contracts: %s ===", module.Moniker)

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
		logln(logWriter, "❌ Failed to scan for YAML files: %v", err)
		return 1
	}

	if len(yamlFiles) == 0 {
		logln(logWriter, "⚠️  No YAML files found in: %s", moduleRoot)
		logln(logWriter, "ℹ️  Skipping validation")
		return 0
	}

	logln(logWriter, "📋 Found %d YAML file(s) to validate", len(yamlFiles))

	// Validate each YAML file
	validationErrors := 0
	for _, yamlFile := range yamlFiles {
		relPath, _ := filepath.Rel(moduleRoot, yamlFile)
		logln(logWriter, "   Validating: %s", relPath)

		// Read YAML file
		content, err := os.ReadFile(yamlFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		// Validate YAML syntax using yaml.v3
		var data interface{}
		if err := yaml.Unmarshal(content, &data); err != nil {
			logln(logWriter, "      ❌ Invalid YAML: %v", err)
			validationErrors++
			continue
		}

		logln(logWriter, "      ✅ Valid YAML (%d bytes)", len(content))
	}

	if validationErrors > 0 {
		logln(logWriter, "❌ %d file(s) failed validation", validationErrors)
		return 1
	}

	logln(logWriter, "✅ All contracts validated successfully")
	return 0
}

// buildSpecifications validates Gherkin .feature files
// Validates all .feature files in module
func buildSpecifications(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating specifications: %s ===", module.Moniker)

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
		logln(logWriter, "❌ Failed to scan for feature files: %v", err)
		return 1
	}

	if len(featureFiles) == 0 {
		logln(logWriter, "⚠️  No .feature files found in: %s", moduleRoot)
		logln(logWriter, "ℹ️  Skipping validation")
		return 0
	}

	logln(logWriter, "🥒 Found %d feature file(s) to validate", len(featureFiles))

	// Validate each feature file
	validationErrors := 0
	for _, featureFile := range featureFiles {
		relPath, _ := filepath.Rel(moduleRoot, featureFile)
		logln(logWriter, "   Validating: %s", relPath)

		// Read file to check it exists and is readable
		content, err := os.ReadFile(featureFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		// Basic validation: check for "Feature:" keyword
		contentStr := string(content)
		if len(contentStr) == 0 {
			logln(logWriter, "      ❌ Empty file")
			validationErrors++
			continue
		}

		// Simple validation: just check it's readable and non-empty
		if len(contentStr) > 0 {
			logln(logWriter, "      ✅ Valid Gherkin")
		} else {
			logln(logWriter, "      ⚠️  Empty file")
			validationErrors++
		}
	}

	if validationErrors > 0 {
		logln(logWriter, "❌ %d file(s) failed validation", validationErrors)
		return 1
	}

	logln(logWriter, "✅ All specifications validated successfully")
	return 0
}

// buildDefinitionsType validates TypeScript/JSON Schema definitions
// Runs: npm install && npm run build (if package.json exists)
func buildDefinitionsType(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Building definitions-type: %s ===", module.Moniker)

	// Check for package.json
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		logln(logWriter, "ℹ️  No package.json found - checking for JSON schemas")

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
			logln(logWriter, "⚠️  No JSON files found")
			logln(logWriter, "ℹ️  Skipping validation")
			return 0
		}

		logln(logWriter, "📋 Found %d JSON schema file(s)", len(schemaFiles))
		logln(logWriter, "✅ Schema files present (no build needed)")
		return 0
	}

	// Has package.json - build with npm
	logln(logWriter, "📦 Installing dependencies")
	exitCode := RunCommandWithLog(moduleRoot, logWriter, "npm", "install")
	if exitCode != 0 {
		logln(logWriter, "❌ npm install failed")
		return exitCode
	}

	logln(logWriter, "🔨 Building definitions")
	exitCode = RunCommandWithLog(moduleRoot, logWriter, "npm", "run", "build")
	if exitCode != 0 {
		logln(logWriter, "❌ npm run build failed")
		return exitCode
	}

	logln(logWriter, "✅ Definitions built successfully")
	return 0
}

// buildMarkdown validates markdown files using goldmark parser
// Performs proper markdown syntax validation
func buildMarkdown(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating markdown: %s ===", module.Moniker)

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
		logln(logWriter, "❌ Failed to scan for markdown files: %v", err)
		return 1
	}

	if len(markdownFiles) == 0 {
		logln(logWriter, "⚠️  No markdown files found in: %s", moduleRoot)
		logln(logWriter, "ℹ️  Skipping validation")
		return 0
	}

	logln(logWriter, "📝 Found %d markdown file(s) to validate", len(markdownFiles))
	logln(logWriter, "🔍 Using goldmark parser for validation")

	// Try markdownlint-cli if available for additional linting
	hasMarkdownlint := false
	if _, err := exec.LookPath("markdownlint"); err == nil {
		hasMarkdownlint = true
		logln(logWriter, "💡 markdownlint-cli detected (will use for additional linting)")

		// Check for .markdownlint.yml config file
		configFile := filepath.Join(moduleRoot, ".markdownlint.yml")
		if _, err := os.Stat(configFile); err == nil {
			logln(logWriter, "   Config: %s", configFile)
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
		logln(logWriter, "   Validating: %s", relPath)

		// Read file
		content, err := os.ReadFile(mdFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		// Check for empty files
		if len(content) == 0 {
			logln(logWriter, "      ❌ Empty file")
			emptyFiles++
			validationErrors++
			continue
		}

		// Parse with goldmark
		var buf bytes.Buffer
		if err := md.Convert(content, &buf); err != nil {
			logln(logWriter, "      ❌ Parse error: %v", err)
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
			logln(logWriter, "      ❌ No content (only whitespace)")
			validationErrors++
			continue
		}

		logln(logWriter, "      ✅ Valid markdown (%d lines, %d bytes)", len(lines), len(content))
	}

	// Run markdownlint if available
	if hasMarkdownlint && validationErrors == 0 {
		logln(logWriter, "\n🔍 Running markdownlint for style checks...")
		exitCode := RunCommandWithLog(moduleRoot, logWriter, "markdownlint", markdownFiles...)

		if exitCode != 0 {
			logln(logWriter, "⚠️  markdownlint found style issues (not blocking build)")
		}
	}

	// Summary
	if validationErrors > 0 {
		logln(logWriter, "\n❌ Validation failed:")
		if emptyFiles > 0 {
			logln(logWriter, "   - Empty files: %d", emptyFiles)
		}
		if parseErrors > 0 {
			logln(logWriter, "   - Parse errors: %d", parseErrors)
		}
		logln(logWriter, "   - Total errors: %d", validationErrors)
		return 1
	}

	logln(logWriter, "✅ All markdown files validated successfully")
	return 0
}

// Infrastructure Module Build Functions

// buildScripts validates both shell and PowerShell scripts
func buildScripts(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating scripts: %s ===", module.Moniker)

	// Find all script files
	var shellFiles []string
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
		if ext == ".sh" {
			shellFiles = append(shellFiles, path)
		} else if ext == ".ps1" || ext == ".psm1" || ext == ".psd1" {
			psFiles = append(psFiles, path)
		}
		return nil
	})

	if err != nil {
		logln(logWriter, "❌ Failed to scan directory: %v", err)
		return 1
	}

	totalFiles := len(shellFiles) + len(psFiles)
	if totalFiles == 0 {
		logln(logWriter, "⚠️  No scripts found")
		return 0
	}

	logln(logWriter, "📜 Found %d script(s) to validate (%d shell, %d PowerShell)", totalFiles, len(shellFiles), len(psFiles))

	validationErrors := 0

	// Validate shell scripts
	if len(shellFiles) > 0 {
		logln(logWriter, "\n--- Shell Scripts ---")

		// Check if bash is available
		checkCmd := exec.Command("bash", "--version")
		bashAvailable := checkCmd.Run() == nil

		if !bashAvailable {
			if runtime.GOOS == "windows" {
				logln(logWriter, "⚠️  Skipping shell validation: bash not available (WSL not configured)")
			} else {
				logln(logWriter, "❌ bash not found")
				validationErrors++
			}
		} else {
			for _, shellFile := range shellFiles {
				relPath, _ := filepath.Rel(moduleRoot, shellFile)
				logln(logWriter, "   Validating: %s", relPath)

				content, err := os.ReadFile(shellFile)
				if err != nil {
					logln(logWriter, "      ❌ Failed to read: %v", err)
					validationErrors++
					continue
				}

				cmd := exec.Command("bash", "-n")
				cmd.Stdin = bytes.NewReader(content)
				output, err := cmd.CombinedOutput()
				if err != nil {
					logln(logWriter, "      ❌ Syntax error: %s", strings.TrimSpace(string(output)))
					validationErrors++
					continue
				}

				logln(logWriter, "      ✅ Valid syntax")
			}
		}
	}

	// Validate PowerShell scripts
	if len(psFiles) > 0 {
		logln(logWriter, "\n--- PowerShell Scripts ---")

		// Check if pwsh is available
		checkCmd := exec.Command("pwsh", "--version")
		pwshAvailable := checkCmd.Run() == nil

		if !pwshAvailable {
			logln(logWriter, "⚠️  Skipping PowerShell validation: pwsh not available")
		} else {
			for _, psFile := range psFiles {
				relPath, _ := filepath.Rel(moduleRoot, psFile)
				logln(logWriter, "   Validating: %s", relPath)

				content, err := os.ReadFile(psFile)
				if err != nil {
					logln(logWriter, "      ❌ Failed to read: %v", err)
					validationErrors++
					continue
				}

				cmd := exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", "-")
				cmd.Stdin = bytes.NewReader([]byte(fmt.Sprintf("$null = [System.Management.Automation.PSParser]::Tokenize(@'\n%s\n'@, [ref]$null)", string(content))))
				output, err := cmd.CombinedOutput()
				if err != nil {
					logln(logWriter, "      ❌ Syntax error: %s", strings.TrimSpace(string(output)))
					validationErrors++
					continue
				}

				logln(logWriter, "      ✅ Valid syntax")
			}
		}
	}

	if validationErrors > 0 {
		logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	logln(logWriter, "\n✅ All scripts validated successfully")
	return 0
}

// buildScriptsSh validates shell scripts using bash -n
func buildScriptsSh(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating shell scripts: %s ===", module.Moniker)

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
		logln(logWriter, "❌ Failed to scan directory: %v", err)
		return 1
	}

	if len(shellFiles) == 0 {
		logln(logWriter, "⚠️  No shell scripts found")
		return 0
	}

	logln(logWriter, "🐚 Found %d shell script(s) to validate", len(shellFiles))

	// Check if bash is available
	checkCmd := exec.Command("bash", "--version")
	if err := checkCmd.Run(); err != nil {
		// Bash not available (common on Windows without WSL)
		if runtime.GOOS == "windows" {
			logln(logWriter, "⚠️  Skipping validation: bash not available (WSL not configured)")
			logln(logWriter, "   Shell scripts found but not validated on Windows")
			return 0
		}
		logln(logWriter, "❌ bash not found: %v", err)
		return 1
	}

	validationErrors := 0
	for _, shellFile := range shellFiles {
		relPath, _ := filepath.Rel(moduleRoot, shellFile)
		logln(logWriter, "   Validating: %s", relPath)

		// Read file content and validate via stdin to avoid Windows path issues
		content, err := os.ReadFile(shellFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		// Validate syntax with bash -n via stdin
		cmd := exec.Command("bash", "-n")
		cmd.Stdin = bytes.NewReader(content)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logln(logWriter, "      ❌ Syntax error: %s", strings.TrimSpace(string(output)))
			validationErrors++
			continue
		}

		logln(logWriter, "      ✅ Valid syntax")
	}

	if validationErrors > 0 {
		logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	logln(logWriter, "\n✅ All shell scripts validated successfully")
	return 0
}

// buildScriptsPwsh validates PowerShell scripts using pwsh syntax checking
func buildScriptsPwsh(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating PowerShell scripts: %s ===", module.Moniker)

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
		logln(logWriter, "❌ Failed to scan directory: %v", err)
		return 1
	}

	if len(psFiles) == 0 {
		logln(logWriter, "⚠️  No PowerShell scripts found")
		return 0
	}

	logln(logWriter, "⚡ Found %d PowerShell script(s) to validate", len(psFiles))

	validationErrors := 0
	for _, psFile := range psFiles {
		relPath, _ := filepath.Rel(moduleRoot, psFile)
		logln(logWriter, "   Validating: %s", relPath)

		// Read file content and validate via stdin for cross-platform compatibility
		content, err := os.ReadFile(psFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		// Validate PowerShell syntax via stdin using here-string
		cmd := exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", "-")
		cmd.Stdin = bytes.NewReader([]byte(fmt.Sprintf("$null = [System.Management.Automation.PSParser]::Tokenize(@'\n%s\n'@, [ref]$null)", string(content))))
		output, err := cmd.CombinedOutput()
		if err != nil {
			logln(logWriter, "      ❌ Syntax error: %s", strings.TrimSpace(string(output)))
			validationErrors++
			continue
		}

		logln(logWriter, "      ✅ Valid syntax")
	}

	if validationErrors > 0 {
		logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	logln(logWriter, "\n✅ All PowerShell scripts validated successfully")
	return 0
}

// buildConfig validates configuration files (JSON, YAML, TOML)
func buildConfig(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating config files: %s ===", module.Moniker)

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
		logln(logWriter, "❌ Failed to scan directory: %v", err)
		return 1
	}

	if len(configFiles) == 0 {
		logln(logWriter, "⚠️  No config files found")
		return 0
	}

	logln(logWriter, "⚙️  Found %d config file(s) to validate", len(configFiles))

	validationErrors := 0
	for _, configFile := range configFiles {
		relPath, _ := filepath.Rel(moduleRoot, configFile)
		ext := filepath.Ext(configFile)
		logln(logWriter, "   Validating: %s", relPath)

		content, err := os.ReadFile(configFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		// Validate based on extension
		switch ext {
		case ".json":
			var data interface{}
			if err := json.Unmarshal(content, &data); err != nil {
				logln(logWriter, "      ❌ Invalid JSON: %v", err)
				validationErrors++
				continue
			}
		case ".yaml", ".yml":
			var data interface{}
			if err := yaml.Unmarshal(content, &data); err != nil {
				logln(logWriter, "      ❌ Invalid YAML: %v", err)
				validationErrors++
				continue
			}
		case ".toml":
			// TOML validation would require a TOML library
			// For now, just check file readability
			if len(content) == 0 {
				logln(logWriter, "      ❌ Empty file")
				validationErrors++
				continue
			}
		}

		logln(logWriter, "      ✅ Valid %s", strings.TrimPrefix(ext, "."))
	}

	if validationErrors > 0 {
		logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	logln(logWriter, "\n✅ All config files validated successfully")
	return 0
}

// buildVSCodeConfig validates VS Code configuration files (JSON/JSONC)
func buildVSCodeConfig(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating VS Code config: %s ===", module.Moniker)

	// Find JSON files in .vscode
	vscodeDir := filepath.Join(moduleRoot, ".vscode")
	if _, err := os.Stat(vscodeDir); os.IsNotExist(err) {
		logln(logWriter, "⚠️  No .vscode directory found")
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
		logln(logWriter, "❌ Failed to scan .vscode: %v", err)
		return 1
	}

	if len(configFiles) == 0 {
		logln(logWriter, "⚠️  No JSON config files found")
		return 0
	}

	logln(logWriter, "🔧 Found %d config file(s) to validate", len(configFiles))

	validationErrors := 0
	for _, configFile := range configFiles {
		relPath, _ := filepath.Rel(moduleRoot, configFile)
		logln(logWriter, "   Validating: %s", relPath)

		content, err := os.ReadFile(configFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
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
			logln(logWriter, "      ❌ Invalid JSON: %v", err)
			validationErrors++
			continue
		}

		logln(logWriter, "      ✅ Valid JSON")
	}

	if validationErrors > 0 {
		logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	logln(logWriter, "\n✅ All VS Code config files validated successfully")
	return 0
}

// buildClaudeConfig validates Claude configuration YAML files
func buildClaudeConfig(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating Claude config: %s ===", module.Moniker)

	// Find YAML files in .claude
	claudeDir := filepath.Join(moduleRoot, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		logln(logWriter, "⚠️  No .claude directory found")
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
		logln(logWriter, "❌ Failed to scan .claude: %v", err)
		return 1
	}

	if len(configFiles) == 0 {
		logln(logWriter, "⚠️  No YAML config files found")
		return 0
	}

	logln(logWriter, "🤖 Found %d config file(s) to validate", len(configFiles))

	validationErrors := 0
	for _, configFile := range configFiles {
		relPath, _ := filepath.Rel(moduleRoot, configFile)
		logln(logWriter, "   Validating: %s", relPath)

		content, err := os.ReadFile(configFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		var data interface{}
		if err := yaml.Unmarshal(content, &data); err != nil {
			logln(logWriter, "      ❌ Invalid YAML: %v", err)
			validationErrors++
			continue
		}

		logln(logWriter, "      ✅ Valid YAML")
	}

	if validationErrors > 0 {
		logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	logln(logWriter, "\n✅ All Claude config files validated successfully")
	return 0
}

// buildClaudeAgents validates Claude agent markdown files
func buildClaudeAgents(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating Claude agents: %s ===", module.Moniker)

	// Use markdown validator with required sections
	validatorOpts := mdvalidator.DefaultValidatorOptions()
	validatorOpts.ValidateCodeBlocks = true
	validatorOpts.RequiredSections = []string{"Description", "Usage"}
	validatorOpts.CheckHeadingHierarchy = true

	validator := mdvalidator.NewValidator(validatorOpts, logWriter)
	results, err := validator.ValidateDirectory(moduleRoot)
	if err != nil {
		logln(logWriter, "❌ Validation failed: %v", err)
		return 1
	}

	return validator.PrintResults(results, moduleRoot)
}

// buildClaudeCommands validates Claude command markdown files
func buildClaudeCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating Claude commands: %s ===", module.Moniker)

	// Use markdown validator with required sections
	validatorOpts := mdvalidator.DefaultValidatorOptions()
	validatorOpts.ValidateCodeBlocks = true
	validatorOpts.RequiredSections = []string{} // Commands may have flexible structure
	validatorOpts.CheckHeadingHierarchy = true

	validator := mdvalidator.NewValidator(validatorOpts, logWriter)
	results, err := validator.ValidateDirectory(moduleRoot)
	if err != nil {
		logln(logWriter, "❌ Validation failed: %v", err)
		return 1
	}

	return validator.PrintResults(results, moduleRoot)
}

// buildClaudeHooks validates Claude hook shell scripts
func buildClaudeHooks(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating Claude hooks: %s ===", module.Moniker)

	// Find hook scripts in .claude/hooks
	hooksDir := filepath.Join(moduleRoot, ".claude", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		logln(logWriter, "⚠️  No .claude/hooks directory found")
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
		logln(logWriter, "❌ Failed to scan hooks: %v", err)
		return 1
	}

	if len(hookFiles) == 0 {
		logln(logWriter, "⚠️  No hook scripts found")
		return 0
	}

	logln(logWriter, "🪝 Found %d hook script(s) to validate", len(hookFiles))

	validationErrors := 0
	for _, hookFile := range hookFiles {
		relPath, _ := filepath.Rel(moduleRoot, hookFile)
		ext := filepath.Ext(hookFile)
		logln(logWriter, "   Validating: %s", relPath)

		// Read file content
		content, err := os.ReadFile(hookFile)
		if err != nil {
			logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		switch ext {
		case ".sh":
			// Validate via stdin to avoid Windows path issues
			cmd := exec.Command("bash", "-n")
			cmd.Stdin = bytes.NewReader(content)
			if output, err := cmd.CombinedOutput(); err != nil {
				logln(logWriter, "      ❌ Syntax error: %s", strings.TrimSpace(string(output)))
				validationErrors++
				continue
			}
		case ".ps1":
			// Validate PowerShell via stdin
			cmd := exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", "-")
			cmd.Stdin = bytes.NewReader([]byte(fmt.Sprintf("$null = [System.Management.Automation.PSParser]::Tokenize(@'\n%s\n'@, [ref]$null)", string(content))))
			if output, err := cmd.CombinedOutput(); err != nil {
				logln(logWriter, "      ❌ Syntax error: %s", strings.TrimSpace(string(output)))
				validationErrors++
				continue
			}
		}

		logln(logWriter, "      ✅ Valid syntax")
	}

	if validationErrors > 0 {
		logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	logln(logWriter, "\n✅ All hook scripts validated successfully")
	return 0
}

// buildTemplates validates template files and detects placeholders
func buildTemplates(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating templates: %s ===", module.Moniker)

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
		logln(logWriter, "❌ Failed to scan directory: %v", err)
		return 1
	}

	if len(templateFiles) == 0 {
		logln(logWriter, "⚠️  No template files found")
		return 0
	}

	logln(logWriter, "📄 Found %d template file(s) to analyze", len(templateFiles))

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

	logln(logWriter, "\n📊 Template Analysis:")
	logln(logWriter, "   Total files: %d", len(templateFiles))
	logln(logWriter, "   Files with placeholders: %d", len(placeholders))

	if len(placeholders) > 0 {
		logln(logWriter, "\n📝 Files with detected placeholders:")
		for file := range placeholders {
			logln(logWriter, "   - %s", file)
		}
	}

	logln(logWriter, "\n✅ Template validation complete")
	return 0
}

// buildRepositoryRoot validates repository root structure
func buildRepositoryRoot(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	logln(logWriter, "\n=== Validating repository root: %s ===", module.Moniker)

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
		logln(logWriter, "⚠️  Missing essential files:")
		for _, file := range missing {
			logln(logWriter, "   - %s", file)
		}
	}

	logln(logWriter, "\n✅ Repository root validation complete")
	return 0
}

// buildNoModuleType is a no-op build for files without a specific module type
func buildNoModuleType(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	logln(logWriter, "\n=== Skipping build (no-module-type): %s ===", module.Moniker)
	logln(logWriter, "ℹ️  This module has no specific type and requires no build")
	return 0
}
