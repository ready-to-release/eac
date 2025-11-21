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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	contractsreports "github.com/ready-to-release/eac/src/core/contracts/reports"
	moduledeps "github.com/ready-to-release/eac/src/core/module-deps"
	"github.com/ready-to-release/eac/src/core/repository"
	systemdeps "github.com/ready-to-release/eac/src/core/system-deps"
	"github.com/ready-to-release/eac/src/core/testing"
)

func init() {
	registry.Register(TestSuite)
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
		} else if strings.HasPrefix(arg, "--") {
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid flags: --skip-deps, --list-only, --sequential, --parallel\n")
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

	// Create markdown summary file
	mdPath := filepath.Join(testRunDir, "test-suite-summary.md")
	mdFile, err := os.Create(mdPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create markdown summary: %v\n", err)
		mdFile = nil // Continue without markdown
	} else {
		defer mdFile.Close()
	}

	// Track start time for duration calculation
	startTime := time.Now()

	// Write markdown header immediately
	if mdFile != nil {
		fmt.Fprintf(mdFile, "# Test Suite Report: %s\n\n", suite.Name)
		fmt.Fprintf(mdFile, "**Suite**: %s  \n", suiteName)
		fmt.Fprintf(mdFile, "**Started**: %s  \n", startTime.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(mdFile, "**Status**: 🔄 In Progress...\n\n")
		fmt.Fprintf(mdFile, "---\n\n")
		mdFile.Sync() // Flush to disk immediately
	}

	// Create multi-writer to log to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Phase 1: Discover all tests (Go + Godog)
	fmt.Fprintf(multiWriter, "=== Phase 1: Test Discovery ===\n")

	allTests, err := testing.DiscoverAllTests(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover tests: %v\n", err)
		return 1
	}

	fmt.Fprintf(multiWriter, "Discovered %d tests\n\n", len(allTests))

	// Update markdown: Phase 1 complete
	if mdFile != nil {
		fmt.Fprintf(mdFile, "## Progress\n\n")
		fmt.Fprintf(mdFile, "- ✅ **Phase 1**: Discovered %d tests\n", len(allTests))
		mdFile.Sync()
	}

	// Phase 2: Apply inference rules
	fmt.Fprintf(multiWriter, "=== Phase 2: Inference Engine ===\n")
	allTests = testing.ApplyInferences(allTests, suite.Inferences)
	fmt.Fprintf(multiWriter, "Applied %d inference rules\n", len(suite.Inferences))

	// Load module registry for module-based inference
	moduleReport, err := contractsreports.GetModuleContracts(workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load module contracts: %v\n", err)
	} else {
		// Infer system dependencies from module dependencies
		allTests = testing.InferSystemDepsFromModuleDeps(allTests, moduleReport.Registry)
		fmt.Fprintf(multiWriter, "Inferred system deps from module types\n")
	}
	fmt.Fprintf(multiWriter, "\n")

	// Update markdown: Phase 2 complete
	if mdFile != nil {
		fmt.Fprintf(mdFile, "- ✅ **Phase 2**: Applied %d inference rules\n", len(suite.Inferences))
		mdFile.Sync()
	}

	// Phase 3: Select tests for suite
	fmt.Fprintf(multiWriter, "=== Phase 3: Suite Selection ===\n")
	selectedTests := suite.SelectTests(allTests)
	fmt.Fprintf(multiWriter, "Selected %d tests for suite '%s'\n", len(selectedTests), suite.Moniker)

	// Phase 3.5: Filter out framework tests (tests about the testing framework itself)
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
		fmt.Fprintf(multiWriter, "INFO: %d framework tests excluded from execution\n", frameworkTestCount)
	}
	fmt.Fprintf(multiWriter, "Running %d production tests\n\n", len(productionTests))

	// Update markdown: Phase 3 complete
	if mdFile != nil {
		fmt.Fprintf(mdFile, "- ✅ **Phase 3**: Selected %d production tests for suite '%s'", len(productionTests), suite.Moniker)
		if frameworkTestCount > 0 {
			fmt.Fprintf(mdFile, " (%d framework tests excluded)", frameworkTestCount)
		}
		fmt.Fprintf(mdFile, "\n")
		mdFile.Sync()
	}

	// If list-only, just show tests and exit
	if listOnly {
		fmt.Fprintf(multiWriter, "=== Production Tests ===\n")
		for i, test := range productionTests {
			fmt.Fprintf(multiWriter, "%d. %s (%s)\n", i+1, test.TestName, test.Type)
			fmt.Fprintf(multiWriter, "   File: %s\n", test.FilePath)
			fmt.Fprintf(multiWriter, "   Tags: %s\n\n", strings.Join(test.Tags, ", "))
		}
		return 0
	}

	// Phase 4: Extract and verify dependencies (system + module)
	fmt.Fprintf(multiWriter, "=== Phase 4: Dependency Verification ===\n")
	systemDeps := testing.GetSystemDependencies(productionTests)
	moduleDeps := testing.GetModuleDependencies(productionTests)

	allDeps := append(append([]string{}, systemDeps...), moduleDeps...)

	if len(allDeps) == 0 {
		fmt.Fprintf(multiWriter, "No dependencies required\n\n")
		// Update markdown: Phase 4 complete (no deps)
		if mdFile != nil {
			fmt.Fprintf(mdFile, "- ✅ **Phase 4**: No system dependencies required\n")
			mdFile.Sync()
		}
	} else {
		fmt.Fprintf(multiWriter, "System dependencies: %s\n", strings.Join(systemDeps, ", "))
		fmt.Fprintf(multiWriter, "Module dependencies: %s\n", strings.Join(moduleDeps, ", "))

		// Update markdown: Start dependencies table
		if mdFile != nil {
			fmt.Fprintf(mdFile, "- 🔄 **Phase 4**: Verifying %d dependencies (%d system, %d module)...\n\n",
				len(allDeps), len(systemDeps), len(moduleDeps))
			fmt.Fprintf(mdFile, "## Dependencies\n\n")
			fmt.Fprintf(mdFile, "| Dependency | Status | Version |\n")
			fmt.Fprintf(mdFile, "|------------|--------|----------|\n")
			mdFile.Sync()
		}

		if !skipDeps {
			hasFailures := false

			// Verify system dependencies
			sysResults := systemdeps.VerifyAll(systemDeps)
			for _, result := range sysResults {
				if result.Available {
					fmt.Fprintf(multiWriter, "✅ %s - %s\n", result.Dependency, result.Version)
					// Update markdown: Add dependency row
					if mdFile != nil {
						fmt.Fprintf(mdFile, "| %s | ✅ Available | %s |\n", result.Dependency, result.Version)
						mdFile.Sync()
					}
				} else {
					fmt.Fprintf(multiWriter, "❌ %s - not available\n", result.Dependency)
					hasFailures = true
					// Update markdown: Add failed dependency row
					if mdFile != nil {
						fmt.Fprintf(mdFile, "| %s | ❌ Not Available | - |\n", result.Dependency)
						mdFile.Sync()
					}
				}
			}

			// Verify module dependencies
			modResults := moduledeps.VerifyAll(moduleDeps)
			for _, result := range modResults {
				if result.Available {
					fmt.Fprintf(multiWriter, "✅ %s - %s\n", result.Dependency, result.Version)
					// Update markdown: Add dependency row
					if mdFile != nil {
						fmt.Fprintf(mdFile, "| %s | ✅ Available | %s |\n", result.Dependency, result.Version)
						mdFile.Sync()
					}
				} else {
					fmt.Fprintf(multiWriter, "❌ %s - not available\n", result.Dependency)
					hasFailures = true
					// Update markdown: Add failed dependency row
					if mdFile != nil {
						fmt.Fprintf(mdFile, "| %s | ❌ Not Available | - |\n", result.Dependency)
						mdFile.Sync()
					}
				}
			}

			fmt.Fprintln(multiWriter)

			// Update markdown: Phase 4 status
			if mdFile != nil {
				fmt.Fprintf(mdFile, "\n")
				if hasFailures {
					fmt.Fprintf(mdFile, "⚠️ **Phase 4 Failed**: Some dependencies are missing\n\n")
				}
				mdFile.Sync()
			}

			if hasFailures {
				fmt.Fprintf(multiWriter, "❌ Error: Required dependencies are missing\n")
				fmt.Fprintf(multiWriter, "Use --skip-deps to run tests anyway\n")
				return 1
			}
		} else {
			fmt.Fprintf(multiWriter, "Dependency check skipped (--skip-deps)\n\n")
			// Update markdown: Phase 4 skipped
			if mdFile != nil {
				fmt.Fprintf(mdFile, "\n⏭️ Dependency check skipped (--skip-deps)\n\n")
				mdFile.Sync()
			}
		}
	}

	// findGodogTestRunner finds the test runner package for a feature file
	// Feature files are in specs/src-<module>/..., test runners are in src/<module>/tests/
	findGodogTestRunner := func(featurePath string) string {
		// Extract module from specs path
		// Example: specs/src-commands/commit-ai/... -> src/commands/tests
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
	fmt.Fprintf(multiWriter, "=== Phase 5: Test Execution ===\n")

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
				fmt.Fprintf(multiWriter, "⚠️  Skipping test %s: unable to compute relative path from %s to %s\n",
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
				fmt.Fprintf(multiWriter, "⚠️  Skipping test %s: unable to compute relative path from %s to %s\n",
					test.TestName, workspaceRoot, absDir)
				continue
			}
			pkgPath = relDir
		}
		testsByPackage[pkgPath] = append(testsByPackage[pkgPath], test)
	}

	// Start global progress tracker with total package count
	StartGlobalTracker(multiWriter, len(testsByPackage))
	defer StopGlobalTracker()

	// Package count removed from console output for cleaner orchestration

	// Update markdown: Phase 5 start
	if mdFile != nil {
		if parallel {
			fmt.Fprintf(mdFile, "- 🔄 **Phase 5**: Running tests from %d packages in parallel...\n\n", len(testsByPackage))
		} else {
			fmt.Fprintf(mdFile, "- 🔄 **Phase 5**: Running tests from %d packages...\n\n", len(testsByPackage))
		}
		fmt.Fprintf(mdFile, "## Test Results\n\n")
		mdFile.Sync()
	}

	totalPassed := 0
	totalFailed := 0

	// Calculate optimal test-level parallelism
	numCPU := runtime.NumCPU()
	var testParallelism int

	if parallel {
		// Package-level parallel: distribute CPU across packages
		// Each package gets a smaller share of CPU cores
		testParallelism = max(2, numCPU/4)
		totalPassed, totalFailed = runTestsParallel(testsByPackage, multiWriter, mdFile, testParallelism, testRunDir)
	} else {
		// Sequential packages: each package gets full CPU power
		testParallelism = numCPU
		totalPassed, totalFailed = runTestsSequential(testsByPackage, multiWriter, mdFile, testParallelism, testRunDir)
	}

	// Parallelism info removed from console for cleaner output

	// Phase 6: Generate summary
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Fprintf(multiWriter, "=== Test Run Summary ===\n")
	fmt.Fprintf(multiWriter, "Suite: %s\n", suite.Name)
	fmt.Fprintf(multiWriter, "Total packages: %d\n", len(testsByPackage))
	fmt.Fprintf(multiWriter, "Packages passed: %d\n", totalPassed)
	fmt.Fprintf(multiWriter, "Packages failed: %d\n", totalFailed)
	fmt.Fprintf(multiWriter, "Individual tests discovered: %d\n", len(allTests))
	fmt.Fprintf(multiWriter, "Production tests: %d\n", len(productionTests))
	if frameworkTestCount > 0 {
		fmt.Fprintf(multiWriter, "Framework tests excluded: %d\n", frameworkTestCount)
	}
	fmt.Fprintf(multiWriter, "Results directory: %s\n", testRunDir)

	// Check for undefined steps in test logs
	undefinedSteps := extractUndefinedSteps(testRunDir)
	if len(undefinedSteps) > 0 {
		fmt.Fprintf(multiWriter, "\n⚠️  WARNING: %d undefined steps found\n", len(undefinedSteps))
		fmt.Fprintf(multiWriter, "Scenarios with undefined steps need step implementations.\n")
		fmt.Fprintf(multiWriter, "Run with verbose logging to see full details.\n")
	}

	// Update markdown: Final summary
	if mdFile != nil {
		fmt.Fprintf(mdFile, "\n---\n\n")
		fmt.Fprintf(mdFile, "## Summary\n\n")

		// Calculate pass rate
		passRate := 0.0
		if len(productionTests) > 0 {
			passRate = float64(totalPassed) / float64(len(productionTests)) * 100
		}

		// Determine final status
		finalStatus := "✅ PASSED"
		if totalFailed > 0 {
			finalStatus = "❌ FAILED"
		}

		// Write summary table
		fmt.Fprintf(mdFile, "| Metric | Value |\n")
		fmt.Fprintf(mdFile, "|--------|-------|\n")
		fmt.Fprintf(mdFile, "| **Status** | **%s** |\n", finalStatus)
		fmt.Fprintf(mdFile, "| Duration | %.1fs |\n", duration.Seconds())
		fmt.Fprintf(mdFile, "| Tests Discovered | %d |\n", len(allTests))
		fmt.Fprintf(mdFile, "| Production Tests | %d |\n", len(productionTests))
		if frameworkTestCount > 0 {
			fmt.Fprintf(mdFile, "| Framework Tests Excluded | %d |\n", frameworkTestCount)
		}
		fmt.Fprintf(mdFile, "| Tests Passed | %d ✅ |\n", totalPassed)
		fmt.Fprintf(mdFile, "| Tests Failed | %d |\n", totalFailed)
		fmt.Fprintf(mdFile, "| Pass Rate | %.1f%% |\n", passRate)
		fmt.Fprintf(mdFile, "\n")

		// Add links
		fmt.Fprintf(mdFile, "## Files\n\n")
		fmt.Fprintf(mdFile, "- **Full Log**: [`test-suite.log`](./test-suite.log)\n")
		fmt.Fprintf(mdFile, "- **Results Directory**: `%s`\n", testRunDir)
		fmt.Fprintf(mdFile, "\n---\n\n")
		fmt.Fprintf(mdFile, "*Generated by `test suite %s` on %s*\n", suite.Moniker, endTime.Format("2006-01-02 15:04:05"))

		// Update the status line at the top (re-write the file from beginning for final status)
		mdFile.Seek(0, 0)
		mdFile.Truncate(0)

		// Write final markdown with complete status
		fmt.Fprintf(mdFile, "# Test Suite Report: %s\n\n", suite.Name)
		fmt.Fprintf(mdFile, "**Suite**: %s  \n", suiteName)
		fmt.Fprintf(mdFile, "**Started**: %s  \n", startTime.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(mdFile, "**Completed**: %s  \n", endTime.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(mdFile, "**Duration**: %.1fs  \n", duration.Seconds())
		fmt.Fprintf(mdFile, "**Status**: %s\n\n", finalStatus)
		fmt.Fprintf(mdFile, "---\n\n")

		// Re-write all the phase information (this could be optimized by buffering, but simpler for now)
		// For now, just write the final summary table

		fmt.Fprintf(mdFile, "## Summary\n\n")
		fmt.Fprintf(mdFile, "| Metric | Value |\n")
		fmt.Fprintf(mdFile, "|--------|-------|\n")
		fmt.Fprintf(mdFile, "| **Status** | **%s** |\n", finalStatus)
		fmt.Fprintf(mdFile, "| Duration | %.1fs |\n", duration.Seconds())
		fmt.Fprintf(mdFile, "| Tests Discovered | %d |\n", len(allTests))
		fmt.Fprintf(mdFile, "| Production Tests | %d |\n", len(productionTests))
		if frameworkTestCount > 0 {
			fmt.Fprintf(mdFile, "| Framework Tests Excluded | %d |\n", frameworkTestCount)
		}
		fmt.Fprintf(mdFile, "| Tests Passed | %d ✅ |\n", totalPassed)
		fmt.Fprintf(mdFile, "| Tests Failed | %d |\n", totalFailed)
		fmt.Fprintf(mdFile, "| Pass Rate | %.1f%% |\n\n", passRate)

		fmt.Fprintf(mdFile, "## Files\n\n")
		fmt.Fprintf(mdFile, "- **Full Log**: [`test-suite.log`](./test-suite.log)\n")
		fmt.Fprintf(mdFile, "- **Summary**: `test-suite-summary.md` (this file)\n")
		fmt.Fprintf(mdFile, "- **Results Directory**: `%s`\n", testRunDir)
		fmt.Fprintf(mdFile, "\n---\n\n")
		fmt.Fprintf(mdFile, "*Generated by `test suite %s` on %s*\n", suite.Moniker, endTime.Format("2006-01-02 15:04:05"))

		mdFile.Sync()
	}

	if totalFailed > 0 {
		return 1
	}

	return 0
}

// fileExists checks if a file exists at the given path
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
func runTestsSequential(testsByPackage map[string][]testing.TestReference, multiWriter io.Writer, mdFile *os.File, testParallelism int, testRunDir string) (int, int) {
	packagesPassed := 0
	packagesFailed := 0
	packageNum := 0

	for pkgPath, tests := range testsByPackage {
		packageNum++
		// Package info removed from console - now only shows results

		// Update markdown: Package starting
		if mdFile != nil {
			pkgName := filepath.Base(pkgPath)
			if pkgName == "" {
				pkgName = pkgPath
			}
			fmt.Fprintf(mdFile, "- 🔄 **[%d/%d]** %s (%d tests)...\n", packageNum, len(testsByPackage), pkgName, len(tests))
			mdFile.Sync()
		}

		passed, failed := runPackageTests(pkgPath, tests, multiWriter, mdFile, testParallelism, testRunDir)
		// Count packages, not individual tests
		if failed > 0 {
			packagesFailed++
		} else if passed > 0 {
			packagesPassed++
		}
	}

	return packagesPassed, packagesFailed
}

// runTestsParallel runs tests across packages in parallel using goroutines
func runTestsParallel(testsByPackage map[string][]testing.TestReference, multiWriter io.Writer, mdFile *os.File, testParallelism int, testRunDir string) (int, int) {
	// Use a mutex to protect shared counters and output
	var mu sync.Mutex
	packagesPassed := 0
	packagesFailed := 0

	// Create a wait group to track all goroutines
	var wg sync.WaitGroup

	// Create a channel to limit concurrent package tests (use number of CPU cores)
	// For now, use a fixed pool size of 4 to avoid overwhelming the system
	semaphore := make(chan struct{}, 4)

	packageNum := 0
	numPackages := len(testsByPackage)

	for pkgPath, tests := range testsByPackage {
		wg.Add(1)
		packageNum++
		currentPkgNum := packageNum

		go func(path string, testList []testing.TestReference, pkgNum int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Package info removed from console - now only shows results
			mu.Lock()

			// Update markdown: Package starting
			if mdFile != nil {
				pkgName := filepath.Base(path)
				if pkgName == "" {
					pkgName = path
				}
				fmt.Fprintf(mdFile, "- 🔄 **[%d/%d]** %s (%d tests)...\n", pkgNum, numPackages, pkgName, len(testList))
				mdFile.Sync()
			}
			mu.Unlock()

			// Run tests for this package
			passed, failed := runPackageTests(path, testList, multiWriter, mdFile, testParallelism, testRunDir)

			// Update totals (thread-safe) - count packages, not individual tests
			mu.Lock()
			if failed > 0 {
				packagesFailed++
			} else if passed > 0 {
				packagesPassed++
			}
			mu.Unlock()
		}(pkgPath, tests, currentPkgNum)
	}

	// Wait for all packages to complete
	wg.Wait()

	return packagesPassed, packagesFailed
}

// runPackageTests runs tests for a single package and returns (passed, failed) counts
func runPackageTests(pkgPath string, tests []testing.TestReference, multiWriter io.Writer, mdFile *os.File, testParallelism int, testRunDir string) (int, int) {
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

	// Track test start/completion with global progress tracker (before early returns)
	TrackTestStart(pkgName)
	defer TrackTestComplete(pkgName)

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
		return len(tests), 0
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

	// Create modular directory structure for test outputs
	// Convert package path to directory structure (e.g., "src/commands/tests" -> "src/commands/tests/")
	// Feature files like "specs/foo/bar/spec.feature" -> "foo/" with output "bar.log"
	var pkgOutputDir string
	var logFileName string

	if relFeatureFile != "" {
		// For feature files, strip "specs/" prefix and use parent dir for output
		// e.g., "specs/src-commands/templates/specification.feature"
		//    -> output dir: "src-commands/", filename: "templates.log"
		featureDir := filepath.Dir(relFeatureFile)
		featureDir = strings.TrimPrefix(featureDir, "specs/")
		featureDir = strings.TrimPrefix(featureDir, "specs\\") // Windows path separator

		// Use the directory name (last component) as the base filename
		dirName := filepath.Base(featureDir)
		// Use the parent directory as the output directory
		parentDir := filepath.Dir(featureDir)

		pkgOutputDir = filepath.Join(testRunDir, parentDir)
		logFileName = dirName + ".log"
	} else {
		// For Go packages, use the package path directly
		pkgOutputDir = filepath.Join(testRunDir, relPkgPath)
		logFileName = "test.log"
	}

	// Create the output directory
	if err := os.MkdirAll(pkgOutputDir, 0755); err != nil {
		fmt.Fprintf(multiWriter, "❌ Failed to create output directory: %v\n\n", err)
		return 0, len(tests)
	}

	logFilePath := filepath.Join(pkgOutputDir, logFileName)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Fprintf(multiWriter, "❌ Failed to create log file: %v\n\n", err)
		return 0, len(tests)
	}
	defer logFile.Close()

	// For Godog packages, write feature summaries to log file
	if isGodogTestPackage {
		displayGodogFeatureSummaries(actualPkgDir, logFile)
	}

	// Run go test for this package with test-level parallelism
	// If testing a specific feature file, pass it via environment variable (GODOG_PATHS)
	var cmd *exec.Cmd

	cmd = exec.Command("go", "test", "-v", "-parallel", fmt.Sprintf("%d", testParallelism))
	cmd.Dir = actualPkgDir

	// Verbose test output goes to log file only, not console
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// For Godog test packages, set GODOG environment variables
	if isGodogTestPackage {
		// Start with base environment
		cmd.Env = os.Environ()

		// Set format for console output
		cmd.Env = append(cmd.Env, "GODOG_FORMAT=progress")

		// Only generate cucumber.json reports when testing specific feature files
		// Test orchestrator packages (src/*/tests) should not generate reports
		// Reports are only generated for individual feature files in specs/
		if relFeatureFile != "" {
			// Set output directory for cucumber.json/junit.xml reports
			// Use the package-specific output directory for modular structure
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_OUTPUT_DIR=%s", pkgOutputDir))

			// Set report format (default to cucumber for BDD tests)
			cmd.Env = append(cmd.Env, "GODOG_REPORT_FORMAT=cucumber")

			// Use the directory name for the report (e.g., "templates.cucumber.json")
			reportName := strings.TrimSuffix(logFileName, ".log") + ".cucumber.json"
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_REPORT_NAME=%s", reportName))
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
			reportName := strings.TrimSuffix(logFileName, ".log") + ".cucumber.json"
			fmt.Fprintf(logFile, "📊 Test report will be saved as: %s\n\n", reportName)
		}
	}

	err = cmd.Run()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
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
				fmt.Fprintf(multiWriter, "❌ Package %s [%s] %s failed due to %s (See %s for details)\n", pkgName, testType, testCountInfo, failureReason, relLogPath)
			} else {
				fmt.Fprintf(multiWriter, "❌ Package %s [%s] failed due to %s (See %s for details)\n", pkgName, testType, failureReason, relLogPath)
			}
			// Update markdown: Package failed
			if mdFile != nil {
				fmt.Fprintf(mdFile, "  - ❌ Failed (exit code: %d) - %s\n", exitErr.ExitCode(), failureReason)
				mdFile.Sync()
			}
			return 0, len(tests)
		} else {
			testType := "go"
			testCountInfo := fmt.Sprintf("(0/%d)", len(tests))
			if isGodogTestPackage {
				testType = "godog"
				testCountInfo = ""
			}

			if testCountInfo != "" {
				fmt.Fprintf(multiWriter, "❌ Package %s [%s] %s failed to run tests: %v\n", pkgName, testType, testCountInfo, err)
			} else {
				fmt.Fprintf(multiWriter, "❌ Package %s [%s] failed to run tests: %v\n", pkgName, testType, err)
			}
			// Update markdown: Package error
			if mdFile != nil {
				fmt.Fprintf(mdFile, "  - ❌ Error: %v\n", err)
				mdFile.Sync()
			}
			return 0, len(tests)
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
		fmt.Fprintf(multiWriter, "✅ Package %s [%s] %s passed (See %s for details)\n", pkgName, testType, testCountInfo, relLogPath)
	} else {
		fmt.Fprintf(multiWriter, "✅ Package %s [%s] passed (See %s for details)\n", pkgName, testType, relLogPath)
	}
	// Update markdown: Package passed
	if mdFile != nil {
		fmt.Fprintf(mdFile, "  - ✅ Passed\n")
		mdFile.Sync()
	}
	return len(tests), 0
}

// extractGodogScenarioCounts parses godog test output to extract scenario counts
// Returns: "(passed/total)" or empty string if counts not found
func extractGodogScenarioCounts(logPath string) string {
	content, err := os.ReadFile(logPath)
	if err != nil {
		return ""
	}

	logStr := string(content)
	lines := strings.Split(logStr, "\n")

	// Look for line like: "60 scenarios (59 passed, 1 failed)"
	// or "60 scenarios (60 passed)"
	// or with ANSI codes: "17 scenarios ([32m17 passed[0m)"
	for _, line := range lines {
		if strings.Contains(line, " scenarios (") {
			// Parse the line to extract counts
			// Format: "N scenarios (X passed)" or "N scenarios (X passed, Y failed)"
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				totalStr := fields[0]
				// Extract passed count from parentheses
				for i, field := range fields {
					// Check for "passed)" with or without ANSI codes
					// Field can be: "passed)", "passed,", "passed[0m)", etc.
					if strings.Contains(field, "passed") {
						if i > 0 {
							passedStr := fields[i-1]
							// Remove ANSI color codes and parentheses
							// Can be: "([32m17" or "[32m17" or "(17" or "17"
							passedStr = strings.TrimPrefix(passedStr, "(")
							// Remove all common ANSI codes
							passedStr = strings.ReplaceAll(passedStr, "\x1b[32m", "")
							passedStr = strings.ReplaceAll(passedStr, "\x1b[0m", "")
							passedStr = strings.ReplaceAll(passedStr, "[32m", "")
							passedStr = strings.ReplaceAll(passedStr, "[0m", "")
							passedStr = strings.ReplaceAll(passedStr, "\x1b", "")
							return fmt.Sprintf("(%s/%s)", passedStr, totalStr)
						}
					}
				}
			}
		}
	}

	return ""
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

