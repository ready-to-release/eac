// Command: test suite
// Short: Run tests for a specific test suite (parallel by default)
// Long: Run tests for a specific test suite (parallel by default).
// Long:
// Long: Test suites are defined in contracts and group related tests for execution.
// Long: By default, tests within a suite run in parallel for faster execution.
// Long:
// Long: Use --skip-deps to skip dependency installation before running tests.
// Long: Use --list-only to see which tests would run without executing them.
// Long: Use --sequential to run tests one at a time instead of in parallel.
// Long:
// Long: Example:
// Long:   test suite integration
// Long:   test suite unit --sequential
// Long:   test suite e2e --list-only
// Flag.skip-deps: type=bool, usage=Skip dependency installation before running tests
// Flag.list-only: type=bool, usage=List tests that would run without executing them
// Flag.sequential: type=bool, usage=Run tests sequentially instead of in parallel
package test

import (
	"bufio"
	"encoding/json"
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
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/reporter"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/testjson"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/testers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	contractsreports "github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	moduledeps "github.com/ready-to-release/eac/go/eac/core/module-deps"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	systemdeps "github.com/ready-to-release/eac/go/eac/core/system-deps"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

func init() {
	registry.Register(TestSuite)
}

// writeln writes a formatted string with platform-specific line ending to the writer
func writeln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

// acquireSuiteLock attempts to acquire an exclusive lock for the test suite.
// Returns the lock handle and nil error on success.
// Returns nil and error if lock is already held (suite is running).
func acquireSuiteLock(suiteName, workspaceRoot string, repoCfg *config.RepositoryConfig) (*flock.Flock, error) {
	// Ensure test output directory exists (parent directory for lock files)
	testDir := filepath.Join(workspaceRoot, repoCfg.Paths.Out.Test)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create test directory: %w", err)
	}

	// Create lock file path in parent directory (so it survives directory purge)
	// Use suite name as the mutex identifier
	lockPath := filepath.Join(testDir, fmt.Sprintf(".lock-%s", suiteName))

	// Create flock instance
	lock := flock.New(lockPath)

	// Try to acquire lock (non-blocking)
	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		return nil, fmt.Errorf("test suite '%s' is already running", suiteName)
	}

	return lock, nil
}

// releaseSuiteLock releases the lock and removes the lock file
func releaseSuiteLock(lock *flock.Flock) {
	if lock == nil {
		return
	}

	lockPath := lock.Path()
	lock.Unlock()

	// Clean up the lock file
	os.Remove(lockPath)
}

// PackageTestResult holds detailed test execution results for a single package
type PackageTestResult struct {
	PackageName   string        // Package name/path for display
	LogFilePath   string        // Path to the test log file
	TestsPassed   int           // Number of individual tests that passed
	TestsFailed   int           // Number of individual tests that failed
	TestsSkipped  int           // Number of individual tests that were skipped
	TestsTotal    int           // Total number of tests in this package
	PackageFailed bool          // Whether the package execution itself failed
	ExpectedFiles []string      // Files that should have been created by this test run
	Duration      time.Duration // Time taken to run the package tests
}

// TestContext holds shared state for parallel test execution via orchestrator
type TestContext struct {
	testsByPackage  map[string][]testing.TestReference
	testParallelism int
	testRunDir      string
	reportFormat    string
	coverage        bool
	suiteTagFilter  string

	// Thread-safe result storage
	mu      sync.Mutex
	results map[string]PackageTestResult
}

// CreateWorker returns an orchestrator worker function for test execution
func (ctx *TestContext) CreateWorker() orchestrator.WorkerFunc {
	return func(pkgPath string, logWriter io.Writer) int {
		tests := ctx.testsByPackage[pkgPath]
		result := runPackageTests(pkgPath, tests, logWriter, ctx.testParallelism,
			ctx.testRunDir, ctx.reportFormat, ctx.coverage, ctx.suiteTagFilter)

		ctx.mu.Lock()
		ctx.results[pkgPath] = result
		ctx.mu.Unlock()

		if result.PackageFailed || result.TestsFailed > 0 {
			return 1
		}
		return 0
	}
}

// CollectResults returns all collected test results
func (ctx *TestContext) CollectResults() []PackageTestResult {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	results := make([]PackageTestResult, 0, len(ctx.results))
	for _, r := range ctx.results {
		results = append(results, r)
	}
	return results
}

// TestSuite runs tests for a specific test suite
func TestSuite() int {
	// Parse arguments and flags
	if len(os.Args) < 4 {
		log.Errorf("missing suite name")
		log.Errorf("Usage: test suite <suite-name> [flags]")
		log.Errorf("")
		log.Errorf("Flags:")
		log.Errorf("  --skip-deps    Skip dependency verification")
		log.Errorf("  --list-only    List tests without running them")
		log.Errorf("  --sequential   Run tests sequentially (for debugging)")
		log.Errorf("  --parallel     Run tests in parallel (DEFAULT, explicit override)")
		log.Errorf("  --as-cucumber  Generate Cucumber JSON reports (DEFAULT)")
		log.Errorf("  --as-junit     Generate JUnit XML reports")
		log.Errorf("  --coverage     Generate coverage reports (coverage.out, coverage.json)")
		log.Errorf("  --timings      Show detailed timing summary")
		log.Errorf("")
		log.Errorf("Default: Tests run in parallel for optimal performance.")
		log.Errorf("Use --sequential if you need deterministic ordering or debugging.")
		log.Errorf("")
		log.Errorf("Available suites:")
		for _, suite := range testing.ListSuites() {
			log.Errorf("  - %s", suite)
		}
		return 1
	}

	suiteName := os.Args[3]

	// Parse flags
	skipDeps := false
	listOnly := false
	parallel := true  // Default to parallel execution for better performance
	reportFormat := "cucumber" // Default report format for BDD tests
	coverage := false // Generate coverage reports
	showTimings := false // Show timing summary
	useTUI := false      // TUI console (reserved for future use)
	tuiHeight := tui.DefaultHeight // TUI height
	var moduleFilters []string // Optional module filters (can be comma-separated)

	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--skip-deps" {
			skipDeps = true
		} else if arg == "--list-only" {
			listOnly = true
		} else if arg == "--sequential" {
			parallel = false  // Opt-out of parallel execution
		} else if arg == "--parallel" {
			parallel = true   // Explicit parallel (redundant but allowed)
		} else if arg == "--as-junit" {
			reportFormat = "junit"
		} else if arg == "--as-cucumber" {
			reportFormat = "cucumber"
		} else if arg == "--coverage" {
			coverage = true
		} else if arg == "--timings" {
			showTimings = true
		} else if arg == "--tui" {
			useTUI = true
		} else if arg == "--tui-height" {
			if i+1 >= len(os.Args) {
				log.Errorf("--tui-height requires a value")
				return 1
			}
			i++
			var err error
			tuiHeight, err = strconv.Atoi(os.Args[i])
			if err != nil || tuiHeight < 3 || tuiHeight > 20 {
				log.Errorf("--tui-height must be a number between 3 and 20")
				return 1
			}
		} else if arg == "--module" {
			// Read module names from next argument (comma-separated)
			if i+1 >= len(os.Args) {
				log.Errorf("--module requires one or more module names")
				log.Errorf("Usage: --module <name> or --module <name1>,<name2>")
				return 1
			}
			i++
			// Split by comma and trim spaces
			modules := strings.Split(os.Args[i], ",")
			for _, mod := range modules {
				trimmed := strings.TrimSpace(mod)
				if trimmed != "" {
					moduleFilters = append(moduleFilters, trimmed)
				}
			}
		} else if strings.HasPrefix(arg, "--") {
			log.Errorf("unknown flag: %s", arg)
			log.Errorf("Valid flags: --skip-deps, --list-only, --sequential, --parallel, --module <name>, --as-junit, --as-cucumber, --coverage, --timings, --tui, --tui-height")
			return 1
		}
	}


	// Get repository root
	workspaceRootNative, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}
	// Codebase uses Unix-style paths throughout - normalize for path comparisons
	workspaceRoot := filepath.ToSlash(workspaceRootNative)

	// Load EAC config (properly merged with defaults) for path resolution
	eacCfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load EAC config: %v", err)
		return 1
	}
	repoCfg := eacCfg.Repository

	// Get the test suite
	suite, err := testing.GetSuite(suiteName)
	if err != nil {
		log.Errorf("%v", err)
		log.Errorf("")
		log.Errorf("Available suites:")
		for _, s := range testing.ListSuites() {
			log.Errorf("  - %s", s)
		}
		return 1
	}

	// Acquire exclusive lock for this test suite FIRST (before any directory operations)
	lockFile, err := acquireSuiteLock(suiteName, workspaceRootNative, repoCfg)
	if err != nil {
		log.Errorf("test suite '%s' is already running", suiteName)
		log.Errorf("Details: %v", err)
		return 1
	}
	defer releaseSuiteLock(lockFile)

	// Purge and recreate test output directory (now protected by lock)
	testRunDir := filepath.Join(workspaceRootNative, repoCfg.TestOutputPath(suiteName))
	if err := os.RemoveAll(testRunDir); err != nil {
		log.Errorf("Warning: failed to purge test directory: %v", err)
	}
	if err := os.MkdirAll(testRunDir, 0755); err != nil {
		log.Errorf("failed to create test run directory: %v", err)
		return 1
	}

	// Create log file
	logPath := filepath.Join(testRunDir, "test-suite.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		log.Errorf("failed to create log file: %v", err)
		return 1
	}
	defer logFile.Close()

	// Track start time for duration calculation
	startTime := time.Now()

	// Create multi-writer to log to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Phase 1: Discover all tests (Go + Godog)
	writeln(multiWriter, "%s", output.PhaseHeader(1, "Test Discovery"))

	allTests, err := testing.DiscoverAllTests(workspaceRoot)
	if err != nil {
		log.Errorf("failed to discover tests: %v", err)
		return 1
	}

	writeln(multiWriter, "Discovered %d tests", len(allTests))
	writeln(multiWriter, "")

	// Phase 2: Apply inference rules
	writeln(multiWriter, "%s", output.PhaseHeader(2, "Tag Inference"))
	allTests = testing.ApplyInferences(allTests, suite.Inferences)
	writeln(multiWriter, "Applied %d inference rules", len(suite.Inferences))

	// Load module registry for module-based inference
	moduleReport, err := contractsreports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("Warning: failed to load module contracts: %v", err)
	} else {
		// Infer system dependencies from module dependencies
		allTests = testing.InferSystemDepsFromModuleDeps(allTests, moduleReport.Registry)
		writeln(multiWriter, "Inferred system deps from module types")
	}

	writeln(multiWriter, "")

	// Phase 3: Select tests for suite
	writeln(multiWriter, "%s", output.PhaseHeader(3, "Suite Selection"))
	selectedTests, selectionStats := suite.SelectTestsWithStats(allTests)
	writeln(multiWriter, "Selected %d tests for suite '%s'", len(selectedTests), suite.Moniker)

	// Phase 3.5: Apply module filter if specified
	if len(moduleFilters) > 0 {
		writeln(multiWriter, "%s", output.ListFormatWithPrefix("Filtering by modules", moduleFilters, 80, 5))
		filteredTests := []testing.TestReference{}

		// Track unique modules found for debugging
		foundModules := make(map[string]int)

		// Build file-to-module cache for efficient lookup
		// Uses contract-based ownership (respects excludes, parent-child hierarchy)
		fileModuleCache := make(map[string]string)

		for _, test := range selectedTests {
			// Check cache first
			testModule, cached := fileModuleCache[test.FilePath]
			if !cached {
				// Use contract-based ownership if registry is available
				if moduleReport != nil && moduleReport.Registry != nil {
					// Normalize path for contract matching
					normalizedPath := filepath.ToSlash(test.FilePath)
					// Make path relative to workspace root
					if strings.HasPrefix(normalizedPath, filepath.ToSlash(workspaceRoot)) {
						normalizedPath = strings.TrimPrefix(normalizedPath, filepath.ToSlash(workspaceRoot))
						normalizedPath = strings.TrimPrefix(normalizedPath, "/")
					}

					// Find owning modules using contract-based matching
					matchingModules := moduleReport.Registry.FindModulesForFile(normalizedPath)
					if len(matchingModules) > 0 {
						// Use the first matching module (most specific due to filtering)
						testModule = matchingModules[0].Moniker
					}
				}

				// Fallback to path-based extraction if no contract match
				if testModule == "" {
					testModule = extractModuleFromPath(test.FilePath)
				}

				fileModuleCache[test.FilePath] = testModule
			}

			// Track modules found
			foundModules[testModule]++

			// Check if test belongs to any of the specified modules
			for _, moduleFilter := range moduleFilters {
				if testModule == moduleFilter {
					filteredTests = append(filteredTests, test)
					break
				}
			}
		}

		// Log found modules for debugging
		if len(filteredTests) == 0 {
			writeln(multiWriter, "No tests matched. Modules found in selected tests:")
			for mod, count := range foundModules {
				if mod != "" {
					writeln(multiWriter, "  - %s (%d tests)", mod, count)
				} else {
					writeln(multiWriter, "  - [empty module name] (%d tests)", count)
				}
			}
			// Show a few sample paths
			if len(selectedTests) > 0 {
				writeln(multiWriter, "Sample file paths (first 5):")
				sampleCount := 5
				if len(selectedTests) < sampleCount {
					sampleCount = len(selectedTests)
				}
				for i := 0; i < sampleCount; i++ {
					writeln(multiWriter, "  - %s", selectedTests[i].FilePath)
				}
			}
		}

		selectedTests = filteredTests
		writeln(multiWriter, "Selected %d tests after module filtering", len(selectedTests))
	}

	// Phase 3.6: Filter out framework tests (tests about the testing framework itself)
	productionTests := []testing.TestReference{}
	frameworkTestCount := 0
	for _, test := range selectedTests {
		if testing.ShouldSkipValidation(test) {
			frameworkTestCount++
		} else {
			productionTests = append(productionTests, test)
		}
	}

	if frameworkTestCount > 0 {
		writeln(multiWriter, "INFO: %d framework tests excluded from execution", frameworkTestCount)
	}

	// Phase 3.7: Filter by OS compatibility
	// Tests with deps:linux, deps:macos, deps:windows are OS-specific
	// Tests without any of these are OS-agnostic and run everywhere
	osCompatibleTests := filterByOSCompatibility(productionTests, multiWriter)
	osFilteredCount := len(productionTests) - len(osCompatibleTests)
	if osFilteredCount > 0 {
		writeln(multiWriter, "INFO: %d tests excluded (incompatible with %s)", osFilteredCount, runtime.GOOS)
	}
	productionTests = osCompatibleTests

	writeln(multiWriter, "Running %d production tests", len(productionTests))
	writeln(multiWriter, "")

	// If list-only, just show tests and exit
	if listOnly {
		writeln(multiWriter, "=== Production Tests ===")
		for i, test := range productionTests {
			writeln(multiWriter, "%d. %s (%s)", i+1, test.TestName, test.Type)
			writeln(multiWriter, "   File: %s", test.FilePath)
			writeln(multiWriter, "   Tags: %s", strings.Join(test.Tags, ", "))
			writeln(multiWriter, "")
		}
		return 0
	}

	// Phase 4: Extract and verify dependencies (system + module)
	writeln(multiWriter, "%s", output.PhaseHeader(4, "Dependency Verification"))
	systemDeps := testing.GetSystemDependencies(productionTests)
	moduleDeps := testing.GetModuleDependencies(productionTests)

	allDeps := append(append([]string{}, systemDeps...), moduleDeps...)

	// Track unavailable module deps - tests requiring them will be skipped
	var unavailableModuleDeps []string

	if len(allDeps) == 0 {
		writeln(multiWriter, "No dependencies required")
		writeln(multiWriter, "")
	} else {
		writeln(multiWriter, "System dependencies: %s", strings.Join(systemDeps, ", "))
		writeln(multiWriter, "Module dependencies: %s", strings.Join(moduleDeps, ", "))

		if !skipDeps {
			hasSystemFailures := false

			// Verify system dependencies - these cause hard failures
			sysResults := systemdeps.VerifyAll(systemDeps)
			for _, result := range sysResults {
				writeln(multiWriter, "%s", output.DependencyLine(result.Available, result.Dependency, result.Version))
				if !result.Available {
					hasSystemFailures = true
				}
			}

			// Verify module dependencies - unavailable ones cause tests to be skipped
			modResults := moduledeps.VerifyAll(moduleDeps)
			for _, result := range modResults {
				writeln(multiWriter, "%s", output.DependencyLine(result.Available, result.Dependency, result.Version))
				if !result.Available {
					unavailableModuleDeps = append(unavailableModuleDeps, result.Dependency)
				}
			}

			writeln(multiWriter, "")

			// Only fail for system dependencies - module deps cause skip instead
			if hasSystemFailures {
				writeln(multiWriter, "%s Error: Required system dependencies are missing", output.IconFail)
				writeln(multiWriter, "Use --skip-deps to run tests anyway")
				return 1
			}

			// Report skipped module dependencies
			if len(unavailableModuleDeps) > 0 {
				writeln(multiWriter, "%s Module dependencies not available, tests will be skipped: %s",
					output.IconSkip, strings.Join(unavailableModuleDeps, ", "))
				writeln(multiWriter, "")
			}
		} else {
			writeln(multiWriter, "Dependency check skipped (--skip-deps)")
			writeln(multiWriter, "")
		}
	}

	// findGodogTestRunner finds the test runner package for a feature file
	// Feature files are in specs/<module>/..., test runners are in {test_impl_root}/<module>/
	findGodogTestRunner := func(featurePath string) string {
		// Extract module from specs path
		// Example: specs/eac-commands/commit/... -> go/eac/specs/impl/eac-commands
		//          specs/r2r-cli/... -> go/eac/specs/impl/r2r-cli
		//          specs/repository/... -> go/eac/specs/impl/repository
		specsPrefix := repoCfg.Paths.SpecsRoot + "/"
		relPath := strings.TrimPrefix(featurePath, specsPrefix)
		relPath = strings.TrimPrefix(relPath, strings.ReplaceAll(specsPrefix, "/", "\\"))

		// Get first path component (e.g., "eac-commands" or "repository")
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		if len(parts) == 0 {
			return repoCfg.TestImplPath("eac-commands") // fallback
		}

		// Test runners are in {test_impl_root}/<module>/
		moniker := parts[0]
		return repoCfg.TestImplPath(moniker)
	}

	// Phase 5: Run tests
	writeln(multiWriter, "%s", output.PhaseHeader(5, "Test Execution"))
	writeln(multiWriter, "%s", output.OutputDir(testRunDir))

	// Group tests by package
	// For Godog tests: need to find their test runner package
	// For Go tests: group by directory as usual
	// IMPORTANT: Convert absolute paths to workspace-relative paths for proper cmd.Dir handling
	testsByPackage := make(map[string][]testing.TestReference)
	for _, test := range productionTests {
		var pkgPath string
		if test.Type == "godog" {
			// Find the test runner package for this feature file
			// Feature files are in specs/, test runners are in go/eac/specs/impl/*/
			// Create synthetic package key: testrunner:featurefile (both relative)

			// Convert absolute feature file path to relative
			relFeaturePath, err := filepath.Rel(workspaceRoot, test.FilePath)
			if err != nil {
				// Skip test if we can't compute relative path
				writeln(multiWriter, "⚠️  Skipping test %s: unable to compute relative path from %s to %s",
					test.TestName, workspaceRoot, test.FilePath)
				continue
			}

			// Find test runner using relative path
			testRunnerPkg := findGodogTestRunner(relFeaturePath)
			absTestRunnerPkg := filepath.Join(workspaceRootNative, testRunnerPkg)

			// Check if test runner directory exists (has godog_test.go)
			if fileExists(filepath.Join(absTestRunnerPkg, "godog_test.go")) {
				// Has test runner - create synthetic key to run feature through it
				pkgPath = testRunnerPkg + ":" + relFeaturePath
			} else {
				// No test runner - group by feature directory (will be skipped as "Godog-only")
				pkgPath = filepath.Dir(relFeaturePath)
			}
		} else {
			// Go tests grouped by directory
			// Convert absolute path to relative
			absDir := filepath.Dir(test.FilePath)
			relDir, err := filepath.Rel(workspaceRoot, absDir)
			if err != nil {
				// Skip test if we can't compute relative path
				writeln(multiWriter, "⚠️  Skipping test %s: unable to compute relative path from %s to %s",
					test.TestName, workspaceRoot, absDir)
				continue
			}
			pkgPath = relDir
		}
		testsByPackage[pkgPath] = append(testsByPackage[pkgPath], test)
	}

	// Aggregate results
	var results []PackageTestResult

	// Calculate optimal test-level parallelism
	numCPU := runtime.NumCPU()
	var testParallelism int

	// Build godog tag filter for suite with skip tags integrated
	// Load config to get skip reasons
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(multiWriter, "❌ Failed to load config: %v\n", err)
		return 1
	}

	skipTags := cfg.TestingTags.GetSkipTagsForSuite()

	// Add unavailable module dependencies as skip tags
	// Tests tagged with @depm:modulename will be skipped if that module isn't built
	for _, dep := range unavailableModuleDeps {
		skipTags = append(skipTags, fmt.Sprintf("@depm:%s", dep))
	}

	suiteTagFilter := suite.BuildGodogTagFilterWithSkipTags(skipTags)

	// Calculate test parallelism based on execution mode
	if parallel {
		// Package-level parallel: distribute CPU across packages
		testParallelism = max(2, numCPU/4)
	} else {
		// Sequential packages: each package gets full CPU power
		testParallelism = numCPU
	}

	// Create test context for orchestrator
	testCtx := &TestContext{
		testsByPackage:  testsByPackage,
		testParallelism: testParallelism,
		testRunDir:      testRunDir,
		reportFormat:    reportFormat,
		coverage:        coverage,
		suiteTagFilter:  suiteTagFilter,
		results:         make(map[string]PackageTestResult),
	}

	// Configure orchestrator
	maxConcurrency := 4
	if !parallel {
		maxConcurrency = 1 // Sequential mode
	}

	// OutputBaseDir needs to be relative to WorkspaceRoot
	relTestRunDir := repoCfg.TestOutputPath(suiteName)

	orchConfig := orchestrator.Config{
		WorkspaceRoot:        workspaceRootNative,
		OutputBaseDir:        relTestRunDir,
		LogFileName:          "test.log",
		OrchestratorLogName:  "orchestrator.log",
		ActionVerb:           "Testing",
		MaxConcurrency:       maxConcurrency,
		StatusUpdateInterval: 2, // seconds
		TUI:                  useTUI,
		TUIHeight:            tuiHeight,
	}

	// Build moniker list from package paths
	monikers := make([]string, 0, len(testsByPackage))
	for pkgPath := range testsByPackage {
		monikers = append(monikers, pkgPath)
	}

	// Execute tests via orchestrator
	orch := orchestrator.New(orchConfig, testCtx.CreateWorker())
	_, orchErr := orch.Run(monikers)
	if orchErr != nil {
		writeln(multiWriter, "❌ Orchestrator error: %v", orchErr)
		return 1
	}
	orch.StopTUI()
	orch.Close()

	// Collect results from test context
	results = testCtx.CollectResults()

	// Calculate totals from results
	packagesPassed := 0
	packagesFailed := 0
	testsPassed := 0
	testsFailed := 0
	testsSkipped := 0
	testsTotal := 0

	for _, result := range results {
		if result.PackageFailed {
			packagesFailed++
		} else {
			packagesPassed++
		}
		testsPassed += result.TestsPassed
		testsFailed += result.TestsFailed
		testsSkipped += result.TestsSkipped
		testsTotal += result.TestsTotal
	}

	// Parallelism info removed from console for cleaner output

	// Phase 6: Generate summary
	endTime := time.Now()

	writeln(multiWriter, "%s", output.SectionHeader("Test Summary"))
	writeln(multiWriter, "Suite: %s", suite.Name)
	writeln(multiWriter, "")
	writeln(multiWriter, "Test Selection Breakdown:")
	writeln(multiWriter, "  Tests discovered:        %d", selectionStats.TotalDiscovered)
	writeln(multiWriter, "  - Skipped (@skip:*):     %d", selectionStats.Skipped)
	writeln(multiWriter, "  - Not matching suite:    %d", selectionStats.NotMatchingSuite)
	writeln(multiWriter, "  = Selected for suite:    %d", selectionStats.Selected)
	if frameworkTestCount > 0 {
		writeln(multiWriter, "  - Framework tests:       %d", frameworkTestCount)
	}
	writeln(multiWriter, "  - OS incompatible:       %d", osFilteredCount)
	writeln(multiWriter, "  = Production tests:      %d", len(productionTests))
	writeln(multiWriter, "")
	writeln(multiWriter, "Test Execution:")
	writeln(multiWriter, "  Total packages: %d", len(testsByPackage))
	writeln(multiWriter, "  Packages passed: %d", packagesPassed)
	writeln(multiWriter, "  Packages failed: %d", packagesFailed)
	writeln(multiWriter, "  Tests total: %d", testsTotal)
	writeln(multiWriter, "  Tests passed: %d", testsPassed)
	writeln(multiWriter, "  Tests failed: %d", testsFailed)
	if testsSkipped > 0 {
		writeln(multiWriter, "  Tests skipped: %d", testsSkipped)
	}
	writeln(multiWriter, "")
	writeln(multiWriter, "Results directory: %s", testRunDir)

	// Calculate timing data (always needed for JSON export)
	var totalDuration time.Duration

	// Build timing data for JSON export
	type TimingEntry struct {
		PackageName   string  `json:"package_name"`
		DisplayName   string  `json:"display_name"`
		FeaturePath   string  `json:"feature_path,omitempty"`
		TestType      string  `json:"test_type"`
		DurationSecs  float64 `json:"duration_seconds"`
		CurrentLevel  string  `json:"current_level,omitempty"`
		ProposedLevel string  `json:"proposed_level"`
		NeedsRetag    bool    `json:"needs_retag"`
	}
	timingEntries := []TimingEntry{}

	// Build a lookup map from package/feature path to L-level tag from discovered tests
	testLevelMap := make(map[string]string)
	for _, test := range allTests {
		// Extract L-level from tags
		level := ""
		for _, tag := range test.Tags {
			if tag == "@L0" || tag == "L0" {
				level = "L0"
				break
			} else if tag == "@L1" || tag == "L1" {
				level = "L1"
				break
			} else if tag == "@L2" || tag == "L2" {
				level = "L2"
				break
			}
		}

		// Store in map by package dir (for go tests) or feature path (for godog)
		if level != "" {
			if test.Type == "godog" {
				// For godog, use feature file path as key
				testLevelMap[test.FilePath] = level
			} else {
				// For go tests, use package directory as key
				// FilePath is like "go/eac/commands/impl/commit/assembly_test.go"
				// We want "go/eac/commands/impl/commit"
				packageDir := filepath.Dir(test.FilePath)
				testLevelMap[packageDir] = level
			}
		}
	}

	// Helper to get current level from the lookup map
	getCurrentLevel := func(packagePath string, featurePath string) string {
		if featurePath != "" {
			// Godog test - lookup by feature path
			if level, ok := testLevelMap[featurePath]; ok {
				return level
			}
		} else if packagePath != "" {
			// Go test - lookup by package path
			if level, ok := testLevelMap[packagePath]; ok {
				return level
			}
		}
		return ""
	}

	for _, result := range results {
		totalDuration += result.Duration
		seconds := result.Duration.Seconds()

		// Normalize path separators to forward slashes
		displayName := strings.ReplaceAll(result.PackageName, "\\", "/")

		// Determine test type and feature path
		testType := "go"
		featurePath := ""
		if strings.Contains(displayName, ":specs/") {
			testType = "godog"
			// Extract feature path from package name
			// Example: go/eac/commands/tests:specs/eac-commands/build-module/specification.feature
			parts := strings.Split(displayName, ":")
			if len(parts) == 2 {
				featurePath = parts[1]
			}
			// Shorten display name
			displayName = strings.ReplaceAll(displayName, ":specs/", ":")
			displayName = strings.TrimSuffix(displayName, "/specification.feature")
		}

		// Determine proposed level based on timing
		var proposedLevel string
		if seconds <= 0.5 {
			proposedLevel = "L0"
		} else if seconds <= 2.0 {
			proposedLevel = "L1"
		} else {
			proposedLevel = "L2"
		}

		// Get current level from discovered tests (works for both go and godog)
		// For go tests, use package path; for godog, use feature path
		packagePath := ""
		if testType == "go" {
			// Extract package path from PackageName
			// Example: "go\\eac\\commands\\impl\\commit" -> "go/eac/commands/impl/commit"
			packagePath = filepath.ToSlash(result.PackageName)
		}
		currentLevel := getCurrentLevel(packagePath, featurePath)

		// Determine if retagging is needed (only for L0-L2 tests)
		needsRetag := false
		if currentLevel != "" {
			// Only retag tests that have L0-L2 tags
			// (currentLevel == "" means no L0-L2 tags found, likely L3+)
			if currentLevel != proposedLevel {
				// Tag present but wrong level
				needsRetag = true
			}
		}

		// Add to timing entries for JSON export
		timingEntries = append(timingEntries, TimingEntry{
			PackageName:   result.PackageName,
			DisplayName:   displayName,
			FeaturePath:   featurePath,
			TestType:      testType,
			DurationSecs:  seconds,
			CurrentLevel:  currentLevel,
			ProposedLevel: proposedLevel,
			NeedsRetag:    needsRetag,
		})
	}

	// Show timing summary table (only with --timings flag)
	if showTimings {
		writeln(multiWriter, "")
		writeln(multiWriter, "%s", output.SectionHeader("Timing Summary"))
		for _, entry := range timingEntries {
			writeln(multiWriter, "%s", output.TimingLine(time.Duration(entry.DurationSecs*float64(time.Second)), entry.DisplayName))
		}
		writeln(multiWriter, "%s", output.TimingTotal(totalDuration))
	}

	// Write timing data to JSON file
	timingsJSONPath := filepath.Join(testRunDir, "timings.json")
	timingsData := map[string]interface{}{
		"suite":         suiteName,
		"total_seconds": totalDuration.Seconds(),
		"entries":       timingEntries,
		"timestamp":     time.Now().Format(time.RFC3339),
	}
	timingsJSON, err := json.MarshalIndent(timingsData, "", "  ")
	if err == nil {
		if writeErr := os.WriteFile(timingsJSONPath, timingsJSON, 0644); writeErr != nil {
			writeln(multiWriter, "⚠️  Warning: Failed to write timings.json: %v", writeErr)
		}
	}

	// Show failed test outputs (top 5)
	if packagesFailed > 0 {
		writeln(multiWriter, "")
		writeln(multiWriter, "%s", output.SectionHeader("Failed Test Outputs"))

		failedResults := []PackageTestResult{}
		for _, result := range results {
			if result.PackageFailed {
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

	// Check for undefined steps in test logs
	undefinedSteps := extractUndefinedSteps(testRunDir)
	if len(undefinedSteps) > 0 {
		writeln(multiWriter, "")
		writeln(multiWriter, "⚠️  WARNING: %d undefined steps found", len(undefinedSteps))
		writeln(multiWriter, "Scenarios with undefined steps need step implementations.")
		writeln(multiWriter, "Run with verbose logging to see full details.")
	}

	// Validate expected output files
	writeln(multiWriter, "")
	writeln(multiWriter, "%s", output.SectionHeader("Output File Validation"))
	fileValidationErrors := validateOutputFiles(results, multiWriter)
	if len(fileValidationErrors) > 0 {
		writeln(multiWriter, "")
		writeln(multiWriter, "❌ %d output file validation errors found", len(fileValidationErrors))
		// Don't fail the build for missing files, just warn
	} else {
		writeln(multiWriter, "✅ All expected output files exist and are non-empty")
	}

	// Generate markdown report using template
	modules, err := reporter.CollectModuleReports(testRunDir)
	if err != nil {
		log.Errorf("Warning: failed to collect module reports: %v", err)
		modules = nil // Continue without module breakdown
	}

	reportData := reporter.BuildReportData(
		suite.Name, suite.Moniker,
		startTime, endTime,
		len(allTests), len(productionTests), frameworkTestCount,
		testsTotal, testsPassed, testsFailed, testsSkipped,
		len(testsByPackage), packagesPassed, packagesFailed,
		modules,
		testRunDir,
	)

	// Get repository root to find template
	templatePath := filepath.Join(workspaceRootNative, repoCfg.Paths.Templates, "reports", "suite-summary.md")
	mdPath := filepath.Join(testRunDir, "test-suite-summary.md")

	renderer := reporter.NewRenderer(templatePath, mdPath, reportData)
	if err := renderer.Render(); err != nil {
		log.Errorf("Warning: failed to generate markdown report: %v", err)
	}

	if packagesFailed > 0 {
		return 1
	}

	return 0
}

// fileExists checks if a file exists at the given path
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readLastLines reads the last N lines from a file, parsing JSON test output
func readLastLines(filePath string, n int) []string {
	file, err := os.Open(filePath)
	if err != nil {
		return []string{"Error reading file: " + err.Error()}
	}
	defer file.Close()

	var outputLines []string
	scanner := bufio.NewScanner(file)

	// Parse JSON test output and extract "Output" fields
	for scanner.Scan() {
		line := scanner.Text()

		// Try to parse as JSON test event
		var event struct {
			Action string `json:"Action"`
			Output string `json:"Output"`
		}

		if err := json.Unmarshal([]byte(line), &event); err == nil {
			// Successfully parsed JSON - extract Output field if it's an output event
			if event.Action == "output" && event.Output != "" {
				// Strip ANSI color codes for cleaner display
				cleaned := stripANSI(event.Output)
				// Trim trailing newlines but keep the content
				cleaned = strings.TrimRight(cleaned, "\n")
				if cleaned != "" {
					outputLines = append(outputLines, cleaned)
				}
			}
		} else {
			// Not JSON - include raw line
			outputLines = append(outputLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return []string{"Error scanning file: " + err.Error()}
	}

	// Return last n lines
	if len(outputLines) <= n {
		return outputLines
	}
	return outputLines[len(outputLines)-n:]
}

// OutputFileError represents a validation error for an expected output file
type OutputFileError struct {
	FilePath string
	Issue    string // "missing" or "empty"
}

// validateOutputFiles checks that all expected output files exist and are non-empty
// Returns a list of validation errors
func validateOutputFiles(results []PackageTestResult, w io.Writer) []OutputFileError {
	var errors []OutputFileError
	var checkedCount, passedCount int

	for _, result := range results {
		for _, expectedFile := range result.ExpectedFiles {
			checkedCount++

			// Check if file exists
			info, err := os.Stat(expectedFile)
			if os.IsNotExist(err) {
				errors = append(errors, OutputFileError{
					FilePath: expectedFile,
					Issue:    "missing",
				})
				writeln(w, "  ❌ Missing: %s", expectedFile)
				continue
			}

			if err != nil {
				errors = append(errors, OutputFileError{
					FilePath: expectedFile,
					Issue:    fmt.Sprintf("error: %v", err),
				})
				writeln(w, "  ❌ Error checking %s: %v", expectedFile, err)
				continue
			}

			// Check if file is empty
			if info.Size() == 0 {
				errors = append(errors, OutputFileError{
					FilePath: expectedFile,
					Issue:    "empty",
				})
				writeln(w, "  ❌ Empty: %s", expectedFile)
				continue
			}

			passedCount++
		}
	}

	if checkedCount > 0 {
		writeln(w, "Checked %d files: %d passed, %d failed", checkedCount, passedCount, len(errors))
	} else {
		writeln(w, "No output files to validate")
	}

	return errors
}

// displayGodogFeatureSummaries parses and displays feature file summaries for a Godog test package
func displayGodogFeatureSummaries(testPkgPath string, w io.Writer) {
	// Determine the specs directory based on the test package path
	// go/r2r/cli/tests -> specs/r2r-cli
	// go/eac/commands/tests -> specs/eac-commands
	var specsDir string
	if strings.Contains(testPkgPath, "cli/tests") || strings.Contains(testPkgPath, "cli\\tests") {
		specsDir = "r2r-cli"
	} else if strings.Contains(testPkgPath, "commands/tests") || strings.Contains(testPkgPath, "commands\\tests") {
		specsDir = "eac-commands"
	} else {
		// Unknown test package, skip feature summary
		return
	}

	// Get repository root to construct absolute path to specs
	repoRoot, err := repository.GetRepositoryRoot(".")
	if err != nil {
		fmt.Fprintf(w, "⚠️  Could not determine repository root: %v\n", err)
		return
	}

	specsPath := filepath.Join(repoRoot, "specs", specsDir)

	// Find all .feature files in the specs directory
	featureFiles, err := testing.FindFeatureFiles(specsPath)
	if err != nil {
		fmt.Fprintf(w, "⚠️  Could not find feature files: %v\n", err)
		return
	}

	if len(featureFiles) == 0 {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Fprintln(w, "📋 GODOG FEATURES")
	fmt.Fprintln(w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Parse and display each feature file
	for _, featurePath := range featureFiles {
		feature, err := testing.ParseFeatureFile(featurePath)
		if err != nil {
			fmt.Fprintf(w, "⚠️  Could not parse %s: %v\n", featurePath, err)
			continue
		}

		displayFeature(feature, w)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Fprintln(w, "Running tests...")
	fmt.Fprintln(w)
}

// displayFeature formats and displays a single feature's metadata
func displayFeature(feature *testing.FeatureFile, w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "📦 MODULE: %s | 🔖 FEATURE: %s\n", feature.Module, feature.FeatureName)

	if feature.Title != "" {
		fmt.Fprintf(w, "   📝 %s\n", feature.Title)
	}

	if feature.Description != "" {
		fmt.Fprintln(w, feature.Description)
	}

	// Display rules if any
	for _, rule := range feature.Rules {
		fmt.Fprintf(w, "   📋 Rule: %s\n", rule.Name)
		if rule.Description != "" {
			fmt.Fprintln(w, rule.Description)
		}
	}

	// Display scenarios
	if len(feature.Scenarios) > 0 {
		fmt.Fprintf(w, "   Scenarios: (%d)\n", len(feature.Scenarios))
		for _, scenario := range feature.Scenarios {
			fmt.Fprintf(w, "     - %s\n", scenario)
		}
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// runPackageTests runs tests for a single package and returns detailed test results
func runPackageTests(pkgPath string, tests []testing.TestReference, multiWriter io.Writer, testParallelism int, testRunDir string, reportFormat string, coverage bool, suiteTagFilter string) (result PackageTestResult) {
	// Track timing for this package - use defer to calculate at end
	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()

	// pkgPath is already workspace-relative (from test grouping phase)
	// Check if this is a synthetic Godog package key (testrunner:featurefile)
	var relPkgPath string      // Relative path for display
	var relFeatureFile string  // Relative feature file path (if Godog)
	if strings.Contains(pkgPath, ":") {
		// Synthetic key for Godog feature file
		parts := strings.SplitN(pkgPath, ":", 2)
		relPkgPath = parts[0]       // Test runner package path (relative)
		relFeatureFile = parts[1]   // Feature file path (relative)
	} else {
		relPkgPath = pkgPath
	}

	// Get workspace root to create absolute paths for cmd.Dir
	workspaceRootNative, err := repository.GetRepositoryRoot("")
	if err != nil {
		workspaceRootNative = "."
	}

	// Use feature file path for display name if available, otherwise package path
	var pkgName string
	if relFeatureFile != "" {
		pkgName = filepath.ToSlash(relFeatureFile)
	} else {
		pkgName = filepath.ToSlash(relPkgPath)
	}

	// Check if this is a spec-only package (godog or tscucumber)
	// These are handled differently than unit tests
	testType := getPackageTestType(tests)
	isSpecOnly := testType == "godog" || testType == "tscucumber"

	if isSpecOnly && testType == "godog" && relFeatureFile == "" {
		// Godog specs without explicit feature file - executed via Go test runner
		// Skip here to avoid duplicate execution
		return PackageTestResult{
			PackageName:   pkgPath,
			LogFilePath:   "",
			TestsPassed:   len(tests),
			TestsFailed:   0,
			TestsSkipped:  0,
			TestsTotal:    len(tests),
			PackageFailed: false,
			ExpectedFiles: []string{},
		}
	}

	// Dispatch to appropriate handler based on test type
	if handler, ok := getTestTypeHandler(testType); ok {
		return handler(pkgPath, tests, multiWriter, testRunDir, reportFormat, suiteTagFilter)
	}

	// Default: Go test execution (handles "go", "godog", and unknown types)

	// Create absolute paths for cmd.Dir by joining relative paths with workspace root
	absPkgPath := filepath.Join(workspaceRootNative, relPkgPath)
	var absFeatureFile string
	if relFeatureFile != "" {
		absFeatureFile = filepath.Join(workspaceRootNative, relFeatureFile)
	}

	// Determine actual directory path for running tests
	// If we have a feature file, use the test runner package directory
	var actualPkgDir string
	if relFeatureFile != "" {
		// Use the test runner package directory (absolute)
		actualPkgDir = absPkgPath
	} else if filepath.Ext(absPkgPath) == ".feature" {
		// pkgPath is a feature file, get its directory
		actualPkgDir = filepath.Dir(absPkgPath)
	} else {
		// pkgPath is already a directory
		actualPkgDir = absPkgPath
	}

	// Check if this is a Godog test package
	isGodogTestPackage := fileExists(filepath.Join(actualPkgDir, "godog_test.go"))

	// Create output file paths for test results
	// Go packages: organized by module like "r2r-cli/internal/conf.log"
	// Feature files: organized like "eac-commands/templates.log"
	var logFilePath string

	if relFeatureFile != "" {
		// For feature files, strip "specs/" prefix and use parent dir for output
		// e.g., "specs/eac-commands/templates/specification.feature"
		//    -> output: "eac-commands/templates.log"
		featureDir := filepath.Dir(relFeatureFile)
		featureDir = strings.TrimPrefix(featureDir, "specs/")
		featureDir = strings.TrimPrefix(featureDir, "specs\\") // Windows path separator

		// Use the directory name (last component) as the base filename
		dirName := filepath.Base(featureDir)
		// Use the parent directory as the output directory
		parentDir := filepath.Dir(featureDir)

		pkgOutputDir := filepath.Join(testRunDir, parentDir)
		// Create the output directory for feature files
		if err := os.MkdirAll(pkgOutputDir, 0755); err != nil {
			writeln(multiWriter, "❌ Failed to create output directory: %v", err)
			writeln(multiWriter, "")
			return PackageTestResult{PackageName: pkgPath, LogFilePath: "", TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: []string{}}
		}

		logFilePath = filepath.Join(pkgOutputDir, dirName+".log")
	} else {
		// For Go packages, organize by module directories with flattened subpaths
		// e.g., "go/r2r/cli/internal/conf" -> "r2r-cli/internal-conf.log"
		// Extract module name from path (e.g., "go/r2r/cli" -> "r2r-cli", "go/eac/commands" -> "eac-commands")
		pathParts := strings.Split(filepath.ToSlash(relPkgPath), "/")

		var moduleName string
		var fileName string

		if len(pathParts) >= 3 && pathParts[0] == "go" {
			// New structure: go/eac/commands or go/r2r/cli
			// Module is second and third components joined with hyphen: "go/eac/commands" -> "eac-commands"
			moduleName = pathParts[1] + "-" + pathParts[2]
			// Remaining path components become flattened filename
			if len(pathParts) > 3 {
				// Join remaining parts with hyphens: "internal/conf" -> "internal-conf"
				fileName = strings.Join(pathParts[3:], "-") + ".log"
			} else {
				// For packages at module root (e.g., "go/eac/commands")
				fileName = "root.log"
			}
		} else {
			// Fallback for unexpected paths
			moduleName = strings.Join(pathParts, "-")
			fileName = "root.log"
		}

		// Create module directory
		moduleDir := filepath.Join(testRunDir, moduleName)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			writeln(multiWriter, "❌ Failed to create output directory: %v", err)
			writeln(multiWriter, "")
			return PackageTestResult{PackageName: pkgPath, LogFilePath: "", TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: []string{}}
		}

		// Create log file path directly under module directory
		logFilePath = filepath.Join(moduleDir, fileName)
	}
	logFile, err := os.Create(logFilePath)
	if err != nil {
		writeln(multiWriter, "❌ Failed to create log file: %v", err)
		writeln(multiWriter, "")
		return PackageTestResult{PackageName: pkgPath, LogFilePath: logFilePath, TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: []string{}}
	}
	defer logFile.Close()

	// Track expected output files for validation
	// Note: We always expect log and JSON files because go test always creates them
	expectedFiles := []string{logFilePath}

	// JSON test results file (always created by go test)
	jsonFilePath := strings.TrimSuffix(logFilePath, ".log") + ".json"
	expectedFiles = append(expectedFiles, jsonFilePath)

	// For Godog packages, write feature summaries to log file
	if isGodogTestPackage {
		displayGodogFeatureSummaries(actualPkgDir, logFile)
	}

	// Run go test for this package with test-level parallelism and JSON output
	// If testing a specific feature file, pass it via environment variable (GODOG_PATHS)
	var cmd *exec.Cmd

	// Build go test arguments
	goTestArgs := []string{"test", "-json", "-parallel", fmt.Sprintf("%d", testParallelism)}

	// Add coverage flags if enabled
	var coverageFile string
	if coverage {
		// Generate coverage file alongside the log file
		coverageFile = strings.TrimSuffix(logFilePath, ".log") + ".coverage.out"
		goTestArgs = append(goTestArgs, "-cover", "-coverprofile="+coverageFile)
		// Note: Coverage files are added to expectedFiles AFTER test execution
		// when we confirm they were actually created (see post-execution logic below)
		// This prevents validation errors when no tests run or all are skipped
	}

	cmd = exec.Command("go", goTestArgs...)
	cmd.Dir = actualPkgDir

	// Create a buffer to capture JSON output
	var jsonOutput strings.Builder
	jsonMultiWriter := io.MultiWriter(logFile, &jsonOutput)

	// JSON output goes to both log file and buffer
	cmd.Stdout = jsonMultiWriter
	cmd.Stderr = logFile

	// Set test run ID environment variable for all tests
	// This allows commands like "build" to redirect outputs to test directory
	cmd.Env = os.Environ()
	testRunID := filepath.Base(testRunDir)
	cmd.Env = append(cmd.Env, fmt.Sprintf("R2R_TEST_RUN_ID=%s", testRunID))

	// For Godog test packages, set GODOG environment variables
	if isGodogTestPackage {

		// Set format for console output
		cmd.Env = append(cmd.Env, "GODOG_FORMAT=progress")

		// Pass suite tag filter to godog for scenario filtering
		// The filter uses exclusion-based logic that accounts for inference rules,
		// so it works correctly even when scenarios lack explicit L-level tags
		if suiteTagFilter != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_SUITE_TAGS=%s", suiteTagFilter))
			fmt.Fprintf(logFile, "🏷️  Suite tag filter: %s\n", suiteTagFilter)
		}

		// Only generate reports when testing specific feature files
		// Test orchestrator packages (go/eac/specs/impl/*) should not generate reports
		// Reports are only generated for individual feature files in specs/
		if relFeatureFile != "" {
			// Set output directory for cucumber.json/junit.xml reports
			// For feature files, this is the parent directory (e.g., paths.OutTestRelPath + "/commit/eac-commands/")
			reportOutputDir := filepath.Dir(logFilePath)
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_OUTPUT_DIR=%s", reportOutputDir))

			// Set report format (cucumber or junit)
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_REPORT_FORMAT=%s", reportFormat))

			// Extract report name from log file path with appropriate extension
			// cucumber -> "templates.cucumber.json", junit -> "templates.junit.xml"
			logBaseName := filepath.Base(logFilePath)
			var reportName string
			if reportFormat == "junit" {
				reportName = strings.TrimSuffix(logBaseName, ".log") + ".junit.xml"
			} else {
				reportName = strings.TrimSuffix(logBaseName, ".log") + ".cucumber.json"
			}
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_REPORT_NAME=%s", reportName))

			// Note: Report file is added to expectedFiles AFTER test execution
			// when we confirm tests actually ran (see post-execution logic below)
			// This prevents validation errors when all scenarios are skipped
		}

		// If testing a specific feature file, set GODOG_PATHS environment variable
		if absFeatureFile != "" {
			// Get relative path from test runner dir to feature file
			relFeaturePath, err := filepath.Rel(actualPkgDir, absFeatureFile)
			if err != nil {
				relFeaturePath = absFeatureFile
			}
			// Convert to forward slashes for godog
			relFeaturePath = filepath.ToSlash(relFeaturePath)
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_PATHS=%s", relFeaturePath))
			fmt.Fprintf(logFile, "🎯 Testing specific feature: %s\n", relFeaturePath)
		}

		// Note: Skip tag filtering is now handled inside godog_test.go via tag contract
		// The godog test loads the tag contract and builds skip filter automatically
		fmt.Fprintf(logFile, "✨ Skip tags will be loaded from tag contract in godog_test.go\n\n")
		if relFeatureFile != "" {
			logBaseName := filepath.Base(logFilePath)
			var reportName string
			if reportFormat == "junit" {
				reportName = strings.TrimSuffix(logBaseName, ".log") + ".junit.xml"
			} else {
				reportName = strings.TrimSuffix(logBaseName, ".log") + ".cucumber.json"
			}
			fmt.Fprintf(logFile, "📊 Test report will be saved as: %s\n\n", reportName)
		}
	}

	err = cmd.Run()

	// Save JSON test results to file
	if jsonOutput.Len() > 0 {
		jsonFilePath := strings.TrimSuffix(logFilePath, ".log") + ".json"
		if writeErr := os.WriteFile(jsonFilePath, []byte(jsonOutput.String()), 0644); writeErr != nil {
			fmt.Fprintf(logFile, "Warning: failed to save JSON results: %v\n", writeErr)
		}
	}

	// Convert coverage profile to JSON if coverage was enabled
	if coverage && coverageFile != "" {
		if _, statErr := os.Stat(coverageFile); statErr == nil {
			// Coverage file exists - add it to expected files for validation
			expectedFiles = append(expectedFiles, coverageFile)

			coverageJSONFile := strings.TrimSuffix(coverageFile, ".out") + ".json"
			if convertErr := convertCoverageToJSON(coverageFile, coverageJSONFile); convertErr != nil {
				fmt.Fprintf(logFile, "Warning: failed to convert coverage to JSON: %v\n", convertErr)
			} else {
				// Conversion succeeded - also expect the JSON file
				expectedFiles = append(expectedFiles, coverageJSONFile)
			}
		}
		// If coverage file doesn't exist, tests didn't run - don't add to expectedFiles
	}

	if err != nil {
		testType := "go"
		resultStr := fmt.Sprintf("0/%d", len(tests))
		if isGodogTestPackage {
			testType = "godog"
			resultStr = "-"
		}

		writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconFail, pkgName, testType, resultStr))
		return PackageTestResult{PackageName: pkgPath, LogFilePath: logFilePath, TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: expectedFiles}
	}

	displayTestType := "go"
	resultStr := fmt.Sprintf("%d/%d", len(tests), len(tests))
	testsPassed := len(tests)
	testsFailed := 0
	testsTotal := len(tests)

	if isGodogTestPackage {
		displayTestType = "godog"
		// Extract scenario counts from godog test output
		passedScenarios, failedScenarios := extractGodogScenarioCountsDetailed(logFilePath)
		testsPassed = passedScenarios
		testsFailed = failedScenarios
		testsTotal = passedScenarios + failedScenarios

		if testsTotal > 0 {
			resultStr = fmt.Sprintf("%d/%d", passedScenarios, testsTotal)

			// Only expect report files if tests actually executed
			// If all scenarios were skipped (testsTotal = 0), report files won't be created
			if relFeatureFile != "" {
				// Add report file to expected files now that we know tests ran
				reportOutputDir := filepath.Dir(logFilePath)
				logBaseName := filepath.Base(logFilePath)
				var reportName string
				if reportFormat == "junit" {
					reportName = strings.TrimSuffix(logBaseName, ".log") + ".junit.xml"
				} else {
					reportName = strings.TrimSuffix(logBaseName, ".log") + ".cucumber.json"
				}
				reportFilePath := filepath.Join(reportOutputDir, reportName)
				expectedFiles = append(expectedFiles, reportFilePath)
			}
		} else {
			resultStr = "-"
		}
	}

	// Use appropriate icon based on failures
	icon := output.IconPass
	if testsFailed > 0 {
		icon = output.IconFail
	}

	writeln(multiWriter, "%s", output.ResultLineNoTime(icon, pkgName, displayTestType, resultStr))
	return PackageTestResult{PackageName: pkgPath, LogFilePath: logFilePath, TestsPassed: testsPassed, TestsFailed: testsFailed, TestsSkipped: 0, TestsTotal: testsTotal, PackageFailed: testsFailed > 0, ExpectedFiles: expectedFiles}
}

// extractGodogScenarioCounts parses godog JSON test results to extract scenario counts
// Returns: "(passed/total)" or empty string if counts not found
func extractGodogScenarioCounts(logPath string) string {
	passed, failed := extractGodogScenarioCountsDetailed(logPath)
	total := passed + failed
	if total == 0 {
		return ""
	}
	return fmt.Sprintf("(%d/%d)", passed, total)
}

// extractGodogScenarioCountsDetailed parses godog test results from go test JSON output
// to extract scenario counts.
// Returns: (passedScenarios, failedScenarios)
func extractGodogScenarioCountsDetailed(logPath string) (int, int) {
	// The log file contains `go test -json` output
	// Parse it to count passed/failed scenarios
	events, err := testjson.ParseJSONFile(logPath)
	if err != nil {
		// Fall back to zero if we can't parse the log
		return 0, 0
	}

	passedScenarios := 0
	failedScenarios := 0

	// For Godog tests, each test event represents a scenario
	// Filter out the top-level test function (e.g., "TestRepositoryFeatures")
	// Scenario names are like "TestRepositoryFeatures/All_module_dependencies_in_repository_are_valid"
	for _, event := range events {
		// Only count sub-tests (scenarios), not the top-level test function
		if event.Test != "" && strings.Contains(event.Test, "/") {
			if event.Action == "pass" {
				passedScenarios++
			} else if event.Action == "fail" {
				failedScenarios++
			}
		}
	}

	return passedScenarios, failedScenarios
}

// extractUndefinedSteps scans test logs for undefined step definitions
func extractUndefinedSteps(testRunDir string) []string {
	var undefinedSteps []string
	uniqueSteps := make(map[string]bool)

	// Find all .log files in test run directory
	entries, err := os.ReadDir(testRunDir)
	if err != nil {
		return undefinedSteps
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}

		logPath := filepath.Join(testRunDir, entry.Name())
		content, err := os.ReadFile(logPath)
		if err != nil {
			continue
		}

		// Parse log for "step is undefined: <step text>" lines
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.Contains(line, "step is undefined:") {
				// Extract step text after "step is undefined: "
				parts := strings.SplitN(line, "step is undefined:", 2)
				if len(parts) == 2 {
					stepText := strings.TrimSpace(parts[1])
					// Remove ANSI color codes
					stepText = stripANSI(stepText)
					if stepText != "" && !uniqueSteps[stepText] {
						uniqueSteps[stepText] = true
						undefinedSteps = append(undefinedSteps, stepText)
					}
				}
			}
		}
	}

	return undefinedSteps
}

// stripANSI removes ANSI color codes from a string
func stripANSI(str string) string {
	// Remove ANSI escape sequences like [33m, [0m, etc.
	result := ""
	inEscape := false
	for i := 0; i < len(str); i++ {
		if str[i] == '\x1b' && i+1 < len(str) && str[i+1] == '[' {
			inEscape = true
			i++ // Skip the '['
			continue
		}
		if inEscape {
			if (str[i] >= 'A' && str[i] <= 'Z') || (str[i] >= 'a' && str[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result += string(str[i])
	}
	return result
}

// analyzeTestFailure analyzes a test log to determine the failure reason
func analyzeTestFailure(logPath string, isGodog bool) string {
	content, err := os.ReadFile(logPath)
	if err != nil {
		return "unknown error"
	}

	logStr := string(content)

	if isGodog {
		// Check for undefined steps
		undefinedCount := strings.Count(logStr, "step is undefined:")
		if undefinedCount > 0 {
			return fmt.Sprintf("missing step implementations (%d undefined)", undefinedCount)
		}

		// Check for ambiguous steps
		if strings.Contains(logStr, "ambiguous step definition") {
			ambiguousCount := strings.Count(logStr, "ambiguous step definition")
			return fmt.Sprintf("ambiguous step definitions (%d)", ambiguousCount)
		}

		// Parse scenario summary
		lines := strings.Split(logStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "scenarios") {
				// Extract failure count from line like "87 scenarios (13 passed, 16 failed, 66 undefined)"
				if strings.Contains(line, "failed") {
					// Try to extract the failed count
					parts := strings.Split(line, ",")
					for _, part := range parts {
						part = stripANSI(strings.TrimSpace(part))
						if strings.Contains(part, "failed") {
							fields := strings.Fields(part)
							if len(fields) >= 1 {
								count := fields[0]
								return fmt.Sprintf("%s failing scenarios", count)
							}
						}
					}
				}
			}
		}

		return "BDD test failures"
	}

	// For regular Go tests, count FAIL lines
	failCount := 0
	lines := strings.Split(logStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "--- FAIL:") {
			failCount++
		}
	}

	if failCount > 0 {
		return fmt.Sprintf("%d failing tests", failCount)
	}

	return "test failures"
}

// extractModuleFromPath extracts the module moniker from a test file path
// Handles go/eac/<module>/..., go/r2r/<module>/..., and specs/<module>/... formats
// Supports both absolute and relative paths
func extractModuleFromPath(filePath string) string {
	// Normalize path separators to forward slashes
	normalizedPath := filepath.ToSlash(filePath)

	// Find "/go/eac/" in the path (handles both absolute and relative paths)
	// For paths like /project/go/eac/commands/..., extract "eac-commands"
	for _, boundary := range []string{"/go/eac/", "/go/r2r/"} {
		idx := strings.Index(normalizedPath, boundary)
		if idx >= 0 {
			// Extract module name from path after boundary
			relativePath := normalizedPath[idx+len(boundary):]
			parts := strings.Split(relativePath, "/")
			if len(parts) >= 1 && parts[0] != "" {
				// Return as eac-<part1> or r2r-<part1>
				prefix := "eac"
				if boundary == "/go/r2r/" {
					prefix = "r2r"
				}
				return prefix + "-" + parts[0]
			}
		}
	}

	// Check if path contains "/specs/" (handles specs/eac-*, specs/github, specs/repository, etc.)
	specsIndex := strings.Index(normalizedPath, "/specs/")
	if specsIndex >= 0 {
		// Extract from "/specs/" onwards
		relativePath := normalizedPath[specsIndex+1:]
		// Format: specs/<module>/...
		parts := strings.Split(strings.TrimPrefix(relativePath, "specs/"), "/")
		if len(parts) >= 1 && parts[0] != "" {
			// Return the first part (e.g., "eac-commands", "github", "repository")
			return parts[0]
		}
	}

	// Also check for paths starting with "go/eac/" or "go/r2r/" (relative paths)
	for _, prefix := range []string{"go/eac/", "go/r2r/"} {
		if strings.HasPrefix(normalizedPath, prefix) {
			parts := strings.Split(normalizedPath[len(prefix):], "/")
			if len(parts) >= 1 && parts[0] != "" {
				monikerPrefix := "eac"
				if prefix == "go/r2r/" {
					monikerPrefix = "r2r"
				}
				return monikerPrefix + "-" + parts[0]
			}
		}
	}

	if strings.HasPrefix(normalizedPath, "specs/") {
		parts := strings.Split(strings.TrimPrefix(normalizedPath, "specs/"), "/")
		if len(parts) >= 1 && parts[0] != "" {
			return parts[0]
		}
	}

	// Fallback: return empty string
	return ""
}

// mapGOOSToDepTag maps runtime.GOOS values to dependency tag names
func mapGOOSToDepTag(goos string) string {
	switch goos {
	case "linux":
		return "linux"
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return goos
	}
}

// filterByOSCompatibility filters tests based on OS-specific dependencies
// Tests with deps:linux only run on Linux, deps:macos only on macOS, deps:windows only on Windows
// Tests without any OS-specific deps run everywhere (OS-agnostic by default)
// Tests with multiple OS deps (e.g., deps:linux AND deps:macos) run on any of those OSes
func filterByOSCompatibility(tests []testing.TestReference, _ io.Writer) []testing.TestReference {
	currentOS := mapGOOSToDepTag(runtime.GOOS)
	compatible := []testing.TestReference{}

	for _, test := range tests {
		// Check if test has any OS-specific dependencies
		hasOSDep := false
		matchesCurrentOS := false

		for _, dep := range test.SystemDependencies {
			// Check if this is an OS dependency
			if testing.IsOSPlatformDep(dep) {
				hasOSDep = true
				if dep == currentOS {
					matchesCurrentOS = true
				}
			}
		}

		// Include test if:
		// 1. It has no OS-specific deps (runs on any OS), OR
		// 2. It has an OS dep that matches the current OS
		if !hasOSDep || matchesCurrentOS {
			compatible = append(compatible, test)
		}
	}

	return compatible
}

// CoverageReport represents the JSON structure for coverage data
type CoverageReport struct {
	Mode     string           `json:"mode"`
	Files    []FileCoverage   `json:"files"`
	Summary  CoverageSummary  `json:"summary"`
}

// FileCoverage represents coverage data for a single file
type FileCoverage struct {
	FileName    string   `json:"fileName"`
	Blocks      []Block  `json:"blocks"`
	TotalLines  int      `json:"totalLines"`
	CoveredLines int     `json:"coveredLines"`
	Coverage    float64  `json:"coverage"`
}

// Block represents a coverage block (start line, end line, statement count, hit count)
type Block struct {
	StartLine int `json:"startLine"`
	StartCol  int `json:"startCol"`
	EndLine   int `json:"endLine"`
	EndCol    int `json:"endCol"`
	NumStmt   int `json:"numStmt"`
	Count     int `json:"count"`
}

// CoverageSummary provides overall coverage statistics
type CoverageSummary struct {
	TotalFiles    int     `json:"totalFiles"`
	TotalLines    int     `json:"totalLines"`
	CoveredLines  int     `json:"coveredLines"`
	TotalCoverage float64 `json:"totalCoverage"`
}

// convertCoverageToJSON converts a Go coverage profile to JSON format
func convertCoverageToJSON(coverageFile, jsonFile string) error {
	file, err := os.Open(coverageFile)
	if err != nil {
		return fmt.Errorf("failed to open coverage file: %w", err)
	}
	defer file.Close()

	report := CoverageReport{
		Files: []FileCoverage{},
	}

	fileMap := make(map[string]*FileCoverage)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// First line is the mode
		if strings.HasPrefix(line, "mode:") {
			report.Mode = strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
			continue
		}

		// Parse coverage line: filename:startLine.startCol,endLine.endCol numStmt count
		// Example: github.com/user/pkg/file.go:10.2,15.5 3 1
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}

		// Parse file:positions
		colonIdx := strings.LastIndex(parts[0], ":")
		if colonIdx == -1 {
			continue
		}

		fileName := parts[0][:colonIdx]
		positions := parts[0][colonIdx+1:]

		// Parse positions: startLine.startCol,endLine.endCol
		positionParts := strings.Split(positions, ",")
		if len(positionParts) != 2 {
			continue
		}

		startParts := strings.Split(positionParts[0], ".")
		endParts := strings.Split(positionParts[1], ".")
		if len(startParts) != 2 || len(endParts) != 2 {
			continue
		}

		startLine, _ := strconv.Atoi(startParts[0])
		startCol, _ := strconv.Atoi(startParts[1])
		endLine, _ := strconv.Atoi(endParts[0])
		endCol, _ := strconv.Atoi(endParts[1])
		numStmt, _ := strconv.Atoi(parts[1])
		count, _ := strconv.Atoi(parts[2])

		block := Block{
			StartLine: startLine,
			StartCol:  startCol,
			EndLine:   endLine,
			EndCol:    endCol,
			NumStmt:   numStmt,
			Count:     count,
		}

		// Get or create file coverage entry
		fileCov, exists := fileMap[fileName]
		if !exists {
			fileCov = &FileCoverage{
				FileName: fileName,
				Blocks:   []Block{},
			}
			fileMap[fileName] = fileCov
		}

		fileCov.Blocks = append(fileCov.Blocks, block)
		fileCov.TotalLines += (endLine - startLine + 1)
		if count > 0 {
			fileCov.CoveredLines += (endLine - startLine + 1)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read coverage file: %w", err)
	}

	// Calculate per-file coverage and add to report
	var totalLines, coveredLines int
	for _, fileCov := range fileMap {
		if fileCov.TotalLines > 0 {
			fileCov.Coverage = float64(fileCov.CoveredLines) / float64(fileCov.TotalLines) * 100
		}
		report.Files = append(report.Files, *fileCov)
		totalLines += fileCov.TotalLines
		coveredLines += fileCov.CoveredLines
	}

	// Calculate summary
	report.Summary = CoverageSummary{
		TotalFiles:   len(report.Files),
		TotalLines:   totalLines,
		CoveredLines: coveredLines,
	}
	if totalLines > 0 {
		report.Summary.TotalCoverage = float64(coveredLines) / float64(totalLines) * 100
	}

	// Write JSON file
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal coverage to JSON: %w", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write coverage JSON: %w", err)
	}

	return nil
}

// TestTypeHandler is a function that runs tests for a specific test type (mocha, jest, etc.)
type TestTypeHandler func(pkgPath string, tests []testing.TestReference, multiWriter io.Writer, testRunDir string, reportFormat string, suiteTagFilter string) PackageTestResult

// testTypeHandlers maps test types to their suite-level handlers.
// Unit tests: mocha (npm)
// Spec tests: tscucumber (npm)
// Go tests (gotest, godog) are handled by default Go execution path.
var testTypeHandlers = map[string]TestTypeHandler{
	"mocha":      runMochaTests,
	"tscucumber": runTsCucumberTests,
}

// getPackageTestType determines the test type for a package.
// If all tests have the same type, returns that type.
// If mixed or empty, returns empty string (use default Go handler).
func getPackageTestType(tests []testing.TestReference) string {
	if len(tests) == 0 {
		return ""
	}

	firstType := tests[0].Type
	for _, test := range tests[1:] {
		if test.Type != firstType {
			return "" // Mixed types, use default
		}
	}

	return firstType
}

// getTestTypeHandler returns the handler for a test type if one is registered.
// Returns false if no handler exists (should use default Go test execution).
func getTestTypeHandler(testType string) (TestTypeHandler, bool) {
	// Go and godog tests are handled by the default Go test execution
	if testType == "" || testType == "go" || testType == "godog" {
		return nil, false
	}

	handler, ok := testTypeHandlers[testType]
	return handler, ok
}

// runMochaTests runs mocha tests for a TypeScript/npm package.
// It dispatches to npm test with Mocha grep pattern for tag filtering.
func runMochaTests(pkgPath string, tests []testing.TestReference, multiWriter io.Writer, testRunDir string, reportFormat string, suiteTagFilter string) PackageTestResult {
	start := time.Now()

	// Get workspace root
	workspaceRootNative, err := repository.GetRepositoryRoot("")
	if err != nil {
		workspaceRootNative = "."
	}

	pkgName := filepath.ToSlash(pkgPath)

	// Find the module root (where package.json is located)
	// pkgPath might be something like ".vscode/extensions/vscode-ext-commit/test"
	// We need to find the parent directory containing package.json
	moduleRoot := filepath.Join(workspaceRootNative, pkgPath)

	// Walk up to find package.json
	for {
		packageJSON := filepath.Join(moduleRoot, "package.json")
		if fileExists(packageJSON) {
			break
		}
		parent := filepath.Dir(moduleRoot)
		if parent == moduleRoot {
			// Reached root without finding package.json
			writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconFail, pkgName, "npm", "no package.json"))
			return PackageTestResult{
				PackageName:   pkgPath,
				LogFilePath:   "",
				TestsPassed:   0,
				TestsFailed:   len(tests),
				TestsTotal:    len(tests),
				PackageFailed: true,
				Duration:      time.Since(start),
			}
		}
		moduleRoot = parent
	}

	// Create log file for mocha output
	// Organize under module name
	relModuleRoot, _ := filepath.Rel(workspaceRootNative, moduleRoot)
	moduleName := strings.ReplaceAll(filepath.ToSlash(relModuleRoot), "/", "-")
	logFilePath := filepath.Join(testRunDir, moduleName+".log")

	logFile, err := os.Create(logFilePath)
	if err != nil {
		writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconFail, pkgName, "npm", "log create failed"))
		return PackageTestResult{
			PackageName:   pkgPath,
			LogFilePath:   "",
			TestsPassed:   0,
			TestsFailed:   len(tests),
			TestsTotal:    len(tests),
			PackageFailed: true,
			Duration:      time.Since(start),
		}
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "=== Testing npm package: %s ===\n", pkgName)
	fmt.Fprintf(logFile, "Module root: %s\n", moduleRoot)
	fmt.Fprintf(logFile, "Tests selected by suite: %d\n", len(tests))
	fmt.Fprintf(logFile, "\n")

	// Run npm test without grep filtering
	// The suite has already selected which tests to run based on discovered tags.
	// Mocha will run all tests, but since we're running from the suite infrastructure,
	// the test count is already correct from discovery.
	args := []string{"test"}

	fmt.Fprintf(logFile, "Running: npm %s\n\n", strings.Join(args, " "))

	// Execute npm test
	wrappedName, wrappedArgs := platform.WrapCommand("npm", args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = moduleRoot
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Run()

	duration := time.Since(start)

	if err != nil {
		writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconFail, pkgName, "npm", fmt.Sprintf("0/%d", len(tests))))
		return PackageTestResult{
			PackageName:   pkgPath,
			LogFilePath:   logFilePath,
			TestsPassed:   0,
			TestsFailed:   len(tests),
			TestsTotal:    len(tests),
			PackageFailed: true,
			Duration:      duration,
			ExpectedFiles: []string{logFilePath},
		}
	}

	writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconPass, pkgName, "npm", fmt.Sprintf("%d/%d", len(tests), len(tests))))
	return PackageTestResult{
		PackageName:   pkgPath,
		LogFilePath:   logFilePath,
		TestsPassed:   len(tests),
		TestsFailed:   0,
		TestsTotal:    len(tests),
		PackageFailed: false,
		Duration:      duration,
		ExpectedFiles: []string{logFilePath},
	}
}

// runTsCucumberTests runs cucumber-js BDD tests for a TypeScript/npm package.
func runTsCucumberTests(pkgPath string, tests []testing.TestReference, multiWriter io.Writer, testRunDir string, reportFormat string, suiteTagFilter string) PackageTestResult {
	start := time.Now()

	workspaceRootNative, err := repository.GetRepositoryRoot("")
	if err != nil {
		workspaceRootNative = "."
	}

	pkgName := filepath.ToSlash(pkgPath)

	// Get module root - try from pkgPath first (format: moduleRoot:featurePath)
	// This is set by findTscucumberTestRunner in groupTestsByPackage
	var moduleRoot string
	if strings.Contains(pkgPath, ":") {
		parts := strings.SplitN(pkgPath, ":", 2)
		moduleRoot = filepath.Join(workspaceRootNative, parts[0])
	}

	// Fallback: get module root from test's module dependency (@depm: tag)
	if moduleRoot == "" && len(tests) > 0 && len(tests[0].ModuleDependencies) > 0 {
		moniker := tests[0].ModuleDependencies[0]
		registry, err := modules.LoadFromWorkspace(workspaceRootNative)
		if err == nil {
			if module, ok := registry.Get(moniker); ok {
				moduleRoot = filepath.Join(workspaceRootNative, module.Files.Root)
			}
		}
	}

	// Verify module root was found
	if moduleRoot == "" {
		writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconFail, pkgName, "tscucumber", "no module"))
		return PackageTestResult{
			PackageName:   pkgPath,
			TestsFailed:   len(tests),
			TestsTotal:    len(tests),
			PackageFailed: true,
			Duration:      time.Since(start),
		}
	}

	packageJSON := filepath.Join(moduleRoot, "package.json")
	if !fileExists(packageJSON) {
		writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconFail, pkgName, "tscucumber", "no package.json"))
		return PackageTestResult{
			PackageName:   pkgPath,
			TestsFailed:   len(tests),
			TestsTotal:    len(tests),
			PackageFailed: true,
			Duration:      time.Since(start),
		}
	}

	// Create log file
	relModuleRoot, _ := filepath.Rel(workspaceRootNative, moduleRoot)
	moduleName := strings.ReplaceAll(filepath.ToSlash(relModuleRoot), "/", "-")
	logFilePath := filepath.Join(testRunDir, moduleName+"-cucumber.log")

	logFile, err := os.Create(logFilePath)
	if err != nil {
		writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconFail, pkgName, "tscucumber", "log create failed"))
		return PackageTestResult{
			PackageName:   pkgPath,
			TestsFailed:   len(tests),
			TestsTotal:    len(tests),
			PackageFailed: true,
			Duration:      time.Since(start),
		}
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "=== Testing cucumber specs: %s ===\n", pkgName)
	fmt.Fprintf(logFile, "Module root: %s\n", moduleRoot)
	fmt.Fprintf(logFile, "Tests selected by suite: %d\n", len(tests))
	fmt.Fprintf(logFile, "\n")

	// Build cucumber-js command
	args := []string{"cucumber-js"}

	// Add tag filter if provided
	if suiteTagFilter != "" {
		// Convert godog tag format to cucumber tag expression
		tagExpr := convertToCucumberTagExpr(suiteTagFilter)
		if tagExpr != "" {
			args = append(args, "--tags", tagExpr)
		}
	}

	fmt.Fprintf(logFile, "Running: npx %s\n\n", strings.Join(args, " "))

	// Execute cucumber-js via npx
	wrappedName, wrappedArgs := platform.WrapCommand("npx", args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = moduleRoot
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Run()
	duration := time.Since(start)

	if err != nil {
		writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconFail, pkgName, "tscucumber", fmt.Sprintf("0/%d", len(tests))))
		return PackageTestResult{
			PackageName:   pkgPath,
			LogFilePath:   logFilePath,
			TestsPassed:   0,
			TestsFailed:   len(tests),
			TestsTotal:    len(tests),
			PackageFailed: true,
			Duration:      duration,
			ExpectedFiles: []string{logFilePath},
		}
	}

	writeln(multiWriter, "%s", output.ResultLineNoTime(output.IconPass, pkgName, "tscucumber", fmt.Sprintf("%d/%d", len(tests), len(tests))))
	return PackageTestResult{
		PackageName:   pkgPath,
		LogFilePath:   logFilePath,
		TestsPassed:   len(tests),
		TestsFailed:   0,
		TestsTotal:    len(tests),
		PackageFailed: false,
		Duration:      duration,
		ExpectedFiles: []string{logFilePath},
	}
}

// convertToCucumberTagExpr converts godog tag format to cucumber-js tag expression.
// Godog: "@L0,@L1 && ~@skip:wip" → Cucumber: "(@L0 or @L1) and not @skip:wip"
func convertToCucumberTagExpr(godogTags string) string {
	if godogTags == "" {
		return ""
	}

	var parts []string
	// Split by && for AND conditions
	andParts := strings.Split(godogTags, " && ")
	for _, part := range andParts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "~") {
			// Negation: ~@tag → not @tag
			parts = append(parts, "not "+strings.TrimPrefix(part, "~"))
		} else if strings.Contains(part, ",") {
			// OR: @L0,@L1 → (@L0 or @L1)
			orTags := strings.Split(part, ",")
			parts = append(parts, "("+strings.Join(orTags, " or ")+")")
		} else {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " and ")
}

// runTestSuiteForModuleInternal runs the test suite command with a module filter.
// This is the internal implementation used by the testers package.
func runTestSuiteForModuleInternal(moniker string, suiteName string) int {
	// Redirect to test suite with module filter
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		oldArgs[0],
		"test",
		"suite",
		suiteName,
		"--module",
		moniker,
	}

	return TestSuite()
}

func init() {
	// Set up the callback for testers package to run test suites
	testers.RunTestSuiteForModule = runTestSuiteForModuleInternal
}

