// go.go - Build handler for Go build system
package builders

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	// Register handler for "go" build dependency
	// All go-* types use this via their build_deps contract
	// Behavior is determined by capabilities, not type names
	RegisterSystem("go", BuildGoModule)
}

// BuildGoModule builds any Go module based on its capabilities from the contract.
// - executable + cross_compile → cross-compiled binaries for multiple platforms
// - executable → single binary for current platform
// - no executable → library, just validate it compiles
func BuildGoModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)

	// Get capabilities from contract
	cfg := config.Global()
	hasExecutable := cfg != nil && cfg.ModuleTypes != nil && cfg.ModuleTypes.HasCapability(module.Type, "executable")
	hasCrossCompile := cfg != nil && cfg.ModuleTypes != nil && cfg.ModuleTypes.HasCapability(module.Type, "cross_compile")
	hasGoModule := cfg != nil && cfg.ModuleTypes != nil && cfg.ModuleTypes.HasCapability(module.Type, "go_module")

	// Skip if not a go_module (shouldn't happen if build_deps is correct, but defensive)
	if !hasGoModule {
		Logln(logWriter, "⚠️  Module type '%s' doesn't have go_module capability", module.Type)
		return 0
	}

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		Logln(logWriter, "Running: go mod tidy")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
			Logln(logWriter, "❌ go mod tidy failed")
			return exitCode
		}
	}

	// Step 2: go generate
	Logln(logWriter, "Running: go generate ./...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "generate", "./..."); exitCode != 0 {
		Logln(logWriter, "❌ go generate failed")
		return exitCode
	}

	// Step 3: Build based on capabilities
	if hasExecutable && hasCrossCompile {
		return buildCrossCompiled(module, moduleRoot, outputDir, logWriter, opts)
	} else if hasExecutable {
		return buildSingleBinary(module, moduleRoot, outputDir, logWriter, opts)
	} else {
		// Library - validate it compiles and write marker
		Logln(logWriter, "Running: go build ./...")
		if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "build", "./..."); exitCode != 0 {
			Logln(logWriter, "❌ go build failed")
			return exitCode
		}

		// Write build-complete marker for dependency verification
		if err := WriteBuildMarker(outputDir); err != nil {
			Logln(logWriter, "⚠️  Could not write build marker: %v", err)
		}

		Logln(logWriter, "✅ Library module built successfully")
		return 0
	}
}

// buildSingleBinary builds a single binary for the current platform
func buildSingleBinary(module *modules.ModuleContract, moduleRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	// Use "commands" as the base name for go-commands type, otherwise use moniker
	binaryName := module.Moniker
	if module.Type == "go-commands" {
		binaryName = "commands"
	}

	// Add platform-specific extension
	if strings.Contains(strings.ToLower(os.Getenv("GOOS")), "windows") ||
		(os.Getenv("GOOS") == "" && filepath.Separator == '\\') {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(outputDir, binaryName)

	Logln(logWriter, "Running: go build -o %s", binaryPath)
	exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "build", "-o", binaryPath)
	if exitCode == 0 {
		Logln(logWriter, "✅ Built executable: %s", binaryName)
	}
	return exitCode
}

// isDockerInDocker detects if we're running inside a Docker container.
// Uses multiple signals for robust detection:
// 1. R2R_DOCKER_MODE env var (set by r2r CLI when launching containers)
// 2. R2R_HOST_REPOROOT with Windows path while running on Linux (indicates container)
// 3. /.dockerenv file exists (standard Docker container indicator)
func isDockerInDocker() bool {
	// Primary check: explicit env var
	if os.Getenv("R2R_DOCKER_MODE") == "true" {
		return true
	}

	// Fallback: R2R_HOST_REPOROOT is set with Windows path but we're on Linux
	// This catches old r2r CLI binaries that don't set R2R_DOCKER_MODE
	hostRoot := os.Getenv("R2R_HOST_REPOROOT")
	if hostRoot != "" && runtime.GOOS == "linux" {
		// Windows paths have backslashes or drive letters
		if strings.Contains(hostRoot, "\\") || (len(hostRoot) >= 2 && hostRoot[1] == ':') {
			return true
		}
	}

	// Final fallback: check for Docker container indicator file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	return false
}

// getHostPlatform returns the host's GOOS and GOARCH when running in a container.
// Uses R2R_HOST_GOOS and R2R_HOST_GOARCH env vars set by r2r CLI.
func getHostPlatform() (goos, goarch string) {
	return os.Getenv("R2R_HOST_GOOS"), os.Getenv("R2R_HOST_GOARCH")
}

// buildCrossCompiled builds binaries for multiple platforms
// In CI mode, builds all configured targets. In local dev mode, builds only for target platform.
// In Docker-in-Docker mode, cross-compiles for the detected host platform.
func buildCrossCompiled(module *modules.ModuleContract, moduleRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	binaryName := module.Moniker

	// Detect if running in CI
	isCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"

	// Get target platforms from handlers config
	cfg := config.Global()
	var configTargets []config.CrossCompileTarget
	if cfg != nil && cfg.Handlers != nil {
		configTargets = cfg.Handlers.GetCrossCompileTargets()
	} else {
		// Fallback to defaults
		configTargets = []config.CrossCompileTarget{
			{OS: "linux", Arch: "amd64", Suffix: ""},
			{OS: "linux", Arch: "arm64", Suffix: ""},
			{OS: "darwin", Arch: "amd64", Suffix: ""},
			{OS: "darwin", Arch: "arm64", Suffix: ""},
			{OS: "windows", Arch: "amd64", Suffix: ".exe"},
		}
	}

	// In local dev mode, build only what's needed
	if !isCI {
		// Determine host platform
		hostOS, hostArch := getHostPlatform()
		if hostOS == "" {
			hostOS = runtime.GOOS
		}
		if hostArch == "" {
			hostArch = runtime.GOARCH
		}

		var filteredTargets []config.CrossCompileTarget

		// Always build Linux for host architecture (works in container)
		for _, t := range configTargets {
			if t.OS == "linux" && t.Arch == hostArch {
				filteredTargets = append(filteredTargets, t)
				break
			}
		}

		// Also build host's native platform if not Linux
		if hostOS != "linux" {
			for _, t := range configTargets {
				if t.OS == hostOS && t.Arch == hostArch {
					filteredTargets = append(filteredTargets, t)
					break
				}
			}
		}

		if len(filteredTargets) > 0 {
			var names []string
			for _, t := range filteredTargets {
				names = append(names, t.OS+"/"+t.Arch)
			}
			Logln(logWriter, "Local dev mode: building for %v", names)
			configTargets = filteredTargets
		}
	}

	// Convert to internal structure with metadata keys
	targets := make([]struct {
		goos        string
		goarch      string
		suffix      string
		metadataKey string
	}, len(configTargets))
	for i, t := range configTargets {
		targets[i] = struct {
			goos        string
			goarch      string
			suffix      string
			metadataKey string
		}{
			goos:        t.OS,
			goarch:      t.Arch,
			suffix:      t.Suffix,
			metadataKey: fmt.Sprintf("exe-%s-%s", t.OS, t.Arch),
		}
	}

	// Build ldflags
	ldflags := ""
	if opts.Compressed {
		ldflags = "-s -w"
	}
	if opts.Version != "" {
		versionFlag := fmt.Sprintf("-X main.version=%s", opts.Version)
		if ldflags != "" {
			ldflags += " " + versionFlag
		} else {
			ldflags = versionFlag
		}
	}

	successCount := 0
	for _, target := range targets {
		// Use custom name from metadata if available, otherwise use default pattern
		outputName := fmt.Sprintf("%s-%s-%s%s", binaryName, target.goos, target.goarch, target.suffix)
		if customName, ok := module.Metadata[target.metadataKey]; ok && customName != "" {
			outputName = customName
		}
		outputPath := filepath.Join(outputDir, outputName)

		Logln(logWriter, "Building: %s/%s → %s", target.goos, target.goarch, outputName)

		args := []string{"build", "-o", outputPath}
		if ldflags != "" {
			args = append(args, "-ldflags", ldflags)
		}

		cmd := exec.Command("go", args...)
		cmd.Dir = moduleRoot
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOOS=%s", target.goos),
			fmt.Sprintf("GOARCH=%s", target.goarch),
			"CGO_ENABLED=0",
		)
		cmd.Stdout = logWriter
		cmd.Stderr = logWriter

		if err := cmd.Run(); err != nil {
			Logln(logWriter, "❌ Failed to build %s/%s: %v", target.goos, target.goarch, err)
			continue
		}

		successCount++

		// Apply UPX compression if requested and available
		// Creates a separate -upx suffixed binary, keeping the original intact
		// Check if platform supports UPX from handlers config
		upxSupported := false
		if cfg != nil && cfg.Handlers != nil {
			upxSupported = cfg.Handlers.IsUPXSupported(target.goos)
		} else {
			// Default: linux and windows support UPX
			upxSupported = target.goos == "linux" || target.goos == "windows"
		}
		if opts.CompressedUPX && upxSupported {
			if upxPath, err := exec.LookPath("upx"); err == nil {
				// Create UPX output name by inserting -upx before the extension
				var upxOutputName string
				if target.suffix == ".exe" {
					upxOutputName = strings.TrimSuffix(outputName, ".exe") + "-upx.exe"
				} else {
					upxOutputName = outputName + "-upx"
				}
				upxOutputPath := filepath.Join(outputDir, upxOutputName)

				// Copy the binary to the UPX target
				Logln(logWriter, "Creating UPX copy: %s", upxOutputName)
				srcFile, err := os.Open(outputPath)
				if err != nil {
					Logln(logWriter, "⚠️  Failed to open source for UPX: %v", err)
				} else {
					dstFile, err := os.Create(upxOutputPath)
					if err != nil {
						srcFile.Close()
						Logln(logWriter, "⚠️  Failed to create UPX target: %v", err)
					} else {
						_, copyErr := io.Copy(dstFile, srcFile)
						srcFile.Close()
						dstFile.Close()
						if copyErr != nil {
							Logln(logWriter, "⚠️  Failed to copy for UPX: %v", copyErr)
						} else {
							// Make executable
							os.Chmod(upxOutputPath, 0755)

							// Compress the copy
							Logln(logWriter, "Compressing with UPX: %s", upxOutputName)
							upxCmd := exec.Command(upxPath, "--best", "--lzma", upxOutputPath)
							upxCmd.Stdout = logWriter
							upxCmd.Stderr = logWriter
							if err := upxCmd.Run(); err != nil {
								Logln(logWriter, "⚠️  UPX compression failed: %v", err)
								os.Remove(upxOutputPath) // Clean up failed UPX file
							}
						}
					}
				}
			}
		}
	}

	if successCount == 0 {
		Logln(logWriter, "❌ All builds failed")
		return 1
	}

	Logln(logWriter, "✅ Built %d/%d targets successfully", successCount, len(targets))

	// Generate checksums
	generateChecksums(outputDir, binaryName, logWriter)

	return 0
}

// generateChecksums creates SHA256 checksums for all built binaries
// It checksums all files that look like executables (no extension or .exe)
// excluding known non-binary files like .txt and .log
func generateChecksums(outputDir string, binaryName string, logWriter io.Writer) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return
	}

	checksumFile := filepath.Join(outputDir, "checksums.txt")
	f, err := os.Create(checksumFile)
	if err != nil {
		Logln(logWriter, "⚠️  Could not create checksums file: %v", err)
		return
	}
	defer f.Close()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip known non-binary files
		if strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".log") {
			continue
		}
		// Include executables (no extension for Unix, .exe for Windows)
		ext := filepath.Ext(name)
		if ext != "" && ext != ".exe" {
			continue
		}

		filePath := filepath.Join(outputDir, name)
		hash, err := computeSHA256(filePath)
		if err != nil {
			continue
		}

		fmt.Fprintf(f, "%s  %s\n", hash, name)
	}

	Logln(logWriter, "✅ Generated checksums.txt")
}

// computeSHA256 computes the SHA256 hash of a file
func computeSHA256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h), nil
}
