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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/runner"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/runners"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	moduledeps "github.com/ready-to-release/eac/go/eac/core/module-deps"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	systemdeps "github.com/ready-to-release/eac/go/eac/core/system-deps"
	"github.com/ready-to-release/eac/go/eac/core/testing"
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
	UseTUI       bool
	TUIHeight    int
	Parallel     bool
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
		case arg == "--tui":
			cfg.UseTUI = true
			tuiExplicitlySet = true
		case arg == "--no-tui":
			cfg.UseTUI = false
			tuiExplicitlySet = true
		case arg == "--sequential":
			cfg.Parallel = false
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
			log.Errorf("Valid flags: --suite, --as-junit, --as-cucumber, --coverage, --skip-deps, --list-only, --timings, --tui, --no-tui, --tui-height, --sequential")
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
	// Show execution context
	log.Infof("Executing test via %s. \"%s\"", logging.GetExecutionContext(), logging.GetFullCommand())
	log.Info("")

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Load repository config for paths
	repoCfg, err := config.LoadRepositoryConfig(workspaceRoot)
	if err != nil {
		log.Errorf("failed to load repository config: %v", err)
		return 1
	}

	// Load EAC config for testing tags and suites
	eacCfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load EAC config: %v", err)
		return 1
	}

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
		ActionVerb:           "testing",
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

	// Suite info
	writeInit("Running test suite: %s", suite.Name)
	writeInit("Description: %s", suite.Description)
	writeInit("")

	// Phase 1: Test Discovery
	writeInit("%s", output.PhaseHeader(1, "Test Discovery"))
	allTests, err := testing.DiscoverAllTests(workspaceRoot)
	if err != nil {
		writeInit("❌ Failed to discover tests: %v", err)
		return 1
	}
	writeInit("Discovered %d tests", len(allTests))
	writeInit("")

	// Phase 2: Tag Inference
	writeInit("%s", output.PhaseHeader(2, "Tag Inference"))
	allTests = testing.ApplyInferences(allTests, suite.Inferences)
	writeInit("Applied %d inference rules", len(suite.Inferences))
	writeInit("Inferred system deps from module types")
	writeInit("")

	// Phase 3: Suite Selection
	writeInit("%s", output.PhaseHeader(3, "Suite Selection"))

	// Track selection stats
	type selectionStats struct {
		TotalDiscovered  int
		Skipped          int
		NotMatchingSuite int
		Selected         int
	}
	stats := selectionStats{TotalDiscovered: len(allTests)}

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

	if stats.Skipped > 0 {
		writeInit("%d tests skipped (tagged with @skip:<reason>)", stats.Skipped)
	}
	writeInit("Selected %d tests for suite '%s'", len(selectedTests), cfg.SuiteName)

	// Filter by OS compatibility
	osCompatibleTests := filterByOSCompatibility(selectedTests, multiWriter)
	osFilteredCount := len(selectedTests) - len(osCompatibleTests)
	if osFilteredCount > 0 {
		writeInit("INFO: %d tests excluded (incompatible with %s)", osFilteredCount, runtime.GOOS)
	}
	selectedTests = osCompatibleTests

	writeInit("Running %d production tests", len(selectedTests))
	writeInit("")

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

	// Phase 4: Dependency Verification
	writeInit("%s", output.PhaseHeader(4, "Dependency Verification"))
	systemDeps := testing.GetSystemDependencies(selectedTests)
	moduleDeps := testing.GetModuleDependencies(selectedTests)

	var unavailableModuleDeps []string

	if len(systemDeps) == 0 && len(moduleDeps) == 0 {
		writeInit("No dependencies required")
	} else {
		writeInit("System dependencies: %s", strings.Join(systemDeps, ", "))
		writeInit("Module dependencies: %s", strings.Join(moduleDeps, ", "))

		if !cfg.SkipDeps {
			hasSystemFailures := false

			// Verify system dependencies
			sysResults := systemdeps.VerifyAll(systemDeps)
			for _, result := range sysResults {
				writeInit("%s", output.DependencyLine(result.Available, result.Dependency, result.Version))
				if !result.Available {
					hasSystemFailures = true
				}
			}

			// Verify module dependencies
			modResults := moduledeps.VerifyAll(moduleDeps)
			for _, result := range modResults {
				writeInit("%s", output.DependencyLine(result.Available, result.Dependency, result.Version))
				if !result.Available {
					unavailableModuleDeps = append(unavailableModuleDeps, result.Dependency)
				}
			}

			if hasSystemFailures {
				writeInit("")
				writeInit("%s Error: Required system dependencies are missing", output.IconFail)
				writeInit("Use --skip-deps to run tests anyway")
				return 1
			}

			if len(unavailableModuleDeps) > 0 {
				writeInit("")
				writeInit("⏭️ Module dependencies not available, tests will be skipped: %s", strings.Join(unavailableModuleDeps, ", "))
			}
		}
	}
	writeInit("")

	// Phase 5: Test Execution (transitions to Run phase when orchestrator starts)
	writeInit("%s", output.PhaseHeader(5, "Test Execution"))
	writeInit("%s", output.OutputDir(testRunDir))

	// Group tests by package
	testsByPackage := groupTestsByPackage(selectedTests, workspaceRoot)

	if len(testsByPackage) == 0 {
		writeInit("No test packages to execute")
		writeInit("")
		return 0
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
	modulePathToPkg := make(map[string]string)
	testsByModulePath := make(map[string][]testing.TestReference)
	for pkgPath, tests := range testsByPackage {
		moduleMoniker := moduleMapper.GetModuleForPackagePath(pkgPath)
		modulePath := moduleMapper.BuildModuleOutputPath(pkgPath, moduleMoniker)
		modulePathToPkg[modulePath] = pkgPath
		testsByModulePath[modulePath] = tests
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

	// Run tests (TUI transitions to Run phase automatically)
	_, orchErr := orch.Run(monikers)
	if orchErr != nil {
		writeInit("❌ Orchestrator error: %v", orchErr)
		return 1
	}

	// Collect results (before stopping TUI)
	results := execCtx.collectResults()

	// Stop TUI first (restores stdout)
	orch.StopTUI()

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

	// Show full summary
	writeln(multiWriter, "%s", output.SectionHeader("Test Summary"))
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

	// Show top 5 failed tests with log excerpts
	if packagesFailed > 0 || testsFailed > 0 {
		writeln(multiWriter, "")
		writeln(multiWriter, "%s", output.SectionHeader("Failed Tests"))

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

// groupTestsByPackage groups tests by their package path using registered runners
func groupTestsByPackage(tests []testing.TestReference, workspaceRoot string) map[string][]testing.TestReference {
	testsByPackage := make(map[string][]testing.TestReference)

	// Load config once for all tests
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return testsByPackage
	}

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

// findGodogTestRunner finds the test runner package for a feature file
// Feature files are in specs/<module>/<feature>/..., test runners are in {test_impl_root}/<module>/ or {test_impl_root}/<module>/<feature>/
// Returns empty string if no test runner with godog_test.go is found
func findGodogTestRunner(featurePath string, workspaceRoot string) string {
	// Load config to get paths
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return ""
	}

	// Extract relative path from specs root
	// e.g., "specs/eac-core/handlers-config/specification.feature" -> "eac-core/handlers-config/specification.feature"
	specsPrefix := cfg.Repository.Paths.SpecsRoot + "/"
	relPath := strings.TrimPrefix(filepath.ToSlash(featurePath), specsPrefix)
	relPath = strings.TrimPrefix(relPath, strings.ReplaceAll(specsPrefix, "/", "\\"))
	relPath = filepath.ToSlash(relPath)

	// Get path components (e.g., ["eac-core", "handlers-config", "specification.feature"])
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}

	// Try progressively deeper paths to find godog_test.go
	// e.g., first try go/eac/specs/impl/eac-core/, then go/eac/specs/impl/eac-core/handlers-config/
	moniker := parts[0]
	basePath := cfg.Repository.TestImplPath(moniker)

	// Check if godog_test.go exists at base path
	if fileExists(filepath.Join(workspaceRoot, basePath, "godog_test.go")) {
		return basePath
	}

	// Try adding subdirectories (skip the filename at the end)
	for i := 1; i < len(parts)-1; i++ {
		subPath := filepath.Join(basePath, strings.Join(parts[1:i+1], "/"))
		subPath = filepath.ToSlash(subPath)
		if fileExists(filepath.Join(workspaceRoot, subPath, "godog_test.go")) {
			return subPath
		}
	}

	// No test runner found
	return ""
}

// findTscucumberTestRunner finds the test runner location for a TypeScript cucumber feature file.
// Feature files are in specs/<module>/..., the test runner is in the module's root directory
// where cucumber.js is located.
// Returns empty string if no matching module is found.
func findTscucumberTestRunner(featurePath string) string {
	// Load config to get modules
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return ""
	}

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
	log.Info("  --no-tui               Disable TUI console (TUI is default for local console)")
	log.Info(fmt.Sprintf("  --tui-height N         Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --sequential           Run tests sequentially instead of in parallel")
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
