// Command: show test-summary
// Description: Generate pretty test summary for a module
// Short: Generate pretty test summary for a module
// Long: The show test-summary command generates a formatted test summary with test results, metrics, and diagnostics.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent, attractive test summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted test summary with emojis and styling, suitable for GitHub Actions $GITHUB_STEP_SUMMARY
// Long: - Success: includes status section, test metrics table (packages, tests, passed/failed/skipped, duration), package breakdown
// Long: - Failure: includes status section, diagnostics with last 100 lines of test log, timing data, and configuration
package show

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowTestSummary)
}

// ShowTestSummary generates a pretty test summary
func ShowTestSummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "test-summary"

	if len(args) < 2 {
		log.Errorf("Usage: show test-summary <module> <suite> [--run-id=<id>]")
		return 1
	}

	module := args[0]
	suite := args[1]
	runID := flags.GetFlagValue(args, "--run-id")

	return generateTestSummary(module, suite, runID)
}

func generateTestSummary(moduleName, suite, runID string) int {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	// Get module contract
	module, ok := cfg.Repository.GetModule(moduleName)
	if !ok {
		log.Errorf("module not found: %s", moduleName)
		return 1
	}

	// Derive status from test manifest
	status := deriveTestStatus(cfg, moduleName)

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
		summary += testMetricsSection(f, module, suite, cfg)
	} else {
		summary += testDiagnosticsSection(f, module, suite, cfg)
	}

	// Test configuration (collapsible)
	summary += testConfigSection(f, module, suite, cfg)

	return summary
}

func testMetricsSection(f *SummaryFormatter, module *config.Module, suite string, cfg *config.EACConfig) string {
	outputDir := cfg.Repository.TestModuleDir(module.Moniker)

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

func testDiagnosticsSection(f *SummaryFormatter, module *config.Module, suite string, cfg *config.EACConfig) string {
	var diagnostics string

	// Read actual test log from the correct output directory
	// Test logs are output to out/test/<module>/packages/<package>/test.log
	// For modules with subpackages, logs may be in subdirectories
	moduleDir := cfg.Repository.TestModuleDir(module.Moniker)
	packagesDir := filepath.Join(moduleDir, "packages")

	// Try to find test.log files in the packages directory
	logContent := findAndReadSubpackageLogs(packagesDir, 100)

	if logContent != "" {
		diagnostics += f.Section(Emoji("diagnostics")+" Test Log (last 100 lines)", f.CodeBlock("", logContent))
	} else {
		diagnostics += f.Section(Emoji("diagnostics")+" Diagnostics", fmt.Sprintf("Tests failed - no log file found in %s", packagesDir))
	}

	// Show test timing if available
	timingPath := cfg.Repository.TestModuleTimingPath(module.Moniker)
	if timing, err := os.ReadFile(timingPath); err == nil {
		diagnostics += f.Section(Emoji("time")+" Timing", string(timing))
	}

	return diagnostics
}

// findAndReadSubpackageLogs searches for test.log files in subdirectories and combines their content
func findAndReadSubpackageLogs(moduleDir string, maxLines int) string {
	var logs []string

	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && info.Name() == "test.log" {
			content := readLogTail(path, 50) // Fewer lines per subpackage
			if content != "" {
				relPath, _ := filepath.Rel(moduleDir, path)
				logs = append(logs, fmt.Sprintf("=== %s ===\n%s", relPath, content))
			}
		}
		return nil
	})

	if err != nil || len(logs) == 0 {
		return ""
	}

	// Combine logs, limiting total output
	combined := ""
	for _, log := range logs {
		combined += log + "\n\n"
	}
	return combined
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
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Output"), Code(cfg.Repository.TestModuleDir(module.Moniker)))

	return f.CollapsibleSection(Emoji("config")+" Test Configuration", configDetails)
}

// deriveTestStatus determines test status from manifest.
// Status is derived as:
// - "success" if manifest exists and AllPassed() returns true
// - "failure" if manifest is missing or has failures
func deriveTestStatus(cfg *config.EACConfig, moduleName string) string {
	// Get workspace root for absolute path
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		// Can't find workspace root - assume failure
		return "failure"
	}

	// Load test manifest
	moduleTestDir := filepath.Join(workspaceRoot, cfg.Repository.TestModuleDir(moduleName))
	manifest, err := implinternal.LoadTestManifest(moduleTestDir)
	if err != nil {
		// No manifest = failure
		return "failure"
	}

	// Check if all tests passed
	if manifest.AllPassed() {
		return "success"
	}

	return "failure"
}
