// Command: test
// Short: Test one or more modules by moniker
// Args: modules
// Long: Test one or more modules by moniker using suite-based filtering.
// Long:
// Long: This command discovers tests, applies inference rules (e.g., Go tests default to @L1),
// Long: filters by suite tags, and runs matching tests with consistent summary output.
// Long:
// Long: Use --suite to select which tests to run. The default runs suites not marked
// Long: as extended_suite in config (typically unit + integration).
// Long:
// Long: Expected Output:
// Long:   - Test execution results with pass/fail status
// Long:   - Detailed test summary table showing modules, packages, and assertions
// Long:   - Test logs written to out/test/<module>/ directory
// Long:   - Exit code 0 if all tests pass, non-zero on failure
// Long:
// Long: Example:
// Long:   test eac-commands                    # Test single module
// Long:   test eac-core r2r-cli                # Test multiple modules
// Long:   test                                 # Test all modules
// Long:   test eac-commands --suite acceptance # Run acceptance tests only
// Flag.suite: type=string, usage=Filter tests by suite (default: non-extended suites from config)
// Flag.coverage: type=bool, usage=Enable coverage reporting
// Flag.skip-deps: type=bool, usage=Skip dependency checks before running tests
// Flag.skip-depm: type=bool, usage=Skip module dependency build artifact validation
// Flag.list-only: type=bool, usage=List tests without running them
// Flag.timings: type=bool, usage=Show detailed timing summary after tests complete
// Flag.debug: type=bool, usage=Enable debug logs to console (file logging always enabled)
// Flag.tui: type=bool, usage=Enable TUI console (default for local console mode)
// Flag.no-tui: type=bool, usage=Disable TUI console (use plain output)
// Flag.sequential: type=bool, usage=Run tests sequentially instead of parallel
// Flag.turbo: type=bool, usage=Enable turbo mode for faster testing (increases parallelism)
// Flag.skip-cache: type=bool, usage=Skip incremental cache, force full test run
// Flag.tui-height: type=int, usage=Set TUI console height (3-20, default: 6)
package test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/manifests"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/runners"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

var log = logging.C()

func init() {
	registry.Register(Test)
}

// TestConfig holds test execution configuration.
type TestConfig struct {
	Monikers    []string
	SuiteName   string
	Coverage    bool
	SkipDeps    bool // Skip system dependency verification (--skip-deps)
	SkipDepm    bool // Skip dependency module build artifact validation (--skip-depm)
	ListOnly    bool
	ShowTimings bool
	DebugMode   bool
	UseTUI      bool
	TUIHeight   int
	Parallel    bool
	Turbo       bool // --turbo flag to increase parallelism
	Roof        int  // --roof flag to limit max parallel capacity
	ForceRetest bool // --skip-cache flag to bypass incremental testing
}

// TestExecutionContext holds shared state for parallel test execution.
type TestExecutionContext struct {
	testsByPackage  map[string][]testing.TestReference // Keyed by module path (e.g., eac-core/config)
	modulePathToPkg map[string]string                  // Maps module path -> original package path (e.g., go/eac/core/config)
	testParallelism int
	testRunDir      string
	coverage        bool
	suiteTagFilter  string
	workspaceRoot   string
	moduleMapper    *ModuleMapper // Maps package paths to module monikers

	// Suite routing
	suiteMoniker string                   // Suite moniker (single or composite like "unit+integration")
	repoCfg      *config.RepositoryConfig // For computing suite output paths

	// Thread-safe result storage
	mu      sync.Mutex
	results map[string]PackageResult
}

// PackageResult is an alias for manifests.PackageResult.
type PackageResult = manifests.PackageResult

// Test is the unified entry point for testing modules.
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

	// Convert to framework config
	cmdCfg := &cmdframework.CommandConfig{
		Type:           cmdframework.CommandTypeTest,
		ActionVerb:     "Testing",
		OutputDir:      "out/test",
		LogFileName:    "test.log",
		Monikers:       cfg.Monikers,
		MaxConcurrency: cfg.Roof, // 0 = auto-detect
		Sequential:     !cfg.Parallel,
		Turbo:          cfg.Turbo,
		SkipDeps:       cfg.SkipDeps,
		SkipDepm:       cfg.SkipDepm,
		ForceRebuild:   cfg.ForceRetest,
		UseTUI:         cfg.UseTUI,
		TUIHeight:      cfg.TUIHeight,
		DebugMode:      cfg.DebugMode,
		ShowTimings:    cfg.ShowTimings,
		SkipResolve:    true, // Test command manages its own execution plan
	}

	testCfg := &TestFrameworkConfig{
		SuiteName:   cfg.SuiteName,
		Coverage:    cfg.Coverage,
		ForceRetest: cfg.ForceRetest,
		Parallel:    cfg.Parallel,
		ListOnly:    cfg.ListOnly,
	}

	return RunTestWithFramework(cmdCfg, testCfg)
}

// parseTestArgs parses command line arguments into TestConfig.
func parseTestArgs(args []string) *TestConfig {
	// Detect execution environment for TUI defaults
	env := environment.Detect()

	// Parse shared flags using flag sets
	shared, err := flags.ParseSharedFlagsWithEnv(flags.TestConfig(), args, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}

	// Parse test-specific flags from remaining args
	testFlags, unknownArgs, err := ParseTestSpecificFlags(shared.Remaining)
	if err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}

	// Check for unknown flags
	for _, arg := range unknownArgs {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			log.Errorf("unknown flag: %s", arg)
			log.Errorf("Valid flags: --suite, --coverage, --skip-deps, --skip-depm, --list-only, --timings, --debug, --tui, --no-tui, --tui-height, --turbo, --roof, --skip-cache, --dry-run")
			return nil
		}
	}

	// Determine parallel mode: --roof 1 means sequential
	parallel := shared.MaxConcurrency != 1

	// Validate --turbo and sequential mutual exclusivity
	if shared.Turbo && !parallel {
		log.Errorf("Error: --turbo and --roof 1 (sequential) cannot be used together")
		return nil
	}

	cfg := &TestConfig{
		// From shared flags
		Monikers:    shared.Monikers,
		SkipDeps:    shared.SkipDeps,
		SkipDepm:    shared.SkipDepm,
		ShowTimings: shared.ShowTimings,
		DebugMode:   shared.Debug,
		UseTUI:      shared.UseTUI,
		TUIHeight:   shared.TUIHeight,
		Turbo:       shared.Turbo,
		Roof:        shared.MaxConcurrency,
		ForceRetest: shared.SkipCache,
		Parallel:    parallel,

		// From test-specific flags
		SuiteName: testFlags.SuiteName,
		Coverage:  testFlags.Coverage,
		ListOnly:  testFlags.ListOnly,
	}

	return cfg
}

// createWorker returns an orchestrator worker function for test execution.
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

// getEffectiveTestRunDir returns the test run directory for a package.
// With the module-based output structure, all tests go to the same root directory
// (out/test) and the module-specific paths are constructed by the module mapper.
func (ctx *TestExecutionContext) getEffectiveTestRunDir(tests []testing.TestReference) string {
	// Module-based structure: always use the test output root
	// Individual module paths are handled by BuildModuleOutputPath
	return ctx.testRunDir
}

// runPackageTests executes tests for a single package with streaming output
// modulePath is the module-based path (e.g., eac-core/config) used for output organization.
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

	// Extract module moniker from modulePath (format: "<moniker>/<subpath>" or just "<moniker>")
	moduleMoniker := modulePath
	if idx := strings.Index(modulePath, "/"); idx > 0 {
		moduleMoniker = modulePath[:idx]
	}

	// Get effective test run dir (routes to correct suite folder for composite suites)
	effectiveTestRunDir := ctx.getEffectiveTestRunDir(tests)

	// Use the module path directly as the output path (orchestrator already uses this)
	cfg := runners.RunConfig{
		WorkspaceRoot:    ctx.workspaceRoot,
		TestRunDir:       effectiveTestRunDir,
		Coverage:         ctx.coverage,
		SuiteTagFilter:   ctx.suiteTagFilter,
		Parallelism:      ctx.testParallelism,
		ModuleMoniker:    moduleMoniker, // For result aggregation
		ModuleOutputPath: modulePath,    // Orchestrator creates this directory
	}
	// Pass original package path to runner for test execution
	runResult := testRunner.Execute(originalPkgPath, tests, tuiWriter, cfg)
	return PackageResult{
		ModuleMoniker: runResult.ModuleMoniker,
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

// runPackageTestsWithOutputDir executes tests for a single package with a pre-specified output directory.
// This is used by component-level execution where the orchestrator creates the output directory.
func (ctx *TestExecutionContext) runPackageTestsWithOutputDir(modulePath string, tests []testing.TestReference, tuiWriter io.Writer, outputDir string) PackageResult {
	// Look up original package path for test execution
	originalPkgPath := ctx.modulePathToPkg[modulePath]
	if originalPkgPath == "" {
		originalPkgPath = modulePath
	}

	// Determine test type and get appropriate runner
	testType := getPackageTestType(tests)
	testRunner := runners.Get(testType)

	// Extract module moniker from modulePath
	moduleMoniker := modulePath
	if idx := strings.Index(modulePath, "/"); idx > 0 {
		moduleMoniker = modulePath[:idx]
	}

	// Get effective test run dir (routes to correct suite folder for composite suites)
	effectiveTestRunDir := ctx.getEffectiveTestRunDir(tests)

	cfg := runners.RunConfig{
		WorkspaceRoot:    ctx.workspaceRoot,
		TestRunDir:       effectiveTestRunDir,
		Coverage:         ctx.coverage,
		SuiteTagFilter:   ctx.suiteTagFilter,
		Parallelism:      ctx.testParallelism,
		ModuleMoniker:    moduleMoniker,
		ModuleOutputPath: modulePath,
		OutputDir:        outputDir, // Pre-created output directory
	}

	runResult := testRunner.Execute(originalPkgPath, tests, tuiWriter, cfg)
	return PackageResult{
		ModuleMoniker: runResult.ModuleMoniker,
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

// runTscucumberPackageTests executes TypeScript cucumber-js tests.
func (ctx *TestExecutionContext) runTscucumberPackageTests(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, relPkgPath, relFeatureFile string) PackageResult {
	start := time.Now()
	result := PackageResult{PackageName: pkgPath}

	// relPkgPath is the module root (e.g., typescript/vscode-ext-commit)
	moduleRoot := filepath.Join(ctx.workspaceRoot, relPkgPath)

	// Create log file for this package
	logDir := filepath.Join(ctx.testRunDir, sanitizePathForLog(pkgPath))
	if err := os.MkdirAll(logDir, 0o755); err != nil {
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
	cmd.Env = append(cmd.Env, "R2R_TEST_LOGGING_ACTIVE=true")

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

// runMochaPackageTests executes TypeScript mocha unit tests.
func (ctx *TestExecutionContext) runMochaPackageTests(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, relPkgPath string) PackageResult {
	start := time.Now()
	result := PackageResult{PackageName: pkgPath}

	// Find module root (parent of test directory)
	// relPkgPath is like "typescript/vscode-ext-commit/test", we need "typescript/vscode-ext-commit"
	moduleRoot := filepath.Dir(filepath.Join(ctx.workspaceRoot, relPkgPath))

	// Create log file for this package
	logDir := filepath.Join(ctx.testRunDir, sanitizePathForLog(pkgPath))
	if err := os.MkdirAll(logDir, 0o755); err != nil {
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
	cmd.Env = append(cmd.Env, "R2R_TEST_LOGGING_ACTIVE=true")

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

// collectResults returns all collected test results.
func (ctx *TestExecutionContext) collectResults() []PackageResult {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	results := make([]PackageResult, 0, len(ctx.results))
	for _, r := range ctx.results {
		results = append(results, r)
	}
	return results
}

// newCommand creates a new exec.Cmd (abstraction for testing).
var newCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func printTestUsage() {
	log.Info("Test one or more modules by moniker")
	log.Info("")
	log.Info("Usage: r2r eac test [module1] [module2] ... [options]")
	log.Info("")
	log.Info("Options:")
	log.Info("  --suite <name>         Filter tests by suite (default: unit+integration)")
	log.Info("  --coverage             Generate coverage reports (coverage.out, coverage.json)")
	log.Info("  --skip-deps            Skip system dependency verification before running tests")
	log.Info("  --skip-depm            Skip module dependency build artifact validation")
	log.Info("  --list-only            List tests that would run without executing them")
	log.Info("  --timings              Show detailed timing summary")
	log.Info("  --debug                Enable debug logs to console (file logging always enabled)")
	log.Info("  --no-tui               Disable TUI console (TUI is default for local console)")
	log.Info(fmt.Sprintf("  --tui-height N         Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --turbo                Enable turbo mode for faster testing (increases parallelism)")
	log.Info("  --roof N               Limit max parallel capacity to N (default: auto-detect from CPU/RAM)")
	log.Info("  --sequential           Run tests sequentially instead of in parallel")
	log.Info("  --skip-cache           Skip incremental cache, force full test run")
	log.Info("")
	log.Info("Available suites:")
	log.Info("  unit                     L0-L1 tests (fast unit tests)")
	log.Info("  integration              L2 tests (Docker-based emulated tests)")
	log.Info("  acceptance               L3 tests (production-like tests)")
	log.Info("  production-verification  L4+PIV tests (production smoke)")
	log.Info("")
	log.Info("Composite suites (use + to combine):")
	log.Info("  unit+integration+acceptance           # Run multiple suites together")
	log.Info("")
	log.Info("Examples:")
	log.Info("  r2r eac test                          # Test all modules (default suites)")
	log.Info("  r2r eac test eac-commands             # Test single module")
	log.Info("  r2r eac test r2r-cli eac-core         # Test multiple modules")
	log.Info("  r2r eac test --suite acceptance       # Run acceptance suite")
	log.Info("  r2r eac test eac-commands --no-tui    # Disable TUI display")
}

// printTestSummary prints unified test summary to a writer (for non-TUI mode)
// Note: suiteName/suiteMoniker kept for API compatibility but suite info is shown during init.
func printTestSummary(w io.Writer, results []PackageResult, suiteName, suiteMoniker string,
	selectedCount, osFilteredCount,
	totalPackages, packagesPassed, packagesFailed,
	testsTotal, testsPassed, testsFailed int,
	testRunDir string,
) {
	// Note: Suite info is displayed during init phase (via writeInitSummary)
	// The suiteName/suiteMoniker params are kept for potential future use
	_ = suiteName
	_ = suiteMoniker

	testsSkipped := testsTotal - testsPassed - testsFailed
	moduleStats := aggregateResultsByModule(results)

	// Init summary line
	initSummary := fmt.Sprintf("%d test cases selected", selectedCount)
	if osFilteredCount > 0 {
		initSummary += fmt.Sprintf(" (%d filtered by OS)", osFilteredCount)
	}

	// Run summary line
	var runSummary string
	if packagesFailed == 0 && testsFailed == 0 {
		assertionInfo := fmt.Sprintf("%d assertions", testsTotal)
		if testsSkipped > 0 {
			assertionInfo += fmt.Sprintf(", %d skipped", testsSkipped)
		}
		runSummary = fmt.Sprintf("%d packages passed (%s)", totalPackages, assertionInfo)
	} else {
		runSummary = fmt.Sprintf("%d/%d packages passed, %d assertions failed",
			packagesPassed, totalPackages, testsFailed)
	}

	// Status icons
	initIcon := "✓"
	runIcon := "✓"
	if packagesFailed > 0 || testsFailed > 0 {
		runIcon = "✗"
	}

	fmt.Fprintf(w, "%s Initialization: %s\n", initIcon, initSummary)
	fmt.Fprintf(w, "%s Testing: %s\n", runIcon, runSummary)
	fmt.Fprintln(w)

	// Module breakdown table
	fmt.Fprintln(w, strings.Repeat("-", 58))
	fmt.Fprintln(w, "Module               Pkgs  Asserts  Fail  Skip  Pass")
	fmt.Fprintln(w, strings.Repeat("-", 58))
	for _, ms := range moduleStats {
		fmt.Fprintf(w, "%-20s %4d  %7d  %4d  %4d  %4d\n",
			truncateString(ms.Module, 20), ms.Packages, ms.Assertions, ms.Failed, ms.Skipped, ms.Passed)
	}
	fmt.Fprintln(w, strings.Repeat("-", 58))
	fmt.Fprintf(w, "%-20s %4d  %7d  %4d  %4d  %4d\n",
		"TOTAL", totalPackages, testsTotal, testsFailed, testsSkipped, testsPassed)
	fmt.Fprintln(w)

	// Results path
	fmt.Fprintf(w, "Results: %s\n", testRunDir)
}

// testTUISummary creates summary data for the TUI Summary pane.
func testTUISummary(
	results []PackageResult, totalTime time.Duration, suiteName, suiteMoniker string,
	osFilteredCount, selectedCount,
	totalPackages, packagesPassed, packagesFailed,
	testsTotal, testsPassed, testsFailed int,
	testRunDir string,
) *tui.SummaryData {
	testsSkipped := testsTotal - testsPassed - testsFailed

	// Format suite name(s) for display
	displaySuiteNames := suiteName
	suiteLabel := "suite"
	if suitesIncluded := manifests.GetSuitesIncluded(suiteMoniker); len(suitesIncluded) > 0 {
		displaySuiteNames = strings.Join(suitesIncluded, ", ")
		suiteLabel = "suite(s)"
	}

	// Init summary: test cases selected (scenarios/functions we chose to run)
	initSummary := fmt.Sprintf("%d test cases selected", selectedCount)
	if osFilteredCount > 0 {
		initSummary += fmt.Sprintf(" (%d filtered by OS)", osFilteredCount)
	}

	// Run summary: packages are the unit of execution, assertions are what ran inside them
	// (testsTotal includes subtests from t.Run, which is why it can exceed selectedCount)
	var runSummary string
	if packagesFailed == 0 && testsFailed == 0 {
		// All passed - show packages passed and total assertions
		assertionInfo := fmt.Sprintf("%d assertions", testsTotal)
		if testsSkipped > 0 {
			assertionInfo += fmt.Sprintf(", %d skipped", testsSkipped)
		}
		runSummary = fmt.Sprintf("%d packages passed (%s)", totalPackages, assertionInfo)
	} else {
		// Failures - show what failed
		runSummary = fmt.Sprintf("%d/%d packages passed, %d assertions failed",
			packagesPassed, totalPackages, testsFailed)
	}

	// Build per-module breakdown
	moduleStats := aggregateResultsByModule(results)

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

	// Add per-module breakdown table
	details = append(details, "")
	details = append(details, strings.Repeat("-", 58))
	details = append(details, "Module               Pkgs  Asserts  Fail  Skip  Pass")
	details = append(details, strings.Repeat("-", 58))
	for _, ms := range moduleStats {
		details = append(details, fmt.Sprintf("%-20s %4d  %7d  %4d  %4d  %4d",
			truncateString(ms.Module, 20), ms.Packages, ms.Assertions, ms.Failed, ms.Skipped, ms.Passed))
	}
	details = append(details, strings.Repeat("-", 58))
	details = append(details, fmt.Sprintf("%-20s %4d  %7d  %4d  %4d  %4d",
		"TOTAL", totalPackages, testsTotal, testsFailed, testsSkipped, testsPassed))

	details = append(details, "")
	details = append(details, fmt.Sprintf("Results: %s", testRunDir))

	nextSteps := ""
	if packagesFailed > 0 || testsFailed > 0 {
		nextSteps = "Review detailed failure output below"
	} else {
		nextSteps = fmt.Sprintf("All tests passed for %s: %s", suiteLabel, displaySuiteNames)
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

// moduleTestStats holds aggregated test statistics for a module.
type moduleTestStats struct {
	Module     string
	Packages   int
	Passed     int
	Failed     int
	Skipped    int
	Assertions int
}

// aggregateResultsByModule groups test results by module moniker.
func aggregateResultsByModule(results []PackageResult) []moduleTestStats {
	moduleMap := make(map[string]*moduleTestStats)

	for _, result := range results {
		// Use ModuleMoniker directly (set by runner via GetTestInfo)
		module := result.ModuleMoniker
		if module == "" {
			// Fallback: extract from package name if ModuleMoniker not set
			module = result.PackageName
			if idx := strings.Index(module, "/"); idx > 0 {
				module = module[:idx]
			}
		}

		stats, exists := moduleMap[module]
		if !exists {
			stats = &moduleTestStats{Module: module}
			moduleMap[module] = stats
		}

		stats.Packages++
		stats.Passed += result.TestsPassed
		stats.Failed += result.TestsFailed
		stats.Skipped += result.TestsSkipped
		stats.Assertions += result.TestsTotal
	}

	// Convert to sorted slice
	var moduleList []moduleTestStats
	for _, stats := range moduleMap {
		moduleList = append(moduleList, *stats)
	}
	sort.Slice(moduleList, func(i, j int) bool {
		return moduleList[i].Module < moduleList[j].Module
	})

	return moduleList
}

// truncateString truncates a string to maxLen, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Terminology for test metrics:
//
// | Framework    | Test Case                | Package              | Assertion              |
// |--------------|--------------------------|----------------------|------------------------|
// | Go (gotest)  | Test* function           | Go package directory | t.Run subtests + main  |
// | Go (godog)   | Gherkin scenario         | Feature file folder  | Scenario steps         |
// | TS (mocha)   | describe() block         | Test file            | it() blocks            |
// | TS (cucumber)| Gherkin scenario         | Feature file folder  | Scenario steps         |
//
// - "Test cases selected" = scenarios + Test* functions discovered and matched by suite filter
// - "Packages" = execution units (directories/files containing test cases)
// - "Assertions" = actual test executions reported by runner (includes subtests)

// getUniqueModulesFromTests extracts unique module monikers from test references
// Uses ModuleMapper for accurate module ownership lookup from registry.
func getUniqueModulesFromTests(tests []testing.TestReference, mapper *ModuleMapper) []string {
	moduleSet := make(map[string]bool)
	for i := range tests {
		test := &tests[i]
		module := mapper.GetModuleForFile(test.FilePath)
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

// incrementalTestInfo holds information about incremental test detection results.
type incrementalTestInfo struct {
	detectionTime   time.Duration
	modulesNeedTest []string
	modulesUpToDate []string
	freshRun        bool
	changeReasons   map[string]string
}
