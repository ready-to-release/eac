package show

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	testview "github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
)

// buildTestTable creates a test table from a list of test entries
// isScenario controls whether the first column is "Test" or "Scenario".
func buildTestTable(tests []testview.TestEntry, moduleName string, isScenario bool) string {
	firstCol := "Test"
	if isScenario {
		firstCol = "Scenario"
	}
	table := render.NewTableBuilder().
		WithHeaders(firstCol, "L-Level", "Verification", "Status", "Other Tags")

	// Build self-referential depm tag to filter out
	selfDepmTag := "@depm:" + moduleName

	for _, t := range tests {
		// Extract L-level from tags
		lLevel := "-"
		verification := "-"
		var allTags []string

		for _, tag := range t.Tags {
			// Skip self-referential depm tags
			if tag == selfDepmTag {
				continue
			}

			// L-level tags - extract but don't add to other tags
			if len(tag) >= 2 && tag[0] == '@' && (tag[1] == 'L' || tag == "@HE2E") {
				if lLevel == "-" {
					lLevel = tag
				}
				continue
			}

			// Verification tags - extract but don't add to other tags
			switch tag {
			case "@ov":
				verification = "ov"
				continue
			case "@iv":
				verification = "iv"
				continue
			case "@pv":
				verification = "pv"
				continue
			case "@piv":
				verification = "piv"
				continue
			case "@ppv":
				verification = "ppv"
				continue
			}

			// Collect other tags (shortened format)
			shortTag := tag
			if len(tag) > 1 && tag[0] == '@' {
				shortTag = tag[1:] // Remove @ prefix for display
			}
			allTags = append(allTags, shortTag)
		}

		// Status with emoji
		status := ""
		switch t.Status {
		case testview.StatusPassed:
			status = Emoji("success")
		case testview.StatusFailed:
			status = Emoji("failure")
		case testview.StatusSkipped:
			status = "⏭️"
		default:
			status = t.Status
		}

		// Join all tags
		tagsStr := strings.Join(allTags, ", ")

		table.AddRow(t.Name, lLevel, verification, status, tagsStr)
	}

	return table.Build()
}

func testMetricsSection(f *SummaryFormatter, module *config.Module, suite string, cfg *config.EACConfig) string {
	outputDir := paths.TestModuleDir(cfg.RepoRoot, module.Moniker)

	// Try to read test summary JSON if available
	summaryFile := filepath.Join(outputDir, fmt.Sprintf("%s-summary.json", module.Moniker))
	if testResults := readTestSummaryJSON(summaryFile); testResults != nil {
		return formatTestResults(f, testResults)
	}

	// Fallback: just show that tests passed
	return f.Section(Emoji("success")+" Test Results", fmt.Sprintf("Tests completed successfully\n\nOutput: %s", Code(outputDir)))
}

// TestResults represents parsed test results.
type TestResults struct {
	Packages int                  `json:"packages"`
	Total    int                  `json:"total"`
	Passed   int                  `json:"passed"`
	Failed   int                  `json:"failed"`
	Skipped  int                  `json:"skipped"`
	Duration float64              `json:"duration"`
	Details  []PackageTestResults `json:"details,omitempty"`
}

// PackageTestResults represents test results for a single package.
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
		{"Components Tested", fmt.Sprintf("%d", results.Packages)},
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

	return f.Section(Emoji("chart")+" Component Breakdown", f.Table(headers, rows))
}

func testDiagnosticsSection(f *SummaryFormatter, module *config.Module, suite string, cfg *config.EACConfig) string {
	var diagnostics string

	// Read actual test log from the module's test output directory
	// Test logs are output to out/test/<module>/<subpackage>/test.log
	moduleDir := paths.TestModuleDir(cfg.RepoRoot, module.Moniker)

	// Try to find test.log files in the module directory and subdirectories
	logContent := findAndReadSubpackageLogs(moduleDir, 100)

	if logContent != "" {
		diagnostics += f.Section(Emoji("diagnostics")+" Test Log (last 100 lines)", f.CodeBlock("", logContent))
	} else {
		diagnostics += f.Section(Emoji("diagnostics")+" Diagnostics", fmt.Sprintf("Tests failed - no log file found in %s", moduleDir))
	}

	// Show test timing if available
	timingPath := paths.TestModuleTimingPath(cfg.RepoRoot, module.Moniker)
	if timing, err := os.ReadFile(timingPath); err == nil {
		diagnostics += f.Section(Emoji("time")+" Timing", string(timing))
	}

	return diagnostics
}

// findAndReadSubpackageLogs searches for test.log files in subdirectories and combines their content.
func findAndReadSubpackageLogs(moduleDir string, maxLines int) string {
	var logs []string

	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && info.Name() == "test.log" {
			content := readLogTail(path, 50) // Fewer lines per subpackage
			if content != "" {
				relPath, relErr := filepath.Rel(moduleDir, path)
				if relErr != nil {
					relPath = path // Fallback to absolute path
				}
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

	// Test framework - derive from enabled packages
	testFramework := "default"
	if module.Components.HasComponentType("go") {
		testFramework = "go test"
	} else if module.Components.HasComponentType("node") {
		testFramework = "mocha"
	}
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Framework"), testFramework)

	// Output directory
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Output"), Code(paths.TestModuleDir(cfg.RepoRoot, module.Moniker)))

	return f.CollapsibleSection(Emoji("config")+" Test Configuration", configDetails)
}
