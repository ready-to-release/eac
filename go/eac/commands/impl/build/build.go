// Command: build
// Short: Build one or more modules by moniker
// Long: Build one or more modules by moniker.
// Long:
// Long: This command builds modules respecting their dependency order.
// Long: If no monikers are specified, all modules in the repository are built.
// Long:
// Long: Expected Output:
// Long:   - Build logs written to 'out/build/<module>/build.log' (one per module)
// Long:   - Build manifest at 'out/build/<module>/build.manifest.json' (with timing data)
// Long:   - Failed builds are clearly marked with error details
// Long:   - Failed builds do not stop execution of remaining modules
// Long:   - Exit code 0 indicates all builds succeeded
// Long:   - Non-zero exit code indicates one or more builds failed
// Long:
// Long: Example:
// Long:   build                           # Build all modules
// Long:   build eac-commands              # Build a single module
// Long:   build eac-core r2r-cli          # Build specific modules
// Long:   build --tidy-first eac-commands # Build with go mod tidy first
// Flag.tidy-first: type=bool, usage=Run 'go mod tidy' before building (default for local)
// Flag.no-tidy: type=bool, usage=Skip 'go mod tidy' (default for CI)
// Flag.rebuild: type=bool, usage=Force full rebuild, ignoring incremental build state
// Flag.layered-build: type=bool, usage=Execute layers sequentially (default: all modules in parallel)
// Flag.skip-depm: type=bool, usage=Only build specified modules, no dependency resolution (CI isolation)
// Flag.no-deps: type=bool, usage=Alias for --skip-depm
// Flag.use-existing-depm: type=bool, usage=Skip building module dependencies if artifacts exist (for CI incremental builds)
// Flag.skip-deps-verification: type=bool, usage=Skip system dependency verification (go, docker, etc.)
// Flag.timings: type=bool, usage=Show detailed timing summary
// Flag.debug: type=bool, usage=Enable debug logs to console (file logging always enabled)
// Flag.tui: type=bool, usage=Enable TUI console (default for local, errors in CI/container)
// Flag.no-tui: type=bool, usage=Disable TUI console (use plain output)
// Flag.tui-height: type=int, usage=Set TUI console height (3-20, default: 6)
// Flag.version: type=string, usage=Inject version string into binary (Go modules with executable artifacts)
// Flag.accept-warnings: type=bool, usage=Don't fail on MkDocs warnings (non-strict mode)
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
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/git"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/buildstate"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/tool"
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

// ensureCommandsBinary rebuilds the commands binary on every local build call.
// This ensures the binary always reflects the latest source code changes.
// CI uses the setup-commands action instead.
func ensureCommandsBinary(workspaceRoot string) error {
	fmt.Println("⚙️  Building commands binary...")

	// Determine binary name for current platform
	binaryName := "commands"
	if runtime.GOOS == "windows" {
		binaryName = "commands.exe"
	}

	// Create tools directory
	toolsDir := filepath.Join(workspaceRoot, paths.OutDir, paths.ToolsDir)
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return fmt.Errorf("create tools dir: %w", err)
	}

	// Build from source: go/eac/commands
	cmdDir := filepath.Join(workspaceRoot, "go", "eac", "commands")
	outputPath := filepath.Join(toolsDir, binaryName)

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", outputPath, ".")
	buildCmd.Dir = cmdDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	fmt.Printf("   ✅ Built: %s\n", outputPath)
	return nil
}

// buildFlags defines valid flags for the build command

// Build command entry point - builds one or more modules.
func Build() int {
	args := os.Args[2:] // Skip program name and "build"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printBuildUsage()
		return 0
	}

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Detect execution environment
	env := environment.Detect()

	// Parse module monikers and flags
	var monikers []string
	tidyFirst := !env.IsCI   // Default: true for local, false for CI
	skipDeps := false        // Skip system dependency verification (go, docker, etc.)
	skipDepm := false        // Skip including transitive module dependencies in execution plan
	useExistingDepm := false // Use existing module dependency artifacts (skip building if present)
	forceRebuild := false    // Force full rebuild, ignoring incremental build state (--rebuild)
	layeredBuild := false    // Execute in layers sequentially (default: all parallel)
	showTimings := false
	debugMode := false // Enable debug logs to console
	turbo := false     // Enable turbo mode (+2 parallel workers)
	version := ""
	listArtifacts := false
	dryRun := false
	buildAll := false            // Include non-default books (those with default: false)
	useTUI := env.ShouldUseTUI() // TUI enabled by default for local console mode
	tuiExplicitlySet := false
	tuiHeight := tui.DefaultHeight // TUI console height
	tuiASCII := false              // Use ASCII-only characters in TUI

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--tidy-first":
			tidyFirst = true
		case "--no-tidy":
			tidyFirst = false
		case "--skip-depm":
			skipDepm = true
		case "--use-existing-depm":
			useExistingDepm = true
		case "--rebuild":
			forceRebuild = true
		case "--layered-build":
			layeredBuild = true
		case "--skip-deps":
			skipDeps = true
		case "--timings":
			showTimings = true
		case "--debug":
			debugMode = true
		case "--turbo":
			turbo = true
		case "--accept-warnings":
			// Flag is handled in mkdocs builder via os.Args check
			// Just accept it here so it doesn't fail as unknown flag
		case "--list-artifacts":
			listArtifacts = true
		case "--dry-run":
			dryRun = true
		case "--all":
			buildAll = true
		case "--tui":
			useTUI = true
			tuiExplicitlySet = true
		case "--no-tui":
			useTUI = false
			tuiExplicitlySet = true
		case "--ascii":
			tuiASCII = true
		case "--tui-height":
			if i+1 >= len(args) {
				log.Errorf("Error: --tui-height requires a value")
				printBuildUsage()
				return 1
			}
			i++
			var err error
			tuiHeight, err = parseIntArg(args[i])
			if err != nil || tuiHeight < 3 || tuiHeight > 20 {
				log.Errorf("Error: --tui-height must be a number between 3 and 20")
				printBuildUsage()
				return 1
			}
		case "--version":
			if i+1 >= len(args) {
				log.Errorf("Error: --version requires a value")
				printBuildUsage()
				return 1
			}
			i++
			version = args[i]
		default:
			if strings.HasPrefix(arg, "--version=") {
				version = strings.TrimPrefix(arg, "--version=")
			} else if strings.HasPrefix(arg, "--tui-height=") {
				heightStr := strings.TrimPrefix(arg, "--tui-height=")
				var err error
				tuiHeight, err = parseIntArg(heightStr)
				if err != nil || tuiHeight < 3 || tuiHeight > 20 {
					log.Errorf("Error: --tui-height must be a number between 3 and 20")
					printBuildUsage()
					return 1
				}
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf("Error: unknown flag: %s", arg)
				printBuildUsage()
				return 1
			} else {
				monikers = append(monikers, arg)
			}
		}
	}

	// Validate TUI usage - error if explicitly enabled in CI or container mode
	if err := env.ValidateTUI(tuiExplicitlySet, useTUI); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Validate --turbo and --layered-build mutual exclusivity
	// (layered-build implies sequential execution, turbo increases parallelism)
	if turbo && layeredBuild {
		log.Errorf("Error: --turbo and --layered-build cannot be used together")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Ensure commands binary exists (devbox only - CI uses setup-commands action)
	if env.IsLocalConsole {
		if err := ensureCommandsBinary(workspaceRoot); err != nil {
			log.Errorf("Error: failed to build commands binary: %v", err)
			return 1
		}
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("Error: failed to load module contracts: %v", err)
		return 1
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

	// Build requested set for --use-existing-depm logic
	requestedSet := make(map[string]bool)
	for _, m := range monikers {
		requestedSet[m] = true
	}

	// Create command config for framework
	cmdCfg := &cmdframework.CommandConfig{
		Type:         cmdframework.CommandTypeBuild,
		ActionVerb:   "Building",
		OutputDir:    paths.OutBuildRelPath,
		LogFileName:  "build.log",
		Monikers:     monikers,
		IncludeDepm:  !skipDepm,
		SkipDeps:     skipDeps,
		SkipDepm:     skipDepm,
		ForceRebuild: forceRebuild,
		Layered:      layeredBuild,
		Turbo:        turbo,
		DryRun:       dryRun,
		UseTUI:       useTUI,
		TUIHeight:    tuiHeight,
		TUIASCIIMode: tuiASCII,
		ShowTimings:  showTimings,
		DebugMode:    debugMode,
	}

	// Create build-specific config
	buildCfg := &BuildConfig{
		TidyFirst:       tidyFirst,
		Version:         version,
		BuildAll:        buildAll,
		UseExistingDepm: useExistingDepm,
		LayeredBuild:    layeredBuild,
		RequestedSet:    requestedSet,
	}

	return RunBuildWithFramework(cmdCfg, buildCfg)
}

// parseIntArg parses a string argument as an integer.
func parseIntArg(s string) (int, error) {
	return strconv.Atoi(s)
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
		for _, ch := range compHandlers {
			artifacts := ch.Handler.ListArtifacts(module, workspaceRoot)
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
func runModuleBuild(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, tidyFirst bool, version string, dryRun, buildAll bool) int {
	// Get all handlers for module's buildable components
	compHandlers := builders.GetHandlersForModule(module)
	if len(compHandlers) == 0 {
		output.Writeln(logWriter, "ℹ️  No buildable components for module: %s", module.Moniker)
		return 0 // Not an error - module just doesn't have buildable components
	}

	// Determine which artifacts to build
	requestedArtifacts := determineRequestedArtifactsForBuild(module, buildAll, workspaceRoot)

	opts := builders.BuildOptions{
		TidyFirst:          tidyFirst,
		Version:            version,
		DryRun:             dryRun,
		RequestedArtifacts: requestedArtifacts,
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
	for _, ch := range compHandlers {
		handler := ch.Handler

		// Log which component/handler we're building
		if len(compHandlers) > 1 {
			output.Writeln(logWriter, "")
			output.Writeln(logWriter, "━━━ Building component: %s (handler: %s) ━━━", ch.Component, handler.Name())
		}

		// Validate module before building
		if err := handler.ValidateModule(module, workspaceRoot, ch.Component); err != nil {
			output.Writeln(logWriter, "❌ Module validation failed for %s: %v", ch.Component, err)
			return 1
		}

		exitCode := handler.Build(module, workspaceRoot, outputDir, logWriter, opts)
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

// verifyBuildDependenciesQuiet checks build dependencies silently and returns status for summary
// No output is written - the caller is responsible for displaying the results via InitSummary.
func verifyBuildDependenciesQuiet(monikers []string, moduleReport *reports.ModuleContractReport) (int, initsummary.DepsStatus) {
	status := initsummary.DepsStatus{Verified: true}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		// Config not loaded, skip verification
		return 0, status
	}

	// Collect unique build dependencies from all modules
	depsMap := make(map[string]bool)
	for _, moniker := range monikers {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			continue
		}

		// Get build requirements from package types
		enabledPackages := module.GetEnabledComponents()
		deps := cfg.ComponentTypes.GetBuildRequirements(enabledPackages)
		for _, dep := range deps {
			if dep != "" {
				depsMap[dep] = true
			}
		}
	}

	// No dependencies to verify
	if len(depsMap) == 0 {
		return 0, status
	}

	// Convert to sorted slice for consistent output
	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)
	status.Required = deps

	// Verify dependencies using tool registry
	registry := tool.GlobalRegistry()
	results := registry.VerifyAll(deps)

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
	log.Info("  --rebuild                 Force full rebuild, ignoring incremental build state")
	log.Info("  --layered-build           Execute layers sequentially (default: all modules in parallel)")
	log.Info("  --skip-depm               Only build specified modules, no module dependency resolution (CI isolation)")
	log.Info("  --use-existing-depm       Skip building module dependencies if artifacts exist (for CI incremental builds)")
	log.Info("  --skip-deps               Skip system dependency verification (go, docker, etc.)")
	log.Info("  --turbo                   Enable turbo mode for faster builds (increases parallelism)")
	log.Info("  --timings                 Show detailed timing summary")
	log.Info("  --debug                   Enable debug logs to console (file logging always enabled)")
	log.Info("  --tui                     Enable TUI console (default for local, errors in CI/container)")
	log.Info("  --no-tui                  Disable TUI console (use plain output)")
	log.Info(fmt.Sprintf("  --tui-height N            Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --ascii                   Use ASCII-only characters in TUI (for terminals with poor Unicode support)")
	log.Info("  --version VERSION         Inject version string into binary (Go modules with executable artifacts)")
	log.Info("  --accept-warnings         Don't fail on MkDocs warnings (non-strict mode)")
	log.Info("  --all                     Include non-default books (those with default: false)")
	log.Info("  -h, --help                Show this help message")
	log.Info("")
	log.Info("MkDocs modules with books (books.yml):")
	log.Info("  Books with 'default: false' are skipped unless --all is used.")
	log.Info("  Output is configured via the book's 'output' field:")
	log.Info("    site       - HTML site only")
	log.Info("    pdf-dark   - PDF with dark theme")
	log.Info("    pdf-light  - PDF with light theme")
	log.Info("    pdf-all    - Both dark and light PDFs")
	log.Info("")
	log.Info("Examples:")
	log.Info("  build                                # Build all modules")
	log.Info("  build r2r-cli                        # Build CLI for current platform")
	log.Info("  build r2r-cli --all                  # Build CLI for all platforms")
	log.Info("  build books                          # Build books (uses books.yml output config)")
	log.Info("  build r2r-cli --list-artifacts       # List artifacts without building")
}

// hasExistingArtifacts checks if a module's build artifacts already exist AND
// were built from the same source inputs.
// Used by --use-existing-depm to skip building modules whose artifacts are present
// (typically downloaded from previous CI runs).
func hasExistingArtifacts(moniker, moduleType, workspaceRoot string, buildAll bool) bool {
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

	// First, verify input hash matches (if manifest exists with input hash)
	// This ensures cached artifacts are from the same source code
	manifest, err := implinternal.LoadModuleManifest(buildDir)
	if err == nil && manifest.InputHash != "" {
		// Load module registry for hash computation
		modRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
		if err == nil {
			if contract, ok := modRegistry.Get(moniker); ok {
				currentHash, err := buildstate.ComputeModuleInputHash(workspaceRoot, contract)
				if err == nil && currentHash != manifest.InputHash {
					log.Debugf("Module %s: input hash mismatch (cached=%s, current=%s) - need rebuild",
						moniker, manifest.InputHash[:16], currentHash[:16])
					return false
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

// generateBuildManifest creates per-module manifest files tracking what was built.
// Each module gets its own immutable manifest at out/build/<module>/build.manifest.json.
func generateBuildManifest(workspaceRoot string, results []orchestrator.WorkResult, moduleTypes map[string]string, executionOrder []string, buildAll bool) error {
	// Get git commit SHA
	gitCommit := git.GetCommitSHA(workspaceRoot)

	// Load config for artifact resolution
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load module registry for input hash computation
	modRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		log.Warnf("Failed to load module registry for input hash: %v", err)
		// Continue without input hashes - not fatal
	}

	// Track current platform
	currentPlatform := implinternal.PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	manifestCount := 0

	// Process each successfully built module - create a per-module manifest
	for _, result := range results {
		// Skip failed builds
		if result.ExitCode != 0 {
			log.Debugf("Skipping %s: build failed (exit code %d)", result.Moniker, result.ExitCode)
			continue
		}

		moniker := result.Moniker
		moduleType, ok := moduleTypes[moniker]
		if !ok {
			log.Debugf("Skipping %s: not found in moduleTypes map", moniker)
			continue
		}

		// Get module config
		module, ok := cfg.Repository.GetModule(moniker)
		if !ok {
			continue
		}

		// Build directory for this module
		moduleBuildDir := cfg.Repository.BuildOutputPathAbs(workspaceRoot, moniker)

		// Resolve artifacts for current platform
		artifacts, _, err := implinternal.ResolveArtifactsForModuleWithConfig(
			module, moduleBuildDir, currentPlatform.OS, currentPlatform.Arch, cfg,
		)
		if err != nil {
			log.Warnf("Failed to resolve artifacts for %s: %v", moniker, err)
			continue
		}

		// Determine which artifacts were requested (filters out UPX in non-CI, etc.)
		requestedArtifactIDs := implinternal.DetermineRequestedArtifacts(module, buildAll, cfg)

		// Build set of requested IDs for fast lookup
		requestedSet := make(map[string]bool, len(requestedArtifactIDs))
		for _, id := range requestedArtifactIDs {
			requestedSet[id] = true
		}

		// Convert to ArtifactInfo, filtering to only include requested artifacts
		artifactInfos := make([]implinternal.ArtifactInfo, 0, len(artifacts))
		for i := range artifacts {
			art := &artifacts[i]
			// Only include artifacts that were actually requested
			if !requestedSet[art.ID] {
				continue
			}

			platform := ""
			if art.Type == "executable" {
				platform = fmt.Sprintf("%s-%s", currentPlatform.OS, currentPlatform.Arch)
			}

			// For image artifacts, use the image reference as the path
			// (images don't have file paths, they have tags/references)
			// For file-based artifacts, store relative path from build dir for portability
			artifactPath := art.ResolvedPath
			if art.Type == "image" && artifactPath == "" {
				artifactPath = art.ResolvedName // e.g., "ext-eac:latest"
			} else if art.Type != "image" && art.ResolvedPath != "" {
				// Make path relative to module build dir for manifest
				if relPath, err := filepath.Rel(moduleBuildDir, art.ResolvedPath); err == nil {
					artifactPath = relPath
				}
			}

			// Compute content hash for file-based artifacts (not images or directories)
			var size int64
			var sha256Hash string
			if art.Type != "image" && art.ResolvedPath != "" {
				// ResolvedPath is already absolute - use it directly for hashing
				if s, h, err := implinternal.HashArtifactFile(art.ResolvedPath); err == nil {
					size = s
					sha256Hash = h
				}
			}

			artifactInfo := implinternal.ArtifactInfo{
				Type:     art.Type,
				ID:       art.ID,
				Name:     art.ResolvedName,
				Path:     artifactPath,
				Platform: platform,
				Size:     size,
				SHA256:   sha256Hash,
			}

			// For image artifacts, enrich with docker_build config info
			if art.Type == "image" {
				enrichImageArtifact(&artifactInfo, module, moniker)
			}

			artifactInfos = append(artifactInfos, artifactInfo)
		}

		// Create per-module manifest (immutable - created once per build)
		manifest := implinternal.NewModuleManifest(moniker, moduleType, gitCommit)
		manifest.DurationSeconds = result.Duration.Seconds()
		manifest.RequestedArtifacts = requestedArtifactIDs
		manifest.Artifacts = artifactInfos
		manifest.Platforms = []implinternal.PlatformInfo{currentPlatform}

		// Compute input hash for CI cache validation
		if modRegistry != nil {
			if contract, ok := modRegistry.Get(moniker); ok {
				inputHash, err := buildstate.ComputeModuleInputHash(workspaceRoot, contract)
				if err != nil {
					log.Debugf("Failed to compute input hash for %s: %v", moniker, err)
				} else {
					manifest.InputHash = inputHash
				}
			}
		}

		// Collect all files in build output
		files, err := implinternal.CollectBuildFiles(moduleBuildDir)
		if err != nil {
			log.Warnf("Failed to collect build files for %s: %v", moniker, err)
		} else {
			manifest.Files = files
		}

		// Validate and save manifest to module's build directory
		if err := manifest.ValidateAndSaveWithRoot(moduleBuildDir, workspaceRoot); err != nil {
			return fmt.Errorf("failed to validate/save manifest for %s: %w", moniker, err)
		}

		manifestCount++
		log.Debugf("Generated manifest for module %s at %s", moniker, moduleBuildDir)
	}

	log.Debugf("Generated %d module manifests", manifestCount)
	return nil
}

// updateSkippedModuleManifests updates the VerifiedUnchangedAt field for modules
// that were skipped because they were already up-to-date.
func updateSkippedModuleManifests(workspaceRoot string, skippedModules []string, gitCommit string) {
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Warnf("Failed to load config for manifest update: %v", err)
		return
	}

	for _, moniker := range skippedModules {
		moduleBuildDir := cfg.Repository.BuildOutputPathAbs(workspaceRoot, moniker)
		manifest, err := implinternal.LoadModuleManifest(moduleBuildDir)
		if err != nil {
			// No manifest exists for this module - that's fine, skip it
			log.Debugf("No manifest for skipped module %s: %v", moniker, err)
			continue
		}

		if err := manifest.UpdateVerifiedUnchangedAt(moduleBuildDir, gitCommit); err != nil {
			log.Warnf("Failed to update verification status for %s: %v", moniker, err)
		} else {
			log.Debugf("Updated verification status for %s to %s", moniker, gitCommit)
		}
	}
}

// validateModuleBuildOutputs validates that a module's build produced the expected artifacts.
func validateModuleBuildOutputs(moniker, moduleType, workspaceRoot string, logWriter io.Writer, buildAll bool) error {
	// Load config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get module
	module, ok := cfg.Repository.GetModule(moniker)
	if !ok {
		return fmt.Errorf("module not found: %s", moniker)
	}

	// Determine which artifacts were requested for this build
	// This uses cfg.GetBuildArtifactIDs which correctly handles module-level artifacts
	requestedArtifacts := implinternal.DetermineRequestedArtifacts(module, buildAll, cfg)

	// If no artifacts to validate, nothing to do
	if len(requestedArtifacts) == 0 {
		fmt.Fprintf(logWriter, "  ℹ️  No artifacts defined for this module\n")
		return nil
	}

	// Resolve and validate artifacts
	// Note: BuildOutputPath returns a relative path, so we need to make it absolute
	buildDirRel := cfg.Repository.BuildOutputPath(moniker)
	buildDir := filepath.Join(workspaceRoot, buildDirRel)
	artifacts, _, err := implinternal.ResolveArtifactsForModuleWithConfig(
		module, buildDir, runtime.GOOS, runtime.GOARCH, cfg,
	)
	if err != nil {
		return fmt.Errorf("failed to resolve artifacts: %w", err)
	}

	// Check if this module has docker_build with push=true
	// If so, image artifacts are pushed to registry and may not exist locally
	dockerConfig := module.GetDockerBuildConfig()
	imagesPushedToRegistry := dockerConfig != nil && dockerConfig.Push

	// Filter artifacts to only those that were requested
	var requestedArtifactList []implinternal.ResolvedArtifact
	var requestedMissing, requestedTotal int

	for i := range artifacts {
		art := &artifacts[i]
		// Check if this artifact was requested
		isRequested := false
		for _, reqID := range requestedArtifacts {
			if art.ID == reqID {
				isRequested = true
				break
			}
		}

		if isRequested {
			requestedArtifactList = append(requestedArtifactList, *art)
			requestedTotal++

			// For image artifacts with push=true, trust that buildx push succeeded
			// (buildx would have failed if push failed). The image may not exist locally.
			if art.Type == "image" && imagesPushedToRegistry {
				// Image was pushed to registry - trust the build
				continue
			}

			if !art.Exists {
				requestedMissing++
			}
		}
	}

	// Check if any requested artifacts are missing
	if requestedMissing > 0 {
		fmt.Fprintf(logWriter, "\n❌ Expected artifacts were not created:\n")
		for i := range requestedArtifactList {
			art := &requestedArtifactList[i]
			if !art.Exists {
				// Skip image artifacts that were pushed to registry
				if art.Type == "image" && imagesPushedToRegistry {
					continue
				}
				// Include ❌ so log parser picks up each missing artifact for summary display
				fmt.Fprintf(logWriter, "  ❌ Missing: %s (%s)\n", art.ID, art.ResolvedPath)
			}
		}
		return fmt.Errorf("build succeeded but %d/%d artifacts missing", requestedMissing, requestedTotal)
	}

	// Success - all requested artifacts present
	fmt.Fprintf(logWriter, "  ✅ Validated %d artifact(s) created\n", requestedTotal)
	return nil
}

// determineRequestedArtifactsForBuild determines which artifact IDs should be built for a module.
func determineRequestedArtifactsForBuild(moduleContract *modules.ModuleContract, buildAll bool, workspaceRoot string) []string {
	// When --all is specified, return "*" to signal builders to include all artifacts
	// This handles cases where module type has no artifacts defined (like container type)
	// but module has books that should all be built
	if buildAll {
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
