// Command: test modules
// Description: Test multiple modules in sequence and collect results in a test run directory
// Usage: test modules <moniker1> <moniker2> ... [--as-cucumber|--as-junit]
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/src/contracts/modules"
	"github.com/ready-to-release/eac/src/contracts/reports"
	"github.com/ready-to-release/eac/src/reports/cucumber"
)

func init() {
	Register("test modules", TestModules)
}

// TestModules tests multiple modules in sequence
func TestModules() int {
	// Parse arguments and flags
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: missing module monikers\n")
		fmt.Fprintf(os.Stderr, "Usage: test modules <moniker1> <moniker2> ... [--as-cucumber|--as-junit]\n")
		return 1
	}

	// Parse module monikers and format flag
	var monikers []string
	reportFormat := "cucumber"

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--as-cucumber" {
			reportFormat = "cucumber"
		} else if arg == "--as-junit" {
			reportFormat = "junit"
		} else if arg[:2] == "--" {
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid formats: --as-cucumber (default), --as-junit\n")
			return 1
		} else {
			monikers = append(monikers, arg)
		}
	}

	if len(monikers) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no module monikers provided\n")
		return 1
	}

	// Get repository root
	workspaceRoot, err := filepath.Abs("../..")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to determine workspace root: %v\n", err)
		return 1
	}

	// Create test-run-id directory
	testRunID := time.Now().Format("2006-01-02-150405")
	testRunDir := filepath.Join(workspaceRoot, "out", "test-results", testRunID)
	if err := os.MkdirAll(testRunDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create test run directory: %v\n", err)
		return 1
	}

	fmt.Printf("Test Run ID: %s\n", testRunID)
	fmt.Printf("Test Run Directory: %s\n", testRunDir)
	fmt.Printf("Testing %d modules: %v\n\n", len(monikers), monikers)

	// Load module contracts
	report, err := reports.GetModuleContracts(workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// Test each module in sequence
	failedModules := []string{}
	for i, moniker := range monikers {
		fmt.Printf("=== [%d/%d] Testing module: %s ===\n", i+1, len(monikers), moniker)

		// Get module from registry
		module, exists := report.Registry.Get(moniker)
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moniker)
			failedModules = append(failedModules, moniker+" (not found)")
			continue
		}

		// Create module output directory within test run
		moduleOutputDir := filepath.Join(testRunDir, moniker)
		if err := os.MkdirAll(moduleOutputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create module output directory: %v\n", err)
			failedModules = append(failedModules, moniker+" (dir error)")
			continue
		}

		// Create test log file
		logPath := filepath.Join(moduleOutputDir, "test.log")
		logFile, err := os.Create(logPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create log file: %v\n", err)
			failedModules = append(failedModules, moniker+" (log error)")
			continue
		}

		// Create multi-writer to log to both console and file
		multiWriter := io.MultiWriter(os.Stdout, logFile)

		// Run tests for this module
		exitCode := runModuleTest(module, workspaceRoot, moduleOutputDir, multiWriter, reportFormat)

		logFile.Close()

		if exitCode != 0 {
			failedModules = append(failedModules, moniker)
			fmt.Printf("❌ Module %s failed with exit code %d\n\n", moniker, exitCode)
		} else {
			fmt.Printf("✅ Module %s passed\n\n", moniker)
		}
	}

	// Print summary
	fmt.Println("===========================================")
	fmt.Printf("Test Run Summary (ID: %s)\n", testRunID)
	fmt.Println("===========================================")
	fmt.Printf("Total modules: %d\n", len(monikers))
	fmt.Printf("Passed: %d\n", len(monikers)-len(failedModules))
	fmt.Printf("Failed: %d\n", len(failedModules))
	if len(failedModules) > 0 {
		fmt.Println("\nFailed modules:")
		for _, m := range failedModules {
			fmt.Printf("  - %s\n", m)
		}
	}
	fmt.Printf("\nResults directory: %s\n", testRunDir)

	// Generate multi-module summary if all tests passed and using cucumber format
	if len(failedModules) == 0 && reportFormat == "cucumber" {
		fmt.Println("\n=== Generating multi-module summary.md ===")
		generateMultiModuleSummary(testRunID, testRunDir, workspaceRoot)
	}

	if len(failedModules) > 0 {
		return 1
	}
	return 0
}

// runModuleTest runs tests for a single module
func runModuleTest(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string) int {
	// Get test function for module type
	testFunc, hasTester := testFunctions[module.Type]
	if !hasTester {
		fmt.Fprintf(logWriter, "Error: no test function for type: %s\n", module.Type)
		return 1
	}

	// Execute the test function
	return testFunc(module, workspaceRoot, outputDir, logWriter, reportFormat)
}

// generateMultiModuleSummary generates a consolidated summary for all modules
func generateMultiModuleSummary(testRunID string, testRunDir string, workspaceRoot string) {
	summaryPath := filepath.Join(testRunDir, "summary.md")

	// Find all module directories with cucumber.json
	modules, err := findModulesWithResults(testRunDir)
	if err != nil {
		fmt.Printf("Warning: failed to find module results: %v\n", err)
		return
	}

	if len(modules) == 0 {
		fmt.Println("Warning: no module results found")
		return
	}

	fmt.Printf("Found %d module(s) with test results\n", len(modules))

	// Generate multi-module summary
	var summary string
	summary += "# Test Summary\n\n"
	summary += fmt.Sprintf("**Test Run ID**: %s\n\n", testRunID)

	// Render each module as a section
	for i, moduleName := range modules {
		moduleDir := filepath.Join(testRunDir, moduleName)
		cucumberPath := filepath.Join(moduleDir, "cucumber.json")

		// Parse cucumber.json
		report, err := cucumber.ParseFile(cucumberPath)
		if err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", cucumberPath, err)
			continue
		}

		// Add module section header
		summary += fmt.Sprintf("### Module: %s\n\n", moduleName)

		// Render features for this module
		summary += cucumber.RenderAllFeatures(report, nil)

		// Add separator between modules (but not after the last one)
		if i < len(modules)-1 {
			summary += "\n---\n\n"
		}
	}

	// Add Appendix A with all specifications
	summary += "\n---\n\n"
	summary += "## Appendix A: Specifications and Test Results\n\n"

	for _, moduleName := range modules {
		moduleDir := filepath.Join(testRunDir, moduleName)
		cucumberPath := filepath.Join(moduleDir, "cucumber.json")

		report, err := cucumber.ParseFile(cucumberPath)
		if err != nil {
			continue
		}

		// Render appendix for this module
		summary += cucumber.RenderAppendixA(report, workspaceRoot)
	}

	// Write summary.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		fmt.Printf("Warning: failed to write summary.md: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated: %s\n", summaryPath)
}

// findModulesWithResults finds all subdirectories containing cucumber.json
func findModulesWithResults(testRunDir string) ([]string, error) {
	entries, err := os.ReadDir(testRunDir)
	if err != nil {
		return nil, err
	}

	var modules []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this directory has cucumber.json
		cucumberPath := filepath.Join(testRunDir, entry.Name(), "cucumber.json")
		if _, err := os.Stat(cucumberPath); err == nil {
			modules = append(modules, entry.Name())
		}
	}

	return modules, nil
}
