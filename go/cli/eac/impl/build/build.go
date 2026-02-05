// Command: build
// Short: Build one or more modules by moniker
// Long: Build one or more modules by moniker.
// Long:
// Long: This command builds modules respecting their dependency order.
// Long: If no monikers are specified, all modules in the repository are built.
// Long:
// Long: Expected Output:
// Long:   - Build logs written to 'out/build/<module>/build.log' (one per module)
// Long:   - Build manifest at 'out/build/<module>/<component>/uow.manifest.json' (with timing data)
// Long:   - Failed builds are clearly marked with error details
// Long:   - Failed builds do not stop execution of remaining modules
// Long:   - Exit code 0 indicates all builds succeeded
// Long:   - Non-zero exit code indicates one or more builds failed
// Long:
// Long: Example:
// Long:   build                           # Build all modules
// Long:   build eac-cli              # Build a single module
// Long:   build core r2r-cli          # Build specific modules
// Long:   build --tidy-first eac-cli # Build with go mod tidy first
// Flag.tidy-first: type=bool, usage=Run 'go mod tidy' before building (default for local)
// Flag.no-tidy: type=bool, usage=Skip 'go mod tidy' (default for CI)
// Flag.skip-cache: type=bool, usage=Skip incremental cache, force full rebuild
// Flag.skip-depm: type=bool, usage=Only build specified modules, no dependency resolution (CI isolation)
// Flag.no-deps: type=bool, usage=Alias for --skip-depm
// Flag.use-existing-depm: type=bool, usage=Skip building module dependencies if artifacts exist (for CI incremental builds)
// Flag.skip-deps: type=bool, usage=Skip system dependency verification (go, docker, etc.)
// Flag.timings: type=bool, usage=Show detailed timing summary
// Flag.debug: type=bool, usage=Enable debug logs to console (file logging always enabled)
// Flag.tui: type=bool, usage=Enable TUI console (default for local, errors in CI/container)
// Flag.no-tui: type=bool, usage=Disable TUI console (use plain output)
// Flag.tui-height: type=int, usage=Set TUI console height (3-20, default: 6)
// Flag.ascii: type=bool, usage=Use ASCII-only characters in TUI (for terminals with poor Unicode support)
// Flag.skip-tui-delay: type=bool, usage=Skip TUI exit delay (exit immediately when done)
// Flag.version: type=string, usage=Inject version string into binary (Go modules with executable artifacts)
// Flag.accept-warnings: type=bool, usage=Don't fail on MkDocs warnings (non-strict mode)
// Flag.reproducible: type=string, usage=MkDocs reproducibility mode (auto/true/false, default: auto)
// Flag.all: type=bool, usage=Include non-default books (those with default: false)
// Flag.list-artifacts: type=bool, usage=List artifacts that would be produced (no build)
// Flag.dry-run: type=bool, usage=Simulate build without running actual commands
// Flag.turbo: type=bool, usage=Enable turbo mode for faster builds (increases parallelism)
// Args: modules
package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/builders"
	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

var log = logging.C()

// writeln writes a formatted message followed by a newline to the writer.
func writeln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+"\n", args...)
}

func init() {
	registry.Register(Build)
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
	buildCmd := exec.Command("go", "build", "-o", outputPath, ".")
	buildCmd.Dir = cmdDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
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
func commandsBinaryNeedsRebuild(workspaceRoot, cmdDir, binaryPath string) (bool, string) {
	// Check if binary exists
	binaryStat, err := os.Stat(binaryPath)
	if err != nil {
		return true, "binary missing"
	}
	binaryModTime := binaryStat.ModTime()

	// Check sentinel files that indicate source changes:
	// 1. go.mod in the commands directory (dependency changes)
	// 2. go.sum in the commands directory (dependency changes)
	// 3. go.work in workspace root (workspace changes)
	sentinelFiles := []string{
		filepath.Join(cmdDir, "go.mod"),
		filepath.Join(cmdDir, "go.sum"),
		filepath.Join(workspaceRoot, "go.work"),
	}

	for _, sentinel := range sentinelFiles {
		if stat, err := os.Stat(sentinel); err == nil {
			if stat.ModTime().After(binaryModTime) {
				return true, filepath.Base(sentinel) + " changed"
			}
		}
	}

	// Check if any .go file in the commands package tree is newer
	// Walk the entire go directory since commands imports from core, adapters, etc.
	goDir := filepath.Join(workspaceRoot, "go")
	newestGoFile := ""
	newestGoTime := time.Time{}

	filepath.Walk(goDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			// Skip vendor, testdata, and hidden directories
			name := info.Name()
			if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		// Only check .go files
		if strings.HasSuffix(path, ".go") {
			if info.ModTime().After(newestGoTime) {
				newestGoTime = info.ModTime()
				newestGoFile = path
			}
		}
		return nil
	})

	if newestGoTime.After(binaryModTime) {
		// Get relative path for cleaner output
		relPath, _ := filepath.Rel(workspaceRoot, newestGoFile)
		if relPath == "" {
			relPath = filepath.Base(newestGoFile)
		}
		return true, relPath + " changed"
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

	// Detect execution environment
	env := environment.Detect()

	// Parse shared flags using flag sets
	shared, err := flags.ParseSharedFlagsWithEnv(flags.BuildConfig(), args, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		printBuildUsage()
		return 1
	}

	// Parse build-specific flags from remaining args
	buildFlags, unknownArgs, err := ParseBuildSpecificFlags(shared.Remaining)
	if err != nil {
		log.Errorf("Error: %v", err)
		printBuildUsage()
		return 1
	}

	// Check for unknown flags
	for _, arg := range unknownArgs {
		if strings.HasPrefix(arg, "--") {
			log.Errorf("Error: unknown flag: %s", arg)
			printBuildUsage()
			return 1
		}
	}

	// Extract values from shared flags
	monikers := shared.Monikers
	skipDeps := shared.SkipDeps
	skipDepm := shared.SkipDepm
	forceRebuild := shared.CacheConfig.ShouldSkipState()
	showTimings := shared.ShowTimings
	debugMode := shared.Debug
	turbo := shared.Turbo
	roof := shared.MaxConcurrency
	dryRun := shared.DryRun
	useTUI := shared.UseTUI
	tuiHeight := shared.TUIHeight
	tuiASCII := shared.TUIASCIIMode
	tuiDemo := shared.TUI3Demo
	skipTUIDelay := shared.SkipTUIDelay

	// Extract build-specific flags
	useExistingDepm := buildFlags.UseExistingDepm
	version := buildFlags.Version
	reproducible := buildFlags.Reproducible
	listArtifacts := buildFlags.ListArtifacts

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
	artifactsMode := env.DefaultArtifactsMode()
	if buildFlags.Artifacts != "" {
		mode, err := environment.ParseArtifactsMode(buildFlags.Artifacts)
		if err != nil {
			log.Errorf("Error: %v", err)
			return 1
		}
		artifactsMode = mode
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Ensure eac binary exists (local only - CI uses setup-commands action)
	if env.IsLocalConsole {
		if err := ensureCommandsBinary(workspaceRoot); err != nil {
			log.Errorf("Error: failed to build eac binary: %v", err)
			return 1
		}
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("Error: failed to load module contracts: %v", err)
		return 1
	}

	// Build requested set BEFORE expanding to all modules
	// RequestedSet tracks modules explicitly specified by the user on the command line.
	// This enables incremental detection to skip unchanged modules when "eac build" is run
	// without arguments, while still building explicitly requested modules unconditionally.
	requestedSet := make(map[string]bool)
	for _, m := range monikers {
		requestedSet[m] = true
	}

	// If no monikers provided, default to all buildable modules (before dependency check)
	if len(monikers) == 0 {
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

	// Handle --list-artifacts flag
	if listArtifacts {
		return listModuleArtifacts(monikers, workspaceRoot, moduleReport)
	}

	// Create command config for framework
	cmdCfg := &cmdframework.CommandConfig{
		Type:           cmdframework.CommandTypeBuild,
		CommandPath:    "build",
		ActionVerb:     "Building",
		OutputDir:      paths.OutBuildRelPath,
		LogFileName:    "build.log",
		Monikers:       monikers,
		IncludeDepm:    !skipDepm,
		SkipDeps:       skipDeps,
		SkipDepm:       skipDepm,
		ForceRebuild:   forceRebuild,
		Turbo:          turbo,
		MaxConcurrency: roof,
		DryRun:         dryRun,
		UseTUI:         useTUI,
		TUIHeight:      tuiHeight,
		TUIASCIIMode:   tuiASCII,
		TUI3Demo:       tuiDemo,
		SkipTUIDelay:   skipTUIDelay,
		ShowTimings:    showTimings,
		DebugMode:      debugMode,
		CacheConfig:    shared.CacheConfig,
	}

	// Create build-specific config
	buildCfg := &BuildConfig{
		TidyFirst:       tidyFirst,
		Version:         version,
		UseExistingDepm: useExistingDepm,
		Reproducible:    reproducible,
		ArtifactsMode:   artifactsMode,
		RequestedSet:    requestedSet,
	}

	return RunBuildWithFramework(cmdCfg, buildCfg)
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
		compHandlers := builders.GetHandlersForModule(module)
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
func runModuleBuild(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, tidyFirst bool, version string, dryRun bool, artifactsMode environment.ArtifactsMode, reproducible bool) int {
	// Get all handlers for module's buildable components
	compHandlers := builders.GetHandlersForModule(module)
	if len(compHandlers) == 0 {
		output.Writeln(logWriter, "ℹ️  No buildable components for module: %s", module.Moniker)
		return 0 // Not an error - module just doesn't have buildable components
	}

	// Determine which artifacts to build
	requestedArtifacts := determineRequestedArtifactsForBuild(module, artifactsMode, workspaceRoot)

	opts := builders.BuildOptions{
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

	// Process artifact derivations (compression, etc.) if all builds succeeded
	// Use merged artifacts from cfg.GetBuildArtifacts to ensure module-level takes priority
	cfg := config.Global()
	if cfg != nil {
		// Get merged artifacts (module-level takes priority over type-level)
		// Pass buildAll=true to include all derived artifacts (UPX variants, etc.)
		mergedArtifacts := cfg.GetBuildArtifacts(module.Moniker, true)
		if err := ProcessArtifactDerivations(module.Moniker, mergedArtifacts, outputDir, opts.RequestedArtifacts, module.Metadata, logWriter); err != nil {
			output.Writeln(logWriter, "❌ Artifact derivation failed: %v", err)
			return 1
		}
	}

	// Note: Post-build steps are executed centrally by framework.go's buildAfterExecute()
	// via processAllArtifactDerivations(). This ensures consistent behavior and prevents
	// duplicate execution when framework.go calls runModuleBuild() as its worker.

	return 0
}

// verifyBuildDependenciesQuiet checks build dependencies silently and returns status for summary.
// No output is written - the caller is responsible for displaying the results via InitSummary.
// Docker is verified from the bootstrap namespace - it's the only true system dependency.
// Tool-level dependencies are handled by the tool system at execution time.
func verifyBuildDependenciesQuiet(monikers []string, moduleReport *reports.ModuleContractReport) (int, initsummary.DepsStatus) {
	status := initsummary.DepsStatus{Verified: true}

	// Get bootstrap tools from the tool registry
	registry := tool.GlobalRegistry()
	bootstrapTools := registry.GetBootstrapTools()

	if len(bootstrapTools) == 0 {
		// No bootstrap tools configured, skip verification
		return 0, status
	}

	status.Required = bootstrapTools

	// Verify bootstrap tools
	results := registry.VerifyAll(bootstrapTools)

	for _, result := range results {
		status.Available = append(status.Available, initsummary.DepsResult{
			Name:      result.ToolID,
			Available: result.Available,
			Version:   result.Version,
		})
		if !result.Available {
			status.Missing = append(status.Missing, result.ToolID)
		}
	}

	if len(status.Missing) > 0 {
		return 1, status
	}

	return 0, status
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
	log.Info("  build r2r-cli                        # Build CLI for current platform")
	log.Info("  build books                          # Build books (uses books.yml output config)")
	log.Info("  build r2r-cli --list-artifacts       # List artifacts without building")
}

// hasExistingArtifacts checks if a module's build artifacts already exist AND
// were built from the same source inputs.
// Used by --use-existing-depm to skip building modules whose artifacts are present
// (typically downloaded from previous CI runs).
func hasExistingArtifacts(moniker, workspaceRoot string, buildAll bool) bool {
	// Load config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return false
	}

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
	requestedArtifacts := implinternal.DetermineRequestedArtifacts(module, buildAll, cfg)
	if len(requestedArtifacts) == 0 {
		return true
	}

	// Resolve artifacts and check if they exist
	buildDirRel := cfg.Repository.BuildOutputPath(moniker)
	buildDir := filepath.Join(workspaceRoot, buildDirRel)

	// First, verify input hash matches using UoW manifests
	// This ensures cached artifacts are from the same source code
	reader := coreoutput.NewReader(workspaceRoot)
	if manifests, err := reader.ListUoWs(workunit.ContextBuild, moniker); err == nil && len(manifests) > 0 {
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

	artifacts, _, err := implinternal.ResolveArtifactsForModuleWithConfig(
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
func determineRequestedArtifactsForBuild(moduleContract *modules.ModuleContract, mode environment.ArtifactsMode, workspaceRoot string) []string {
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
	return implinternal.DetermineRequestedArtifacts(module, false, cfg)
}

// enrichImageArtifact populates image-specific fields (Tags, Registry) from docker_build config.
// Uses module-level docker_build config.
func enrichImageArtifact(artifactInfo *implinternal.ArtifactInfo, module *config.Module, moniker string) {
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
