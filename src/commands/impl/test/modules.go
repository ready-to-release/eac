// Command: test modules
// Short: Test multiple modules in sequence and collect results in a test run directory
// Long: Test multiple modules in sequence and collect results in a test run directory.
// Long:
// Long: This command tests multiple modules respecting their dependency order.
// Long: If no monikers are specified, all modules in the repository are tested.
// Long:
// Long: Test results are collected in a timestamped directory under 'out/test-runs/'
// Long: containing logs, reports, and summary information for each module. Failed tests
// Long: are clearly marked and do not stop the execution of remaining modules.
// Long:
// Long: Example:
// Long:   test modules                     # Test all modules
// Long:   test modules src-core src-cli    # Test specific modules
// Long:   test modules --as-junit          # Generate JUnit XML reports
// Flag.as-cucumber: type=bool, usage=Output test results in Cucumber JSON format
// Flag.as-junit: type=bool, usage=Output test results in JUnit XML format
// HasSideEffects: false
package test

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	// Parse module monikers and flags (default: cucumber format)
	var monikers []string
	reportFormat := "cucumber"
	generateOnly := false

	// Parse arguments starting from index 3 (skip "binary", "test", "modules")
	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--as-cucumber" {
			reportFormat = "cucumber"
		} else if arg == "--as-junit" {
			reportFormat = "junit"
		} else if arg == "--no-generate" {
			// Legacy flag - no longer used
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

	// Create orchestrator log file
	orchestratorLogPath := filepath.Join(workspaceRoot, "out", "test", "orchestrator.log")
	if err := os.MkdirAll(filepath.Dir(orchestratorLogPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create orchestrator log directory: %v\n", err)
		return 1
	}
	orchestratorLog, err := os.Create(orchestratorLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create orchestrator log: %v\n", err)
		return 1
	}
	defer orchestratorLog.Close()

	// Use buffered writer for orchestrator log
	orchestratorLogBuf := bufio.NewWriter(orchestratorLog)
	defer orchestratorLogBuf.Flush()

	// Create multi-writer for orchestrator output (console + log file)
	orchestratorOut := io.MultiWriter(os.Stdout, orchestratorLogBuf)

	fmt.Fprintf(orchestratorOut, "Testing %d modules: %v\n\n", len(monikers), monikers)

	// Start global progress tracker
	StartGlobalTracker(orchestratorOut, len(monikers))
	defer StopGlobalTracker()

	// Test each module in sequence
	var mu sync.Mutex
	failedModules := []string{}
	testedModules := []*modules.ModuleContract{}
	for _, moniker := range monikers {

		// Get module from registry
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			statusLine := fmt.Sprintf("[testing] %s (Module not found) ........ Failed\r\n", moniker)
			orchestratorOut.Write([]byte(statusLine))
			os.Stdout.Sync()
			failedModules = append(failedModules, moniker+" (not found)")
			continue
		}

		// Purge and create module output directory
		moduleOutputDir := filepath.Join(workspaceRoot, "out", "test", moniker)
		if err := os.RemoveAll(moduleOutputDir); err != nil {
			// Silently continue - logged to orchestrator log only
		}
		if err := os.MkdirAll(moduleOutputDir, 0755); err != nil {
			statusLine := fmt.Sprintf("[testing] %s (Failed to create directory) ........ Failed\r\n", moniker)
			orchestratorOut.Write([]byte(statusLine))
			os.Stdout.Sync()
			failedModules = append(failedModules, moniker+" (dir error)")
			continue
		}

		// Create test log file
		logPath := filepath.Join(moduleOutputDir, "test.log")
		logFile, err := os.Create(logPath)
		if err != nil {
			statusLine := fmt.Sprintf("[testing] %s (Failed to create log) ........ Failed\r\n", moniker)
			orchestratorOut.Write([]byte(statusLine))
			os.Stdout.Sync()
			failedModules = append(failedModules, moniker+" (log error)")
			continue
		}

		// Module output goes to file only (not console)
		multiWriter := io.MultiWriter(logFile)

		// Track test start with global progress tracker
		TrackTestStart(moniker)

		// Run tests for this module
		exitCode := runModuleTest(module, workspaceRoot, moduleOutputDir, multiWriter, reportFormat)

		// Track test completion
		TrackTestComplete(moniker)

		logFile.Close()

		// Track tested modules
		testedModules = append(testedModules, module)

		// Print clean status line (thread-safe)
		mu.Lock()
		relLogPath := filepath.Join("out", "test", moniker, "test.log")
		var statusLine string
		if exitCode != 0 {
			statusLine = fmt.Sprintf("[testing] %s (See %s for details) ........ Failed\r\n", moniker, relLogPath)
			failedModules = append(failedModules, moniker)
		} else {
			statusLine = fmt.Sprintf("[testing] %s (See %s for details) ........ Done\r\n", moniker, relLogPath)
		}
		orchestratorOut.Write([]byte(statusLine))
		os.Stdout.Sync()
		mu.Unlock()
	}

	// Print summary
	fmt.Fprintf(orchestratorOut, "\n===========================================\n")
	fmt.Fprintf(orchestratorOut, "Test Run Summary\n")
	fmt.Fprintf(orchestratorOut, "===========================================\n")
	fmt.Fprintf(orchestratorOut, "Total modules: %d\n", len(monikers))
	fmt.Fprintf(orchestratorOut, "Passed: %d\n", len(monikers)-len(failedModules))
	fmt.Fprintf(orchestratorOut, "Failed: %d\n", len(failedModules))
	if len(failedModules) > 0 {
		fmt.Fprintf(orchestratorOut, "\n❌ Failed modules:\n")
		for _, m := range failedModules {
			fmt.Fprintf(orchestratorOut, "  - %s\n", m)
		}
	}
	fmt.Fprintf(orchestratorOut, "\nOrchestrator log: out/test/orchestrator.log\n")
	fmt.Fprintf(orchestratorOut, "Module logs: out/test/<module>/test.log\n")

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

// generateMultiModuleGherkinSummary generates a consolidated BDD summary for all modules
func generateMultiModuleGherkinSummary(testRunID string, testRunDir string, workspaceRoot string) {
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

// generateMultiModuleUnitTestSummary generates a consolidated test summary for all modules
func generateMultiModuleUnitTestSummary(testRunID string, testRunDir string, workspaceRoot string, modules []*modules.ModuleContract) {
	summaryPath := filepath.Join(testRunDir, "summary_unit.md")

	var summary string
	summary += "## Unit Test Summary\n\n"
	summary += fmt.Sprintf("**Test Run ID**: %s\n\n", testRunID)

	// Process each module
	passedCount := 0
	failedCount := 0

	for _, module := range modules {
		moduleOutputDir := filepath.Join(testRunDir, module.Moniker)
		testSummaryPath := filepath.Join(moduleOutputDir, "summary_unit.md")

		// Check if this module has a unit test summary (not all modules generate one - only unit tests)
		if _, err := os.Stat(testSummaryPath); err != nil {
			continue // Skip Gherkin-only modules
		}

		// Read the individual test summary
		content, err := os.ReadFile(testSummaryPath)
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", testSummaryPath, err)
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

// formatDuration formats a duration as "1m 23s" or "45s"
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
