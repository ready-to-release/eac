// Command: test modules
// Description: Test multiple modules in sequence and collect results in a test run directory
// Usage: test modules [moniker1] [moniker2] ... [--as-cucumber|--as-junit]
// Default: Tests all modules if no monikers specified
// HasSideEffects: false
package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/commands/impl/test/internal/cucumber"
)

func init() {
	registry.Register(TestModules)
}

// TestModules tests multiple modules in sequence (defaults to all modules)
func TestModules() int {
	// Parse module monikers and flags (default: cucumber format, generate summary enabled)
	var monikers []string
	reportFormat := "cucumber"
	generateSummaryEnabled := true
	generateOnly := false

	// Parse arguments starting from index 3 (skip "binary", "test", "modules")
	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--as-cucumber" {
			reportFormat = "cucumber"
		} else if arg == "--as-junit" {
			reportFormat = "junit"
		} else if arg == "--no-generate" {
			generateSummaryEnabled = false
		} else if arg == "--generate-only" {
			generateOnly = true
		} else if strings.HasPrefix(arg, "--as-") {
			fmt.Fprintf(os.Stderr, "Error: unknown format flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid formats: --as-cucumber (default), --as-junit\n")
			return 1
		} else if strings.HasPrefix(arg, "--") {
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Valid flags: --as-cucumber, --as-junit, --no-generate, --generate-only\n")
			return 1
		} else {
			monikers = append(monikers, arg)
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Handle --generate-only flag (requires existing test-run-id)
	if generateOnly {
		if len(monikers) != 1 {
			fmt.Fprintf(os.Stderr, "Error: --generate-only requires exactly one test-run-id\n")
			fmt.Fprintf(os.Stderr, "Usage: test modules <test-run-id> --generate-only\n")
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

	// Test each module in sequence
	failedModules := []string{}
	testedModules := []*modules.ModuleContract{}
	for i, moniker := range monikers {
		fmt.Printf("=== [%d/%d] Testing module: %s ===\n", i+1, len(monikers), moniker)

		// Get module from registry
		module, exists := moduleReport.Registry.Get(moniker)
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

		// Track tested modules
		testedModules = append(testedModules, module)

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

	// Generate summary if enabled and using cucumber format
	if generateSummaryEnabled && reportFormat == "cucumber" && len(failedModules) == 0 {
		fmt.Println("\n📊 Generating test summary...")
		if err := generateSummaryMulti(testRunID); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to generate summary: %v\n", err)
			// Don't fail the test run, just warn
		}
	}

	// Also generate individual module summaries (BDD and TDD) if all tests passed
	if len(failedModules) == 0 {
		// Generate BDD summary if using cucumber format
		if reportFormat == "cucumber" {
			fmt.Println("\n=== Generating multi-module summary_acceptance.md ===")
			generateMultiModuleBDDSummary(testRunID, testRunDir, workspaceRoot)
		}

		// Generate TDD summary for all modules
		fmt.Println("\n=== Generating multi-module summary_unit.md ===")
		generateMultiModuleTDDSummary(testRunID, testRunDir, workspaceRoot, testedModules)
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

// generateMultiModuleBDDSummary generates a consolidated BDD summary for all modules
func generateMultiModuleBDDSummary(testRunID string, testRunDir string, workspaceRoot string) {
	summaryPath := filepath.Join(testRunDir, "summary_acceptance.md")
	appendixPath := filepath.Join(testRunDir, "appendix_a.md")

	// Find all module directories with cucumber.json
	modules, err := FindModulesWithResults(testRunDir)
	if err != nil {
		fmt.Printf("Warning: failed to find module results: %v\n", err)
		return
	}

	if len(modules) == 0 {
		fmt.Println("Warning: no module results found")
		return
	}

	fmt.Printf("Found %d module(s) with test results\n", len(modules))

	// Generate multi-module summary (fragment starting at level 2)
	var summary string
	summary += "## Acceptance Test Summary\n\n"
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
		summary += fmt.Sprintf("#### Module: %s\n\n", moduleName)

		// Render features for this module
		summary += cucumber.RenderAllFeatures(report, nil)

		// Add separator between modules (but not after the last one)
		if i < len(modules)-1 {
			summary += "\n---\n\n"
		}
	}

	// Write summary_acceptance.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		fmt.Printf("Warning: failed to write summary_acceptance.md: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated: %s\n", summaryPath)

	// Generate Appendix A with all specifications as separate file (fragment starting at level 2)
	var appendix string
	appendix += "## Appendix A: Specifications and Test Results\n\n"

	for _, moduleName := range modules {
		moduleDir := filepath.Join(testRunDir, moduleName)
		cucumberPath := filepath.Join(moduleDir, "cucumber.json")

		report, err := cucumber.ParseFile(cucumberPath)
		if err != nil {
			continue
		}

		// Render appendix for this module
		appendix += cucumber.RenderAppendixA(report, workspaceRoot)
	}

	// Write appendix_a.md
	if err := os.WriteFile(appendixPath, []byte(appendix), 0644); err != nil {
		fmt.Printf("Warning: failed to write appendix_a.md: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated: %s\n", appendixPath)
}

// generateMultiModuleTDDSummary generates a consolidated TDD summary for all modules
func generateMultiModuleTDDSummary(testRunID string, testRunDir string, workspaceRoot string, modules []*modules.ModuleContract) {
	summaryPath := filepath.Join(testRunDir, "summary_unit.md")

	var summary string
	summary += "## Unit Test Summary\n\n"
	summary += fmt.Sprintf("**Test Run ID**: %s\n\n", testRunID)

	// Process each module
	passedCount := 0
	failedCount := 0

	for _, module := range modules {
		moduleOutputDir := filepath.Join(testRunDir, module.Moniker)
		tddSummaryPath := filepath.Join(moduleOutputDir, "summary_unit.md")

		// Check if this module has a unit test summary (not all modules generate one - only non-BDD tests)
		if _, err := os.Stat(tddSummaryPath); err != nil {
			continue // Skip BDD-only modules
		}

		// Read the individual TDD summary
		content, err := os.ReadFile(tddSummaryPath)
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", tddSummaryPath, err)
			continue
		}

		// Check if module passed or failed
		contentStr := string(content)
		if strings.Contains(contentStr, "**Status**: ✅ Passed") {
			passedCount++
		} else if strings.Contains(contentStr, "**Status**: ❌ Failed") {
			failedCount++
		}

		// Add module section header
		summary += fmt.Sprintf("#### Module: %s\n\n", module.Moniker)
		summary += fmt.Sprintf("**Type**: %s\n", module.Type)

		// Extract status and test output from the individual summary
		lines := strings.Split(contentStr, "\n")
		inTestOutput := false
		for _, line := range lines {
			if strings.HasPrefix(line, "**Status**:") {
				summary += line + "\n"
			} else if strings.HasPrefix(line, "### Test Output") {
				inTestOutput = true
				summary += "\n" + line + "\n"
			} else if inTestOutput {
				summary += line + "\n"
			}
		}

		summary += "\n---\n\n"
	}

	// Add overall summary at the top
	overallStatus := "✅ Passed"
	if failedCount > 0 {
		overallStatus = "❌ Failed"
	}

	// Prepend overall summary
	header := "## Unit Test Summary\n\n"
	header += fmt.Sprintf("**Test Run ID**: %s\n", testRunID)
	header += fmt.Sprintf("**Overall Status**: %s\n", overallStatus)
	header += fmt.Sprintf("**Total Modules**: %d\n", passedCount+failedCount)
	header += fmt.Sprintf("**Passed**: %d\n", passedCount)
	header += fmt.Sprintf("**Failed**: %d\n\n", failedCount)
	header += "---\n\n"

	summary = header + strings.TrimPrefix(summary, "## Unit Test Summary\n\n"+fmt.Sprintf("**Test Run ID**: %s\n\n", testRunID))

	// Write summary_unit.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		fmt.Printf("Warning: failed to write summary_unit.md: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated: %s\n", summaryPath)
}

// findModulesWithResults finds all subdirectories containing cucumber.json
func FindModulesWithResults(testRunDir string) ([]string, error) {
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
