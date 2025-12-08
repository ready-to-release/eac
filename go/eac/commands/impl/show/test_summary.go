// Command: show test-summary
// Description: Generate pretty test summary for a module
// Short: Generate pretty test summary for a module
// Long: The show test-summary command generates a formatted test summary with test results, metrics, and diagnostics.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent, attractive test summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
package show

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(ShowTestSummary)
}

// ShowTestSummary generates a pretty test summary
func ShowTestSummary() int {
	args := os.Args[3:] // Skip program name, "show", and "test-summary"

	if len(args) < 2 {
		log.Errorf("Usage: show test-summary <module> <suite> [--status=success|failure] [--run-id=<id>]")
		return 1
	}

	module := args[0]
	suite := args[1]
	status := "success"
	runID := ""

	// Parse flags
	for i := 2; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 9 && arg[:9] == "--status=" {
			status = arg[9:]
		} else if len(arg) > 9 && arg[:9] == "--run-id=" {
			runID = arg[9:]
		}
	}

	return generateTestSummary(module, suite, status, runID)
}

func generateTestSummary(moduleName, suite, status, runID string) int {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	// Get module contract
	module, ok := cfg.Modules.GetModule(moduleName)
	if !ok {
		log.Errorf("module not found: %s", moduleName)
		return 1
	}

	// Create formatter
	formatter := NewSummaryFormatter(moduleName, status)

	// Generate summary
	summary := testSummaryContent(formatter, module, suite, status, cfg)

	// Add footer with duration
	duration := time.Since(startTime)
	summary += formatter.Footer(duration)

	// Output to stdout (caller redirects to GITHUB_STEP_SUMMARY)
	fmt.Print(summary)

	return 0
}

func testSummaryContent(f *SummaryFormatter, module *config.Module, suite, status string, cfg *config.EACConfig) string {
	var summary string

	// Header
	summary += f.Header("test")
	summary += fmt.Sprintf("**Suite**: %s\n\n", Code(suite))

	// Status section
	if status == "success" {
		summary += f.StatusSection("All tests passed")
	} else {
		summary += f.StatusSection("Tests failed")
	}

	// Test results (if available)
	if status == "success" {
		summary += testMetricsSection(f, module, suite)
	} else {
		summary += testDiagnosticsSection(f, module)
	}

	// Test configuration (collapsible)
	summary += testConfigSection(f, module, suite, cfg)

	return summary
}

func testMetricsSection(f *SummaryFormatter, module *config.Module, suite string) string {
	outputDir := filepath.Join("out", "test", suite)

	// Try to read test summary JSON if available
	summaryFile := filepath.Join(outputDir, fmt.Sprintf("%s-summary.json", module.Moniker))
	if testResults := readTestSummaryJSON(summaryFile); testResults != nil {
		return formatTestResults(f, testResults)
	}

	// Fallback: just show that tests passed
	return f.Section(Emoji("success")+" Test Results", fmt.Sprintf("Tests completed successfully\n\nOutput: %s", Code(outputDir)))
}

// TestResults represents parsed test results
type TestResults struct {
	Packages int                  `json:"packages"`
	Total    int                  `json:"total"`
	Passed   int                  `json:"passed"`
	Failed   int                  `json:"failed"`
	Skipped  int                  `json:"skipped"`
	Duration float64              `json:"duration"`
	Details  []PackageTestResults `json:"details,omitempty"`
}

// PackageTestResults represents test results for a single package
type PackageTestResults struct {
	Package  string  `json:"package"`
	Tests    int     `json:"tests"`
	Duration float64 `json:"duration"`
	Status   string  `json:"status"`
}

func readTestSummaryJSON(path string) *TestResults {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var results TestResults
	if err := json.Unmarshal(data, &results); err != nil {
		return nil
	}

	return &results
}

func formatTestResults(f *SummaryFormatter, results *TestResults) string {
	var summary string

	// Overall metrics
	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Packages Tested", fmt.Sprintf("%d", results.Packages)},
		{"Total Tests", fmt.Sprintf("%d", results.Total)},
		{"Passed", fmt.Sprintf("%d %s", results.Passed, Emoji("success"))},
		{"Failed", fmt.Sprintf("%d", results.Failed)},
		{"Skipped", fmt.Sprintf("%d", results.Skipped)},
		{"Duration", fmt.Sprintf("%.1fs", results.Duration)},
	}

	summary += f.Section(Emoji("metrics")+" Test Results", f.Table(headers, rows))

	// Package breakdown (if available)
	if len(results.Details) > 0 {
		summary += packageBreakdown(f, results.Details)
	}

	return summary
}

func packageBreakdown(f *SummaryFormatter, details []PackageTestResults) string {
	headers := []string{"Package", "Tests", "Duration", "Status"}
	var rows [][]string

	for _, pkg := range details {
		status := Emoji("success")
		if pkg.Status == "failed" {
			status = Emoji("failure")
		}
		rows = append(rows, []string{
			pkg.Package,
			fmt.Sprintf("%d", pkg.Tests),
			fmt.Sprintf("%.1fs", pkg.Duration),
			status,
		})
	}

	return f.Section(Emoji("chart")+" Package Breakdown", f.Table(headers, rows))
}

func testDiagnosticsSection(f *SummaryFormatter, module *config.Module) string {
	var diagnostics string

	// Read actual test log
	logPath := filepath.Join("out", "logs", fmt.Sprintf("%s-test.log", module.Moniker))
	logContent := readLogTail(logPath, 100) // Last 100 lines for test failures

	if logContent != "" {
		diagnostics += f.Section(Emoji("diagnostics")+" Test Log (last 100 lines)", f.CodeBlock("", logContent))
	} else {
		diagnostics += f.Section(Emoji("diagnostics")+" Diagnostics", "Tests failed - no log file found")
	}

	// Show test timing if available
	timingPath := filepath.Join("out", "test", "test-timing.txt")
	if timing, err := os.ReadFile(timingPath); err == nil {
		diagnostics += f.Section(Emoji("time")+" Timing", string(timing))
	}

	return diagnostics
}

func testConfigSection(f *SummaryFormatter, module *config.Module, suite string, cfg *config.EACConfig) string {
	var configDetails string

	// Suite info
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Suite"), suite)

	// Test framework (from module type)
	if cfg.ModuleTypes != nil {
		testFramework := cfg.ModuleTypes.GetTestFramework(module.Type)
		if testFramework == "" {
			testFramework = "default"
		}
		configDetails += fmt.Sprintf("- %s: %s\n", Bold("Framework"), testFramework)
	}

	// Output directory
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Output"), Code(fmt.Sprintf("out/test/%s", suite)))

	return f.CollapsibleSection(Emoji("config")+" Test Configuration", configDetails)
}
