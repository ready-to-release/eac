// go.go - Build functions for Go module types
package builders

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// BuildGoDefault is the default build handler for Go modules without specific handlers.
// Runs go generate and validates the module compiles.
func BuildGoDefault(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)
	Logln(logWriter, "\n=== Building Go module: %s (type: %s) ===", module.Moniker, module.Type)

	if opts.TidyFirst {
		Logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	Logln(logWriter, "Running: go generate ./...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	Logln(logWriter, "Running: go build ./...")
	return RunCommandWithLog(moduleRoot, logWriter, "go", "build", "./...")
}

// BuildGoCLI builds a CLI binary with cross-platform support
func BuildGoCLI(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== Building go-cli: %s ===", module.Moniker)

	// Log compression mode
	if opts.CompressedUPX {
		Logln(logWriter, "Compression: UPX (--compressed-upx)")
	} else if opts.Compressed {
		Logln(logWriter, "Compression: stripped (--compressed)")
	} else {
		Logln(logWriter, "Compression: none (dev build with debug info)")
	}

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		Logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	// Step 2: go generate
	Logln(logWriter, "Running: go generate ./...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	// Check for UPX if needed
	if opts.CompressedUPX {
		if _, err := exec.LookPath("upx"); err != nil {
			Logln(logWriter, "❌ UPX not found in PATH. Install UPX for --compressed-upx support.")
			Logln(logWriter, "   See: https://upx.github.io/")
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

	var builtBinaries []string

	// Build for each target platform
	for _, platform := range platforms {
		var binaryName string
		if platform.GOOS == "darwin" {
			binaryName = fmt.Sprintf("r2r-%s-%s%s", platform.GOOS, platform.GOARCH, platform.Ext)
		} else {
			binaryName = fmt.Sprintf("r2r-%s%s", platform.GOOS, platform.Ext)
		}
		binaryPath := filepath.Join(outputDir, binaryName)

		Logln(logWriter, "\n--- Building for %s (%s/%s) ---", platform.Name, platform.GOOS, platform.GOARCH)
		Logln(logWriter, "Output: %s", binaryPath)

		buildArgs := []string{"build", "-o", binaryPath}

		// Build ldflags
		var ldflags string
		if opts.Compressed || opts.CompressedUPX {
			ldflags = "-s -w"
		}
		if opts.Version != "" {
			if ldflags != "" {
				ldflags += " "
			}
			ldflags += fmt.Sprintf("-X 'github.com/ready-to-release/eac/src/cli/cmd.Version=%s'", opts.Version)
			Logln(logWriter, "Version: %s", opts.Version)
		}
		if ldflags != "" {
			buildArgs = append(buildArgs, "-ldflags", ldflags)
		}

		env := []string{
			fmt.Sprintf("GOOS=%s", platform.GOOS),
			fmt.Sprintf("GOARCH=%s", platform.GOARCH),
		}

		if exitCode := RunCommandWithEnv(moduleRoot, logWriter, env, "go", buildArgs...); exitCode != 0 {
			Logln(logWriter, "❌ Build failed for %s", platform.Name)
			return 1
		}

		Logln(logWriter, "✅ Built successfully: %s", binaryPath)

		if runtime.GOOS != "windows" && platform.GOOS != "windows" {
			if exitCode := RunCommandWithLog(moduleRoot, logWriter, "chmod", "+x", binaryPath); exitCode != 0 {
				Logln(logWriter, "⚠️  Warning: could not set executable permissions")
			}
		}

		builtBinaries = append(builtBinaries, binaryPath)
	}

	// Apply UPX compression if requested
	if opts.CompressedUPX {
		Logln(logWriter, "\n--- Applying UPX compression ---")
		Logln(logWriter, "Note: UPX compression skipped for Darwin binaries (not supported when cross-compiling)")

		for _, binaryPath := range builtBinaries {
			baseName := filepath.Base(binaryPath)

			if strings.Contains(baseName, "darwin") {
				Logln(logWriter, "⏭️  Skipping %s (Darwin binaries not supported by UPX cross-compile)", baseName)
				continue
			}

			originalInfo, err := os.Stat(binaryPath)
			if err != nil {
				Logln(logWriter, "⚠️  Warning: could not stat %s: %v", binaryPath, err)
				continue
			}
			originalSize := originalInfo.Size()

			ext := filepath.Ext(baseName)
			nameWithoutExt := strings.TrimSuffix(baseName, ext)
			upxName := nameWithoutExt + "-upx" + ext
			upxPath := filepath.Join(outputDir, upxName)

			if err := CopyFile(binaryPath, upxPath); err != nil {
				Logln(logWriter, "❌ Failed to copy %s for UPX: %v", baseName, err)
				return 1
			}

			Logln(logWriter, "Compressing: %s -> %s", baseName, upxName)

			cmd := exec.Command("upx", "--best", "-q", upxPath)
			cmd.Stdout = logWriter
			cmd.Stderr = logWriter

			if err := cmd.Run(); err != nil {
				Logln(logWriter, "⚠️  UPX compression failed for %s: %v", upxName, err)
				os.Remove(upxPath)
				continue
			}

			compressedInfo, err := os.Stat(upxPath)
			if err != nil {
				Logln(logWriter, "⚠️  Warning: could not stat compressed file: %v", err)
				continue
			}
			compressedSize := compressedInfo.Size()

			ratio := float64(compressedSize) / float64(originalSize) * 100
			Logln(logWriter, "✅ %s: %.1f MB -> %.1f MB (%.0f%%)",
				upxName,
				float64(originalSize)/1024/1024,
				float64(compressedSize)/1024/1024,
				ratio)
		}
	}

	Logln(logWriter, "\n✅ All builds completed successfully")
	return 0
}

// BuildGoCommands builds the runtime command dispatcher
func BuildGoCommands(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== go-commands: %s ===", module.Moniker)

	if opts.TidyFirst {
		Logln(logWriter, "🔄 Running go mod tidy...")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			Logln(logWriter, "❌ go mod tidy failed")
			return exitCode
		}
		Logln(logWriter, "✅ go mod tidy completed")
	}

	Logln(logWriter, "🔄 Running go generate...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		Logln(logWriter, "❌ go generate failed")
		return exitCode
	}
	Logln(logWriter, "✅ go generate completed")

	Logln(logWriter, "\nℹ️  This module uses 'go run .' and is never compiled to a binary")
	Logln(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)")
	return 0
}

// BuildGoMCP builds an MCP JSON-RPC server binary
func BuildGoMCP(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	serverName := module.Moniker
	if len(serverName) > 8 && serverName[:8] == "src-mcp-" {
		serverName = serverName[8:]
	}

	binaryName := fmt.Sprintf("mcp-server-%s", serverName)
	binaryPath := filepath.Join(outputDir, binaryName)

	Logln(logWriter, "\n=== Building go-mcp: %s ===", module.Moniker)

	if opts.TidyFirst {
		Logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	Logln(logWriter, "Running: go build -o %s", binaryPath)
	return RunCommandWithLog(moduleRoot, logWriter, "go", "build", "-o", binaryPath)
}

// BuildGoLibrary builds a Go library module
func BuildGoLibrary(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== go-library: %s ===", module.Moniker)

	if opts.TidyFirst {
		Logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			return exitCode
		}
	}

	Logln(logWriter, "Running: go generate ./...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		return exitCode
	}

	Logln(logWriter, "ℹ️  This is a library module (no binary to build)")
	Logln(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)")
	return 0
}

// BuildGoTests builds a Go test module
func BuildGoTests(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== go-tests: %s ===", module.Moniker)
	Logln(logWriter, "ℹ️  This is a test module (use 'test module' command to run tests)")
	Logln(logWriter, "ℹ️  Auto-built during testing (no explicit build needed)")
	return 0
}
