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

	// Phase 2: Apply inference rules
	fmt.Fprintf(multiWriter, "=== Phase 2: Inference Engine ===\n")
	allTests = testing.ApplyInferences(allTests, suite.Inferences)
	fmt.Fprintf(multiWriter, "Applied %d inference rules\n\n", len(suite.Inferences))

	// Phase 3: Select tests for suite
	fmt.Fprintf(multiWriter, "=== Phase 3: Suite Selection ===\n")
	selectedTests := suite.SelectTests(allTests)
	fmt.Fprintf(multiWriter, "Selected %d tests for suite '%s'\n\n", len(selectedTests), suite.Moniker)

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
	} else {
		fmt.Fprintf(multiWriter, "Required dependencies: %s\n", strings.Join(dependencies, ", "))

		if !skipDeps {
			results := systemdeps.VerifyAll(dependencies)

			hasFailures := false
			for _, result := range results {
				if result.Available {
					fmt.Fprintf(multiWriter, "✅ %s - %s\n", result.Dependency, result.Version)
				} else {
					fmt.Fprintf(multiWriter, "❌ %s - not available\n", result.Dependency)
					hasFailures = true
				}
			}
			fmt.Fprintln(multiWriter)

			if hasFailures {
				fmt.Fprintf(multiWriter, "❌ Error: Required dependencies are missing\n")
				fmt.Fprintf(multiWriter, "Use --skip-deps to run tests anyway\n")
				return 1
			}
		} else {
			fmt.Fprintf(multiWriter, "Dependency check skipped (--skip-deps)\n\n")
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

	totalPassed := 0
	totalFailed := 0

	for pkgPath, tests := range testsByPackage {
		fmt.Fprintf(multiWriter, "📦 Package: %s\n", pkgPath)
		fmt.Fprintf(multiWriter, "   Tests: %d\n", len(tests))

		// Run go test for this package
		cmd := exec.Command("go", "test", "-v")
		cmd.Dir = pkgPath
		cmd.Stdout = multiWriter
		cmd.Stderr = multiWriter

		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(multiWriter, "❌ Package tests failed (exit code: %d)\n\n", exitErr.ExitCode())
				totalFailed += len(tests)
			} else {
				fmt.Fprintf(multiWriter, "❌ Failed to run tests: %v\n\n", err)
				totalFailed += len(tests)
			}
		} else {
			fmt.Fprintf(multiWriter, "✅ Package tests passed\n\n")
			totalPassed += len(tests)
		}
	}

	// Phase 6: Generate summary
	fmt.Fprintf(multiWriter, "=== Test Run Summary ===\n")
	fmt.Fprintf(multiWriter, "Suite: %s\n", suite.Name)
	fmt.Fprintf(multiWriter, "Run ID: %s\n", testRunID)
	fmt.Fprintf(multiWriter, "Total discovered: %d\n", len(allTests))
	fmt.Fprintf(multiWriter, "Total selected: %d\n", len(selectedTests))
	fmt.Fprintf(multiWriter, "Total passed: %d\n", totalPassed)
	fmt.Fprintf(multiWriter, "Total failed: %d\n", totalFailed)
	fmt.Fprintf(multiWriter, "Results directory: %s\n", testRunDir)

	if totalFailed > 0 {
		return 1
	}

	return 0
}
