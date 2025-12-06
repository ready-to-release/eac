// Command: build
// Short: Build one or more modules by moniker
// Long: Build one or more modules by moniker.
// Long:
// Long: This command builds modules respecting their dependency order.
// Long: If no monikers are specified, all modules in the repository are built.
// Long:
// Long: Build results are collected in 'out/build/' with per-module logs and
// Long: a summary orchestrator log. Failed builds are clearly marked but do not
// Long: stop the execution of remaining modules.
// Long:
// Long: Example:
// Long:   build                           # Build all modules
// Long:   build eac-commands              # Build a single module
// Long:   build eac-core r2r-cli          # Build specific modules
// Long:   build --tidy-first eac-commands # Build with go mod tidy first
// Args: modules
package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gofrs/flock"
	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	"github.com/ready-to-release/eac/go/eac/commands/impl/show"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	systemdeps "github.com/ready-to-release/eac/go/eac/core/system-deps"
)

var log = logging.C()

// writeln writes a formatted string with platform-specific line ending to the writer
func writeln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

func init() {
	registry.Register(Build)
}

// BuildResult captures the outcome of a module build
type BuildResult struct {
	Moniker  string
	ExitCode int
	Warnings []string
	Errors   []string
}

// acquireModuleBuildLock attempts to acquire an exclusive lock for building a module.
// Returns the lock handle and nil error on success.
// Returns nil and error if lock is already held (module is being built).
func acquireModuleBuildLock(moniker, workspaceRoot string) (*flock.Flock, error) {
	// Ensure out/build directory exists (parent directory for lock files)
	buildDir := filepath.Join(workspaceRoot, "out", "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create build directory: %w", err)
	}

	// Create lock file path in parent directory (so it survives directory purge)
	// Use module moniker as the mutex identifier
	lockPath := filepath.Join(buildDir, fmt.Sprintf(".lock-%s", moniker))

	// Create flock instance
	lock := flock.New(lockPath)

	// Try to acquire lock (non-blocking)
	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		return nil, fmt.Errorf("module '%s' is already being built", moniker)
	}

	return lock, nil
}

// releaseModuleBuildLock releases the lock and removes the lock file
func releaseModuleBuildLock(lock *flock.Flock) {
	if lock == nil {
		return
	}

	lockPath := lock.Path()
	lock.Unlock()

	// Clean up the lock file
	os.Remove(lockPath)
}

// Build command entry point - builds one or more modules
func Build() int {
	args := os.Args[2:] // Skip program name and "build"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printBuildUsage()
		return 0
	}

	// Detect execution environment
	isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != ""
	isContainer := logging.GetExecutionContext() == logging.ContextR2RCLI
	// Detect test context: R2R_TEST_RUN_ID (test runner), GODOG_FORMAT (godog tests), R2R_MOCK_SECURITY (spec test subprocess)
	isTestContext := os.Getenv("R2R_TEST_RUN_ID") != "" || os.Getenv("GODOG_FORMAT") != "" || os.Getenv("R2R_MOCK_SECURITY") != ""
	isLocalConsole := !isCI && !isContainer && !isTestContext

	// Parse module monikers and flags
	var monikers []string
	tidyFirst := !isCI // Default: true for local, false for CI
	tidyExplicitlySet := false
	compressed := false
	compressedUPX := false
	skipVerification := false // Skip system dependency verification (go, docker, etc.)
	skipModuleDeps := false   // Skip including transitive module dependencies
	showTimings := false
	debugMode := false        // Enable debug logs to console
	version := ""
	listArtifacts := false
	dryRun := false
	buildAll := false         // Include non-default books (those with default: false)
	useTUI := isLocalConsole  // TUI enabled by default for local console mode
	tuiExplicitlySet := false
	tuiHeight := tui.DefaultHeight // TUI console height

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--tidy-first":
			tidyFirst = true
			tidyExplicitlySet = true
		case "--no-tidy":
			tidyFirst = false
			tidyExplicitlySet = true
		case "--compressed":
			compressed = true
		case "--compressed-upx":
			compressedUPX = true
			compressed = true // UPX implies stripped
		case "--skip-deps":
			skipModuleDeps = true
		case "--skip-verification":
			skipVerification = true
		case "--timings":
			showTimings = true
		case "--debug":
			debugMode = true
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
	if tuiExplicitlySet && useTUI && (isCI || isContainer) {
		if isCI {
			log.Errorf("Error: --tui cannot be used in CI environments")
		} else {
			log.Errorf("Error: --tui cannot be used in container/extension mode (use local console instead)")
		}
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
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

	// Enable debug logging if requested
	if debugMode {
		logging.EnableDebug()
	}

	// Run build (single or multiple modules) - phases are handled inside
	return buildMultipleModules(monikers, workspaceRoot, moduleReport, tidyFirst, tidyExplicitlySet, compressed, compressedUPX, version, skipVerification, skipModuleDeps, showTimings, debugMode, dryRun, buildAll, useTUI, tuiHeight)
}

// parseIntArg parses a string argument as an integer
func parseIntArg(s string) (int, error) {
	return strconv.Atoi(s)
}

// listModuleArtifacts lists the artifacts that would be produced by building the specified modules
func listModuleArtifacts(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport) int {
	// Sort monikers for consistent output
	sort.Strings(monikers)

	for _, moniker := range monikers {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			log.Errorf("Error: module not found: %s", moniker)
			continue
		}

		listFunc := builders.GetListArtifactsFunc(module.Type)
		if listFunc == nil {
			// No artifacts for this module type
			continue
		}

		artifacts := listFunc(module, workspaceRoot)
		outputDir := paths.BuildOutputPath(workspaceRoot, moniker)

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

	return 0
}


// buildMultipleModules builds multiple modules in parallel using the orchestrator
func buildMultipleModules(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport, tidyFirst bool, tidyExplicitlySet bool, compressed bool, compressedUPX bool, version string, skipVerification bool, skipModuleDeps bool, showTimings bool, debugMode bool, dryRun bool, buildAll bool, useTUI bool, tuiHeight int) int {
	// Build module type lookup for ALL modules (will be populated after execution plan)
	moduleTypes := make(map[string]string)

	// Configure orchestrator early so we can use it for Init phase output
	orchConfig := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        "out/build",
		LogFileName:          "build.log",
		OrchestratorLogName:  "orchestrator.log",
		ActionVerb:           "building",
		MaxConcurrency:       0, // Use default (number of CPUs)
		StatusUpdateInterval: 2, // Update every 2 seconds
		ModuleTypes:          moduleTypes,
		ShowTimings:          showTimings,
		DryRun:               dryRun,
		TUI:                  useTUI,
		TUIHeight:            tuiHeight,
	}

	// Create orchestrator early for phase management
	orch := orchestrator.New(orchConfig, nil) // Worker set later
	defer orch.Close()

	// Initialize and start TUI if enabled (for Init phase output)
	if useTUI {
		if err := orch.Init(); err != nil {
			log.Errorf("Error initializing orchestrator: %v", err)
			return 1
		}
		orch.StartTUI()
	}

	// Helper to write output to console OR TUI Init phase
	writeInit := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		if useTUI {
			orch.SendInitLine(msg)
		} else {
			log.Info(msg)
		}
	}

	// Show execution context
	writeInit("Executing build via %s. \"%s\"", logging.GetExecutionContext(), logging.GetFullCommand())
	writeInit("")

	// Phase 1: Module Discovery
	writeInit(output.PhaseHeader(1, "Module Discovery"))
	writeInit("Requested: %d modules:%s", len(monikers), output.ListFormat(monikers, 60, 5))

	// Calculate execution order to determine dependencies early
	includeDependencies := !skipModuleDeps
	executionPlan, err := repository.CalculateExecutionOrder(monikers, workspaceRoot, includeDependencies)
	if err != nil {
		log.Errorf("Failed to calculate execution order: %v", err)
		return 1
	}

	// Show dependencies if any were added
	if includeDependencies && len(executionPlan.ExecutionOrder) > len(monikers) {
		// Find which modules were added as dependencies
		requestedSet := make(map[string]bool)
		for _, m := range monikers {
			requestedSet[m] = true
		}
		var addedDeps []string
		for _, m := range executionPlan.ExecutionOrder {
			if !requestedSet[m] {
				addedDeps = append(addedDeps, m)
			}
		}
		writeInit("Dependencies: %d modules:%s", len(addedDeps), output.ListFormat(addedDeps, 60, 5))
		writeInit("Total: %d modules to build", len(executionPlan.ExecutionOrder))
	}

	// Check if any requested modules are Go modules (for tidy mode logging)
	hasGoModules := false
	for _, mon := range monikers {
		if module, exists := moduleReport.Registry.Get(mon); exists {
			if builders.IsGoModuleType(module.Type) {
				hasGoModules = true
				break
			}
		}
	}

	if hasGoModules {
		if tidyFirst {
			if tidyExplicitlySet {
				writeInit("Tidy mode: enabled (explicit flag)")
			} else {
				writeInit("Tidy mode: enabled (default for local builds)")
			}
		} else {
			if tidyExplicitlySet {
				writeInit("Tidy mode: disabled (explicit flag)")
			} else {
				writeInit("Tidy mode: disabled (CI environment detected)")
			}
		}
	}
	writeInit("")

	// Phase 2: Dependency Verification (system dependencies like go, docker, etc.)
	// Use the expanded execution order to verify deps for ALL modules that will be built
	if !skipVerification {
		if exitCode := verifyBuildDependencies(executionPlan.ExecutionOrder, moduleReport, writeInit); exitCode != 0 {
			return exitCode
		}
	}

	// Phase 3: Build Execution (still part of Init in TUI, but transitions to Run when execution starts)
	writeInit(output.PhaseHeader(3, "Build Execution"))
	writeInit(output.OutputDir("out/build/"))

	// Build module type lookup for ALL modules in execution plan (including dependencies)
	for _, mon := range executionPlan.ExecutionOrder {
		if module, exists := moduleReport.Registry.Get(mon); exists {
			moduleTypes[mon] = module.Type
		}
	}
	orch.SetModuleTypes(moduleTypes)

	// Create worker function that builds a single module and returns type info
	worker := func(moniker string, logWriter io.Writer) int {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			fmt.Fprintf(logWriter, "Error: module not found: %s\n", moniker)
			return 1
		}

		// Skip lock acquisition in dry-run mode since we're not actually building
		if !dryRun {
			// Acquire exclusive lock for this module FIRST (before any directory operations)
			lockFile, err := acquireModuleBuildLock(moniker, workspaceRoot)
			if err != nil {
				fmt.Fprintf(logWriter, "Error: module '%s' is already being built\n", moniker)
				fmt.Fprintf(logWriter, "Details: %v\n", err)
				return 1
			}
			defer releaseModuleBuildLock(lockFile)
		}

		moduleOutputDir := paths.BuildOutputPath(workspaceRoot, moniker)
		return runModuleBuild(module, workspaceRoot, moduleOutputDir, logWriter, tidyFirst, compressed, compressedUPX, version, dryRun, buildAll)
	}
	orch.SetWorker(worker)

	// Run orchestrator with layered execution (TUI transitions to Run phase automatically)
	results, err := orch.RunLayered(executionPlan.Layers)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Stop TUI first (restores stdout), then print full summary
	orch.StopTUI()
	orch.PrintSummary(results)

	// Show rich timing analysis if requested
	if showTimings {
		log.Info("")
		// Get the list of modules that were built
		builtModules := executionPlan.ExecutionOrder

		// Display rich timing analysis for the modules that were built
		buildOutputDir := filepath.Join(workspaceRoot, "out", "build")
		show.ShowBuildTimesForModules(builtModules, 10, buildOutputDir)
	}

	// Return exit code based on results
	return orchestrator.GetExitCode(results)
}

// runModuleBuild runs build for a single module
func runModuleBuild(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, tidyFirst bool, compressed bool, compressedUPX bool, version string, dryRun bool, buildAll bool) int {
	// Get build function for module type
	buildFunc := builders.GetBuildFunc(module.Type)

	opts := builders.BuildOptions{
		TidyFirst:     tidyFirst,
		Compressed:    compressed,
		CompressedUPX: compressedUPX,
		Version:       version,
		DryRun:        dryRun,
		BuildAll:      buildAll,
	}

	// In dry-run mode, simulate a successful build
	if dryRun {
		writeln(logWriter, "Build: %s (dry-run)", module.Moniker)
		writeln(logWriter, "Type: %s", module.Type)
		writeln(logWriter, "Root: %s", module.Files.Root)
		writeln(logWriter, "")
		writeln(logWriter, "Dry-run mode: skipping actual build")
		return 0
	}

	exitCode := buildFunc(module, workspaceRoot, outputDir, logWriter, opts)
	if exitCode != 0 {
		return exitCode
	}

	// Execute post-build steps if build succeeded
	return builders.ExecutePostBuildSteps(module.Type, module.Moniker, workspaceRoot, outputDir, logWriter)
}

// verifyBuildDependencies checks that all required build dependencies are available
func verifyBuildDependencies(monikers []string, moduleReport *reports.ModuleContractReport, writeFn func(format string, args ...interface{})) int {
	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		// Config not loaded, skip verification
		return 0
	}

	// Collect unique build dependencies from all modules
	depsMap := make(map[string]bool)
	for _, moniker := range monikers {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			continue
		}

		deps := cfg.ModuleTypes.GetBuildDeps(module.Type)
		for _, dep := range deps {
			if dep != "" {
				depsMap[dep] = true
			}
		}
	}

	// No dependencies to verify
	if len(depsMap) == 0 {
		return 0
	}

	// Convert to sorted slice for consistent output
	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)

	// Print phase header
	writeFn(output.PhaseHeader(2, "Dependency Verification"))
	writeFn("Required: %s", strings.Join(deps, ", "))

	// Verify all dependencies
	results := systemdeps.VerifyAll(deps)
	var missing []string

	for _, result := range results {
		writeFn(output.DependencyLine(result.Available, result.Name, result.Version))
		if !result.Available {
			missing = append(missing, result.Moniker)
		}
	}

	writeFn("")

	if len(missing) > 0 {
		log.Errorf("❌ Error: Required build dependencies are missing: %s", strings.Join(missing, ", "))
		log.Errorf("   Use --skip-verification to bypass this check")
		return 1
	}

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
	log.Info("  --skip-deps               Only build specified modules (skip transitive dependencies)")
	log.Info("  --skip-verification       Skip system dependency verification (go, docker, etc.)")
	log.Info("  --timings                 Show detailed timing summary")
	log.Info("  --debug                   Enable debug logs to console (file logging always enabled)")
	log.Info("  --tui                     Enable TUI console (default for local, errors in CI/container)")
	log.Info("  --no-tui                  Disable TUI console (use plain output)")
	log.Info(fmt.Sprintf("  --tui-height N            Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --compressed              Strip debug info for smaller binaries (go-cli only)")
	log.Info("  --compressed-upx          Also apply UPX compression for maximum size reduction")
	log.Info("  --version VERSION         Inject version string into binary (go-cli only)")
	log.Info("  --accept-warnings         Don't fail on MkDocs warnings (non-strict mode)")
	log.Info("  --all                     Include non-default books (those with default: false)")
	log.Info("  -h, --help                Show this help message")
	log.Info("")
	log.Info("Compression (go-cli only):")
	log.Info("  Default (dev):     Full debug info for debugging (~39 MB)")
	log.Info("  --compressed:      Strip debug info with -ldflags \"-s -w\" (~26 MB, ~30% smaller)")
	log.Info("  --compressed-upx:  Also UPX compress (~10 MB, ~70% smaller total)")
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
	log.Info("  build r2r-cli                        # Build CLI with debug info")
	log.Info("  build r2r-cli --compressed           # Build CLI for release")
	log.Info("  build books                          # Build books (uses books.yml output config)")
	log.Info("  build r2r-cli --list-artifacts       # List artifacts without building")
}
