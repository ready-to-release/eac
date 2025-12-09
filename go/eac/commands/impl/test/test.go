// Command: test
// Short: Test one or more modules by moniker
// Args: modules
// Long: Test one or more modules by moniker using suite-based filtering.
// Long:
// Long: This command discovers tests, applies inference rules (e.g., Go tests default to @L1),
// Long: filters by suite tags, and runs matching tests with consistent summary output.
// Long:
// Long: Use --suite to select which tests to run. The default suite is "commit" which
// Long: includes L0, L1, and L2 tests (fast tests for pre-commit/MR validation).
// Long:
// Long: Example:
// Long:   test eac-commands                    # Test single module
// Long:   test eac-core r2r-cli                # Test multiple modules
// Long:   test                                 # Test all modules
// Long:   test eac-commands --suite acceptance # Run acceptance tests only
// Flag.suite: type=string, usage=Filter tests by suite (default: "commit")
package test

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
	"sync"
	"time"

	"github.com/gofrs/flock"
	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/impl/show"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/runner"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/runners"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/buildstate"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	moduledeps "github.com/ready-to-release/eac/go/eac/core/module-deps"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/teststate"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	systemdeps "github.com/ready-to-release/eac/go/eac/core/system-deps"
	"github.com/ready-to-release/eac/go/eac/core/testing"
	"go.uber.org/zap/zapcore"
)

var log = logging.C()

func init() {
	registry.Register(Test)
}

// TestConfig holds test execution configuration
type TestConfig struct {
	Monikers     []string
	SuiteName    string
	ReportFormat string
	Coverage     bool
	SkipDeps     bool
	ListOnly     bool
	ShowTimings  bool
	DebugMode    bool
	UseTUI       bool
	TUIHeight    int
	Parallel     bool
	ForceRetest  bool // --retest flag to bypass incremental testing
}

// TestExecutionContext holds shared state for parallel test execution
type TestExecutionContext struct {
	testsByPackage  map[string][]testing.TestReference // Keyed by module path (e.g., eac-core/config)
	modulePathToPkg map[string]string                  // Maps module path -> original package path (e.g., go/eac/core/config)
	testParallelism int
	testRunDir      string
	reportFormat    string
	coverage        bool
	suiteTagFilter  string
	workspaceRoot   string
	moduleMapper    *ModuleMapper // Maps package paths to module monikers

	// Thread-safe result storage
	mu      sync.Mutex
	results map[string]PackageResult
}

// PackageResult holds test results for a single package
type PackageResult struct {
	PackageName   string
	LogFilePath   string
	TestsPassed   int
	TestsFailed   int
	TestsSkipped  int
	TestsTotal    int
	PackageFailed bool
	Duration      time.Duration
}

// Test is the unified entry point for testing modules
func Test() int {
	args := os.Args[2:] // Skip program name and "test"

	// Check for subcommands that should be handled separately
	if len(args) > 0 {
		switch args[0] {
		case "suite", "list-suites", "debug":
			// These are handled by their own registered commands
			return 0
		case "--help", "-h":
			printTestUsage()
			return 0
		}
	}

	// Parse arguments
	cfg := parseTestArgs(args)
	if cfg == nil {
		return 1
	}

	// Execute tests directly (like build command)
	return executeTests(cfg)
}

// parseTestArgs parses command line arguments into TestConfig
func parseTestArgs(args []string) *TestConfig {
	// Detect execution environment for TUI defaults
	isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != ""
	isContainer := logging.GetExecutionContext() == logging.ContextR2RCLI
	isLocalConsole := !isCI && !isContainer

	cfg := &TestConfig{
		SuiteName:    "commit",
		ReportFormat: "cucumber",
		TUIHeight:    tui.DefaultHeight,
		Parallel:     true,
		UseTUI:       isLocalConsole, // TUI enabled by default for local console mode
	}

	tuiExplicitlySet := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--suite":
			if i+1 >= len(args) {
				log.Errorf("--suite requires a suite name")
				return nil
			}
			i++
			cfg.SuiteName = args[i]
		case arg == "--as-junit":
			cfg.ReportFormat = "junit"
		case arg == "--as-cucumber":
			cfg.ReportFormat = "cucumber"
		case arg == "--coverage":
			cfg.Coverage = true
		case arg == "--skip-deps":
			cfg.SkipDeps = true
		case arg == "--list-only":
			cfg.ListOnly = true
		case arg == "--timings":
			cfg.ShowTimings = true
		case arg == "--debug":
			cfg.DebugMode = true
		case arg == "--tui":
			cfg.UseTUI = true
			tuiExplicitlySet = true
		case arg == "--no-tui":
			cfg.UseTUI = false
			tuiExplicitlySet = true
		case arg == "--sequential":
			cfg.Parallel = false
		case arg == "--retest":
			cfg.ForceRetest = true
		case arg == "--tui-height":
			if i+1 >= len(args) {
				log.Errorf("--tui-height requires a value")
				return nil
			}
			i++
			var err error
			cfg.TUIHeight, err = strconv.Atoi(args[i])
			if err != nil || cfg.TUIHeight < 3 || cfg.TUIHeight > 20 {
				log.Errorf("--tui-height must be a number between 3 and 20")
				return nil
			}
		case strings.HasPrefix(arg, "--tui-height="):
			heightStr := strings.TrimPrefix(arg, "--tui-height=")
			var err error
			cfg.TUIHeight, err = strconv.Atoi(heightStr)
			if err != nil || cfg.TUIHeight < 3 || cfg.TUIHeight > 20 {
				log.Errorf("--tui-height must be a number between 3 and 20")
				return nil
			}
		case strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"):
			log.Errorf("unknown flag: %s", arg)
			log.Errorf("Valid flags: --suite, --as-junit, --as-cucumber, --coverage, --skip-deps, --list-only, --timings, --debug, --tui, --no-tui, --tui-height, --sequential, --retest")
			return nil
		default:
			cfg.Monikers = append(cfg.Monikers, arg)
		}
	}

	// Validate TUI usage in CI/container environments
	if tuiExplicitlySet && cfg.UseTUI && (isCI || isContainer) {
		if isCI {
			log.Errorf("Error: --tui cannot be used in CI environments")
		} else {
			log.Errorf("Error: --tui cannot be used in container/extension mode (use local console instead)")
		}
		return nil
	}

	return cfg
}

// executeTests runs tests directly using orchestrator (like buildMultipleModules)
func executeTests(cfg *TestConfig) int {
	// Get workspace root first (needed for file logging)
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Configure logging for test command
	// Debug always goes to file, also to console if --debug flag set
	if err := logging.ConfigureLoggingSimple(workspaceRoot, "test", cfg.DebugMode); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	// Load EAC config for testing tags, suites, and repository paths
	eacCfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load EAC config: %v", err)
		return 1
	}

	// Use repository config from EAC config (properly merged with defaults)
	repoCfg := eacCfg.Repository

	// Load suite configuration
	suite, err := testing.GetSuite(cfg.SuiteName)
	if err != nil {
		log.Errorf("suite not found: %s", cfg.SuiteName)
		log.Info("Available suites:")
		for _, s := range eacCfg.TestSuites.Suites {
			log.Infof("  - %s", s.Moniker)
		}
		return 1
	}

	// Acquire suite lock to prevent concurrent executions
	lock, err := acquireTestLock(cfg.SuiteName, workspaceRoot, repoCfg)
	if err != nil {
		log.Errorf("Failed to acquire lock: %v", err)
		return 1
	}
	defer releaseTestLock(lock)

	// Create test output directory
	testRunDir := filepath.Join(workspaceRoot, repoCfg.TestOutputPath(cfg.SuiteName))
	if err := os.MkdirAll(testRunDir, 0755); err != nil {
		log.Errorf("failed to create test directory: %v", err)
		return 1
	}

	// Create log file for test output
	logFilePath := filepath.Join(testRunDir, "test-suite.log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Errorf("failed to create log file: %v", err)
		return 1
	}
	defer logFile.Close()

	// Multi-writer for console and log file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Configure orchestrator early for phase management
	maxConcurrency := 4
	if !cfg.Parallel {
		maxConcurrency = 1
	}
	orchConfig := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        repoCfg.TestOutputPath(cfg.SuiteName),
		LogFileName:          "test.log",
		OrchestratorLogName:  "orchestrator.log",
		ActionVerb:           "Testing",
		MaxConcurrency:       maxConcurrency,
		StatusUpdateInterval: 2,
		TUI:                  cfg.UseTUI,
		TUIHeight:            cfg.TUIHeight,
	}

	// Create orchestrator early for phase management
	orch := orchestrator.New(orchConfig, nil) // Worker set later
	defer orch.Close()

	// Initialize and start TUI if enabled (for Init phase output)
	if cfg.UseTUI {
		if err := orch.Init(); err != nil {
			log.Errorf("Error initializing orchestrator: %v", err)
			return 1
		}
		orch.StartTUI()

		// Configure component logger to send output to TUI Init pane
		if tuiWriter := orch.GetTUIWriter(tui.PhaseInit); tuiWriter != nil {
			// Determine which log levels should go to TUI based on debug mode
			tuiLevel := zapcore.InfoLevel
			if cfg.DebugMode {
				tuiLevel = zapcore.DebugLevel
			}

			// Enable TUI output for all component loggers
			if err := logging.EnableTUIForComponentLogger(tuiWriter, tuiLevel); err != nil {
				log.Errorf("Error enabling TUI for logger: %v", err)
			}
		}

		// Give TUI time to fully initialize and render first frame
		// before sending messages to prevent partial renders
		time.Sleep(50 * time.Millisecond)
	}

	// Helper to write output to console/log OR TUI Init phase
	writeInit := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		if cfg.UseTUI {
			orch.SendInitLine(msg)
		} else {
			writeln(multiWriter, "%s", msg)
		}
	}

	// writeInitSummary outputs the full initialization summary
	writeInitSummary := func(summary *initsummary.Summary) {
		var formatted string
		if cfg.UseTUI {
			formatted = initsummary.FormatCompact(summary)
		} else {
			formatted = initsummary.FormatDetailed(summary)
		}
		for _, line := range strings.Split(strings.TrimSpace(formatted), "\n") {
			writeInit("%s", line)
		}
	}

	// Test Discovery
	allTests, err := testing.DiscoverAllTests(workspaceRoot)
	if err != nil {
		log.Errorf("Failed to discover tests: %v", err)
		return 1
	}
	log.Debugf("Discovered %d tests", len(allTests))

	// Tag Inference
	allTests = testing.ApplyInferences(allTests, suite.Inferences)
	log.Debugf("Applied %d inference rules", len(suite.Inferences))

	// Suite Selection

	// Track selection stats
	type selectionStats struct {
		TotalDiscovered  int
		Skipped          int
		NotMatchingSuite int
		Selected         int
		OSFiltered       int
		ModulesRequested []string
		ModulesInScope   []string
		ModulesNoTests   []string
	}
	stats := selectionStats{
		TotalDiscovered:  len(allTests),
		ModulesRequested: cfg.Monikers,
	}

	// Filter tests by suite and modules
	var selectedTests []testing.TestReference
	skipReasons := eacCfg.TestingTags.GetSkipReasons()

	for _, test := range allTests {
		// Check for @skip tags
		isSkipped := false
		for _, tag := range test.Tags {
			for reason := range skipReasons {
				if tag == "@skip:"+reason {
					isSkipped = true
					break
				}
			}
			if isSkipped {
				break
			}
		}
		if isSkipped {
			stats.Skipped++
			continue
		}

		// Check if test matches suite
		if !suite.Matches(test) {
			stats.NotMatchingSuite++
			continue
		}

		// Check module filter if specified
		if len(cfg.Monikers) > 0 {
			matchesModule := false
			testModule := extractModuleFromPath(test.FilePath)
			for _, moniker := range cfg.Monikers {
				if testModule == moniker {
					matchesModule = true
					break
				}
			}
			if !matchesModule {
				continue
			}
		}

		selectedTests = append(selectedTests, test)
	}
	stats.Selected = len(selectedTests)

	// Filter by OS compatibility
	osCompatibleTests := filterByOSCompatibility(selectedTests, multiWriter)
	osFilteredCount := len(selectedTests) - len(osCompatibleTests)
	selectedTests = osCompatibleTests
	stats.OSFiltered = osFilteredCount

	// Calculate module stats
	stats.ModulesInScope = getUniqueModulesFromTests(selectedTests)

	// Find modules that have no tests
	inScopeSet := make(map[string]bool)
	for _, m := range stats.ModulesInScope {
		inScopeSet[m] = true
	}

	// Determine which modules to check for no-tests
	var modulesToCheck []string
	if len(cfg.Monikers) > 0 {
		modulesToCheck = cfg.Monikers
	} else {
		// Default: check all modules in config
		modulesToCheck = eacCfg.Modules.AllMonikers()
	}

	for _, m := range modulesToCheck {
		if !inScopeSet[m] {
			stats.ModulesNoTests = append(stats.ModulesNoTests, m)
		}
	}

	// If list-only, just show tests and exit
	if cfg.ListOnly {
		writeInit("=== Selected Tests ===")
		for i, test := range selectedTests {
			writeInit("%d. %s (%s)", i+1, test.TestName, test.Type)
			writeInit("   File: %s", test.FilePath)
			writeInit("   Tags: %s", strings.Join(test.Tags, ", "))
			writeInit("")
		}
		return 0
	}

	// Dependency Verification
	systemDeps := testing.GetSystemDependencies(selectedTests)
	moduleDeps := testing.GetModuleDependencies(selectedTests)

	var unavailableModuleDeps []string
	var depsStatus initsummary.DepsStatus

	if len(systemDeps) == 0 && len(moduleDeps) == 0 {
		depsStatus = initsummary.DepsStatus{Verified: true}
	} else {
		depsStatus = initsummary.DepsStatus{
			Verified: !cfg.SkipDeps,
			Skipped:  cfg.SkipDeps,
			Required: append(systemDeps, moduleDeps...),
		}

		if !cfg.SkipDeps {
			hasSystemFailures := false

			// Verify system dependencies
			sysResults := systemdeps.VerifyAll(systemDeps)
			for _, result := range sysResults {
				depsStatus.Available = append(depsStatus.Available, initsummary.DepsResult{
					Name:      result.Name,
					Available: result.Available,
					Version:   result.Version,
				})
				if !result.Available {
					hasSystemFailures = true
					depsStatus.Missing = append(depsStatus.Missing, result.Name)
				}
			}

			// Verify module dependencies
			modResults := moduledeps.VerifyAll(moduleDeps)
			for _, result := range modResults {
				depsStatus.Available = append(depsStatus.Available, initsummary.DepsResult{
					Name:      result.Dependency,
					Available: result.Available,
					Version:   result.Version,
				})
				if !result.Available {
					unavailableModuleDeps = append(unavailableModuleDeps, result.Dependency)
				}
			}

			if hasSystemFailures {
				// Output summary showing what failed
				initSummary := initsummary.New("test").
					SetRequest(cfg.Monikers, cfg.Monikers).
					SetExecutionContext(string(logging.GetExecutionContext())).
					SetDepsStatus(depsStatus).
					SetTestInfo(&initsummary.TestInfo{
						SuiteName:        suite.Name,
						SuiteDescription: suite.Description,
						TotalDiscovered:  stats.TotalDiscovered,
						Skipped:          stats.Skipped,
						NotMatchingSuite: stats.NotMatchingSuite,
						OSFiltered:       stats.OSFiltered,
						Selected:         len(selectedTests),
						ModulesRequested: stats.ModulesRequested,
						ModulesInScope:   stats.ModulesInScope,
						ModulesNoTests:   stats.ModulesNoTests,
					}).
					SetOutputDir(testRunDir)
				writeInitSummary(initSummary)
				writeInit("")
				writeInit("❌ Required system dependencies are missing: %s", strings.Join(depsStatus.Missing, ", "))
				writeInit("   Use --skip-deps to run tests anyway")
				return 1
			}
		}
	}

	// Load module registry early - needed for artifact validation and incremental testing
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		log.Warnf("Failed to load module registry: %v", err)
	}

	// Build Artifact Validation (includes staleness check)
	var artifactValidation *initsummary.ArtifactValidationInfo

	if len(stats.ModulesInScope) > 0 {
		artifactValidation = validateBuildArtifactsQuiet(stats.ModulesInScope, eacCfg, workspaceRoot, moduleRegistry)

		if !artifactValidation.AllValid() {
			// Output summary showing what failed
			initSummary := initsummary.New("test").
				SetRequest(cfg.Monikers, cfg.Monikers).
				SetExecutionContext(string(logging.GetExecutionContext())).
				SetDepsStatus(depsStatus).
				SetTestInfo(&initsummary.TestInfo{
					SuiteName:        suite.Name,
					SuiteDescription: suite.Description,
					TotalDiscovered:  stats.TotalDiscovered,
					Skipped:          stats.Skipped,
					NotMatchingSuite: stats.NotMatchingSuite,
					OSFiltered:       stats.OSFiltered,
					Selected:         len(selectedTests),
					ModulesRequested: stats.ModulesRequested,
					ModulesInScope:   stats.ModulesInScope,
					ModulesNoTests:   stats.ModulesNoTests,
				}).
				SetArtifactValidation(artifactValidation).
				SetOutputDir(testRunDir)
			writeInitSummary(initSummary)
			writeInit("")

			if !artifactValidation.AllPresent {
				writeInit("❌ Build artifacts are missing")
				for _, moduleName := range artifactValidation.MissingFrom {
					writeInit("  - %s", moduleName)
				}
			}

			if !artifactValidation.AllCurrent {
				if !artifactValidation.AllPresent {
					writeInit("")
				}
				writeInit("❌ Build artifacts are stale (source changed since build)")
				for _, moduleName := range artifactValidation.StaleModules {
					reason := artifactValidation.StaleReasons[moduleName]
					writeInit("  - %s: %s", moduleName, reason)
				}
			}

			writeInit("")
			writeInit("Resolution: Run 'build <module>' for each module to generate up-to-date artifacts")
			return 1
		}
	}

	// Build and output the structured initialization summary
	initSummary := initsummary.New("test").
		SetRequest(cfg.Monikers, cfg.Monikers).
		SetExecutionContext(string(logging.GetExecutionContext())).
		SetFlags(initsummary.Flags{
			ListOnly:    cfg.ListOnly,
			ShowTimings: cfg.ShowTimings,
			DebugMode:   cfg.DebugMode,
			UseTUI:      cfg.UseTUI,
		}).
		SetDepsStatus(depsStatus).
		SetTestInfo(&initsummary.TestInfo{
			SuiteName:             suite.Name,
			SuiteDescription:      suite.Description,
			TotalDiscovered:       stats.TotalDiscovered,
			Skipped:               stats.Skipped,
			NotMatchingSuite:      stats.NotMatchingSuite,
			OSFiltered:            stats.OSFiltered,
			Selected:              len(selectedTests),
			ModulesRequested:      stats.ModulesRequested,
			ModulesInScope:        stats.ModulesInScope,
			ModulesNoTests:        stats.ModulesNoTests,
			InferenceRulesApplied: len(suite.Inferences),
		}).
		SetOutputDir(testRunDir)

	if artifactValidation != nil {
		initSummary.SetArtifactValidation(artifactValidation)
	}

	// Output the structured initialization summary
	writeInitSummary(initSummary)

	// Group tests by package
	testsByPackage := groupTestsByPackage(selectedTests, workspaceRoot, eacCfg)

	if len(testsByPackage) == 0 {
		writeInit("No test packages to execute")
		writeInit("")
		return 0
	}

	// Incremental Test Detection
	// (moduleRegistry already loaded earlier for artifact validation)
	// Only enabled for local (devbox) mode, not CI
	isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != ""
	useIncremental := !isCI && !cfg.ForceRetest && moduleRegistry != nil

	log.Debugf("Incremental test setup: isCI=%v, forceRetest=%v, registryNil=%v, useIncremental=%v",
		isCI, cfg.ForceRetest, moduleRegistry == nil, useIncremental)

	// Track which modules need testing and which are skipped
	var modulesNeedingTest map[string]bool
	var incrementalInfo *incrementalTestInfo

	if useIncremental {
		log.Debugf("Incremental test detection: isCI=%v, forceRetest=%v, useIncremental=%v", isCI, cfg.ForceRetest, useIncremental)

		// Build module test info from tests
		moduleTestInfo, uniqueModules := buildModuleTestInfo(testsByPackage, moduleRegistry, eacCfg, workspaceRoot)
		log.Debugf("Unique modules for incremental detection: %v", uniqueModules)

		// Create a BuildID loader for dependencies not in the current test run
		// This ensures that if a dependency was rebuilt, dependents are retested
		loadDepBuildID := func(moniker string) string {
			moduleBuildDir := eacCfg.Repository.BuildOutputPathAbs(workspaceRoot, moniker)
			if manifest, err := implinternal.LoadModuleManifest(moduleBuildDir); err == nil {
				return manifest.BuildID
			}
			return ""
		}

		// Detect which modules need testing
		changeResult, err := teststate.DetectChangesWithLoader(workspaceRoot, moduleTestInfo, loadDepBuildID)
		if err != nil {
			log.Warnf("Incremental test detection failed, running all tests: %v", err)
		} else {
			incrementalInfo = &incrementalTestInfo{
				detectionTime:   changeResult.DetectionTime,
				modulesNeedTest: changeResult.ModulesNeedingTest,
				modulesUpToDate: changeResult.UpToDateModules,
				freshRun:        changeResult.FreshRun,
				changeReasons:   changeResult.ChangeReasons,
			}

			// Build set of modules needing test
			modulesNeedingTest = make(map[string]bool)
			for _, m := range changeResult.ModulesNeedingTest {
				modulesNeedingTest[m] = true
			}

			log.Debugf("Change detection result: %d need testing, %d up-to-date, detection time=%v",
				len(changeResult.ModulesNeedingTest), len(changeResult.UpToDateModules), changeResult.DetectionTime)

			// If all modules are up-to-date, we can skip testing entirely
			if len(changeResult.ModulesNeedingTest) == 0 && !changeResult.FreshRun {
				writeInit("")
				writeInit("✅ All tests up-to-date (nothing to run)")
				writeInit("   Use --retest to force a full test run")
				return 0
			}
		}
	}

	// If incremental detected, show info
	if incrementalInfo != nil && !incrementalInfo.freshRun {
		writeInit("")
		writeInit("=== Incremental Test Detection ===")
		writeInit("Detection time: %v", incrementalInfo.detectionTime.Round(time.Millisecond))
		if len(incrementalInfo.modulesNeedTest) > 0 {
			writeInit("Need testing: %d modules", len(incrementalInfo.modulesNeedTest))
			for _, m := range incrementalInfo.modulesNeedTest {
				if reason, ok := incrementalInfo.changeReasons[m]; ok {
					writeInit("  - %s (%s)", m, reason)
				} else {
					writeInit("  - %s", m)
				}
			}
		}
		if len(incrementalInfo.modulesUpToDate) > 0 {
			writeInit("Up-to-date: %d modules (skipping)", len(incrementalInfo.modulesUpToDate))
		}
	}

	// Build skip tags for godog filter
	skipTags := eacCfg.TestingTags.GetSkipTagsForSuite()
	for _, dep := range unavailableModuleDeps {
		skipTags = append(skipTags, fmt.Sprintf("@depm:%s", dep))
	}
	suiteTagFilter := suite.BuildGodogTagFilterWithSkipTags(skipTags)

	// Calculate test parallelism
	numCPU := runtime.NumCPU()
	var testParallelism int
	if cfg.Parallel {
		testParallelism = max(2, numCPU/4)
	} else {
		testParallelism = numCPU
	}

	// Create module mapper for output path organization
	moduleMapper := NewModuleMapper(eacCfg, workspaceRoot)

	// Create module-based moniker mapping
	// Maps module output path -> original package path for test lookup
	// Also filter by incremental detection if enabled
	modulePathToPkg := make(map[string]string)
	testsByModulePath := make(map[string][]testing.TestReference)
	skippedPackages := 0
	for pkgPath, tests := range testsByPackage {
		moduleMoniker := moduleMapper.GetModuleForPackagePath(pkgPath)

		// Filter by incremental detection
		if modulesNeedingTest != nil && !modulesNeedingTest[moduleMoniker] {
			skippedPackages++
			continue
		}

		modulePath := moduleMapper.BuildModuleOutputPath(pkgPath, moduleMoniker)
		modulePathToPkg[modulePath] = pkgPath
		testsByModulePath[modulePath] = tests
	}

	if skippedPackages > 0 {
		writeInit("")
		writeInit("Skipped %d packages (modules up-to-date)", skippedPackages)
	}

	// Create test execution context using module-based paths
	execCtx := &TestExecutionContext{
		testsByPackage:    testsByModulePath,    // Now keyed by module path
		modulePathToPkg:   modulePathToPkg,      // Reverse mapping for runner
		testParallelism:   testParallelism,
		testRunDir:        testRunDir,
		reportFormat:      cfg.ReportFormat,
		coverage:          cfg.Coverage,
		suiteTagFilter:    suiteTagFilter,
		workspaceRoot:     workspaceRoot,
		moduleMapper:      moduleMapper,
		results:           make(map[string]PackageResult),
	}

	// Build moniker list and type map using module-based paths
	monikers := make([]string, 0, len(testsByModulePath))
	moduleTypes := make(map[string]string)
	for modulePath, tests := range testsByModulePath {
		monikers = append(monikers, modulePath)
		// Use the type of the first test in the package
		if len(tests) > 0 {
			moduleTypes[modulePath] = tests[0].Type
		}
	}

	// Set worker and module types on the orchestrator
	orch.SetWorker(execCtx.createWorker())
	orch.SetModuleTypes(moduleTypes)

	// Track test execution time
	testStartTime := time.Now()

	// Run tests (TUI transitions to Run phase automatically)
	_, orchErr := orch.Run(monikers)
	if orchErr != nil {
		writeInit("❌ Orchestrator error: %v", orchErr)
		return 1
	}

	// Collect results (before stopping TUI)
	results := execCtx.collectResults()

	// Calculate totals
	packagesPassed := 0
	packagesFailed := 0
	testsPassed := 0
	testsFailed := 0
	testsSkipped := 0
	testsTotal := 0

	for _, result := range results {
		if result.PackageFailed || result.TestsFailed > 0 {
			packagesFailed++
		} else {
			packagesPassed++
		}
		testsPassed += result.TestsPassed
		testsFailed += result.TestsFailed
		testsSkipped += result.TestsSkipped
		testsTotal += result.TestsTotal
	}

	// Update incremental test state (only in local devbox mode)
	log.Debugf("State update check: useIncremental=%v, registryNil=%v", useIncremental, moduleRegistry == nil)
	if useIncremental && moduleRegistry != nil {
		log.Debugf("Updating incremental test state for %d module paths", len(testsByModulePath))
		// Build map of module -> pass/fail from results
		testedModuleResults := make(map[string]bool)

		for modulePath := range testsByModulePath {
			// Extract module moniker from path
			moduleMoniker := strings.Split(modulePath, "/")[0]

			// Check if any test in this module failed
			result, exists := execCtx.results[modulePath]
			if !exists {
				continue
			}

			// If module already failed, keep it failed
			if existing, ok := testedModuleResults[moduleMoniker]; ok && !existing {
				continue
			}

			passed := !result.PackageFailed && result.TestsFailed == 0
			testedModuleResults[moduleMoniker] = passed
		}

		// Rebuild module test info for saving state
		moduleTestInfo, _ := buildModuleTestInfo(testsByPackage, moduleRegistry, eacCfg, workspaceRoot)

		// Update test state
		if err := teststate.UpdateModuleState(workspaceRoot, testedModuleResults, moduleTestInfo); err != nil {
			log.Warnf("Failed to update test state: %v", err)
		} else {
			log.Debugf("Updated test state for %d modules", len(testedModuleResults))
		}
	}

	// Send summary data to TUI and wait for user to exit
	if cfg.UseTUI {
		testSummary := testTUISummary(
			results, time.Since(testStartTime), suite.Name,
			osFilteredCount, len(selectedTests),
			len(testsByPackage), packagesPassed, packagesFailed,
			testsTotal, testsPassed, testsFailed,
			testRunDir,
		)
		orch.SendSummary(testSummary)
		// Wait for user to press any key to exit
		orch.WaitTUI()
	} else {
		// Stop TUI and print plain text summary
		orch.StopTUI()
	}

	// Show full summary (when not using TUI or after TUI exits)
	if !cfg.UseTUI {
		writeln(multiWriter, "══════════════════════════════════")
		writeln(multiWriter, "  Test Summary")
		writeln(multiWriter, "══════════════════════════════════")
	writeln(multiWriter, "Suite: %s", suite.Name)
	writeln(multiWriter, "")
	writeln(multiWriter, "Test Selection Breakdown:")
	writeln(multiWriter, "  Tests discovered:        %d", stats.TotalDiscovered)
	writeln(multiWriter, "  - Skipped (@skip:*):     %d", stats.Skipped)
	writeln(multiWriter, "  - Not matching suite:    %d", stats.NotMatchingSuite)
	writeln(multiWriter, "  = Selected for suite:    %d", stats.Selected)
	writeln(multiWriter, "  - OS incompatible:       %d", osFilteredCount)
	writeln(multiWriter, "  = Production tests:      %d", len(selectedTests))
	writeln(multiWriter, "")
	writeln(multiWriter, "Test Execution:")
	writeln(multiWriter, "  Total packages: %d", len(testsByPackage))
	writeln(multiWriter, "  Packages passed: %d", packagesPassed)
	writeln(multiWriter, "  Packages failed: %d", packagesFailed)
	writeln(multiWriter, "  Tests total: %d", testsTotal)
	writeln(multiWriter, "  Tests passed: %d", testsPassed)
	writeln(multiWriter, "  Tests failed: %d", testsFailed)
	writeln(multiWriter, "")
	writeln(multiWriter, "Results directory: %s", testRunDir)
	}

	// Show top 5 failed tests with log excerpts (always, even when using TUI)
	// These "roll off" after the TUI exits, remaining visible for review
	if packagesFailed > 0 || testsFailed > 0 {
		writeln(multiWriter, "")
		writeln(multiWriter, "══════════════════════════════════")
		writeln(multiWriter, "  Failed Tests")
		writeln(multiWriter, "══════════════════════════════════")

		// Collect failed results
		failedResults := []PackageResult{}
		for _, result := range results {
			if result.PackageFailed || result.TestsFailed > 0 {
				failedResults = append(failedResults, result)
			}
		}

		// Show top 5 failed tests
		maxToShow := 5
		if len(failedResults) < maxToShow {
			maxToShow = len(failedResults)
		}

		for i := 0; i < maxToShow; i++ {
			result := failedResults[i]
			writeln(multiWriter, "")
			writeln(multiWriter, "❌ %s", result.PackageName)

			// Read last 15 lines from log file
			if result.LogFilePath != "" && fileExists(result.LogFilePath) {
				lines := readLastLines(result.LogFilePath, 15)
				for _, line := range lines {
					writeln(multiWriter, "  %s", line)
				}
			} else {
				writeln(multiWriter, "  (log file not available)")
			}
		}

		if len(failedResults) > maxToShow {
			writeln(multiWriter, "")
			writeln(multiWriter, "... and %d more failed tests", len(failedResults)-maxToShow)
		}
	}

	writeln(multiWriter, "")

	// Show timing analysis if requested
	if cfg.ShowTimings {
		writeln(multiWriter, "")
		// Extract unique module monikers from tested packages
		testedModules := make(map[string]bool)
		for pkgPath := range testsByPackage {
			moduleMoniker := moduleMapper.GetModuleForPackagePath(pkgPath)
			if moduleMoniker != "" {
				testedModules[moduleMoniker] = true
			}
		}

		// Convert to slice
		moduleList := make([]string, 0, len(testedModules))
		for module := range testedModules {
			moduleList = append(moduleList, module)
		}

		// Calculate wall-clock time (test execution duration)
		wallClockSeconds := time.Since(testStartTime).Seconds()

		// Display timing analysis for just the modules that were tested
		// Pass the suite-specific output directory and wall-clock time
		show.ShowTestTimingsForModules(moduleList, 5, testRunDir, wallClockSeconds)
	}

	// Return exit code based on failures
	if packagesFailed > 0 || testsFailed > 0 {
		return 1
	}
	return 0
}

// createWorker returns an orchestrator worker function for test execution
func (ctx *TestExecutionContext) createWorker() orchestrator.WorkerFunc {
	return func(pkgPath string, tuiWriter io.Writer) int {
		tests := ctx.testsByPackage[pkgPath]
		result := ctx.runPackageTests(pkgPath, tests, tuiWriter)

		ctx.mu.Lock()
		ctx.results[pkgPath] = result
		ctx.mu.Unlock()

		if result.PackageFailed || result.TestsFailed > 0 {
			return 1
		}
		return 0
	}
}

// runPackageTests executes tests for a single package with streaming output
// modulePath is the module-based path (e.g., eac-core/config) used for output organization
func (ctx *TestExecutionContext) runPackageTests(modulePath string, tests []testing.TestReference, tuiWriter io.Writer) PackageResult {
	// Look up original package path for test execution
	// modulePath is what the orchestrator uses for output directories
	// originalPkgPath is what the runner uses to find test files
	originalPkgPath := ctx.modulePathToPkg[modulePath]
	if originalPkgPath == "" {
		originalPkgPath = modulePath // Fallback if not found
	}

	// Determine test type and get appropriate runner
	testType := getPackageTestType(tests)
	testRunner := runners.Get(testType)

	// If we have a registered runner, use it
	if testRunner != nil {
		// Use the module path directly as the output path (orchestrator already uses this)
		cfg := runners.RunConfig{
			WorkspaceRoot:    ctx.workspaceRoot,
			TestRunDir:       ctx.testRunDir,
			ReportFormat:     ctx.reportFormat,
			Coverage:         ctx.coverage,
			SuiteTagFilter:   ctx.suiteTagFilter,
			Parallelism:      ctx.testParallelism,
			ModuleOutputPath: modulePath, // Orchestrator creates this directory
		}
		// Pass original package path to runner for test execution
		runResult := testRunner.Execute(originalPkgPath, tests, tuiWriter, cfg)
		return PackageResult{
			PackageName:   runResult.PackageName,
			LogFilePath:   runResult.LogFilePath,
			TestsPassed:   runResult.TestsPassed,
			TestsFailed:   runResult.TestsFailed,
			TestsSkipped:  runResult.TestsSkipped,
			TestsTotal:    runResult.TestsTotal,
			PackageFailed: runResult.PackageFailed,
			Duration:      runResult.Duration,
		}
	}

	// Fallback to inline Go test execution (legacy path, should not be reached)
	start := time.Now()
	result := PackageResult{PackageName: modulePath}

	// Determine package directory from original package path
	var relPkgPath, relFeatureFile string
	if strings.Contains(originalPkgPath, ":") {
		parts := strings.SplitN(originalPkgPath, ":", 2)
		relPkgPath = parts[0]
		relFeatureFile = parts[1]
	} else {
		relPkgPath = originalPkgPath
	}

	_ = relFeatureFile // Used in fallback godog setup below

	actualPkgDir := filepath.Join(ctx.workspaceRoot, relPkgPath)

	// Create log file using module path (orchestrator creates this directory)
	logDir := filepath.Join(ctx.testRunDir, modulePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(tuiWriter, "❌ Failed to create log directory: %v\n", err)
		result.PackageFailed = true
		return result
	}

	logFilePath := filepath.Join(logDir, "test.log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Fprintf(tuiWriter, "❌ Failed to create log file: %v\n", err)
		result.PackageFailed = true
		return result
	}
	defer logFile.Close()
	result.LogFilePath = logFilePath

	// Create streaming test runner
	streamingRunner := runner.NewStreamingRunner(tuiWriter, logFile)

	// Build go test command
	goTestArgs := []string{"test", "-json", "-v", "-parallel", fmt.Sprintf("%d", ctx.testParallelism)}

	// Add coverage if enabled
	if ctx.coverage {
		coverageFile := filepath.Join(logDir, "coverage.out")
		goTestArgs = append(goTestArgs, "-cover", "-coverprofile="+coverageFile)
	}

	// Add package path
	goTestArgs = append(goTestArgs, ".")

	cmd := newCommand("go", goTestArgs...)
	cmd.Dir = actualPkgDir
	cmd.Env = os.Environ()

	// Set test run ID for nested commands
	testRunID := filepath.Base(ctx.testRunDir)
	cmd.Env = append(cmd.Env, fmt.Sprintf("R2R_TEST_RUN_ID=%s", testRunID))

	// Set godog environment variables if this is a godog test
	isGodogTest := fileExists(filepath.Join(actualPkgDir, "godog_test.go"))
	if isGodogTest {
		cmd.Env = append(cmd.Env, "GODOG_FORMAT=progress")
		if ctx.suiteTagFilter != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_SUITE_TAGS=%s", ctx.suiteTagFilter))
		}
		if relFeatureFile != "" {
			relFeaturePath, _ := filepath.Rel(actualPkgDir, filepath.Join(ctx.workspaceRoot, relFeatureFile))
			relFeaturePath = filepath.ToSlash(relFeaturePath)
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_PATHS=%s", relFeaturePath))

			// Set report output for feature files
			reportDir := logDir
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_OUTPUT_DIR=%s", reportDir))
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_REPORT_FORMAT=%s", ctx.reportFormat))
		}
	}

	// Run tests with streaming output
	testResult, runErr := streamingRunner.Run(cmd)

	result.TestsPassed = testResult.TestsPassed
	result.TestsFailed = testResult.TestsFailed
	result.TestsSkipped = testResult.TestsSkipped
	result.TestsTotal = testResult.TestsTotal
	result.PackageFailed = testResult.PackageFailed || runErr != nil
	result.Duration = time.Since(start)

	return result
}

// runTscucumberPackageTests executes TypeScript cucumber-js tests
func (ctx *TestExecutionContext) runTscucumberPackageTests(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, relPkgPath, relFeatureFile string) PackageResult {
	start := time.Now()
	result := PackageResult{PackageName: pkgPath}

	// relPkgPath is the module root (e.g., typescript/vscode-ext-commit)
	moduleRoot := filepath.Join(ctx.workspaceRoot, relPkgPath)

	// Create log file for this package
	logDir := filepath.Join(ctx.testRunDir, sanitizePathForLog(pkgPath))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(tuiWriter, "❌ Failed to create log directory: %v\n", err)
		result.PackageFailed = true
		return result
	}

	logFilePath := filepath.Join(logDir, "test.log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Fprintf(tuiWriter, "❌ Failed to create log file: %v\n", err)
		result.PackageFailed = true
		return result
	}
	defer logFile.Close()
	result.LogFilePath = logFilePath

	// Check if package.json exists
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(tuiWriter, "❌ No package.json found\n")
		fmt.Fprintf(logFile, "No package.json found at %s\n", packageJSON)
		result.PackageFailed = true
		return result
	}

	// Build cucumber-js command with feature file path
	args := []string{"cucumber-js"}

	// Add tag filter if provided
	if ctx.suiteTagFilter != "" {
		// Convert godog tag format to cucumber tag expression
		tagExpr := convertToCucumberTagExpr(ctx.suiteTagFilter)
		if tagExpr != "" {
			args = append(args, "--tags", tagExpr)
		}
	}

	// Add the specific feature file if provided
	if relFeatureFile != "" {
		// Convert to path relative to module root
		featurePath := filepath.Join(ctx.workspaceRoot, relFeatureFile)
		relPath, err := filepath.Rel(moduleRoot, featurePath)
		if err == nil {
			args = append(args, relPath)
		}
	}

	// Log command
	fmt.Fprintf(logFile, "=== Testing TypeScript cucumber specs ===\n")
	fmt.Fprintf(logFile, "Module root: %s\n", moduleRoot)
	fmt.Fprintf(logFile, "Command: npx %s\n\n", strings.Join(args, " "))

	// Execute npx cucumber-js
	wrappedName, wrappedArgs := platform.WrapCommand("npx", args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = moduleRoot
	cmd.Env = os.Environ()

	// Capture output
	output, runErr := cmd.CombinedOutput()
	fmt.Fprintf(logFile, "%s\n", output)

	// Parse results
	if runErr != nil {
		result.PackageFailed = true
		result.TestsFailed = len(tests)
		fmt.Fprintf(tuiWriter, "❌ cucumber-js failed\n")
	} else {
		result.TestsPassed = len(tests)
		fmt.Fprintf(tuiWriter, "✅ cucumber-js passed\n")
	}

	result.TestsTotal = len(tests)
	result.Duration = time.Since(start)

	return result
}

// runMochaPackageTests executes TypeScript mocha unit tests
func (ctx *TestExecutionContext) runMochaPackageTests(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, relPkgPath string) PackageResult {
	start := time.Now()
	result := PackageResult{PackageName: pkgPath}

	// Find module root (parent of test directory)
	// relPkgPath is like "typescript/vscode-ext-commit/test", we need "typescript/vscode-ext-commit"
	moduleRoot := filepath.Dir(filepath.Join(ctx.workspaceRoot, relPkgPath))

	// Create log file for this package
	logDir := filepath.Join(ctx.testRunDir, sanitizePathForLog(pkgPath))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(tuiWriter, "❌ Failed to create log directory: %v\n", err)
		result.PackageFailed = true
		return result
	}

	logFilePath := filepath.Join(logDir, "test.log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Fprintf(tuiWriter, "❌ Failed to create log file: %v\n", err)
		result.PackageFailed = true
		return result
	}
	defer logFile.Close()
	result.LogFilePath = logFilePath

	// Check if package.json exists
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(tuiWriter, "❌ No package.json found\n")
		fmt.Fprintf(logFile, "No package.json found at %s\n", packageJSON)
		result.PackageFailed = true
		return result
	}

	// Build npm test command
	args := []string{"test"}

	// Log command
	fmt.Fprintf(logFile, "=== Testing TypeScript mocha tests ===\n")
	fmt.Fprintf(logFile, "Module root: %s\n", moduleRoot)
	fmt.Fprintf(logFile, "Command: npm %s\n\n", strings.Join(args, " "))

	// Execute npm test
	wrappedName, wrappedArgs := platform.WrapCommand("npm", args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = moduleRoot
	cmd.Env = os.Environ()

	// Capture output
	output, runErr := cmd.CombinedOutput()
	fmt.Fprintf(logFile, "%s\n", output)

	// Parse results
	if runErr != nil {
		result.PackageFailed = true
		result.TestsFailed = len(tests)
		fmt.Fprintf(tuiWriter, "❌ mocha tests failed\n")
	} else {
		result.TestsPassed = len(tests)
		fmt.Fprintf(tuiWriter, "✅ mocha tests passed\n")
	}

	result.TestsTotal = len(tests)
	result.Duration = time.Since(start)

	return result
}

// collectResults returns all collected test results
func (ctx *TestExecutionContext) collectResults() []PackageResult {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	results := make([]PackageResult, 0, len(ctx.results))
	for _, r := range ctx.results {
		results = append(results, r)
	}
	return results
}

// groupTestsByPackage groups tests by their package path using registered runners.
// Uses the provided merged config to get correct test_impl_root path.
func groupTestsByPackage(tests []testing.TestReference, workspaceRoot string, cfg *config.EACConfig) map[string][]testing.TestReference {
	testsByPackage := make(map[string][]testing.TestReference)

	for _, test := range tests {
		var pkgPath string

		// Get the runner for this test type
		testRunner := runners.Get(test.Type)

		// Calculate relative path from workspace root
		relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
		if err != nil {
			continue
		}
		relPath = filepath.ToSlash(relPath)

		if testRunner != nil && (test.Type == "godog" || test.Type == "tscucumber") {
			// BDD tests: use runner to find test root
			testRoot := testRunner.FindTestRoot(relPath, cfg)
			if testRoot == "" {
				// No test runner found - skip this test
				continue
			}
			pkgPath = testRunner.BuildPackagePath(testRoot, relPath)
		} else if test.Type == "mocha" {
			// Mocha tests: group by test directory
			absDir := filepath.Dir(test.FilePath)
			relDir, err := filepath.Rel(workspaceRoot, absDir)
			if err != nil {
				continue
			}
			pkgPath = filepath.ToSlash(relDir)
		} else {
			// Go tests (gotest): group by directory
			absDir := filepath.Dir(test.FilePath)
			relDir, err := filepath.Rel(workspaceRoot, absDir)
			if err != nil {
				continue
			}
			pkgPath = filepath.ToSlash(relDir)
		}
		testsByPackage[pkgPath] = append(testsByPackage[pkgPath], test)
	}

	return testsByPackage
}

// sanitizePathForLog converts a package path to a safe directory name
func sanitizePathForLog(pkgPath string) string {
	// Replace colons and other special chars
	safe := strings.ReplaceAll(pkgPath, ":", "_")
	safe = strings.ReplaceAll(safe, "\\", "/")
	return safe
}

// acquireTestLock acquires an exclusive lock for test execution
func acquireTestLock(suiteName, workspaceRoot string, repoCfg *config.RepositoryConfig) (*flock.Flock, error) {
	testDir := filepath.Join(workspaceRoot, repoCfg.Paths.Out.Test)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create test directory: %w", err)
	}

	lockPath := filepath.Join(testDir, fmt.Sprintf(".lock-%s", suiteName))
	lock := flock.New(lockPath)

	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return nil, fmt.Errorf("test suite '%s' is already running", suiteName)
	}

	return lock, nil
}

// releaseTestLock releases the test lock
func releaseTestLock(lock *flock.Flock) {
	if lock == nil {
		return
	}
	lockPath := lock.Path()
	lock.Unlock()
	os.Remove(lockPath)
}

// newCommand creates a new exec.Cmd (abstraction for testing)
var newCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// findTscucumberTestRunner finds the test runner location for a TypeScript cucumber feature file.
// Feature files are in specs/<module>/..., the test runner is in the module's root directory
// where cucumber.js is located.
// Uses the provided merged config to get correct paths.
// Returns empty string if no matching module is found.
func findTscucumberTestRunner(featurePath string, cfg *config.EACConfig) string {
	// Extract module moniker from specs path
	// e.g., "specs/vscode-ext-commit/progress-buffer/specification.feature" -> "vscode-ext-commit"
	specsPrefix := cfg.Repository.Paths.SpecsRoot + "/"
	relPath := strings.TrimPrefix(filepath.ToSlash(featurePath), specsPrefix)
	relPath = strings.TrimPrefix(relPath, strings.ReplaceAll(specsPrefix, "/", "\\"))
	relPath = filepath.ToSlash(relPath)

	// Get module moniker (first path component)
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}
	moniker := parts[0]

	// Look up the module by moniker
	module, ok := cfg.Modules.GetModule(moniker)
	if !ok {
		return ""
	}

	// Return the module's root directory where cucumber.js should be
	return filepath.ToSlash(module.Files.Root)
}

func printTestUsage() {
	log.Info("Test one or more modules by moniker")
	log.Info("")
	log.Info("Usage: r2r eac test [module1] [module2] ... [options]")
	log.Info("")
	log.Info("Options:")
	log.Info("  --suite <name>         Filter tests by suite (default: \"commit\")")
	log.Info("  --as-cucumber          Generate Cucumber JSON reports (default)")
	log.Info("  --as-junit             Generate JUnit XML reports")
	log.Info("  --coverage             Generate coverage reports (coverage.out, coverage.json)")
	log.Info("  --skip-deps            Skip dependency verification before running tests")
	log.Info("  --list-only            List tests that would run without executing them")
	log.Info("  --timings              Show detailed timing summary")
	log.Info("  --debug                Enable debug logs to console (file logging always enabled)")
	log.Info("  --no-tui               Disable TUI console (TUI is default for local console)")
	log.Info(fmt.Sprintf("  --tui-height N         Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --sequential           Run tests sequentially instead of in parallel")
	log.Info("  --retest               Force full test run, bypassing incremental detection")
	log.Info("")
	log.Info("Available suites:")
	log.Info("  commit                 L0-L2 tests (fast, pre-commit)")
	log.Info("  acceptance             IV/OV/PV tests (PLTE acceptance)")
	log.Info("  production-verification  L4+PIV tests (production smoke)")
	log.Info("")
	log.Info("Examples:")
	log.Info("  r2r eac test                          # Test all modules")
	log.Info("  r2r eac test eac-commands             # Test single module")
	log.Info("  r2r eac test r2r-cli eac-core         # Test multiple modules")
	log.Info("  r2r eac test eac-commands --suite acceptance")
	log.Info("  r2r eac test eac-commands --no-tui    # Disable TUI display")
	log.Info("")
	log.Info("Related commands:")
	log.Info("  r2r eac test suite <name>             # Run a specific test suite")
	log.Info("  r2r eac test list-suites              # List all available test suites")
}

// testTUISummary creates summary data for the TUI Summary pane
func testTUISummary(
	results []PackageResult, totalTime time.Duration, suiteName string,
	osFilteredCount, selectedCount,
	totalPackages, packagesPassed, packagesFailed,
	testsTotal, testsPassed, testsFailed int,
	testRunDir string,
) *tui.SummaryData {
	// Calculate skipped tests
	testsSkipped := testsTotal - testsPassed - testsFailed

	// Init summary with OS-filtered count if applicable
	initSummary := fmt.Sprintf("%d tests selected", selectedCount)
	if osFilteredCount > 0 {
		initSummary += fmt.Sprintf(" (%d filtered by OS)", osFilteredCount)
	}

	// Run summary with skipped count if applicable
	runSummary := fmt.Sprintf("%d/%d packages passed, %d/%d tests passed",
		packagesPassed, totalPackages, testsPassed, testsTotal)
	if testsSkipped > 0 {
		runSummary += fmt.Sprintf(", %d skipped", testsSkipped)
	}

	var details []string
	if packagesFailed > 0 || testsFailed > 0 {
		failedPackages := []string{}
		for _, result := range results {
			if result.PackageFailed || result.TestsFailed > 0 {
				failedPackages = append(failedPackages, result.PackageName)
			}
		}
		details = append(details, fmt.Sprintf("Failed: %d packages, %d tests", packagesFailed, testsFailed))
		if len(failedPackages) <= 3 {
			details = append(details, fmt.Sprintf("  (%s)", strings.Join(failedPackages, ", ")))
		} else {
			details = append(details, fmt.Sprintf("  (%s, +%d more)", failedPackages[0], len(failedPackages)-1))
		}
	}
	details = append(details, fmt.Sprintf("Results: %s", testRunDir))

	nextSteps := ""
	if packagesFailed > 0 || testsFailed > 0 {
		nextSteps = "Review detailed failure output below"
	} else {
		nextSteps = fmt.Sprintf("All tests passed for suite: %s", suiteName)
	}

	return &tui.SummaryData{
		Success:     packagesFailed == 0 && testsFailed == 0,
		TotalTime:   totalTime,
		InitSummary: initSummary,
		RunSummary:  runSummary,
		Details:     details,
		NextSteps:   nextSteps,
	}
}

// getUniqueModulesFromTests extracts unique module monikers from test references
func getUniqueModulesFromTests(tests []testing.TestReference) []string {
	moduleSet := make(map[string]bool)
	for _, test := range tests {
		module := extractModuleFromPath(test.FilePath)
		if module != "" {
			moduleSet[module] = true
		}
	}

	modules := make([]string, 0, len(moduleSet))
	for module := range moduleSet {
		modules = append(modules, module)
	}

	// Sort for consistent output
	sort.Strings(modules)
	return modules
}

// validateBuildArtifactsQuiet validates that build artifacts exist and are up-to-date for the given modules.
// It performs:
// 1. Manifest schema validation against the build-manifest contract
// 2. Artifact existence validation (files actually exist on disk)
// 3. Staleness check (source files unchanged since build)
// Returns ArtifactValidationInfo with details about missing/stale artifacts
func validateBuildArtifactsQuiet(
	moduleList []string,
	cfg *config.EACConfig,
	workspaceRoot string,
	moduleRegistry *modules.Registry,
) *initsummary.ArtifactValidationInfo {
	// Use the manifest loader to validate manifests against schema and check artifacts
	summary, err := implinternal.LoadAndValidateManifests(workspaceRoot, moduleList, cfg)
	if err != nil {
		log.Debugf("Manifest validation error: %v", err)
		// If we can't validate, report all modules as missing
		return &initsummary.ArtifactValidationInfo{
			Validated:      true,
			ModulesChecked: moduleList,
			AllPresent:     false,
			MissingFrom:    moduleList,
		}
	}

	var missingFrom []string
	missingDetails := make(map[string][]string)

	for _, result := range summary.Results {
		if result.Error != "" {
			log.Debugf("Module %s: %s", result.Moniker, result.Error)
			missingFrom = append(missingFrom, result.Moniker)
			missingDetails[result.Moniker] = []string{result.Error}
		} else if !result.SchemaValid {
			log.Debugf("Module %s: manifest schema invalid", result.Moniker)
			missingFrom = append(missingFrom, result.Moniker)
			missingDetails[result.Moniker] = []string{"manifest schema invalid"}
		} else if !result.ArtifactsValid {
			log.Debugf("Module %s: missing artifacts %v", result.Moniker, result.MissingArtifacts)
			missingFrom = append(missingFrom, result.Moniker)
			missingDetails[result.Moniker] = result.MissingArtifacts
		}
	}

	// Check for staleness using buildstate change detection
	var staleModules []string
	staleReasons := make(map[string]string)

	if moduleRegistry != nil {
		// Build modules map for change detection
		modulesMap := make(map[string]buildstate.ModuleFileGetter)
		for _, moniker := range moduleList {
			if contract, ok := moduleRegistry.Get(moniker); ok {
				modulesMap[moniker] = contract
			}
		}

		if len(modulesMap) > 0 {
			moduleFiles, err := buildstate.GetModuleSourceFiles(workspaceRoot, modulesMap)
			if err != nil {
				log.Debugf("Failed to get module source files for staleness check: %v", err)
			} else {
				changeResult, err := buildstate.DetectChanges(workspaceRoot, moduleFiles)
				if err != nil {
					log.Debugf("Failed to detect changes for staleness check: %v", err)
				} else if !changeResult.FreshBuild {
					// Report changed modules as stale (they need rebuild)
					for _, moniker := range changeResult.ChangedModules {
						// Only report as stale if artifacts are present (otherwise it's already in missingFrom)
						if !contains(missingFrom, moniker) {
							staleModules = append(staleModules, moniker)
							if reason, ok := changeResult.ChangeReasons[moniker]; ok {
								staleReasons[moniker] = reason
							} else {
								staleReasons[moniker] = "source files changed since build"
							}
						}
					}
				}
			}
		}
	}

	return &initsummary.ArtifactValidationInfo{
		Validated:              true,
		ModulesChecked:         moduleList,
		AllPresent:             len(missingFrom) == 0,
		MissingFrom:            missingFrom,
		MissingArtifactDetails: missingDetails,
		AllCurrent:             len(staleModules) == 0,
		StaleModules:           staleModules,
		StaleReasons:           staleReasons,
	}
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// incrementalTestInfo holds information about incremental test detection results
type incrementalTestInfo struct {
	detectionTime   time.Duration
	modulesNeedTest []string
	modulesUpToDate []string
	freshRun        bool
	changeReasons   map[string]string
}

// buildModuleTestInfo builds the ModuleTestFiles map for incremental test detection.
// Returns the moduleInfo map and a list of unique module monikers.
func buildModuleTestInfo(
	testsByPackage map[string][]testing.TestReference,
	moduleRegistry *modules.Registry,
	eacCfg *config.EACConfig,
	workspaceRoot string,
) (map[string]teststate.ModuleTestFiles, []string) {
	moduleInfo := make(map[string]teststate.ModuleTestFiles)
	uniqueModules := make(map[string]bool)

	// Create a module mapper to find which module owns each package
	moduleMapper := NewModuleMapper(eacCfg, workspaceRoot)

	// Collect all test package paths per module
	testPackagesByModule := make(map[string][]string)
	for pkgPath := range testsByPackage {
		moduleMoniker := moduleMapper.GetModuleForPackagePath(pkgPath)
		if moduleMoniker != "" {
			uniqueModules[moduleMoniker] = true
			testPackagesByModule[moduleMoniker] = append(testPackagesByModule[moduleMoniker], pkgPath)
		}
	}

	// Build ModuleTestFiles for each module
	for moniker := range uniqueModules {
		module, exists := moduleRegistry.Get(moniker)
		if !exists {
			continue
		}

		info := teststate.ModuleTestFiles{
			Dependencies: module.GetDependencies(),
		}

		// Load build manifest to get BuildID (links tests to specific build)
		// This ensures `build --rebuild` triggers retesting
		moduleBuildDir := eacCfg.Repository.BuildOutputPathAbs(workspaceRoot, moniker)
		if manifest, err := implinternal.LoadModuleManifest(moduleBuildDir); err == nil {
			info.BuildID = manifest.BuildID
		}

		// Get source files from module definition
		sourcePatterns := module.GetGlobPatterns()
		sourceFiles, err := teststate.ExpandGlobPatterns(workspaceRoot, sourcePatterns)
		if err == nil {
			// Filter to only include actual source files (not test files)
			for _, f := range sourceFiles {
				if !isTestFile(f) {
					info.SourceFiles = append(info.SourceFiles, f)
				}
			}
		}

		// Get test files from the test packages
		for _, pkgPath := range testPackagesByModule[moniker] {
			// Extract actual path from godog-style paths
			actualPath := pkgPath
			if idx := strings.Index(pkgPath, ":"); idx >= 0 {
				parts := strings.SplitN(pkgPath, ":", 3)
				if len(parts) >= 2 {
					actualPath = parts[1] // testRoot for godog paths
				} else {
					actualPath = parts[0]
				}
			}

			// Find test files in this package
			testGlobs := []string{
				filepath.Join(actualPath, "*_test.go"),
				filepath.Join(actualPath, "*.feature"),
			}
			testFiles, err := teststate.ExpandGlobPatterns(workspaceRoot, testGlobs)
			if err == nil {
				info.TestFiles = append(info.TestFiles, testFiles...)
			}
		}

		moduleInfo[moniker] = info
	}

	// Convert uniqueModules map to slice
	moduleList := make([]string, 0, len(uniqueModules))
	for m := range uniqueModules {
		moduleList = append(moduleList, m)
	}
	sort.Strings(moduleList)

	return moduleInfo, moduleList
}

// isTestFile returns true if the file path looks like a test file
func isTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go") ||
		strings.HasSuffix(path, ".feature") ||
		strings.HasSuffix(path, ".test.ts") ||
		strings.HasSuffix(path, ".spec.ts")
}
