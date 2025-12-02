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
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

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
	showTimings := false       // Show timing summary

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--suite" {
			if i+1 >= len(args) {
				log.Errorf("--suite requires a suite name")
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
		} else if arg == "--timings" {
			showTimings = true
		} else if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			log.Errorf("unknown flag: %s", arg)
			log.Errorf("Valid flags: --suite <name>, --as-junit, --as-cucumber, --coverage, --skip-deps, --list-only, --timings")
			return 1
		} else {
			monikers = append(monikers, arg)
		}
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("failed to load module contracts: %v", err)
		return 1
	}

	// If no monikers provided, default to all modules
	if len(monikers) == 0 {
		log.Info("No modules specified, testing all modules...")
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

	// Run test suite with module filter(s)
	// This provides consistent output format whether testing 1 or N modules
	return testModulesViaSuite(monikers, suiteName, reportFormat, coverage, skipDeps, listOnly, showTimings)
}

// testModulesViaSuite runs tests for one or more modules using the test suite infrastructure
// This provides consistent output format with summary, whether testing 1 or N modules
func testModulesViaSuite(monikers []string, suiteName string, reportFormat string, coverage bool, skipDeps bool, listOnly bool, showTimings bool) int {
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

	// Add timings flag
	if showTimings {
		args = append(args, "--timings")
	}

	os.Args = args

	return TestSuite()
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
	log.Info("  r2r eac test eac-commands --as-junit  # Generate JUnit XML")
	log.Info("  r2r eac test eac-commands --coverage  # Generate coverage reports")
	log.Info("")
	log.Info("Related commands:")
	log.Info("  r2r eac test suite <name>             # Run a specific test suite")
	log.Info("  r2r eac test list-suites              # List all available test suites")
}
