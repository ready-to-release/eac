// Command: test
// Short: Test one or more modules by moniker
// Args: modules
// Long: Test one or more modules by moniker using type-based dispatch.
// Long:
// Long: This command identifies the module type from the module contract and dispatches
// Long: to the appropriate test handler. Supported module types include Go modules with
// Long: unit tests, specifications, and other testable components.
// Long:
// Long: When a single module is specified, output is verbose. When multiple modules
// Long: are specified, tests run in parallel with orchestrator output.
// Long:
// Long: Test output formats can be controlled with flags. The default output shows test
// Long: results in real-time. Use --as-cucumber for cucumber-style JSON output or --as-junit
// Long: for JUnit XML format suitable for CI/CD systems.
// Long:
// Long: Use --suite to filter tests by suite. The default suite is "commit".
// Long:
// Long: Example:
// Long:   test src-commands                    # Test single module (verbose)
// Long:   test src-core src-cli                # Test multiple modules (parallel)
// Long:   test                                 # Test all modules
// Long:   test src-commands --as-junit         # Test with JUnit output
// Long:   test src-commands --suite integration
// Flag.as-cucumber: type=bool, usage=Output test results in Cucumber JSON format
// Flag.as-junit: type=bool, usage=Output test results in JUnit XML format
// Flag.suite: type=string, usage=Filter tests by suite (default: "commit")
// HasSideEffects: false
package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
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
	reportFormat := "cucumber"
	generateOnly := false

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
		} else if arg == "--no-generate" {
			// Legacy flag - no longer used
		} else if arg == "--generate-only" {
			generateOnly = true
		} else if strings.HasPrefix(arg, "--as-") {
			fmt.Fprintf(os.Stderr, "Error: unknown format flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid formats: --as-cucumber (default), --as-junit\n")
			return 1
		} else if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid flags: --as-cucumber, --as-junit, --suite <name>\n")
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

	// Handle --generate-only flag (requires existing test-run-id)
	if generateOnly {
		if len(monikers) != 1 {
			fmt.Fprintf(os.Stderr, "Error: --generate-only requires exactly one test-run-id\n")
			fmt.Fprintf(os.Stderr, "Usage: test <test-run-id> --generate-only\n")
			return 1
		}
		testRunID := monikers[0]
		fmt.Printf("📊 Generating summary for test run: %s (skipping tests)\n", testRunID)
		if err := generateSummaryMulti(testRunID); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot, "0.1.0")
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

	// If exactly one module specified, run single module test (verbose output)
	if len(monikers) == 1 {
		return testSingleModule(monikers[0], workspaceRoot, moduleReport, reportFormat, suiteName)
	}

	// Multiple modules - run parallel test with orchestrator
	return testMultipleModules(monikers, workspaceRoot, moduleReport, reportFormat, suiteName)
}

// testSingleModule tests a single module with verbose output
func testSingleModule(moniker string, workspaceRoot string, moduleReport *reports.ModuleContractReport, reportFormat string, suiteName string) int {
	// Load module contract
	moduleContract, err := modules.LoadSingleModule(workspaceRoot, moniker, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contract for '%s': %v\n", moniker, err)
		return 1
	}

	// Check if module has a type-specific test function
	testFunc, exists := testFunctions[moduleContract.Type]
	if exists {
		// Use type-based dispatch for modules with specific test functions
		outputDir := filepath.Join(workspaceRoot, "out", "test", suiteName, moniker)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create output directory: %v\n", err)
			return 1
		}

		logWriter := os.Stdout
		return testFunc(moduleContract, workspaceRoot, outputDir, logWriter, reportFormat, suiteName)
	}

	// For modules without specific test functions, redirect to test suite
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	newArgs := []string{
		os.Args[0],
		"test",
		"suite",
		suiteName,
		"--module",
		moniker,
	}

	os.Args = newArgs

	return TestSuite()
}

// testMultipleModules tests multiple modules in parallel using the orchestrator
func testMultipleModules(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport, reportFormat string, suiteName string) int {
	// Configure orchestrator
	config := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        "out/test",
		LogFileName:          "test.log",
		OrchestratorLogName:  "orchestrator.log",
		ActionVerb:           "testing",
		MaxConcurrency:       0, // Use default (number of CPUs)
		StatusUpdateInterval: 2, // Update every 2 seconds
	}

	// Create worker function that tests a single module
	worker := func(moniker string, logWriter io.Writer) int {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			fmt.Fprintf(logWriter, "Error: module not found: %s\n", moniker)
			return 1
		}

		moduleOutputDir := filepath.Join(workspaceRoot, "out", "test", moniker)
		return runModuleTest(module, workspaceRoot, moduleOutputDir, logWriter, reportFormat, suiteName)
	}

	// Create and run orchestrator
	orch := orchestrator.New(config, worker)
	results, err := orch.Run(monikers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Print summary and close orchestrator
	orch.PrintSummary(results)
	orch.Close()

	// Return exit code based on results
	return orchestrator.GetExitCode(results)
}

func printTestUsage() {
	fmt.Println("Test one or more modules by moniker")
	fmt.Println()
	fmt.Println("Usage: r2r eac test [module1] [module2] ... [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --as-cucumber          Output in Cucumber JSON format (default)")
	fmt.Println("  --as-junit             Output in JUnit XML format")
	fmt.Println("  --suite <name>         Filter tests by suite (default: \"commit\")")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  r2r eac test                          # Test all modules")
	fmt.Println("  r2r eac test src-commands             # Test single module (verbose)")
	fmt.Println("  r2r eac test src-cli src-core         # Test multiple modules (parallel)")
	fmt.Println("  r2r eac test src-commands --as-junit  # Test with JUnit output")
	fmt.Println()
	fmt.Println("Related commands:")
	fmt.Println("  r2r eac test suite <name>             # Run a specific test suite")
	fmt.Println("  r2r eac test list-suites              # List all available test suites")
	fmt.Println()
	fmt.Println("Use 'r2r eac test <subcommand> --help' for more information about a command.")
}
