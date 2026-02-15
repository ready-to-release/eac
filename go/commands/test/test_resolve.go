package test

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/core/testing"
)

// printTestSummary prints unified test summary to a writer (for non-TUI mode)
// Note: suiteName/suiteMoniker kept for API compatibility but suite info is shown during init.
func printTestSummary(w io.Writer, results []PackageResult, suiteName, suiteMoniker string,
	selectedCount, osFilteredCount,
	totalPackages, packagesPassed, packagesFailed,
	testsTotal, testsPassed, testsFailed int,
	testRunDir string,
) {
	// Note: Suite info is displayed during init phase (via writeInitSummary)
	// The suiteName/suiteMoniker params are kept for potential future use
	_ = suiteName
	_ = suiteMoniker

	testsSkipped := testsTotal - testsPassed - testsFailed
	moduleStats := aggregateResultsByModule(results)

	// Init summary line
	initSummary := fmt.Sprintf("%d test cases selected", selectedCount)
	if osFilteredCount > 0 {
		initSummary += fmt.Sprintf(" (%d filtered by OS)", osFilteredCount)
	}

	// Run summary line
	var runSummary string
	if packagesFailed == 0 && testsFailed == 0 {
		assertionInfo := fmt.Sprintf("%d assertions", testsTotal)
		if testsSkipped > 0 {
			assertionInfo += fmt.Sprintf(", %d skipped", testsSkipped)
		}
		runSummary = fmt.Sprintf("%d packages passed (%s)", totalPackages, assertionInfo)
	} else {
		runSummary = fmt.Sprintf("%d/%d packages passed, %d assertions failed",
			packagesPassed, totalPackages, testsFailed)
	}

	// Status icons
	initIcon := "✓"
	runIcon := "✓"
	if packagesFailed > 0 || testsFailed > 0 {
		runIcon = "✗"
	}

	fmt.Fprintf(w, "%s Initialization: %s\n", initIcon, initSummary)
	fmt.Fprintf(w, "%s Testing: %s\n", runIcon, runSummary)
	fmt.Fprintln(w)

	// Module breakdown table
	fmt.Fprintln(w, strings.Repeat("-", 58))
	fmt.Fprintln(w, "Module               Pkgs  Asserts  Fail  Skip  Pass")
	fmt.Fprintln(w, strings.Repeat("-", 58))
	for _, ms := range moduleStats {
		fmt.Fprintf(w, "%-20s %4d  %7d  %4d  %4d  %4d\n",
			truncateString(ms.Module, 20), ms.Packages, ms.Assertions, ms.Failed, ms.Skipped, ms.Passed)
	}
	fmt.Fprintln(w, strings.Repeat("-", 58))
	fmt.Fprintf(w, "%-20s %4d  %7d  %4d  %4d  %4d\n",
		"TOTAL", totalPackages, testsTotal, testsFailed, testsSkipped, testsPassed)
	fmt.Fprintln(w)

	// Results path
	fmt.Fprintf(w, "Results: %s\n", testRunDir)
}

// testTUISummary creates summary data for the TUI Summary pane.
func testTUISummary(
	results []PackageResult, totalTime time.Duration, suiteName, suiteMoniker string,
	osFilteredCount, selectedCount,
	totalPackages, packagesPassed, packagesFailed,
	testsTotal, testsPassed, testsFailed int,
	testRunDir string,
) *tui.SummaryData {
	testsSkipped := testsTotal - testsPassed - testsFailed

	// Format suite name(s) for display
	displaySuiteNames := suiteName
	suiteLabel := "suite"
	if suitesIncluded := getSuitesIncluded(suiteMoniker); len(suitesIncluded) > 0 {
		displaySuiteNames = strings.Join(suitesIncluded, ", ")
		suiteLabel = "suite(s)"
	}

	// Init summary: test cases selected (scenarios/functions we chose to run)
	initSummary := fmt.Sprintf("%d test cases selected", selectedCount)
	if osFilteredCount > 0 {
		initSummary += fmt.Sprintf(" (%d filtered by OS)", osFilteredCount)
	}

	// Run summary: packages are the unit of execution, assertions are what ran inside them
	// (testsTotal includes subtests from t.Run, which is why it can exceed selectedCount)
	var runSummary string
	if packagesFailed == 0 && testsFailed == 0 {
		// All passed - show packages passed and total assertions
		assertionInfo := fmt.Sprintf("%d assertions", testsTotal)
		if testsSkipped > 0 {
			assertionInfo += fmt.Sprintf(", %d skipped", testsSkipped)
		}
		runSummary = fmt.Sprintf("%d packages passed (%s)", totalPackages, assertionInfo)
	} else {
		// Failures - show what failed
		runSummary = fmt.Sprintf("%d/%d packages passed, %d assertions failed",
			packagesPassed, totalPackages, testsFailed)
	}

	// Build per-module breakdown
	moduleStats := aggregateResultsByModule(results)

	var details []string
	if packagesFailed > 0 || testsFailed > 0 {
		failedPackages := []string{}
		for _, result := range results {
			if result.PackageFailed || result.TestsFailed > 0 {
				failedPackages = append(failedPackages, result.PackageName)
			}
		}
		details = append(details, fmt.Sprintf("Failed: %d packages, %d tests", packagesFailed, testsFailed))
		if len(failedPackages) <= 3 {
			details = append(details, fmt.Sprintf("  (%s)", strings.Join(failedPackages, ", ")))
		} else {
			details = append(details, fmt.Sprintf("  (%s, +%d more)", failedPackages[0], len(failedPackages)-1))
		}
	}

	// Add per-module breakdown table
	details = append(details, "")
	details = append(details, strings.Repeat("-", 58))
	details = append(details, "Module               Pkgs  Asserts  Fail  Skip  Pass")
	details = append(details, strings.Repeat("-", 58))
	for _, ms := range moduleStats {
		details = append(details, fmt.Sprintf("%-20s %4d  %7d  %4d  %4d  %4d",
			truncateString(ms.Module, 20), ms.Packages, ms.Assertions, ms.Failed, ms.Skipped, ms.Passed))
	}
	details = append(details, strings.Repeat("-", 58))
	details = append(details, fmt.Sprintf("%-20s %4d  %7d  %4d  %4d  %4d",
		"TOTAL", totalPackages, testsTotal, testsFailed, testsSkipped, testsPassed))

	details = append(details, "")
	details = append(details, fmt.Sprintf("Results: %s", testRunDir))

	nextSteps := ""
	if packagesFailed > 0 || testsFailed > 0 {
		nextSteps = "Review detailed failure output below"
	} else {
		nextSteps = fmt.Sprintf("All tests passed for %s: %s", suiteLabel, displaySuiteNames)
	}

	return &tui.SummaryData{
		Success:     packagesFailed == 0 && testsFailed == 0,
		TotalTime:   totalTime,
		InitSummary: initSummary,
		RunSummary:  runSummary,
		Details:     details,
		NextSteps:   nextSteps,
	}
}

// moduleTestStats holds aggregated test statistics for a module.
type moduleTestStats struct {
	Module     string
	Packages   int
	Passed     int
	Failed     int
	Skipped    int
	Assertions int
}

// aggregateResultsByModule groups test results by module moniker.
func aggregateResultsByModule(results []PackageResult) []moduleTestStats {
	moduleMap := make(map[string]*moduleTestStats)

	for _, result := range results {
		// Use ModuleMoniker directly (set by runner via GetTestInfo)
		module := result.ModuleMoniker
		if module == "" {
			// Fallback: extract from package name if ModuleMoniker not set
			module = result.PackageName
			if idx := strings.Index(module, "/"); idx > 0 {
				module = module[:idx]
			}
		}

		stats, exists := moduleMap[module]
		if !exists {
			stats = &moduleTestStats{Module: module}
			moduleMap[module] = stats
		}

		stats.Packages++
		stats.Passed += result.TestsPassed
		stats.Failed += result.TestsFailed
		stats.Skipped += result.TestsSkipped
		stats.Assertions += result.TestsTotal
	}

	// Convert to sorted slice
	var moduleList []moduleTestStats
	for _, stats := range moduleMap {
		moduleList = append(moduleList, *stats)
	}
	sort.Slice(moduleList, func(i, j int) bool {
		return moduleList[i].Module < moduleList[j].Module
	})

	return moduleList
}

// truncateString truncates a string to maxLen, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Terminology for test metrics:
//
// | Framework    | Test Case                | Package              | Assertion              |
// |--------------|--------------------------|----------------------|------------------------|
// | Go (gotest)  | Test* function           | Go package directory | t.Run subtests + main  |
// | Go (godog)   | Gherkin scenario         | Feature file folder  | Scenario steps         |
// | TS (mocha)   | describe() block         | Test file            | it() blocks            |
// | TS (cucumber)| Gherkin scenario         | Feature file folder  | Scenario steps         |
//
// - "Test cases selected" = scenarios + Test* functions discovered and matched by suite filter
// - "Packages" = execution units (directories/files containing test cases)
// - "Assertions" = actual test executions reported by runner (includes subtests)

// getUniqueModulesFromTests extracts unique module monikers from test references
// Uses ModuleMapper for accurate module ownership lookup from registry.
func getUniqueModulesFromTests(tests []testing.TestReference, mapper *ModuleMapper) []string {
	moduleSet := make(map[string]bool)
	for i := range tests {
		test := &tests[i]
		module := mapper.GetModuleForFile(test.FilePath)
		if module != "" {
			moduleSet[module] = true
		}
	}

	modules := make([]string, 0, len(moduleSet))
	for module := range moduleSet {
		modules = append(modules, module)
	}

	// Sort for consistent output
	sort.Strings(modules)
	return modules
}

// incrementalTestInfo holds information about incremental test detection results.
type incrementalTestInfo struct {
	detectionTime   time.Duration
	modulesNeedTest []string
	modulesUpToDate []string
	freshRun        bool
	changeReasons   map[string]string
}
