// Command: test suite
// Description: Run tests for a specific test suite
// Usage: test suite <suite-name> [--skip-deps] [--list-only]
// HasSideEffects: false
package test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/core/system-deps"
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
		fmt.Fprintf(os.Stderr, "Usage: test suite <suite-name> [--skip-deps] [--list-only]\n")
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

	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--skip-deps" {
			skipDeps = true
		} else if arg == "--list-only" {
			listOnly = true
		} else if strings.HasPrefix(arg, "--") {
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid flags: --skip-deps, --list-only\n")
			return 1
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

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

	// Create test-run-id directory (timestamp-based)
	testRunID := time.Now().Format("2006-01-02-150405")
	testRunDir := filepath.Join(workspaceRoot, "out", "test-results", testRunID)
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
		fmt.Fprintf(mdFile, "**Run ID**: %s  \n", testRunID)
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
	fmt.Fprintf(multiWriter, "Applied %d inference rules\n\n", len(suite.Inferences))

	// Update markdown: Phase 2 complete
	if mdFile != nil {
		fmt.Fprintf(mdFile, "- ✅ **Phase 2**: Applied %d inference rules\n", len(suite.Inferences))
		mdFile.Sync()
	}

	// Phase 3: Select tests for suite
	fmt.Fprintf(multiWriter, "=== Phase 3: Suite Selection ===\n")
	selectedTests := suite.SelectTests(allTests)
	fmt.Fprintf(multiWriter, "Selected %d tests for suite '%s'\n\n", len(selectedTests), suite.Moniker)

	// Update markdown: Phase 3 complete
	if mdFile != nil {
		fmt.Fprintf(mdFile, "- ✅ **Phase 3**: Selected %d tests for suite '%s'\n", len(selectedTests), suite.Moniker)
		mdFile.Sync()
	}

	// If list-only, just show tests and exit
	if listOnly {
		fmt.Fprintf(multiWriter, "=== Selected Tests ===\n")
		for i, test := range selectedTests {
			fmt.Fprintf(multiWriter, "%d. %s (%s)\n", i+1, test.TestName, test.Type)
			fmt.Fprintf(multiWriter, "   File: %s\n", test.FilePath)
			fmt.Fprintf(multiWriter, "   Tags: %s\n\n", strings.Join(test.Tags, ", "))
		}
		return 0
	}

	// Phase 4: Extract and verify system dependencies
	fmt.Fprintf(multiWriter, "=== Phase 4: Dependency Verification ===\n")
	dependencies := testing.GetSystemDependencies(selectedTests)

	if len(dependencies) == 0 {
		fmt.Fprintf(multiWriter, "No system dependencies required\n\n")
		// Update markdown: Phase 4 complete (no deps)
		if mdFile != nil {
			fmt.Fprintf(mdFile, "- ✅ **Phase 4**: No system dependencies required\n")
			mdFile.Sync()
		}
	} else {
		fmt.Fprintf(multiWriter, "Required dependencies: %s\n", strings.Join(dependencies, ", "))

		// Update markdown: Start dependencies table
		if mdFile != nil {
			fmt.Fprintf(mdFile, "- 🔄 **Phase 4**: Verifying %d dependencies...\n\n", len(dependencies))
			fmt.Fprintf(mdFile, "## Dependencies\n\n")
			fmt.Fprintf(mdFile, "| Dependency | Status | Version |\n")
			fmt.Fprintf(mdFile, "|------------|--------|----------|\n")
			mdFile.Sync()
		}

		if !skipDeps {
			results := systemdeps.VerifyAll(dependencies)

			hasFailures := false
			for _, result := range results {
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

	// Phase 5: Run tests
	fmt.Fprintf(multiWriter, "=== Phase 5: Test Execution ===\n")

	// Group tests by package
	testsByPackage := make(map[string][]testing.TestReference)
	for _, test := range selectedTests {
		pkgPath := filepath.Dir(test.FilePath)
		testsByPackage[pkgPath] = append(testsByPackage[pkgPath], test)
	}

	fmt.Fprintf(multiWriter, "Running tests from %d packages\n\n", len(testsByPackage))

	// Update markdown: Phase 5 start
	if mdFile != nil {
		fmt.Fprintf(mdFile, "- 🔄 **Phase 5**: Running tests from %d packages...\n\n", len(testsByPackage))
		fmt.Fprintf(mdFile, "## Test Results\n\n")
		mdFile.Sync()
	}

	totalPassed := 0
	totalFailed := 0

	packageNum := 0
	for pkgPath, tests := range testsByPackage {
		packageNum++
		fmt.Fprintf(multiWriter, "📦 Package: %s\n", pkgPath)
		fmt.Fprintf(multiWriter, "   Tests: %d\n", len(tests))

		// Update markdown: Package starting
		if mdFile != nil {
			// Extract package name from path
			pkgName := filepath.Base(pkgPath)
			if pkgName == "" {
				pkgName = pkgPath
			}
			fmt.Fprintf(mdFile, "- 🔄 **[%d/%d]** %s (%d tests)...\n", packageNum, len(testsByPackage), pkgName, len(tests))
			mdFile.Sync()
		}

		// Check if this package contains only Godog features (no Go test files)
		// Godog features are tested by their test packages (src/cli/tests, src/commands/tests)
		// and should not be run directly with `go test`
		isGodogOnly := true
		for _, test := range tests {
			if test.Type != "godog" {
				isGodogOnly = false
				break
			}
		}

		if isGodogOnly {
			// Skip running go test for spec directories - they're tested elsewhere
			fmt.Fprintf(multiWriter, "⏭️  Godog features (tested by test packages)\n\n")
			totalPassed += len(tests)
			continue
		}

		// Check if this is a Godog test package (contains godog_test.go)
		isGodogTestPackage := fileExists(filepath.Join(pkgPath, "godog_test.go"))

		if isGodogTestPackage {
			// Display feature file summaries before running Godog tests
			displayGodogFeatureSummaries(pkgPath, multiWriter)
		}

		// Run go test for this package
		cmd := exec.Command("go", "test", "-v")
		cmd.Dir = pkgPath
		cmd.Stdout = multiWriter
		cmd.Stderr = multiWriter

		// For Godog test packages, set GODOG_FORMAT=progress for concise output
		if isGodogTestPackage {
			cmd.Env = append(os.Environ(), "GODOG_FORMAT=progress")
		}

		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(multiWriter, "❌ Package tests failed (exit code: %d)\n\n", exitErr.ExitCode())
				totalFailed += len(tests)
				// Update markdown: Package failed
				if mdFile != nil {
					fmt.Fprintf(mdFile, "  - ❌ Failed (exit code: %d)\n", exitErr.ExitCode())
					mdFile.Sync()
				}
			} else {
				fmt.Fprintf(multiWriter, "❌ Failed to run tests: %v\n\n", err)
				totalFailed += len(tests)
				// Update markdown: Package error
				if mdFile != nil {
					fmt.Fprintf(mdFile, "  - ❌ Error: %v\n", err)
					mdFile.Sync()
				}
			}
		} else {
			fmt.Fprintf(multiWriter, "✅ Package tests passed\n\n")
			totalPassed += len(tests)
			// Update markdown: Package passed
			if mdFile != nil {
				fmt.Fprintf(mdFile, "  - ✅ Passed\n")
				mdFile.Sync()
			}
		}
	}

	// Phase 6: Generate summary
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Fprintf(multiWriter, "=== Test Run Summary ===\n")
	fmt.Fprintf(multiWriter, "Suite: %s\n", suite.Name)
	fmt.Fprintf(multiWriter, "Run ID: %s\n", testRunID)
	fmt.Fprintf(multiWriter, "Total discovered: %d\n", len(allTests))
	fmt.Fprintf(multiWriter, "Total selected: %d\n", len(selectedTests))
	fmt.Fprintf(multiWriter, "Total passed: %d\n", totalPassed)
	fmt.Fprintf(multiWriter, "Total failed: %d\n", totalFailed)
	fmt.Fprintf(multiWriter, "Results directory: %s\n", testRunDir)

	// Update markdown: Final summary
	if mdFile != nil {
		fmt.Fprintf(mdFile, "\n---\n\n")
		fmt.Fprintf(mdFile, "## Summary\n\n")

		// Calculate pass rate
		passRate := 0.0
		if len(selectedTests) > 0 {
			passRate = float64(totalPassed) / float64(len(selectedTests)) * 100
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
		fmt.Fprintf(mdFile, "| Tests Selected | %d |\n", len(selectedTests))
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
		fmt.Fprintf(mdFile, "**Run ID**: %s  \n", testRunID)
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
		fmt.Fprintf(mdFile, "| Tests Selected | %d |\n", len(selectedTests))
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
