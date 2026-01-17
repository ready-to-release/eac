// go.go - Build handler for Go build system
package builders

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	RegisterHandler(&GoHandler{})
}

// GoHandler builds Go modules (libraries, CLIs, tests).
type GoHandler struct{}

func (h *GoHandler) Name() string { return "go" }

func (h *GoHandler) Capabilities() []string { return []string{"go_module", "cross_compile"} }

func (h *GoHandler) Requirements() []string { return []string{"go"} }

func (h *GoHandler) ValidateModule(module *modules.ModuleContract, workspaceRoot string) error {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)
	goMod := filepath.Join(moduleRoot, "go.mod")
	if _, err := os.Stat(goMod); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found at %s", goMod)
	}
	return nil
}

func (h *GoHandler) ListArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	return listGoModuleArtifacts(module, workspaceRoot)
}

func (h *GoHandler) Build(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	return buildGoModule(module, workspaceRoot, outputDir, logWriter, opts)
}

// listGoModuleArtifacts returns the artifacts that would be produced by building this Go module.
func listGoModuleArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	cfg := config.Global()
	hasGoModule := cfg != nil && cfg.ModuleTypes != nil && cfg.ModuleTypes.HasCapability(module.Type, "go_module")

	if !hasGoModule {
		return nil
	}

	// Check for per-module artifact definitions
	if module.HasBuildArtifacts() {
		return listModuleArtifacts(module)
	}

	// No artifacts = library (compile-only verification)
	return []string{".build-complete"}
}

// listModuleArtifacts returns artifacts based on per-module definitions.
func listModuleArtifacts(module *modules.ModuleContract) []string {
	var artifacts []string

	for _, artifact := range module.GetBuildArtifacts() {
		switch artifact.Type {
		case "executable":
			// Resolve pattern with current platform
			resolver := config.NewArtifactResolverWithPlatform(module.Moniker, "", runtime.GOOS, runtime.GOARCH)
			name := resolver.ResolvePattern(artifact.Pattern)
			artifacts = append(artifacts, name)
		case "test":
			// Test artifact
			artifacts = append(artifacts, artifact.Pattern)
		default:
			// Other artifact types - resolve as-is
			resolver := config.NewArtifactResolver(module.Moniker, "")
			name := resolver.ResolvePattern(artifact.Pattern)
			artifacts = append(artifacts, name)
		}
	}

	// Add checksums for cross-platform builds (multiple executables)
	execCount := 0
	for _, a := range module.GetBuildArtifacts() {
		if a.Type == "executable" {
			execCount++
		}
	}
	if execCount > 1 {
		artifacts = append(artifacts, "checksums.txt")
	}

	return artifacts
}

// buildGoModule builds any Go module based on per-module artifact definitions.
// Behavior is driven by artifacts defined in repository.yml:
//   - No artifacts: library (compile-only verification)
//   - Single executable: builds binary for current platform
//   - Multiple executables: cross-compiled binaries
//   - Test artifacts: runs tests and captures results
func buildGoModule(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)

	// Get capabilities from contract
	cfg := config.Global()
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

	// Step 3: Build based on per-module artifact definitions
	if !module.HasBuildArtifacts() {
		// No artifacts = library (compile-only verification)
		return buildLibrary(module, moduleRoot, outputDir, logWriter)
	}

	// Check artifact types
	hasExecutables := module.HasExecutableArtifacts()
	hasTests := module.HasTestArtifacts()

	if hasTests {
		// Test module - run tests and capture results
		return buildTestModule(module, moduleRoot, outputDir, logWriter, opts)
	}

	if hasExecutables {
		execArtifacts := module.GetArtifactsByType("executable")
		if len(execArtifacts) == 1 {
			// Single executable - build for current platform
			return buildSingleBinaryFromArtifact(module, moduleRoot, workspaceRoot, outputDir, logWriter, opts, execArtifacts[0])
		}
		// Multiple executables - cross-compile
		return buildCrossCompiledFromArtifacts(module, moduleRoot, workspaceRoot, outputDir, logWriter, opts, execArtifacts)
	}

	// Fallback: library behavior
	return buildLibrary(module, moduleRoot, outputDir, logWriter)
}

// buildLibrary builds a library module (compile-only verification).
func buildLibrary(module *modules.ModuleContract, moduleRoot, outputDir string, logWriter io.Writer) int {
	Logln(logWriter, "Running: go build ./...")
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "build", "./..."); exitCode != 0 {
		Logln(logWriter, "❌ go build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Library module built successfully")
	return 0
}

// buildTestModule runs tests and captures results.
func buildTestModule(module *modules.ModuleContract, moduleRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "Running: go test ./... -json")

	// First verify it compiles
	if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "build", "./..."); exitCode != 0 {
		Logln(logWriter, "❌ go build failed")
		return exitCode
	}

	// Run tests with JSON output for results capture
	resultsPath := filepath.Join(outputDir, "results.json")
	resultsFile, err := os.Create(resultsPath)
	if err != nil {
		Logln(logWriter, "❌ Failed to create results file: %v", err)
		return 1
	}
	defer resultsFile.Close()

	cmd := exec.Command("go", "test", "./...", "-json")
	cmd.Dir = moduleRoot
	cmd.Stdout = resultsFile
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		// Test failures are expected - capture the results anyway
		Logln(logWriter, "⚠️  Tests completed with failures")
	}

	Logln(logWriter, "✅ Test module built and results captured")
	return 0
}

// buildSingleBinaryFromArtifact builds a single binary from a per-module artifact definition.
func buildSingleBinaryFromArtifact(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions, artifact contracts.ModuleArtifact) int {
	// Resolve artifact pattern to binary name
	resolver := config.NewArtifactResolverWithPlatform(module.Moniker, "", runtime.GOOS, runtime.GOARCH)
	binaryName := resolver.ResolvePattern(artifact.Pattern)
	binaryPath := filepath.Join(outputDir, binaryName)

	Logln(logWriter, "Running: go build -o %s", binaryPath)
	exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "build", "-o", binaryPath)
	if exitCode == 0 {
		Logln(logWriter, "✅ Built executable: %s", binaryName)
	}
	return exitCode
}

// buildCrossCompiledFromArtifacts builds binaries for multiple platforms from per-module artifact definitions.
func buildCrossCompiledFromArtifacts(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions, artifacts []contracts.ModuleArtifact) int {
	// Extract platform targets from artifact IDs
	type buildTarget struct {
		goos        string
		goarch      string
		pattern     string
		compression string
		deriveFrom  string
	}

	var targets []buildTarget
	for _, artifact := range artifacts {
		// Parse platform from artifact ID (e.g., "linux-amd64", "windows-amd64-upx")
		parts := strings.Split(artifact.ID, "-")
		if len(parts) < 2 {
			continue
		}
		goos := parts[0]
		goarch := parts[1]

		targets = append(targets, buildTarget{
			goos:        goos,
			goarch:      goarch,
			pattern:     artifact.Pattern,
			compression: artifact.Compression,
			deriveFrom:  artifact.DeriveFrom,
		})
	}

	if len(targets) == 0 {
		Logln(logWriter, "❌ No valid platform targets found in artifacts")
		return 1
	}

	// Build ldflags for version injection
	ldflags := buildLdflags(module, moduleRoot, workspaceRoot, opts.Version, logWriter)

	successCount := 0
	for _, target := range targets {
		// Skip derived artifacts (UPX) - they're processed after base builds
		if target.deriveFrom != "" {
			continue
		}

		resolver := config.NewArtifactResolverWithPlatform(module.Moniker, "", target.goos, target.goarch)
		outputName := resolver.ResolvePattern(target.pattern)
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
	}

	if successCount == 0 {
		Logln(logWriter, "❌ All builds failed")
		return 1
	}

	Logln(logWriter, "✅ Built %d/%d targets successfully", successCount, len(targets))

	// Generate checksums
	generateChecksums(outputDir, module.Moniker, logWriter)

	return 0
}

// buildLdflags builds ldflags for version injection.
func buildLdflags(module *modules.ModuleContract, moduleRoot, workspaceRoot, explicitVersion string, logWriter io.Writer) string {
	ldflags := ""

	// Detect if running in CI (for version detection)
	isCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"

	// Determine version to inject
	version := explicitVersion
	if version == "" {
		if isCI {
			// CI: Auto-detect from changelog for release builds
			version = getVersionFromChangelog(moduleRoot, workspaceRoot, module.Moniker)
		} else {
			// Local dev: Use high version number to always be "newer" than releases
			version = "666.666.666-local"
		}
	}

	// Inject version if available
	if version != "" {
		// Get module import path for correct ldflags
		modulePath := getGoModulePath(moduleRoot)
		if modulePath != "" {
			// Use module/cmd.Version pattern (standard for CLI tools)
			versionFlag := fmt.Sprintf("-X %s/cmd.Version=%s", modulePath, version)
			ldflags = versionFlag
			Logln(logWriter, "Injecting version: %s", version)
		}
	}

	return ldflags
}

// buildSingleBinary builds a single binary for the current platform.
func buildSingleBinary(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	binaryName := module.Moniker

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

// getHostPlatform returns the host's GOOS and GOARCH when running in a container.
// Uses R2R_HOST_GOOS and R2R_HOST_GOARCH env vars set by r2r CLI.
func getHostPlatform() (goos, goarch string) {
	return os.Getenv("R2R_HOST_GOOS"), os.Getenv("R2R_HOST_GOARCH")
}

// CrossCompileTarget represents a cross-compilation target platform.
type CrossCompileTarget struct {
	OS     string
	Arch   string
	Suffix string
}

// Default cross-compile targets for Go executables.
var defaultCrossCompileTargets = []CrossCompileTarget{
	{"linux", "amd64", ""},
	{"linux", "arm64", ""},
	{"darwin", "amd64", ""},
	{"darwin", "arm64", ""},
	{"windows", "amd64", ".exe"},
}

// buildCrossCompiled builds binaries for multiple platforms
// With --all flag, builds all configured targets. In default mode, builds only current platform.
func buildCrossCompiled(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	binaryName := module.Moniker

	// Use default cross-compile targets
	configTargets := defaultCrossCompileTargets

	// Filter targets based on requested artifacts
	if len(opts.RequestedArtifacts) > 0 {
		var filteredTargets []CrossCompileTarget
		for _, target := range configTargets {
			// Artifact ID format for executables: {os}-{arch}
			artifactID := fmt.Sprintf("%s-%s", target.OS, target.Arch)
			// Check if this artifact was requested
			for _, requestedID := range opts.RequestedArtifacts {
				if requestedID == artifactID {
					filteredTargets = append(filteredTargets, target)
					break
				}
			}
		}
		if len(filteredTargets) == 0 {
			Logln(logWriter, "No platforms match requested artifacts - skipping build")
			return 0
		}
		configTargets = filteredTargets
		var platformNames []string
		for _, t := range configTargets {
			platformNames = append(platformNames, fmt.Sprintf("%s/%s", t.OS, t.Arch))
		}
		Logln(logWriter, "Building requested platforms: %v", platformNames)
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
			metadataKey: fmt.Sprintf("executable-%s-%s", t.OS, t.Arch),
		}
	}

	// Build ldflags for version injection
	ldflags := ""

	// Detect if running in CI (for version detection)
	isCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"

	// Determine version to inject
	version := opts.Version
	if version == "" {
		if isCI {
			// CI: Auto-detect from changelog for release builds
			version = getVersionFromChangelog(moduleRoot, workspaceRoot, module.Moniker)
		} else {
			// Local dev: Use high version number to always be "newer" than releases
			// This makes it clear the binary is a local non-production build
			version = "666.666.666-local"
		}
	}

	// Inject version if available
	if version != "" {
		// Get module import path for correct ldflags
		modulePath := getGoModulePath(moduleRoot)
		if modulePath != "" {
			// Use module/cmd.Version pattern (standard for CLI tools)
			versionFlag := fmt.Sprintf("-X %s/cmd.Version=%s", modulePath, version)
			if ldflags != "" {
				ldflags += " " + versionFlag
			} else {
				ldflags = versionFlag
			}
			Logln(logWriter, "Injecting version: %s", version)
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

// getGoModulePath reads the module path from go.mod file.
func getGoModulePath(moduleRoot string) string {
	goModPath := filepath.Join(moduleRoot, "go.mod")
	f, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}
	return ""
}

// getVersionFromChangelog attempts to extract version from CHANGELOG.md
// Checks both module root and release/{moniker}/ directories.
func getVersionFromChangelog(moduleRoot, workspaceRoot, moniker string) string {
	// Try multiple locations for changelog
	// TODO(path-migration): Add cfg.Repository.ReleasePathAbs() method to config
	// For now using hardcoded "release" until release paths are added to repository.schema.json
	paths := []string{
		filepath.Join(moduleRoot, "CHANGELOG.md"),
		filepath.Join(workspaceRoot, "release", moniker, "CHANGELOG.md"),
	}

	for _, changelogPath := range paths {
		if version := extractVersionFromFile(changelogPath); version != "" {
			return version
		}
	}
	return ""
}

// extractVersionFromFile extracts the first version from a changelog file.
func extractVersionFromFile(changelogPath string) string {
	f, err := os.Open(changelogPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Look for version headers like "## [0.0.17]" or "## 0.0.17"
		if strings.HasPrefix(line, "## ") {
			version := strings.TrimPrefix(line, "## ")
			// Remove brackets if present: [0.0.17] -> 0.0.17
			version = strings.TrimPrefix(version, "[")
			version = strings.Split(version, "]")[0]
			// Skip [Unreleased]
			if version != "Unreleased" && version != "" {
				return version
			}
		}
	}
	return ""
}

// generateChecksums creates SHA256 checksums for all built binaries
// It checksums all files that look like executables (no extension or .exe)
// excluding known non-binary files like .txt and .log.
func generateChecksums(outputDir, binaryName string, logWriter io.Writer) {
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

// computeSHA256 computes the SHA256 hash of a file.
func computeSHA256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h), nil
}
