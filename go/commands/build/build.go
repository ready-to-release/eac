package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/commands/base"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

var log = logging.C()

// writeln writes a formatted message followed by a newline to the writer.
func writeln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+"\n", args...)
}

type buildCommand struct{}

var _ core.SimpleCommandPort = (*buildCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&buildCommand{},
	}
}

func (c *buildCommand) Name() string { return "build" }

func (c *buildCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "build",
		Short:         "Build one or more modules by moniker",
		Long:          "Build one or more modules by moniker.\n\nThis command builds modules respecting their dependency order.\nIf no monikers are specified, all modules in the repository are built.\n\nExpected Output:\n  - Build logs written to 'out/build/<module>/build.log' (one per module)\n  - Build manifest at 'out/build/<module>/<component>/uow.manifest.json' (with timing data)\n  - Failed builds are clearly marked with error details\n  - Failed builds do not stop execution of remaining modules\n  - Exit code 0 indicates all builds succeeded\n  - Non-zero exit code indicates one or more builds failed\n\nExample:\n  build                           # Build all modules\n  build eac-cli              # Build a single module\n  build core clie          # Build specific modules\n  build --tidy-first eac-cli # Build with go mod tidy first",
		Args:          "modules",
		Flags: []core.FlagSpec{
			{Name: "tidy-first", Type: "bool", Usage: "Run 'go mod tidy' before building (default for local)"},
			{Name: "no-tidy", Type: "bool", Usage: "Skip 'go mod tidy' (default for CI)"},
			{Name: "skip-cache", Type: "bool", Usage: "Skip incremental cache, force full rebuild"},
			{Name: "skip-depm", Type: "bool", Usage: "Only build specified modules, no dependency resolution (CI isolation)"},
			{Name: "no-deps", Type: "bool", Usage: "Alias for --skip-depm"},
			{Name: "use-existing-depm", Type: "bool", Usage: "Skip building module dependencies if artifacts exist (for CI incremental builds)"},
			{Name: "skip-deps", Type: "bool", Usage: "Skip system dependency verification (go, docker, etc.)"},
			{Name: "timings", Type: "bool", Usage: "Show detailed timing summary"},
			{Name: "debug", Type: "bool", Usage: "Enable debug logs to console (file logging always enabled)"},
			{Name: "tui", Type: "bool", Usage: "Enable TUI console (default for local, errors in CI/container)"},
			{Name: "no-tui", Type: "bool", Usage: "Disable TUI console (use plain output)"},
			{Name: "tui-height", Type: "int", Usage: "Set TUI console height (3-20, default: 6)"},
			{Name: "ascii", Type: "bool", Usage: "Use ASCII-only characters in TUI (for terminals with poor Unicode support)"},
			{Name: "skip-tui-delay", Type: "bool", Usage: "Skip TUI exit delay (exit immediately when done)"},
			{Name: "version", Type: "string", Usage: "Inject version string into binary (Go modules with executable artifacts)"},
			{Name: "accept-warnings", Type: "bool", Usage: "Don't fail on MkDocs warnings (non-strict mode)"},
			{Name: "reproducible", Type: "string", Usage: "MkDocs reproducibility mode (auto/true/false, default: auto)"},
			{Name: "all", Type: "bool", Usage: "Include non-default books (those with default: false)"},
			{Name: "list-artifacts", Type: "bool", Usage: "List artifacts that would be produced (no build)"},
			{Name: "dry-run", Type: "bool", Usage: "Simulate build without running actual commands"},
			{Name: "turbo", Type: "bool", Usage: "Enable turbo mode for faster builds (increases parallelism)"},
			{Name: "component", Type: "string", Usage: "Only build specific component(s) within each module (repeatable)"},
		},
	}
}

func (c *buildCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Build()
}

// BuildResult captures the outcome of a module build.
type BuildResult struct {
	Moniker  string
	ExitCode int
	Warnings []string
	Errors   []string
}

// ensureCommandsBinary rebuilds the eac binary if source has changed.
// Uses go.mod as a proxy for source changes - faster than checking all files.
// CI uses the setup-commands action instead.
func ensureCommandsBinary(workspaceRoot string) error {
	buildStart := time.Now()

	// Determine binary name for current platform
	binaryName := "eac"
	if runtime.GOOS == "windows" {
		binaryName = "eac.exe"
	}

	// Create tools directory
	toolsDir := filepath.Join(workspaceRoot, paths.OutDir, paths.ToolsDir)
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return fmt.Errorf("create tools dir: %w", err)
	}

	// Build from source: go/cli/eac
	cmdDir := filepath.Join(workspaceRoot, "go", "cli", "eac")
	outputPath := filepath.Join(toolsDir, binaryName)

	// Check if rebuild is needed
	needsBuild, reason := commandsBinaryNeedsRebuild(workspaceRoot, cmdDir, outputPath)
	if !needsBuild {
		fmt.Printf("⏭️  EAC binary up-to-date (%v)\n", time.Since(buildStart))
		return nil
	}

	fmt.Printf("⚙️  Building eac binary (%s)...\n", reason)

	// Build the binary
	goBuildStart := time.Now()
	toolDef := tool.GlobalRegistry().GetOrAdhoc("go")
	execCtx := &tool.ExecutionContext{
		ModuleRoot:    cmdDir,
		StdoutWriter:  os.Stdout,
		StderrWriter:  os.Stderr,
		ArgsOverrides: []string{"build", "-o", outputPath, "."},
	}
	buildResult, err := tool.GlobalExecutor().Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		return fmt.Errorf("go build: %w", err)
	}
	if buildResult.ExitCode != 0 {
		return fmt.Errorf("go build: exited with code %d", buildResult.ExitCode)
	}

	totalDuration := time.Since(buildStart)
	if totalDuration > 500*time.Millisecond {
		goBuildDuration := time.Since(goBuildStart)
		fmt.Printf("   ✅ Built: %s (go build: %05.2fs, total: %05.2fs)\n", outputPath, goBuildDuration.Seconds(), totalDuration.Seconds())
	} else {
		fmt.Printf("   ✅ Built: %s\n", outputPath)
	}
	return nil
}

// commandsBinaryNeedsRebuild checks if the eac binary needs rebuilding.
// Returns (needsBuild, reason) where reason explains why rebuild is needed.
//
// Uses sentinel files (go.mod, go.sum, go.work, go.work.sum) as proxies for
// source changes instead of walking the entire go/ directory tree. This reduces
// the check from ~50-200ms (hundreds of stat calls) to <5ms (4 stat calls).
// The sentinel files reliably cover dependency and workspace changes; source-only
// changes are caught on the next `go run` invocation which rebuilds anyway.
func commandsBinaryNeedsRebuild(workspaceRoot, cmdDir, binaryPath string) (bool, string) {
	// Check if binary exists
	binaryStat, err := os.Stat(binaryPath)
	if err != nil {
		return true, "binary missing"
	}
	binaryModTime := binaryStat.ModTime()

	// Check sentinel files that indicate source or dependency changes:
	// 1. go.mod in the commands directory (dependency changes)
	// 2. go.sum in the commands directory (dependency changes)
	// 3. go.work in workspace root (workspace module changes)
	// 4. go.work.sum in workspace root (workspace dependency changes)
	sentinelFiles := []string{
		filepath.Join(cmdDir, "go.mod"),
		filepath.Join(cmdDir, "go.sum"),
		filepath.Join(workspaceRoot, "go.work"),
		filepath.Join(workspaceRoot, "go.work.sum"),
	}

	for _, sentinel := range sentinelFiles {
		if stat, err := os.Stat(sentinel); err == nil {
			if stat.ModTime().After(binaryModTime) {
				return true, filepath.Base(sentinel) + " changed"
			}
		}
	}

	return false, ""
}

// Build command entry point - builds one or more modules.
func Build() int {
	args := os.Args[2:] // Skip program name and "build"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printBuildUsage()
		return 0
	}

	// Parse all flags (shared + build-specific)
	env := environment.Detect()
	shared, buildFlags, monikers, err := parseAllBuildFlags(args, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		printBuildUsage()
		return 1
	}

	// Resolve build settings (tidy, artifacts mode, workspace root)
	tidyFirst, artifactsMode, workspaceRoot, err := resolveBuildSettings(env, buildFlags)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Ensure eac binary exists (local only - CI uses setup-commands action)
	if env.IsLocalConsole {
		if err := ensureCommandsBinary(workspaceRoot); err != nil {
			log.Errorf("Error: failed to build eac binary: %v", err)
			return 1
		}
	}

	// Handle --list-artifacts flag (early exit path)
	if buildFlags.ListArtifacts {
		return handleListArtifacts(monikers, workspaceRoot)
	}

	// Build requested set BEFORE expanding to all modules.
	// RequestedSet tracks modules explicitly specified by the user on the command line.
	requestedSet := make(map[string]bool)
	for _, m := range monikers {
		requestedSet[m] = true
	}

	// Create command config and build-specific config for framework
	cmdCfg := buildCommandConfig(shared, monikers)
	buildCfg := buildBuildConfig(buildFlags, tidyFirst, artifactsMode, requestedSet)

	return RunBuildWithFramework(cmdCfg, buildCfg)
}

// parseAllBuildFlags parses shared flags and build-specific flags from command-line args.
// Returns shared flags, build-specific flags, extracted monikers, and any error.
func parseAllBuildFlags(args []string, env *environment.Env) (*flags.SharedFlags, *BuildSpecificFlags, []string, error) {
	shared, err := flags.ParseSharedFlagsWithEnv(flags.BuildConfig(), args, env)
	if err != nil {
		return nil, nil, nil, err
	}

	// Rebuild unconsumed args in original order for build-specific parsing.
	// The shared parser splits unknown flags (Remaining) from positional args (Monikers),
	// but build-specific value-taking flags like --component need their values preserved
	// in the correct position (e.g., "--component site" must stay together).
	buildArgs := rebuildUnconsumedArgs(args, shared.Remaining, shared.Monikers)

	buildFlags, unknownArgs, err := ParseBuildSpecificFlags(buildArgs)
	if err != nil {
		return nil, nil, nil, err
	}

	// Check for unknown flags and extract monikers from remaining args
	var monikers []string
	for _, arg := range unknownArgs {
		if strings.HasPrefix(arg, "--") {
			return nil, nil, nil, fmt.Errorf("unknown flag: %s", arg)
		}
		monikers = append(monikers, arg)
	}

	return shared, buildFlags, monikers, nil
}

// resolveBuildSettings resolves tidy-first mode, artifacts mode, and workspace root
// based on the environment and build-specific flags.
func resolveBuildSettings(env *environment.Env, buildFlags *BuildSpecificFlags) (bool, environments.ArtifactsMode, string, error) {
	// Handle tidy flag: --tidy-first sets it true, --no-tidy sets it false
	// Default is based on environment (true for local, false for CI)
	tidyFirst := !env.IsCI
	if buildFlags.TidyFirst {
		tidyFirst = true
	}
	if buildFlags.NoTidy {
		tidyFirst = false
	}

	// Handle artifacts mode: default based on environment (all for CI, reduced for local)
	artifactsMode := environments.DefaultArtifactsMode()
	if buildFlags.Artifacts != "" {
		mode, err := environments.ParseArtifactsMode(buildFlags.Artifacts)
		if err != nil {
			return false, artifactsMode, "", err
		}
		artifactsMode = mode
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return false, artifactsMode, "", fmt.Errorf("failed to find repository root: %w", err)
	}

	return tidyFirst, artifactsMode, workspaceRoot, nil
}

// handleListArtifacts handles the --list-artifacts flag, loading module contracts
// and listing artifacts without performing a build.
func handleListArtifacts(monikers []string, workspaceRoot string) int {
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("Error: failed to load module contracts: %v", err)
		return 1
	}
	if len(monikers) == 0 {
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}
	return listModuleArtifacts(monikers, workspaceRoot, moduleReport)
}

// buildCommandConfig constructs a CommandConfig from parsed shared flags and monikers.
func buildCommandConfig(shared *flags.SharedFlags, monikers []string) *cmdframework.CommandConfig {
	return &cmdframework.CommandConfig{
		Type:           core.ActionBuild,
		CommandPath:    "build",
		OutputDir:      paths.OutBuildRelPath,
		Monikers:       monikers,
		IncludeDepm:    !shared.SkipDepm,
		SkipDeps:       shared.SkipDeps,
		SkipDepm:       shared.SkipDepm,
		ForceRebuild:   shared.CacheConfig.ShouldSkipState(),
		Turbo:          shared.Turbo,
		MaxConcurrency: shared.MaxConcurrency,
		DryRun:         shared.DryRun,
		UseTUI:         shared.UseTUI,
		TUIHeight:      shared.TUIHeight,
		TUIASCIIMode:   shared.TUIASCIIMode,
		TUI3Demo:       shared.TUI3Demo,
		SkipTUIDelay:   shared.SkipTUIDelay,
		ShowTimings:    shared.ShowTimings,
		DebugMode:      shared.Debug,
		CacheConfig:    shared.CacheConfig,
	}
}

// buildBuildConfig constructs a BuildConfig from resolved build settings.
func buildBuildConfig(buildFlags *BuildSpecificFlags, tidyFirst bool, artifactsMode environments.ArtifactsMode, requestedSet map[string]bool) *BuildConfig {
	return &BuildConfig{
		TidyFirst:       tidyFirst,
		Version:         buildFlags.Version,
		UseExistingDepm: buildFlags.UseExistingDepm,
		Reproducible:    buildFlags.Reproducible,
		ArtifactsMode:   artifactsMode,
		Components:      buildFlags.Components,
		RequestedSet:    requestedSet,
	}
}

// rebuildUnconsumedArgs reconstructs remaining and positional args in their
// original order. The shared parser separates unknown flags (remaining) from
// positional args (positional), but this loses ordering information needed for
// value-taking build-specific flags like --component, --version, --reproducible.
// By restoring original order, ParseBuildSpecificFlags can correctly pair flags
// with their values (e.g., "--component site" stays together).
func rebuildUnconsumedArgs(originalArgs, remaining, positional []string) []string {
	// Count occurrences of each unconsumed arg
	unconsumed := make(map[string]int)
	for _, r := range remaining {
		unconsumed[r]++
	}
	for _, p := range positional {
		unconsumed[p]++
	}

	// Scan original args, keeping only unconsumed ones in original order
	var result []string
	for _, arg := range originalArgs {
		if unconsumed[arg] > 0 {
			result = append(result, arg)
			unconsumed[arg]--
		}
	}
	return result
}

// isValidReproducible checks if a reproducible flag value is valid.
func isValidReproducible(value string) bool {
	return value == "auto" || value == "true" || value == "false"
}

// listModuleArtifacts lists the artifacts that would be produced by building the specified modules.
// Collects artifacts from all buildable packages for each module.
func listModuleArtifacts(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport) int {
	// Sort monikers for consistent output
	sort.Strings(monikers)

	for _, moniker := range monikers {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			log.Errorf("Error: module not found: %s", moniker)
			continue
		}

		// Get all handlers for module's buildable components
		compHandlers := tool.GlobalBuildBridge().GetHandlersForModule(module)
		if len(compHandlers) == 0 {
			// No buildable components for this module
			continue
		}

		outputDir := paths.BuildOutputPath(workspaceRoot, moniker)

		// Collect artifacts from all handlers
		modulePort := adapters.AdaptModule(module)
		for _, ch := range compHandlers {
			artifacts := ch.Handler.ListArtifacts(modulePort, workspaceRoot)
			for _, artifact := range artifacts {
				// Output full path relative to workspace root
				fullPath := filepath.Join(outputDir, artifact)
				relPath, err := filepath.Rel(workspaceRoot, fullPath)
				if err != nil {
					relPath = fullPath
				}
				fmt.Println(relPath)
			}
		}
	}

	return 0
}

// runModuleBuild runs build for a single module.
// Executes all handlers for the module's buildable components in sequence.
func runModuleBuild(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, tidyFirst bool, version string, dryRun bool, artifactsMode environments.ArtifactsMode, reproducible bool) int {
	// Get all handlers for module's buildable components
	compHandlers := tool.GlobalBuildBridge().GetHandlersForModule(module)
	if len(compHandlers) == 0 {
		output.Writeln(logWriter, "ℹ️  No buildable components for module: %s", module.Moniker)
		return 0 // Not an error - module just doesn't have buildable components
	}

	// Determine which artifacts to build
	requestedArtifacts := determineRequestedArtifactsForBuild(module, artifactsMode, workspaceRoot)

	opts := tool.BuildOptions{
		TidyFirst:          tidyFirst,
		Version:            version,
		DryRun:             dryRun,
		RequestedArtifacts: requestedArtifacts,
		Reproducible:       reproducible,
	}

	// In dry-run mode, simulate a successful build
	if dryRun {
		output.Writeln(logWriter, "Build: %s (dry-run)", module.Moniker)
		output.Writeln(logWriter, "Components: %v", module.GetEnabledComponents())
		output.Writeln(logWriter, "Handlers: %d", len(compHandlers))
		for _, ch := range compHandlers {
			output.Writeln(logWriter, "  - %s: %s", ch.Component, ch.Handler.Name())
		}
		// Show component roots
		for compType, root := range module.GetComponentRoots() {
			output.Writeln(logWriter, "Root[%s]: %s", compType, root)
		}
		output.Writeln(logWriter, "")
		output.Writeln(logWriter, "Dry-run mode: skipping actual build")
		return 0
	}

	// Run each handler in sequence
	modulePort := adapters.AdaptModule(module)
	for _, ch := range compHandlers {
		handler := ch.Handler

		// Log which component/handler we're building
		if len(compHandlers) > 1 {
			output.Writeln(logWriter, "")
			output.Writeln(logWriter, "━━━ Building component: %s (handler: %s) ━━━", ch.Component, handler.Name())
		}

		// Validate module before building
		if err := handler.ValidateModule(modulePort, workspaceRoot, ch.Component); err != nil {
			output.Writeln(logWriter, "❌ Module validation failed for %s: %v", ch.Component, err)
			return 1
		}

		exitCode := handler.Build(modulePort, workspaceRoot, outputDir, logWriter, opts)
		if exitCode != 0 {
			output.Writeln(logWriter, "❌ Build failed for component: %s", ch.Component)
			return exitCode
		}
	}

	// Note: Artifact derivations are executed centrally by framework.go's buildAfterExecute()
	// via processAllArtifactDerivations(). This avoids duplicate derivation runs when
	// framework.go calls runModuleBuild() as its worker.

	return 0
}

func printBuildUsage() {
	log.Info("Build one or more modules by moniker")
	log.Info("")
	log.Info("Usage: build [flags] [module1] [module2] ...")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  module1, module2, ...     Module monikers to build (builds all if none specified)")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --dry-run                 Simulate build without running actual commands")
	log.Info("  --list-artifacts          List artifacts that would be produced (no build)")
	log.Info("  --tidy-first              Run 'go mod tidy' before building (default for local)")
	log.Info("  --no-tidy                 Skip 'go mod tidy' (default for CI)")
	log.Info("  --skip-cache              Skip incremental cache, force full rebuild")
	log.Info("  --skip-depm               Only build specified modules, no module dependency resolution (CI isolation)")
	log.Info("  --use-existing-depm       Skip building module dependencies if artifacts exist (for CI incremental builds)")
	log.Info("  --skip-deps               Skip system dependency verification (go, docker, etc.)")
	log.Info("  --turbo                   Enable turbo mode for faster builds (increases parallelism)")
	log.Info("  --roof N                  Limit max parallel capacity to N (default: auto-detect from CPU/RAM)")
	log.Info("  --timings                 Show detailed timing summary")
	log.Info("  --debug                   Enable debug logs to console (file logging always enabled)")
	log.Info("  --tui                     Enable TUI console (default for local, errors in CI/container)")
	log.Info("  --no-tui                  Disable TUI console (use plain output)")
	log.Info(fmt.Sprintf("  --tui-height N            Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --ascii                   Use ASCII-only characters in TUI (for terminals with poor Unicode support)")
	log.Info("  --skip-tui-delay          Skip TUI exit delay (exit immediately when done)")
	log.Info("  --version VERSION         Inject version string into binary (Go modules with executable artifacts)")
	log.Info("  --accept-warnings         Don't fail on MkDocs warnings (non-strict mode)")
	log.Info("  --reproducible MODE       MkDocs reproducibility mode: auto (default), true, false")
	log.Info("                            auto: CI uses true, local uses false")
	log.Info("                            true: Always rebuild HTML from staging")
	log.Info("                            false: Skip MkDocs if staging unchanged")
	log.Info("  --artifacts MODE          Artifact scope mode: all, reduced")
	log.Info("                            all: Build all artifacts for all platforms (CI default)")
	log.Info("                            reduced: Reduced artifacts for faster local builds (local default)")
	log.Info("  --all                     Alias for --artifacts all")
	log.Info("  --component NAME          Only build specific component(s) within each module (repeatable)")
	log.Info("  -h, --help                Show this help message")
	log.Info("")
	log.Info("MkDocs modules with books (books.yml):")
	log.Info("  Books with 'default: false' are skipped.")
	log.Info("  Output is configured via the book's 'output' field:")
	log.Info("    site       - HTML site only")
	log.Info("    pdf-dark   - PDF with dark theme")
	log.Info("    pdf-light  - PDF with light theme")
	log.Info("    pdf-all    - Both dark and light PDFs")
	log.Info("")
	log.Info("Examples:")
	log.Info("  build                                # Build all modules")
	log.Info("  build clie                        # Build CLI for current platform")
	log.Info("  build books                          # Build books (uses books.yml output config)")
	log.Info("  build clie --list-artifacts       # List artifacts without building")
}

// hasExistingArtifacts checks if a module's build artifacts already exist AND
// were built from the same source inputs.
// Used by --use-existing-depm to skip building modules whose artifacts are present
// (typically downloaded from previous CI runs).
func hasExistingArtifacts(moniker, workspaceRoot string, buildAll bool, cfg *config.EACConfig) bool {
	// Get module
	module, ok := cfg.Repository.GetModule(moniker)
	if !ok {
		return false
	}

	// Check if module has any artifacts defined in packages
	hasModuleArtifacts := false
	for _, pkg := range module.Components {
		if pkg != nil && pkg.Build != nil && len(pkg.Build.Artifacts) > 0 {
			hasModuleArtifacts = true
			break
		}
	}

	// If no artifacts defined, consider it as "exists" (nothing to build)
	if !hasModuleArtifacts {
		return true
	}

	// Determine which artifacts are requested
	requestedArtifacts := base.DetermineRequestedArtifacts(module, buildAll, cfg)
	if len(requestedArtifacts) == 0 {
		return true
	}

	// Resolve artifacts and check if they exist
	buildDir := paths.BuildOutputPath(workspaceRoot, moniker)

	// First, verify input hash matches using UoW manifests
	// This ensures cached artifacts are from the same source code
	reader := coreoutput.NewReader(workspaceRoot)
	if manifests, err := reader.ListUoWs(core.ActionBuild, moniker); err == nil && len(manifests) > 0 {
		// Get input hash from first UoW (they should all have the same source hash)
		var cachedHash string
		for _, m := range manifests {
			if m.InputHash != "" {
				cachedHash = m.InputHash
				break
			}
		}
		if cachedHash != "" {
			// Load module registry for hash computation
			modRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
			if err == nil {
				if contract, ok := modRegistry.Get(moniker); ok {
					currentHash, err := hash.ComputeFromPatterns(workspaceRoot, contract)
					if err == nil && currentHash != cachedHash {
						log.Debugf("Module %s: input hash mismatch (cached=%s, current=%s) - need rebuild",
							moniker, cachedHash[:16], currentHash[:16])
						return false
					}
				}
			}
		}
	}

	artifacts, _, err := config.ResolveArtifactsForModuleWithConfig(
		module, buildDir, runtime.GOOS, runtime.GOARCH, cfg,
	)
	if err != nil {
		return false
	}

	// Check if all requested artifacts exist
	for i := range artifacts {
		art := &artifacts[i]
		// Only check requested artifacts
		isRequested := false
		for _, reqID := range requestedArtifacts {
			if art.ID == reqID {
				isRequested = true
				break
			}
		}

		if isRequested && !art.Exists {
			return false
		}
	}

	return true
}

// determineRequestedArtifactsForBuild determines which artifact IDs should be built for a module.
func determineRequestedArtifactsForBuild(moduleContract *modules.ModuleContract, mode environments.ArtifactsMode, workspaceRoot string) []string {
	// When --artifacts all is specified, return "*" to signal builders to include all artifacts
	// This handles cases where module type has no artifacts defined (like container type)
	// but module has books that should all be built
	if mode.AllArtifactsRequested() {
		return []string{"*"}
	}

	// Load config to get module and type definitions
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Debugf("Failed to load config for artifact determination: %v", err)
		return []string{} // Empty list - builders will use defaults
	}

	// Get module from config
	module, ok := cfg.Repository.GetModule(moduleContract.Moniker)
	if !ok {
		log.Debugf("Module %s not found in config", moduleContract.Moniker)
		return []string{}
	}

	// Use the DetermineRequestedArtifacts function
	return base.DetermineRequestedArtifacts(module, false, cfg)
}

// enrichImageArtifact populates image-specific fields (Tags, Registry) from docker_build config.
// Uses module-level docker_build config.
func enrichImageArtifact(artifactInfo *coreoutput.ArtifactInfo, module *config.Module, moniker string) {
	// Get docker_build config from module
	var dockerConfig *config.DockerBuildConfig
	if module != nil {
		dockerConfig = module.GetDockerBuildConfig()
	}

	if dockerConfig == nil {
		return
	}

	// Expand template variables in tags
	expandedTags := make([]string, 0, len(dockerConfig.Tags))
	for _, tag := range dockerConfig.Tags {
		expanded := expandImageTag(tag, moniker)
		expandedTags = append(expandedTags, expanded)
	}

	artifactInfo.Tags = expandedTags
	artifactInfo.Registry = dockerConfig.Registry
}

// expandImageTag expands template variables in an image tag.
func expandImageTag(tag, moniker string) string {
	result := tag
	result = strings.ReplaceAll(result, "{moniker}", moniker)
	result = strings.ReplaceAll(result, "{container}", moniker)

	// Get short SHA if available
	if sha := os.Getenv("GITHUB_SHA"); sha != "" && len(sha) >= 7 {
		result = strings.ReplaceAll(result, "{short_sha}", sha[:7])
	}

	return result
}
