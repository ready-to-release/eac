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
// HasSideEffects: false
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

	"github.com/ready-to-release/eac/src/commands/impl/test/internal/reporter"
	"github.com/ready-to-release/eac/src/commands/impl/test/internal/testjson"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	contractsreports "github.com/ready-to-release/eac/src/core/contracts/reports"
	moduledeps "github.com/ready-to-release/eac/src/core/module-deps"
	"github.com/ready-to-release/eac/src/core/platform"
	"github.com/ready-to-release/eac/src/core/repository"
	systemdeps "github.com/ready-to-release/eac/src/core/system-deps"
	"github.com/ready-to-release/eac/src/core/testing"
)

func init() {
	registry.Register(TestSuite)
}

// writeln writes a formatted string with platform-specific line ending to the writer
func writeln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

// PackageTestResult holds detailed test execution results for a single package
type PackageTestResult struct {
	TestsPassed   int      // Number of individual tests that passed
	TestsFailed   int      // Number of individual tests that failed
	TestsSkipped  int      // Number of individual tests that were skipped
	TestsTotal    int      // Total number of tests in this package
	PackageFailed bool     // Whether the package execution itself failed
	ExpectedFiles []string // Files that should have been created by this test run
}

// TestSuite runs tests for a specific test suite
func TestSuite() int {
	// Parse arguments and flags
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: missing suite name\n")
		fmt.Fprintf(os.Stderr, "Usage: test suite <suite-name> [flags]\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		fmt.Fprintf(os.Stderr, "  --skip-deps    Skip dependency verification\n")
		fmt.Fprintf(os.Stderr, "  --list-only    List tests without running them\n")
		fmt.Fprintf(os.Stderr, "  --sequential   Run tests sequentially (for debugging)\n")
		fmt.Fprintf(os.Stderr, "  --parallel     Run tests in parallel (DEFAULT, explicit override)\n")
		fmt.Fprintf(os.Stderr, "  --as-cucumber  Generate Cucumber JSON reports (DEFAULT)\n")
		fmt.Fprintf(os.Stderr, "  --as-junit     Generate JUnit XML reports\n")
		fmt.Fprintf(os.Stderr, "  --coverage     Generate coverage reports (coverage.out, coverage.json)\n")
		fmt.Fprintf(os.Stderr, "\nDefault: Tests run in parallel for optimal performance.\n")
		fmt.Fprintf(os.Stderr, "Use --sequential if you need deterministic ordering or debugging.\n")
		fmt.Fprintf(os.Stderr, "\nAvailable suites:\n")
		for _, suite := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", suite)
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
		} else if arg == "--module" {
			// Read module names from next argument (comma-separated)
			if i+1 >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "Error: --module requires one or more module names\n")
				fmt.Fprintf(os.Stderr, "Usage: --module <name> or --module <name1>,<name2>\n")
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
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid flags: --skip-deps, --list-only, --sequential, --parallel, --module <name>, --as-junit, --as-cucumber, --coverage\n")
			return 1
		}
	}

	// Get repository root
	workspaceRootNative, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}
	// Codebase uses Unix-style paths throughout - normalize for path comparisons
	workspaceRoot := filepath.ToSlash(workspaceRootNative)

	// Get the test suite
	suite, err := testing.GetSuite(suiteName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nAvailable suites:\n")
		for _, s := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", s)
		}
		return 1
	}

	fmt.Printf("🧪 Running test suite: %s\n", suite.Name)
	fmt.Printf("Description: %s\n\n", suite.Description)

	// Purge and create suite output directory
	testRunDir := filepath.Join(workspaceRootNative, "out", "test", suiteName)
	if err := os.RemoveAll(testRunDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to purge test directory: %v\n", err)
	}
	if err := os.MkdirAll(testRunDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create test run directory: %v\n", err)
		return 1
	}

	// Create log file
	logPath := filepath.Join(testRunDir, "test-suite.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create log file: %v\n", err)
		return 1
	}
	defer logFile.Close()

	// Track start time for duration calculation
	startTime := time.Now()

	// Create multi-writer to log to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Phase 1: Discover all tests (Go + Godog)
	writeln(multiWriter, "=== Phase 1: Test Discovery ===")

	allTests, err := testing.DiscoverAllTests(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover tests: %v\n", err)
		return 1
	}

	writeln(multiWriter, "Discovered %d tests", len(allTests))
	writeln(multiWriter, "")

	// Phase 2: Apply inference rules
	writeln(multiWriter, "=== Phase 2: Inference Engine ===")
	allTests = testing.ApplyInferences(allTests, suite.Inferences)
	writeln(multiWriter, "Applied %d inference rules", len(suite.Inferences))

	// Load module registry for module-based inference
	moduleReport, err := contractsreports.GetModuleContracts(workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load module contracts: %v\n", err)
	} else {
		// Infer system dependencies from module dependencies
		allTests = testing.InferSystemDepsFromModuleDeps(allTests, moduleReport.Registry)
		writeln(multiWriter, "Inferred system deps from module types")
	}

	// Infer OS platform: add @deps:os-agnostic to tests without OS-specific deps
	allTests = testing.InferOSPlatform(allTests)
	writeln(multiWriter, "Inferred OS platform dependencies")
	writeln(multiWriter, "")

	// Phase 3: Select tests for suite
	writeln(multiWriter, "=== Phase 3: Suite Selection ===")
	selectedTests, selectionStats := suite.SelectTestsWithStats(allTests)
	writeln(multiWriter, "Selected %d tests for suite '%s'", len(selectedTests), suite.Moniker)

	// Phase 3.5: Apply module filter if specified
	if len(moduleFilters) > 0 {
		writeln(multiWriter, "Filtering by modules: %s", strings.Join(moduleFilters, ", "))
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
			writeln(multiWriter, "DEBUG: No tests matched. Modules found in selected tests:")
			for mod, count := range foundModules {
				if mod != "" {
					writeln(multiWriter, "  - %s (%d tests)", mod, count)
				} else {
					writeln(multiWriter, "  - [empty module name] (%d tests)", count)
				}
			}
			// Show a few sample paths
			if len(selectedTests) > 0 {
				writeln(multiWriter, "DEBUG: Sample file paths (first 5):")
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
	writeln(multiWriter, "=== Phase 4: Dependency Verification ===")
	systemDeps := testing.GetSystemDependencies(productionTests)
	moduleDeps := testing.GetModuleDependencies(productionTests)

	allDeps := append(append([]string{}, systemDeps...), moduleDeps...)

	// Set TEST_AI_MOCK for tests that use @deps:ai - they use mock AI providers
	for _, dep := range systemDeps {
		if dep == "@deps:ai" {
			os.Setenv("TEST_AI_MOCK", "true")
			break
		}
	}

	if len(allDeps) == 0 {
		writeln(multiWriter, "No dependencies required")
		writeln(multiWriter, "")
	} else {
		writeln(multiWriter, "System dependencies: %s", strings.Join(systemDeps, ", "))
		writeln(multiWriter, "Module dependencies: %s", strings.Join(moduleDeps, ", "))

		if !skipDeps {
			hasFailures := false

			// Verify system dependencies
			sysResults := systemdeps.VerifyAll(systemDeps)
			for _, result := range sysResults {
				if result.Available {
					writeln(multiWriter, "✅ %s - %s", result.Dependency, result.Version)
				} else {
					writeln(multiWriter, "❌ %s - not available", result.Dependency)
					hasFailures = true
				}
			}

			// Verify module dependencies
			modResults := moduledeps.VerifyAll(moduleDeps)
			for _, result := range modResults {
				if result.Available {
					writeln(multiWriter, "✅ %s - %s", result.Dependency, result.Version)
				} else {
					writeln(multiWriter, "❌ %s - not available", result.Dependency)
					hasFailures = true
				}
			}

			fmt.Fprintln(multiWriter)

			if hasFailures {
				writeln(multiWriter, "❌ Error: Required dependencies are missing")
				writeln(multiWriter, "Use --skip-deps to run tests anyway")
				return 1
			}
		} else {
			writeln(multiWriter, "Dependency check skipped (--skip-deps)")
			writeln(multiWriter, "")
		}
	}

	// findGodogTestRunner finds the test runner package for a feature file
	// Feature files are in specs/src-<module>/..., test runners are in src/<module>/tests/
	findGodogTestRunner := func(featurePath string) string {
		// Extract module from specs path
		// Example: specs/src-commands/commit/... -> src/commands/tests
		//          specs/src-cli/... -> src/cli/tests
		relPath := strings.TrimPrefix(featurePath, "specs/")
		relPath = strings.TrimPrefix(relPath, "specs\\")

		// Get first path component (e.g., "src-commands")
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		if len(parts) == 0 {
			return "src/commands/tests" // fallback
		}

		// Convert "src-commands" -> "src/commands"
		module := strings.Replace(parts[0], "-", "/", 1)
		return filepath.Join(module, "tests")
	}

	// Phase 5: Run tests
	writeln(multiWriter, "=== Phase 5: Test Execution ===")

	// Group tests by package
	// For Godog tests: need to find their test runner package
	// For Go tests: group by directory as usual
	// IMPORTANT: Convert absolute paths to workspace-relative paths for proper cmd.Dir handling
	testsByPackage := make(map[string][]testing.TestReference)
	for _, test := range productionTests {
		var pkgPath string
		if test.Type == "godog" {
			// Find the test runner package for this feature file
			// Feature files are in specs/, test runners are in src/*/tests/
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

	// Build godog tag filter for suite
	// All tests must have explicit L-tags, so we can use simple tag matching
	// e.g., commit suite: "@L0,@L1,@L2"
	suiteTagFilter := suite.BuildGodogTagFilter()

	if parallel {
		// Package-level parallel: distribute CPU across packages
		// Each package gets a smaller share of CPU cores
		testParallelism = max(2, numCPU/4)
		results = runTestsParallel(testsByPackage, multiWriter, testParallelism, testRunDir, reportFormat, coverage, suiteTagFilter)
	} else {
		// Sequential packages: each package gets full CPU power
		testParallelism = numCPU
		results = runTestsSequential(testsByPackage, multiWriter, testParallelism, testRunDir, reportFormat, coverage, suiteTagFilter)
	}

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

	writeln(multiWriter, "=== Test Run Summary ===")
	writeln(multiWriter, "Suite: %s", suite.Name)
	writeln(multiWriter, "")
	writeln(multiWriter, "Test Selection Breakdown:")
	writeln(multiWriter, "  Tests discovered:        %d", selectionStats.TotalDiscovered)
	writeln(multiWriter, "  - Ignored (@ignore):     %d", selectionStats.Ignored)
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
	writeln(multiWriter, "=== Output File Validation ===")
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
		fmt.Fprintf(os.Stderr, "Warning: failed to collect module reports: %v\n", err)
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
	templatePath := filepath.Join(workspaceRootNative, "templates", "test-reports", "suite-summary.md")
	mdPath := filepath.Join(testRunDir, "test-suite-summary.md")

	renderer := reporter.NewRenderer(templatePath, mdPath, reportData)
	if err := renderer.Render(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to generate markdown report: %v\n", err)
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
	// src/cli/tests -> specs/src-cli
	// src/commands/tests -> specs/src-commands
	var specsDir string
	if strings.Contains(testPkgPath, "cli/tests") || strings.Contains(testPkgPath, "cli\\tests") {
		specsDir = "src-cli"
	} else if strings.Contains(testPkgPath, "commands/tests") || strings.Contains(testPkgPath, "commands\\tests") {
		specsDir = "src-commands"
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

// runTestsSequential runs tests package by package sequentially
func runTestsSequential(testsByPackage map[string][]testing.TestReference, multiWriter io.Writer, testParallelism int, testRunDir string, reportFormat string, coverage bool, suiteTagFilter string) []PackageTestResult {
	results := []PackageTestResult{}

	for pkgPath, tests := range testsByPackage {
		result := runPackageTests(pkgPath, tests, multiWriter, testParallelism, testRunDir, reportFormat, coverage, suiteTagFilter)
		results = append(results, result)
	}

	return results
}

// runTestsParallel runs tests across packages in parallel using goroutines
func runTestsParallel(testsByPackage map[string][]testing.TestReference, multiWriter io.Writer, testParallelism int, testRunDir string, reportFormat string, coverage bool, suiteTagFilter string) []PackageTestResult {
	// Use a mutex to protect shared results
	var mu sync.Mutex
	results := []PackageTestResult{}

	// Create a wait group to track all goroutines
	var wg sync.WaitGroup

	// Create a channel to limit concurrent package tests (use number of CPU cores)
	// For now, use a fixed pool size of 4 to avoid overwhelming the system
	semaphore := make(chan struct{}, 4)

	for pkgPath, tests := range testsByPackage {
		wg.Add(1)

		go func(path string, testList []testing.TestReference) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Run tests for this package
			result := runPackageTests(path, testList, multiWriter, testParallelism, testRunDir, reportFormat, coverage, suiteTagFilter)

			// Append result (thread-safe)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(pkgPath, tests)
	}

	// Wait for all packages to complete
	wg.Wait()

	return results
}

// runPackageTests runs tests for a single package and returns detailed test results
func runPackageTests(pkgPath string, tests []testing.TestReference, multiWriter io.Writer, testParallelism int, testRunDir string, reportFormat string, coverage bool, suiteTagFilter string) PackageTestResult {
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

	// Check if this package contains only Godog features
	// BUT: if we have a relFeatureFile, we're explicitly running it through its test runner
	isGodogOnly := relFeatureFile == "" // If we have a specific feature, not Godog-only
	if isGodogOnly {
		for _, test := range tests {
			if test.Type != "godog" {
				isGodogOnly = false
				break
			}
		}
	}

	if isGodogOnly {
		// Skip running go test for spec directories (features only, no test code)
		// These specs are executed by their corresponding test packages
		// No console output needed - reduces noise in orchestrator output
		return PackageTestResult{
			TestsPassed:   len(tests),
			TestsFailed:   0,
			TestsSkipped:  0,
			TestsTotal:    len(tests),
			PackageFailed: false,
			ExpectedFiles: []string{}, // No files expected for godog-only packages
		}
	}

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
	// Go packages: organized by module like "src-cli/internal/conf.log"
	// Feature files: organized like "src-commands/templates.log"
	var logFilePath string

	if relFeatureFile != "" {
		// For feature files, strip "specs/" prefix and use parent dir for output
		// e.g., "specs/src-commands/templates/specification.feature"
		//    -> output: "src-commands/templates.log"
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
			return PackageTestResult{TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: []string{}}
		}

		logFilePath = filepath.Join(pkgOutputDir, dirName+".log")
	} else {
		// For Go packages, organize by module directories with flattened subpaths
		// e.g., "src/cli/internal/conf" -> "src-cli/internal-conf.log"
		// Extract first two path components as module name (e.g., "src/cli" -> "src-cli")
		pathParts := strings.Split(filepath.ToSlash(relPkgPath), "/")

		var moduleName string
		var fileName string

		if len(pathParts) >= 2 {
			// Module is first two components joined with hyphen: "src/cli" -> "src-cli"
			moduleName = pathParts[0] + "-" + pathParts[1]
			// Remaining path components become flattened filename
			if len(pathParts) > 2 {
				// Join remaining parts with hyphens: "internal/conf" -> "internal-conf"
				fileName = strings.Join(pathParts[2:], "-") + ".log"
			} else {
				// For packages at module root (e.g., "src/commands")
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
			return PackageTestResult{TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: []string{}}
		}

		// Create log file path directly under module directory
		logFilePath = filepath.Join(moduleDir, fileName)
	}
	logFile, err := os.Create(logFilePath)
	if err != nil {
		writeln(multiWriter, "❌ Failed to create log file: %v", err)
		writeln(multiWriter, "")
		return PackageTestResult{TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: []string{}}
	}
	defer logFile.Close()

	// Track expected output files for validation
	expectedFiles := []string{logFilePath}

	// JSON test results file (always created)
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
		expectedFiles = append(expectedFiles, coverageFile)
		// Also expect JSON coverage file
		coverageJSONFile := strings.TrimSuffix(coverageFile, ".out") + ".json"
		expectedFiles = append(expectedFiles, coverageJSONFile)
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
		// Test orchestrator packages (src/*/tests) should not generate reports
		// Reports are only generated for individual feature files in specs/
		if relFeatureFile != "" {
			// Set output directory for cucumber.json/junit.xml reports
			// For feature files, this is the parent directory (e.g., "out/test/commit/src-commands/")
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

			// Add report file to expected files
			reportFilePath := filepath.Join(reportOutputDir, reportName)
			expectedFiles = append(expectedFiles, reportFilePath)
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
			coverageJSONFile := strings.TrimSuffix(coverageFile, ".out") + ".json"
			if convertErr := convertCoverageToJSON(coverageFile, coverageJSONFile); convertErr != nil {
				fmt.Fprintf(logFile, "Warning: failed to convert coverage to JSON: %v\n", convertErr)
			}
		}
	}

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Analyze log to get failure reason
			failureReason := analyzeTestFailure(logFilePath, isGodogTestPackage)

			testType := "go"
			testCountInfo := fmt.Sprintf("(0/%d)", len(tests))
			if isGodogTestPackage {
				testType = "godog"
				// For Godog, len(tests) is test files not scenarios, so omit count
				testCountInfo = ""
			}

			// Create relative log path from test run directory
			relLogPath, _ := filepath.Rel(filepath.Dir(testRunDir), logFilePath)
			if testCountInfo != "" {
				writeln(multiWriter, "❌ Package %s [%s] %s failed due to %s (See %s for details)", pkgName, testType, testCountInfo, failureReason, relLogPath)
			} else {
				writeln(multiWriter, "❌ Package %s [%s] failed due to %s (See %s for details)", pkgName, testType, failureReason, relLogPath)
			}
			return PackageTestResult{TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: expectedFiles}
		} else {
			testType := "go"
			testCountInfo := fmt.Sprintf("(0/%d)", len(tests))
			if isGodogTestPackage {
				testType = "godog"
				testCountInfo = ""
			}

			if testCountInfo != "" {
				writeln(multiWriter, "❌ Package %s [%s] %s failed to run tests: %v", pkgName, testType, testCountInfo, err)
			} else {
				writeln(multiWriter, "❌ Package %s [%s] failed to run tests: %v", pkgName, testType, err)
			}
			return PackageTestResult{TestsPassed: 0, TestsFailed: len(tests), TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: true, ExpectedFiles: expectedFiles}
		}
	}

	testType := "go"
	testCountInfo := fmt.Sprintf("(%d/%d)", len(tests), len(tests))
	if isGodogTestPackage {
		testType = "godog"
		// Extract scenario counts from godog test output
		testCountInfo = extractGodogScenarioCounts(logFilePath)
	}

	// Create relative log path from test run directory
	relLogPath, _ := filepath.Rel(filepath.Dir(testRunDir), logFilePath)
	if testCountInfo != "" {
		writeln(multiWriter, "✅ Package %s [%s] %s passed (See %s for details)", pkgName, testType, testCountInfo, relLogPath)
	} else {
		writeln(multiWriter, "✅ Package %s [%s] passed (See %s for details)", pkgName, testType, relLogPath)
	}
	return PackageTestResult{TestsPassed: len(tests), TestsFailed: 0, TestsSkipped: 0, TestsTotal: len(tests), PackageFailed: false, ExpectedFiles: expectedFiles}
}

// extractGodogScenarioCounts parses godog JSON test results to extract scenario counts
// Returns: "(passed/total)" or empty string if counts not found
func extractGodogScenarioCounts(logPath string) string {
	// Replace .log extension with .json to get JSON results file
	jsonPath := strings.TrimSuffix(logPath, ".log") + ".json"

	events, err := testjson.ParseJSONFile(jsonPath)
	if err != nil {
		// Fall back to empty string if JSON file doesn't exist
		return ""
	}

	passedScenarios := 0
	failedScenarios := 0

	// For Godog tests, each test event represents a scenario
	// Filter out the TestFeatures wrapper function
	for _, event := range events {
		if event.Test != "" && event.Test != "TestFeatures" {
			if event.Action == "pass" {
				passedScenarios++
			} else if event.Action == "fail" {
				failedScenarios++
			}
		}
	}

	total := passedScenarios + failedScenarios
	if total == 0 {
		return ""
	}

	return fmt.Sprintf("(%d/%d)", passedScenarios, total)
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
// Handles both src/<module>/... and specs/<module>/... formats
// Supports both absolute and relative paths
func extractModuleFromPath(filePath string) string {
	// Normalize path separators to forward slashes
	normalizedPath := filepath.ToSlash(filePath)

	// Find "src/" in the path (handles both absolute and relative paths)
	srcIndex := strings.Index(normalizedPath, "/src/")
	if srcIndex >= 0 {
		// Extract from "/src/" onwards
		relativePath := normalizedPath[srcIndex+1:] // Skip the leading "/"
		// Format: src/<part1>/<part2>/...
		parts := strings.Split(relativePath, "/")
		if len(parts) >= 2 {
			// Return as src-<part1>
			return "src-" + parts[1]
		}
	}

	// Check if path contains "/specs/" (handles specs/src-*, specs/github, specs/repository, etc.)
	specsIndex := strings.Index(normalizedPath, "/specs/")
	if specsIndex >= 0 {
		// Extract from "/specs/" onwards
		relativePath := normalizedPath[specsIndex+1:]
		// Format: specs/<module>/...
		parts := strings.Split(strings.TrimPrefix(relativePath, "specs/"), "/")
		if len(parts) >= 1 && parts[0] != "" {
			// Return the first part (e.g., "src-commands", "github", "repository")
			return parts[0]
		}
	}

	// Also check for paths starting with "src/" or "specs/" (relative paths)
	if strings.HasPrefix(normalizedPath, "src/") {
		parts := strings.Split(normalizedPath, "/")
		if len(parts) >= 2 {
			return "src-" + parts[1]
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
// Tests without any OS-specific deps are considered OS-agnostic and run everywhere
// Tests with multiple OS deps (e.g., deps:linux AND deps:macos) run on any of those OSes
func filterByOSCompatibility(tests []testing.TestReference, w io.Writer) []testing.TestReference {
	currentOS := mapGOOSToDepTag(runtime.GOOS)
	compatible := []testing.TestReference{}

	for _, test := range tests {
		// Check if test has any OS-specific dependencies
		hasOSDep := false
		matchesCurrentOS := false

		for _, dep := range test.SystemDependencies {
			// Check if this is an OS dependency (uses exported list from testing package)
			for _, osDep := range testing.OsPlatformTags {
				if dep == osDep {
					hasOSDep = true
					if dep == currentOS {
						matchesCurrentOS = true
					}
					break
				}
			}
		}

		// Include test if:
		// 1. It has no OS-specific deps (OS-agnostic), OR
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

