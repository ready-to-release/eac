// go.go - Build handler for Go build system
package builders

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&GoHandler{})
}

// GoHandler builds Go modules (libraries, CLIs, tests).
type GoHandler struct{}

// goToolOptions controls which tool (system go or container go-oci) to use.
type goToolOptions struct {
	// UseContainer forces use of container tool (go-oci) instead of system go.
	UseContainer bool

	// RaceDetection enables race detector (requires CGO, uses container).
	RaceDetection bool

	// CGOEnabled forces CGO support (uses container for reliable cross-platform CGO).
	CGOEnabled bool
}

// resolveToolName determines whether to use system 'go' or container 'go-oci'.
// Container is used when CGO or race detection is needed.
func (h *GoHandler) resolveToolName(opts goToolOptions) string {
	// Use container for specific scenarios
	if opts.UseContainer || opts.RaceDetection || opts.CGOEnabled {
		return "go-oci"
	}
	// Default: native go through tool system
	return "go"
}

// executeTool runs the go tool with given args and env overrides.
// Uses the tool system instead of direct exec.Command.
func (h *GoHandler) executeTool(ctx context.Context, toolName, moduleRoot, workspaceRoot string,
	logWriter io.Writer, args []string, envOverrides map[string]string) int {

	registry := tool.GlobalRegistry()
	executor := tool.GlobalExecutor()

	// Get tool definition
	toolDef, ok := registry.Get(toolName)
	if !ok {
		Logln(logWriter, "Tool not found: %s", toolName)
		return 1
	}

	// Build placeholders for mount path resolution
	placeholders := map[string]string{
		"{workspace}": workspaceRoot,
		"{module}":    moduleRoot,
	}

	// Build execution context
	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		LogWriter:     logWriter,
		Operation:     core.ActionBuild,
		ArgsOverrides: args,
		EnvOverrides:  envOverrides,
		Placeholders:  placeholders,
	}

	result, err := executor.Execute(ctx, toolDef, execCtx)
	if err != nil {
		Logln(logWriter, "Tool execution error: %v", err)
		return 1
	}

	return result.ExitCode
}

// executeGoTool is a package-level helper that executes go commands via the tool system.
// It uses the default "go" system tool unless specific scenarios (CGO, race) require "go-oci".
// When the module is not listed in go.work, GOWORK=off is set automatically so that
// ./... pattern resolution works within the standalone module.
func executeGoTool(moduleRoot, workspaceRoot string, logWriter io.Writer, args []string, envOverrides map[string]string) int {
	if !isModuleInGoWork(moduleRoot, workspaceRoot) {
		if envOverrides == nil {
			envOverrides = map[string]string{}
		}
		if _, ok := envOverrides["GOWORK"]; !ok {
			envOverrides["GOWORK"] = "off"
		}
	}
	h := &GoHandler{}
	return h.executeTool(context.Background(), "go", moduleRoot, workspaceRoot, logWriter, args, envOverrides)
}

// isModuleInGoWork checks whether moduleRoot is listed in the workspace go.work file.
func isModuleInGoWork(moduleRoot, workspaceRoot string) bool {
	goWorkPath := filepath.Join(workspaceRoot, "go.work")
	data, err := os.ReadFile(goWorkPath)
	if err != nil {
		return false // No go.work = standalone mode, no issue
	}

	// Normalize moduleRoot for comparison
	absModule, err := filepath.Abs(moduleRoot)
	if err != nil {
		return false
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "./") || strings.HasPrefix(line, ".\\") {
			usePath := filepath.Join(workspaceRoot, filepath.FromSlash(line))
			absUse, err := filepath.Abs(usePath)
			if err != nil {
				continue
			}
			if strings.EqualFold(absModule, absUse) {
				return true
			}
		}
	}
	return false
}

// executeGoToolWithEnv is like executeGoTool but for cross-compilation with GOOS/GOARCH.
func executeGoToolWithEnv(moduleRoot, workspaceRoot string, logWriter io.Writer, args []string, goos, goarch string) int {
	envOverrides := map[string]string{
		"GOOS":        goos,
		"GOARCH":      goarch,
		"CGO_ENABLED": "0",
	}
	return executeGoTool(moduleRoot, workspaceRoot, logWriter, args, envOverrides)
}

func (h *GoHandler) Name() string { return "go" }

func (h *GoHandler) Capabilities() []string { return []string{"go_module", "cross_compile"} }

func (h *GoHandler) Requirements() []string { return []string{"go"} }

func (h *GoHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	componentRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(component))
	// Walk up from component root to find go.mod (may be in parent directory)
	goMod := findGoMod(componentRoot, workspaceRoot)
	if goMod == "" {
		return fmt.Errorf("go.mod not found for component %s (searched from %s)", component, componentRoot)
	}
	return nil
}

// IsContainer returns false as Go builds run using the local go toolchain.
func (h *GoHandler) IsContainer() bool { return false }

// IsHostInstalled returns true as Go builds use the local go toolchain.
func (h *GoHandler) IsHostInstalled() bool { return true }

// findGoMod walks up from dir to find go.mod, stopping at workspaceRoot.
func findGoMod(dir, workspaceRoot string) string {
	for {
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return goMod
		}
		parent := filepath.Dir(dir)
		if parent == dir || len(dir) < len(workspaceRoot) {
			return ""
		}
		dir = parent
	}
}

func (h *GoHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		return nil
	}
	return listGoModuleArtifacts(concrete, workspaceRoot)
}

func (h *GoHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}
	return buildGoModule(concrete, workspaceRoot, outputDir, logWriter, opts)
}

// listGoModuleArtifacts returns the artifacts that would be produced by building this Go module.
func listGoModuleArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	hasGoModule := module.HasComponent("go")

	if !hasGoModule {
		return nil
	}

	// Check for per-module artifact definitions
	if module.HasBuildArtifacts() {
		return listModuleArtifacts(module)
	}

	// No artifacts = library (compile-only verification)
	// The UoW manifest itself proves the build succeeded (exit_code, timestamp, hash)
	return nil
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
	// Use the actual component name from build options (supports bundle sub-components like "ai", "docker")
	// Falls back to "go" for standard single-component modules
	componentName := opts.Component
	if componentName == "" {
		componentName = "go"
	}
	moduleRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(componentName))

	Logln(logWriter, "\n=== Building go: %s ===", module.Moniker)

	// Check if module has this go component
	hasGoModule := module.HasComponent(componentName)

	// Skip if not a go module - this is expected for script-only modules (pwsh, bash)
	// that have test-impl components triggering Go builds
	if !hasGoModule {
		if !isScriptOnlyModule(module) {
			Logln(logWriter, "  Skipping: module '%s' has no Go package", module.Moniker)
		}
		return 0
	}

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		Logln(logWriter, "Running: go mod tidy")
		if exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"mod", "tidy"}, nil); exitCode != 0 {
			Logln(logWriter, "❌ go mod tidy failed")
			return exitCode
		}
	}

	// Step 2: go generate
	Logln(logWriter, "Running: go generate ./...")
	if exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"generate", "./..."}, nil); exitCode != 0 {
		Logln(logWriter, "❌ go generate failed")
		return exitCode
	}

	// Step 3: Build based on per-module artifact definitions
	if !module.HasBuildArtifacts() {
		// No artifacts = library (compile-only verification)
		return buildLibrary(module, moduleRoot, workspaceRoot, outputDir, logWriter)
	}

	// Check artifact types
	hasExecutables := module.HasExecutableArtifacts()
	hasTests := module.HasTestArtifacts()

	if hasTests {
		// Test module - run tests and capture results
		return buildTestModule(module, moduleRoot, workspaceRoot, outputDir, logWriter, opts)
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
	return buildLibrary(module, moduleRoot, workspaceRoot, outputDir, logWriter)
}

// buildLibrary builds a library module (compile-only verification).
func buildLibrary(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer) int {
	Logln(logWriter, "Running: go build ./...")
	if exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"build", "./..."}, nil); exitCode != 0 {
		Logln(logWriter, "❌ go build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Library module built successfully")
	return 0
}

// buildTestModule runs tests and captures results.
func buildTestModule(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "Running: go test ./... -json")

	// First verify it compiles
	if exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"build", "./..."}, nil); exitCode != 0 {
		Logln(logWriter, "❌ go build failed")
		return exitCode
	}

	// Run tests with JSON output for results capture
	// StdoutWriter redirects test JSON output to results file,
	// StderrWriter sends diagnostic output to logWriter
	resultsPath := filepath.Join(outputDir, "results.json")
	resultsFile, err := os.Create(resultsPath)
	if err != nil {
		Logln(logWriter, "❌ Failed to create results file: %v", err)
		return 1
	}
	defer resultsFile.Close()

	registry := tool.GlobalRegistry()
	executor := tool.GlobalExecutor()
	toolDef := registry.GetOrAdhoc("go")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		LogWriter:     logWriter,
		StdoutWriter:  resultsFile,
		StderrWriter:  logWriter,
		Operation:     core.ActionBuild,
		ArgsOverrides: []string{"test", "./...", "-json"},
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    moduleRoot,
		},
	}

	result, err := executor.Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		// Test failures are expected - capture the results anyway
		Logln(logWriter, "⚠️  Tests completed with failures")
	} else if result.ExitCode != 0 {
		Logln(logWriter, "⚠️  Tests completed with failures")
	}

	Logln(logWriter, "✅ Test module built and results captured")
	return 0
}

// buildSingleBinaryFromArtifact builds a single binary from a per-module artifact definition.
func buildSingleBinaryFromArtifact(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions, artifact domain.ModuleArtifact) int {
	// Resolve artifact pattern to binary name
	resolver := config.NewArtifactResolverWithPlatform(module.Moniker, "", runtime.GOOS, runtime.GOARCH)
	binaryName := resolver.ResolvePattern(artifact.Pattern)
	binaryPath := filepath.Join(outputDir, binaryName)

	Logln(logWriter, "Running: go build -o %s", binaryPath)
	exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"build", "-o", binaryPath}, nil)
	if exitCode == 0 {
		Logln(logWriter, "✅ Built executable: %s", binaryName)
	}
	return exitCode
}

// buildCrossCompiledFromArtifacts builds binaries for multiple platforms from per-module artifact definitions.
func buildCrossCompiledFromArtifacts(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions, artifacts []domain.ModuleArtifact) int {
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

	// Filter targets based on artifacts mode (reduced vs all)
	if opts.ArtifactsMode.IsValid() {
		var filteredTargets []buildTarget
		for _, target := range targets {
			if opts.ArtifactsMode.ShouldBuildTarget(target.goos, target.goarch, target.compression) {
				filteredTargets = append(filteredTargets, target)
			}
		}
		if len(filteredTargets) == 0 {
			Logln(logWriter, "⚠️  No targets match artifacts mode '%s' - skipping build", opts.ArtifactsMode)
			return 0
		}
		if len(filteredTargets) < len(targets) {
			Logln(logWriter, "📦 Artifacts mode '%s': building %d/%d targets", opts.ArtifactsMode, len(filteredTargets), len(targets))
		}
		targets = filteredTargets
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

		// Use tool system with env overrides for cross-compilation
		exitCode := executeGoToolWithEnv(moduleRoot, workspaceRoot, logWriter, args, target.goos, target.goarch)
		if exitCode != 0 {
			Logln(logWriter, "❌ Failed to build %s/%s", target.goos, target.goarch)
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
	// NOTE: config.RepositoryConfig.ReleaseModulePathAbs() is now available but this
	// function doesn't have access to config. Uses the same conventional "release/" path.
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

// isScriptOnlyModule returns true if module only has script components (pwsh/bash).
// These modules are expected to not have Go packages.
func isScriptOnlyModule(module *modules.ModuleContract) bool {
	return module.HasComponent("pwsh") || module.HasComponent("bash")
}
