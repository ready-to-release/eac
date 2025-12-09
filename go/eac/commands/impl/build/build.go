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
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/flock"
	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	"github.com/ready-to-release/eac/go/eac/commands/impl/show"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/buildstate"
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
	buildDir := filepath.Join(workspaceRoot, paths.OutBuildRelPath)
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
	skipDepsVerification := false // Skip system dependency verification (go, docker, etc.)
	skipDepm := false             // Skip including transitive module dependencies in execution plan
	useExistingDepm := false      // Use existing module dependency artifacts (skip building if present)
	forceRebuild := false         // Force full rebuild, ignoring incremental build state (--rebuild)
	showTimings := false
	debugMode := false // Enable debug logs to console
	version := ""
	listArtifacts := false
	dryRun := false
	buildAll := false        // Include non-default books (those with default: false)
	useTUI := isLocalConsole // TUI enabled by default for local console mode
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
		case "--skip-depm":
			skipDepm = true
		case "--use-existing-depm":
			useExistingDepm = true
		case "--rebuild":
			forceRebuild = true
		case "--skip-deps-verification":
			skipDepsVerification = true
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

	// Track if user explicitly requested specific modules (vs building all)
	explicitlyRequested := len(monikers) > 0

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

	// Run build (single or multiple modules) - phases are handled inside
	// Note: Logging is configured inside buildMultipleModules after TUI is initialized
	return buildMultipleModules(monikers, workspaceRoot, moduleReport, tidyFirst, tidyExplicitlySet, version, skipDepsVerification, skipDepm, useExistingDepm, forceRebuild, showTimings, debugMode, dryRun, buildAll, useTUI, tuiHeight, explicitlyRequested)
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

		handler := builders.GetHandlerForModule(module, module.Type)
		if handler == nil {
			// No handler for this module type
			continue
		}

		artifacts := handler.ListArtifacts(module, workspaceRoot)
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
func buildMultipleModules(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport, tidyFirst bool, tidyExplicitlySet bool, version string, skipDepsVerification bool, skipDepm bool, useExistingDepm bool, forceRebuild bool, showTimings bool, debugMode bool, dryRun bool, buildAll bool, useTUI bool, tuiHeight int, explicitlyRequested bool) int {
	// Build module type lookup for ALL modules (will be populated after execution plan)
	moduleTypes := make(map[string]string)

	// Configure orchestrator early so we can use it for Init phase output
	orchConfig := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        paths.OutBuildRelPath,
		LogFileName:          "build.log",
		OrchestratorLogName:  "orchestrator.log",
		ActionVerb:           "Building",
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

	// Track execution start time for summary
	executionStart := time.Now()

	// Initialize and start TUI if enabled (for Init phase output)
	if useTUI {
		if err := orch.Init(); err != nil {
			log.Errorf("Error initializing orchestrator: %v", err)
			return 1
		}
		orch.StartTUI()

		// Give TUI time to fully initialize and render first frame
		// before sending messages to prevent partial renders
		time.Sleep(50 * time.Millisecond)
	}

	// Configure unified logging system
	// - Debug logs ALWAYS go to file (out/logs/build/debug.log)
	// - Debug logs go to console only if debugMode=true
	// - TUI output if TUI is enabled
	var tuiWriter io.Writer
	if useTUI {
		tuiWriter = orch.GetTUIWriter(tui.PhaseInit)
	}
	if err := logging.ConfigureLogging(workspaceRoot, "build", debugMode, tuiWriter); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	log.Debugf("Build logging configured: debugMode=%v, useTUI=%v", debugMode, useTUI)

	// Helper to write output to console OR TUI Init phase
	writeInit := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		if useTUI {
			orch.SendInitLine(msg)
		} else {
			log.Info(msg)
		}
	}

	// writeInitSummary outputs the full initialization summary
	writeInitSummary := func(summary *initsummary.Summary) {
		var formatted string
		if useTUI {
			formatted = initsummary.FormatCompact(summary)
		} else {
			formatted = initsummary.FormatDetailed(summary)
		}
		for _, line := range strings.Split(strings.TrimSpace(formatted), "\n") {
			if line != "" {
				writeInit("%s", line)
			}
		}
	}

	// Calculate execution order to determine module dependencies early
	includeDepm := !skipDepm
	executionPlan, err := repository.CalculateExecutionOrder(monikers, workspaceRoot, includeDepm)
	if err != nil {
		log.Errorf("Failed to calculate execution order: %v", err)
		return 1
	}

	// Track total modules for summary
	totalModules := len(executionPlan.ExecutionOrder)

	// Calculate added depm (modules added as dependencies)
	var addedDepm []string
	if includeDepm {
		requestedSet := make(map[string]bool)
		for _, m := range monikers {
			requestedSet[m] = true
		}
		for _, m := range executionPlan.ExecutionOrder {
			if !requestedSet[m] {
				addedDepm = append(addedDepm, m)
			}
		}
	}

	// Incremental Build Detection (devbox only)
	// For local builds, detect which modules actually need rebuilding
	// CI always does full builds (controlled by --use-existing-depm for layer skipping)
	isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != ""
	useIncremental := !isCI && !forceRebuild && !dryRun
	log.Debugf("Incremental build detection: isCI=%v, forceRebuild=%v, dryRun=%v, useIncremental=%v", isCI, forceRebuild, dryRun, useIncremental)

	var skippedModules []string
	var incrementalInfo *initsummary.IncrementalInfo
	var incrementalDetectionTime time.Duration

	if useIncremental {
		log.Debugf("Getting source files for %d modules", len(executionPlan.ExecutionOrder))
		// Build modules map for change detection using shared interface
		modulesMap := make(map[string]buildstate.ModuleFileGetter)
		for _, moniker := range executionPlan.ExecutionOrder {
			if contract, ok := moduleReport.Registry.Get(moniker); ok {
				modulesMap[moniker] = contract
			}
		}

		log.Debugf("Detecting changes...")
		moduleFiles, err := buildstate.GetModuleSourceFiles(workspaceRoot, modulesMap)
		if err != nil {
			log.Debugf("Failed to get module source files: %v", err)
			log.Warnf("Failed to get module files: %v - falling back to full rebuild", err)
		}
		for moniker, files := range moduleFiles {
			log.Debugf("Module %s has %d source files for change detection", moniker, len(files))
		}

		var changeResult *buildstate.ChangeResult
		if err == nil {
			changeResult, err = buildstate.DetectChanges(workspaceRoot, moduleFiles)
		}
		if err != nil {
			// On error, fall back to full build
			log.Debugf("Incremental detection error: %v", err)
			log.Warnf("Incremental detection failed: %v - falling back to full rebuild", err)
		} else if changeResult.FreshBuild {
			log.Debugf("Fresh build detected (no prior state)")
			incrementalInfo = &initsummary.IncrementalInfo{
				Enabled:    true,
				FreshBuild: true,
			}
		} else {
			log.Debugf("Change detection result: %d changed, %d up-to-date, detection time=%v",
				len(changeResult.ChangedModules), len(changeResult.UpToDateModules), changeResult.DetectionTime)
			incrementalDetectionTime = changeResult.DetectionTime

			// Build set of explicitly requested modules - these should always be built
			requestedSet := make(map[string]bool)
			if explicitlyRequested {
				for _, m := range monikers {
					requestedSet[m] = true
				}
			}

			// Filter execution plan to only changed modules
			// But don't skip if user explicitly requested specific modules
			if len(changeResult.ChangedModules) == 0 && !explicitlyRequested {
				// All modules up-to-date and no specific modules requested - output summary and exit
				initSummary := initsummary.New("build").
					SetRequest(monikers, executionPlan.ExecutionOrder).
					SetExecutionPlan(executionPlan.Layers).
					SetExecutionContext(string(logging.GetExecutionContext())).
					SetIncremental(&initsummary.IncrementalInfo{
						Enabled:       true,
						DetectionTime: changeResult.DetectionTime,
						Changed:       []string{},
						UpToDate:      changeResult.UpToDateModules,
					}).
					SetOutputDir(paths.OutBuildRelPath + "/")

				writeInitSummary(initSummary)
				writeInit("")
				writeInit("✅ All modules up-to-date (nothing to build)")
				writeInit("   Use --rebuild to force a full rebuild")

				// Stop TUI cleanly before exiting
				if useTUI {
					orch.StopTUI()
				}
				return 0
			}

			// Propagate changes through dependency graph:
			// If module A changed and module B depends on A, then B also needs rebuilding
			// This ensures dependents are rebuilt when their dependencies change
			propagatedModules := propagateChangesToDependents(changeResult.ChangedModules, moduleReport.Registry)
			log.Debugf("Change propagation: %d direct changes -> %d total (including dependents)",
				len(changeResult.ChangedModules), len(propagatedModules))

			// Filter skipped modules to exclude explicitly requested ones AND propagated ones
			propagatedSet := make(map[string]bool)
			for _, m := range propagatedModules {
				propagatedSet[m] = true
			}
			for _, m := range changeResult.UpToDateModules {
				if !requestedSet[m] && !propagatedSet[m] {
					skippedModules = append(skippedModules, m)
				}
			}

			// Update execution plan to only include changed modules + propagated dependents + explicitly requested
			changedSet := make(map[string]bool)
			for _, m := range propagatedModules {
				changedSet[m] = true
			}
			// Always include explicitly requested modules
			for _, m := range monikers {
				changedSet[m] = true
			}

			// Filter execution order to only include modules that need building
			var filteredOrder []string
			var filteredLayers [][]string
			for _, layer := range executionPlan.Layers {
				var filteredLayer []string
				for _, m := range layer {
					if changedSet[m] {
						filteredLayer = append(filteredLayer, m)
						filteredOrder = append(filteredOrder, m)
					}
				}
				if len(filteredLayer) > 0 {
					filteredLayers = append(filteredLayers, filteredLayer)
				}
			}
			executionPlan.ExecutionOrder = filteredOrder
			executionPlan.Layers = filteredLayers
			totalModules = len(filteredOrder)

			incrementalInfo = &initsummary.IncrementalInfo{
				Enabled:       true,
				DetectionTime: incrementalDetectionTime,
				Changed:       filteredOrder, // Includes both changed and explicitly requested
				UpToDate:      skippedModules,
			}
		}
	} else if forceRebuild {
		// Clear build state to ensure fresh state after this build
		buildstate.ClearState(workspaceRoot)
		incrementalInfo = &initsummary.IncrementalInfo{
			Enabled: false, // Disabled due to --rebuild
		}
	}

	// Dependency Verification (system dependencies like go, docker, etc.)
	// Use the expanded execution order to verify deps for ALL modules that will be built
	var depsStatus initsummary.DepsStatus
	if skipDepsVerification {
		depsStatus = initsummary.DepsStatus{Skipped: true}
	} else {
		var exitCode int
		exitCode, depsStatus = verifyBuildDependenciesQuiet(executionPlan.ExecutionOrder, moduleReport)
		if exitCode != 0 {
			// Output summary showing what failed before returning
			initSummary := initsummary.New("build").
				SetRequest(monikers, executionPlan.ExecutionOrder).
				SetExecutionPlan(executionPlan.Layers).
				SetExecutionContext(string(logging.GetExecutionContext())).
				SetDepsStatus(depsStatus).
				SetOutputDir(paths.OutBuildRelPath + "/")
			writeInitSummary(initSummary)
			log.Errorf("")
			log.Errorf("❌ Required build dependencies are missing: %s", strings.Join(depsStatus.Missing, ", "))
			log.Errorf("   Use --skip-deps-verification to bypass this check")
			return exitCode
		}
	}

	// Build the structured initialization summary
	initSummary := initsummary.New("build").
		SetRequest(monikers, executionPlan.ExecutionOrder).
		SetExecutionPlan(executionPlan.Layers).
		SetExecutionContext(string(logging.GetExecutionContext())).
		SetFlags(initsummary.Flags{
			TidyFirst:            tidyFirst,
			TidyExplicit:         tidyExplicitlySet,
			SkipDepm:             skipDepm,
			UseExistingDepm:      useExistingDepm,
			SkipDepsVerification: skipDepsVerification,
			ForceRebuild:         forceRebuild,
			DryRun:               dryRun,
			BuildAll:             buildAll,
			ShowTimings:          showTimings,
			DebugMode:            debugMode,
			UseTUI:               useTUI,
			Version:              version,
		}).
		SetDepmStatus(initsummary.DepmStatus{
			Verified: includeDepm,
			Skipped:  skipDepm,
			Total:    len(addedDepm),
			Resolved: addedDepm,
		}).
		SetDepsStatus(depsStatus).
		SetOutputDir(paths.OutBuildRelPath + "/")

	// Set incremental info if applicable
	if incrementalInfo != nil {
		initSummary.SetIncremental(incrementalInfo)
	}

	// Output the structured initialization summary
	writeInitSummary(initSummary)

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

		// With --use-existing-depm: skip building if module dependency artifacts already exist
		// This enables incremental CI where dependencies are downloaded from previous runs
		if useExistingDepm && !dryRun {
			if hasExistingArtifacts(moniker, moduleTypes[moniker], workspaceRoot, buildAll) {
				fmt.Fprintf(logWriter, "⏭️  Skipping %s (module dependency artifacts exist)\n", moniker)
				return 0
			}
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
		exitCode := runModuleBuild(module, workspaceRoot, moduleOutputDir, logWriter, tidyFirst, version, dryRun, buildAll)

		// If build succeeded and not dry-run, validate artifacts were created
		if exitCode == 0 && !dryRun {
			modType, ok := moduleTypes[moniker]
			if ok {
				if err := validateModuleBuildOutputs(moniker, modType, workspaceRoot, logWriter, buildAll); err != nil {
					fmt.Fprintf(logWriter, "\n❌ Build artifact validation failed: %v\n", err)
					return 1
				}
			}
		}

		return exitCode
	}
	orch.SetWorker(worker)

	// Run orchestrator with layered execution (TUI transitions to Run phase automatically)
	results, err := orch.RunLayered(executionPlan.Layers)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Generate build manifest with successfully built modules
	// Manifests are also generated in dry-run mode to record what would be built
	if err := generateBuildManifest(workspaceRoot, results, moduleTypes, executionPlan.ExecutionOrder, buildAll); err != nil {
		log.Warnf("Failed to generate build manifest: %v", err)
	}

	// Update verification timestamp for skipped (unchanged) modules
	if !dryRun && len(skippedModules) > 0 {
		gitCommit := getGitCommitSHA(workspaceRoot)
		updateSkippedModuleManifests(workspaceRoot, skippedModules, gitCommit)
	}

	// Update incremental build state for successfully built modules (devbox only)
	// Also update after --rebuild so next build can use incremental detection
	if (useIncremental || forceRebuild) && !dryRun {
		log.Debugf("Incremental mode: updating build state")

		// Collect successfully built modules
		var successfulModules []string
		for _, result := range results {
			if result.ExitCode == 0 {
				successfulModules = append(successfulModules, result.Moniker)
			}
		}
		log.Debugf("Successfully built modules: %v", successfulModules)

		// Also include skipped modules (they were already up-to-date)
		allSuccessful := append(successfulModules, skippedModules...)
		log.Debugf("Total modules for state update (built + skipped): %d", len(allSuccessful))

		// Build modules map for state update using shared interface
		modulesMap := make(map[string]buildstate.ModuleFileGetter)
		for _, moniker := range allSuccessful {
			if contract, ok := moduleReport.Registry.Get(moniker); ok {
				modulesMap[moniker] = contract
			}
		}

		// Get module files for state update
		moduleFiles, err := buildstate.GetModuleSourceFiles(workspaceRoot, modulesMap)
		if err != nil {
			log.Warnf("Failed to get module files for state update: %v", err)
		} else {
			for moniker, files := range moduleFiles {
				log.Debugf("Module %s: %d source files", moniker, len(files))
			}

			if err := buildstate.UpdateModuleState(workspaceRoot, allSuccessful, moduleFiles); err != nil {
				log.Warnf("Failed to update build state: %v", err)
			} else {
				log.Debugf("Build state updated successfully")
			}
		}
	} else {
		log.Debugf("Skipping build state update (useIncremental=%v, dryRun=%v)", useIncremental, dryRun)
	}

	// Build and send summary data to TUI, then wait for user to exit
	if useTUI {
		buildSummary := buildTUISummary(results, time.Since(executionStart), totalModules, dryRun)
		orch.SendSummary(buildSummary)
		// Wait for user to press any key to exit
		orch.WaitTUI()
	} else {
		// Stop TUI and print plain text summary
		orch.StopTUI()
		orch.PrintSummary(results)
	}

	// Show rich timing analysis if requested
	if showTimings {
		log.Info("")
		// Get the list of modules that were built
		builtModules := executionPlan.ExecutionOrder

		// Display rich timing analysis for the modules that were built
		buildOutputDir := filepath.Join(workspaceRoot, paths.OutBuildRelPath)
		show.ShowBuildTimesForModules(builtModules, 10, buildOutputDir)
	}

	// Return exit code based on results
	return orchestrator.GetExitCode(results)
}

// buildTUISummary creates summary data for the TUI Summary pane
func buildTUISummary(results []orchestrator.WorkResult, totalTime time.Duration, totalModules int, dryRun bool) *tui.SummaryData {
	// Count successes and failures
	successCount := 0
	failCount := 0
	var failedModules []string

	for _, result := range results {
		if result.ExitCode == 0 {
			successCount++
		} else {
			failCount++
			failedModules = append(failedModules, result.Moniker)
		}
	}

	// Build init summary
	initSummary := fmt.Sprintf("%d modules prepared", totalModules)

	// Build run summary
	runSummary := fmt.Sprintf("%d/%d modules built", successCount, totalModules)
	if dryRun {
		runSummary += " (dry-run)"
	}

	// Build details
	var details []string
	if failCount > 0 {
		details = append(details, fmt.Sprintf("Failed: %d (%s)", failCount, strings.Join(failedModules, ", ")))
		details = append(details, "")  // Blank line

		// Add error messages from failed modules (top 5 per module)
		for _, result := range results {
			if result.ExitCode != 0 {
				if len(result.Errors) > 0 {
					// Show module name with log path
					details = append(details, fmt.Sprintf("%s (%s):", result.Moniker, result.LogPath))
					// Show up to 5 errors
					errorCount := len(result.Errors)
					if errorCount > 5 {
						errorCount = 5
					}
					for i := 0; i < errorCount; i++ {
						errMsg := result.Errors[i]
						// Truncate long error messages
						if len(errMsg) > 100 {
							errMsg = errMsg[:97] + "..."
						}
						details = append(details, fmt.Sprintf("  • %s", errMsg))
					}
					if len(result.Errors) > 5 {
						details = append(details, fmt.Sprintf("  ...and %d more errors", len(result.Errors)-5))
					}
				}
			}
		}
	}
	details = append(details, "")  // Blank line
	details = append(details, "Output: out/build/")

	// Build next steps
	nextSteps := ""
	if failCount > 0 {
		nextSteps = fmt.Sprintf("Review full logs: out/build/%s/build.log", failedModules[0])
	} else if !dryRun {
		nextSteps = "Run 'test' to verify builds"
	}

	return &tui.SummaryData{
		Success:     failCount == 0,
		TotalTime:   totalTime,
		InitSummary: initSummary,
		RunSummary:  runSummary,
		Details:     details,
		NextSteps:   nextSteps,
	}
}

// runModuleBuild runs build for a single module
func runModuleBuild(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, tidyFirst bool, version string, dryRun bool, buildAll bool) int {
	// Get handler for module (checks per-module handler first, then type)
	handler := builders.GetHandlerForModule(module, module.Type)
	if handler == nil {
		writeln(logWriter, "❌ No handler found for module type: %s", module.Type)
		return 1
	}

	// Validate module before building
	if err := handler.ValidateModule(module, workspaceRoot); err != nil {
		writeln(logWriter, "❌ Module validation failed: %v", err)
		return 1
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
		writeln(logWriter, "Build: %s (dry-run)", module.Moniker)
		writeln(logWriter, "Type: %s", module.Type)
		writeln(logWriter, "Handler: %s", handler.Name())
		writeln(logWriter, "Root: %s", module.Files.Root)
		writeln(logWriter, "")
		writeln(logWriter, "Dry-run mode: skipping actual build")
		return 0
	}

	exitCode := handler.Build(module, workspaceRoot, outputDir, logWriter, opts)
	if exitCode != 0 {
		return exitCode
	}

	// Process artifact derivations (compression, etc.) if build succeeded
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		moduleTypeDef := cfg.ModuleTypes.Get(module.Type)
		if moduleTypeDef != nil {
			if err := ProcessArtifactDerivations(module.Moniker, moduleTypeDef, outputDir, opts.RequestedArtifacts, module.Metadata, logWriter); err != nil {
				writeln(logWriter, "❌ Artifact derivation failed: %v", err)
				return 1
			}
		}
	}

	// Execute post-build steps if build succeeded
	return builders.ExecutePostBuildSteps(module.Type, module.Moniker, workspaceRoot, outputDir, logWriter)
}

// verifyBuildDependenciesQuiet checks build dependencies silently and returns status for summary
// No output is written - the caller is responsible for displaying the results via InitSummary
func verifyBuildDependenciesQuiet(monikers []string, moduleReport *reports.ModuleContractReport) (int, initsummary.DepsStatus) {
	status := initsummary.DepsStatus{Verified: true}

	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
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

		deps := cfg.ModuleTypes.GetBuildDepsFromCapabilities(module.Type, cfg.SystemDependencies)
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

	// Verify all dependencies
	results := systemdeps.VerifyAll(deps)

	for _, result := range results {
		status.Available = append(status.Available, initsummary.DepsResult{
			Name:      result.Name,
			Available: result.Available,
			Version:   result.Version,
		})
		if !result.Available {
			status.Missing = append(status.Missing, result.Moniker)
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
	log.Info("  --skip-depm               Only build specified modules (skip transitive module dependencies)")
	log.Info("  --use-existing-depm       Skip building module dependencies if artifacts exist (for CI incremental builds)")
	log.Info("  --skip-deps-verification  Skip system dependency verification (go, docker, etc.)")
	log.Info("  --timings                 Show detailed timing summary")
	log.Info("  --debug                   Enable debug logs to console (file logging always enabled)")
	log.Info("  --tui                     Enable TUI console (default for local, errors in CI/container)")
	log.Info("  --no-tui                  Disable TUI console (use plain output)")
	log.Info(fmt.Sprintf("  --tui-height N            Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --version VERSION         Inject version string into binary (go-cli only)")
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
// (typically downloaded from previous CI runs)
func hasExistingArtifacts(moniker, moduleType, workspaceRoot string, buildAll bool) bool {
	// Load config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return false
	}

	// Get module
	module, ok := cfg.Modules.GetModule(moniker)
	if !ok {
		return false
	}

	// Get module type definition
	moduleTypeDef := cfg.ModuleTypes.Get(moduleType)
	if moduleTypeDef == nil {
		return false
	}

	// If no artifacts defined, consider it as "exists" (nothing to build)
	if moduleTypeDef.Build == nil || len(moduleTypeDef.Build.Artifacts) == 0 {
		return true
	}

	// Determine which artifacts are requested
	requestedArtifacts := implinternal.DetermineRequestedArtifacts(module, moduleTypeDef, buildAll, cfg)
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
		registry, err := modules.LoadFromWorkspace(workspaceRoot)
		if err == nil {
			if contract, ok := registry.Get(moniker); ok {
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
		module, moduleTypeDef, buildDir, runtime.GOOS, runtime.GOARCH, cfg,
	)
	if err != nil {
		return false
	}

	// Check if all requested artifacts exist
	for _, art := range artifacts {
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
// Each module gets its own immutable manifest at out/build/<module>/build.manifest.json
func generateBuildManifest(workspaceRoot string, results []orchestrator.WorkResult, moduleTypes map[string]string, executionOrder []string, buildAll bool) error {
	// Get git commit SHA
	gitCommit := getGitCommitSHA(workspaceRoot)

	// Load config for artifact resolution
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load module registry for input hash computation
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
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
			continue
		}

		moniker := result.Moniker
		moduleType, ok := moduleTypes[moniker]
		if !ok {
			continue
		}

		// Get module config
		module, ok := cfg.Modules.GetModule(moniker)
		if !ok {
			continue
		}

		// Get module type definition
		moduleTypeDef := cfg.ModuleTypes.Get(moduleType)
		if moduleTypeDef == nil {
			continue
		}

		// Build directory for this module
		moduleBuildDir := cfg.Repository.BuildOutputPathAbs(workspaceRoot, moniker)

		// Resolve artifacts for current platform
		artifacts, _, err := implinternal.ResolveArtifactsForModuleWithConfig(
			module, moduleTypeDef, moduleBuildDir, currentPlatform.OS, currentPlatform.Arch, cfg,
		)
		if err != nil {
			log.Warnf("Failed to resolve artifacts for %s: %v", moniker, err)
			continue
		}

		// Determine which artifacts were requested (filters out UPX in non-CI, etc.)
		requestedArtifactIDs := implinternal.DetermineRequestedArtifacts(module, moduleTypeDef, buildAll, cfg)

		// Build set of requested IDs for fast lookup
		requestedSet := make(map[string]bool, len(requestedArtifactIDs))
		for _, id := range requestedArtifactIDs {
			requestedSet[id] = true
		}

		// Convert to ArtifactInfo, filtering to only include requested artifacts
		artifactInfos := make([]implinternal.ArtifactInfo, 0, len(artifacts))
		for _, art := range artifacts {
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
			artifactPath := art.ResolvedPath
			if art.Type == "image" && artifactPath == "" {
				artifactPath = art.ResolvedName // e.g., "ext-eac:latest"
			}

			// Compute content hash for file-based artifacts (not images or directories)
			var size int64
			var sha256Hash string
			if art.Type != "image" && art.ResolvedPath != "" {
				absPath := filepath.Join(workspaceRoot, art.ResolvedPath)
				if s, h, err := implinternal.HashArtifactFile(absPath); err == nil {
					size = s
					sha256Hash = h
				}
			}

			artifactInfos = append(artifactInfos, implinternal.ArtifactInfo{
				Type:     art.Type,
				ID:       art.ID,
				Name:     art.ResolvedName,
				Path:     artifactPath,
				Platform: platform,
				Size:     size,
				SHA256:   sha256Hash,
			})
		}

		// Create per-module manifest (immutable - created once per build)
		manifest := implinternal.NewModuleManifest(moniker, moduleType, gitCommit)
		manifest.RequestedArtifacts = requestedArtifactIDs
		manifest.Artifacts = artifactInfos
		manifest.Platforms = []implinternal.PlatformInfo{currentPlatform}

		// Compute input hash for CI cache validation
		if registry != nil {
			if contract, ok := registry.Get(moniker); ok {
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
		if err := manifest.ValidateAndSave(moduleBuildDir); err != nil {
			return fmt.Errorf("failed to validate/save manifest for %s: %w", moniker, err)
		}

		manifestCount++
		log.Debugf("Generated manifest for module %s at %s", moniker, moduleBuildDir)
	}

	log.Debugf("Generated %d module manifests", manifestCount)
	return nil
}

// updateSkippedModuleManifests updates the VerifiedUnchangedAt field for modules
// that were skipped because they were already up-to-date
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

// getGitCommitSHA gets the current git commit SHA
func getGitCommitSHA(workspaceRoot string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workspaceRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// propagateChangesToDependents expands a list of changed modules to include all
// modules that transitively depend on them. This ensures that when a library changes,
// all executables/modules that use it are also rebuilt.
//
// Example: If eac-core changes and eac-commands depends on eac-core,
// this returns [eac-core, eac-commands] (plus any other dependents).
func propagateChangesToDependents(changedModules []string, registry *modules.Registry) []string {
	if len(changedModules) == 0 {
		return changedModules
	}

	// Build reverse dependency graph: module -> modules that depend on it
	dependentsGraph := registry.GetReverseDependencyGraph()

	// Start with directly changed modules
	result := make(map[string]bool)
	for _, m := range changedModules {
		result[m] = true
	}

	// Recursively add all transitive dependents
	for _, changedModule := range changedModules {
		addTransitiveDependentsLocal(changedModule, dependentsGraph, result)
	}

	// Convert to sorted slice for consistent output
	modules := make([]string, 0, len(result))
	for m := range result {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	return modules
}

// addTransitiveDependentsLocal recursively adds all modules that depend on the given module
func addTransitiveDependentsLocal(module string, dependentsGraph map[string][]string, result map[string]bool) {
	for _, dependent := range dependentsGraph[module] {
		if !result[dependent] {
			result[dependent] = true
			addTransitiveDependentsLocal(dependent, dependentsGraph, result)
		}
	}
}

// validateModuleBuildOutputs validates that a module's build produced the expected artifacts
func validateModuleBuildOutputs(moniker, moduleType, workspaceRoot string, logWriter io.Writer, buildAll bool) error {
	// Load config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get module
	module, ok := cfg.Modules.GetModule(moniker)
	if !ok {
		return fmt.Errorf("module not found: %s", moniker)
	}

	// Get module type definition
	moduleTypeDef := cfg.ModuleTypes.Get(moduleType)
	if moduleTypeDef == nil {
		return fmt.Errorf("module type not found: %s", moduleType)
	}

	// If no artifacts defined, nothing to validate
	if moduleTypeDef.Build == nil || len(moduleTypeDef.Build.Artifacts) == 0 {
		fmt.Fprintf(logWriter, "  ℹ️  No artifacts defined for this module type\n")
		return nil
	}

	// Determine which artifacts were requested for this build
	requestedArtifacts := implinternal.DetermineRequestedArtifacts(module, moduleTypeDef, buildAll, cfg)

	// Resolve and validate artifacts
	// Note: BuildOutputPath returns a relative path, so we need to make it absolute
	buildDirRel := cfg.Repository.BuildOutputPath(moniker)
	buildDir := filepath.Join(workspaceRoot, buildDirRel)
	artifacts, _, err := implinternal.ResolveArtifactsForModuleWithConfig(
		module, moduleTypeDef, buildDir, runtime.GOOS, runtime.GOARCH, cfg,
	)
	if err != nil {
		return fmt.Errorf("failed to resolve artifacts: %w", err)
	}

	// Filter artifacts to only those that were requested
	var requestedArtifactList []implinternal.ResolvedArtifact
	var requestedMissing, requestedTotal int

	for _, art := range artifacts {
		// Check if this artifact was requested
		isRequested := false
		for _, reqID := range requestedArtifacts {
			if art.ID == reqID {
				isRequested = true
				break
			}
		}

		if isRequested {
			requestedArtifactList = append(requestedArtifactList, art)
			requestedTotal++
			if !art.Exists {
				requestedMissing++
			}
		}
	}

	// Check if any requested artifacts are missing
	if requestedMissing > 0 {
		fmt.Fprintf(logWriter, "\n❌ Expected artifacts were not created:\n")
		for _, art := range requestedArtifactList {
			if !art.Exists {
				fmt.Fprintf(logWriter, "  - %s: %s\n", art.ID, art.ResolvedPath)
			}
		}
		return fmt.Errorf("build succeeded but %d/%d artifacts missing", requestedMissing, requestedTotal)
	}

	// Success - all requested artifacts present
	fmt.Fprintf(logWriter, "  ✅ Validated %d artifact(s) created\n", requestedTotal)
	return nil
}

// determineRequestedArtifactsForBuild determines which artifact IDs should be built for a module
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
	module, ok := cfg.Modules.GetModule(moduleContract.Moniker)
	if !ok {
		log.Debugf("Module %s not found in config", moduleContract.Moniker)
		return []string{}
	}

	// Get module type definition
	moduleType := cfg.ModuleTypes.Get(module.Type)
	if moduleType == nil {
		log.Debugf("Module type %s not found", module.Type)
		return []string{}
	}

	// Use the DetermineRequestedArtifacts function
	return implinternal.DetermineRequestedArtifacts(module, moduleType, false, cfg)
}

