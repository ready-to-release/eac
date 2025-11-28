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
// Long:   test src-commands                    # Test single module
// Long:   test src-core src-cli                # Test multiple modules
// Long:   test                                 # Test all modules
// Long:   test src-commands --suite acceptance # Run acceptance tests only
// Flag.suite: type=string, usage=Filter tests by suite (default: "commit")
// HasSideEffects: false
package test

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(Test)
}

// Test is the unified entry point for testing modules
func Test() int {
	args := os.Args[2:] // Skip program name and "test"

	// Check for subcommands that should be handled separately
	if len(args) > 0 {
		switch args[0] {
		case "suite", "list-suites", "debug":
			// These are handled by their own registered commands
			// This case shouldn't normally be reached due to command registry lookup order
			return 0
		case "--help", "-h":
			printTestUsage()
			return 0
		}
	}

	// Parse arguments - separate monikers from flags
	var monikers []string
	suiteName := "commit"
	reportFormat := "cucumber" // Default report format
	coverage := false          // Generate coverage reports
	skipDeps := false          // Skip dependency verification
	listOnly := false          // List tests without running

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--suite" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --suite requires a suite name\n")
				return 1
			}
			i++
			suiteName = args[i]
		} else if arg == "--as-junit" {
			reportFormat = "junit"
		} else if arg == "--as-cucumber" {
			reportFormat = "cucumber"
		} else if arg == "--coverage" {
			coverage = true
		} else if arg == "--skip-deps" {
			skipDeps = true
		} else if arg == "--list-only" {
			listOnly = true
		} else if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid flags: --suite <name>, --as-junit, --as-cucumber, --coverage, --skip-deps, --list-only\n")
			return 1
		} else {
			monikers = append(monikers, arg)
		}
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// If no monikers provided, default to all modules
	if len(monikers) == 0 {
		fmt.Println("ℹ️  No modules specified, testing all modules...")
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

	// Run test suite with module filter(s)
	// This provides consistent output format whether testing 1 or N modules
	return testModulesViaSuite(monikers, suiteName, reportFormat, coverage, skipDeps, listOnly)
}

// testModulesViaSuite runs tests for one or more modules using the test suite infrastructure
// This provides consistent output format with summary, whether testing 1 or N modules
func testModulesViaSuite(monikers []string, suiteName string, reportFormat string, coverage bool, skipDeps bool, listOnly bool) int {
	// Build module filter argument (comma-separated)
	moduleFilter := strings.Join(monikers, ",")

	// Redirect to test suite with module filter
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Build args with format flag
	args := []string{
		oldArgs[0],
		"test",
		"suite",
		suiteName,
		"--module",
		moduleFilter,
	}

	// Add format flag
	if reportFormat == "junit" {
		args = append(args, "--as-junit")
	} else {
		args = append(args, "--as-cucumber")
	}

	// Add coverage flag
	if coverage {
		args = append(args, "--coverage")
	}

	// Add skip-deps flag
	if skipDeps {
		args = append(args, "--skip-deps")
	}

	// Add list-only flag
	if listOnly {
		args = append(args, "--list-only")
	}

	os.Args = args

	return TestSuite()
}

func printTestUsage() {
	fmt.Println("Test one or more modules by moniker")
	fmt.Println()
	fmt.Println("Usage: r2r eac test [module1] [module2] ... [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --suite <name>         Filter tests by suite (default: \"commit\")")
	fmt.Println("  --as-cucumber          Generate Cucumber JSON reports (default)")
	fmt.Println("  --as-junit             Generate JUnit XML reports")
	fmt.Println("  --coverage             Generate coverage reports (coverage.out, coverage.json)")
	fmt.Println("  --skip-deps            Skip dependency verification before running tests")
	fmt.Println("  --list-only            List tests that would run without executing them")
	fmt.Println()
	fmt.Println("Available suites:")
	fmt.Println("  commit                 L0-L2 tests (fast, pre-commit)")
	fmt.Println("  acceptance             IV/OV/PV tests (PLTE acceptance)")
	fmt.Println("  production-verification  L4+PIV tests (production smoke)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  r2r eac test                          # Test all modules")
	fmt.Println("  r2r eac test src-commands             # Test single module")
	fmt.Println("  r2r eac test src-cli src-core         # Test multiple modules")
	fmt.Println("  r2r eac test src-commands --suite acceptance")
	fmt.Println("  r2r eac test src-commands --as-junit  # Generate JUnit XML")
	fmt.Println("  r2r eac test src-commands --coverage  # Generate coverage reports")
	fmt.Println()
	fmt.Println("Related commands:")
	fmt.Println("  r2r eac test suite <name>             # Run a specific test suite")
	fmt.Println("  r2r eac test list-suites              # List all available test suites")
}
