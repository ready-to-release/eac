// Command: show test-summary
// Short: Generate pretty test summary for a module
// Long: The show test-summary command generates a formatted test summary with test results, metrics, and diagnostics.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent, attractive test summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: When run without arguments, shows a summary for all modules that have test manifests.
// Long: When run with just a module, shows summary for that module across all suites.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted test summary with emojis and styling, suitable for GitHub Actions $GITHUB_STEP_SUMMARY
// Long: - Success: includes status section, test metrics table (packages, tests, passed/failed/skipped, duration), package breakdown
// Long: - Failure: includes status section, diagnostics with last 100 lines of test log, timing data, and configuration
// Flag.status: type=string, usage=Test status override (success or failure)
// Flag.run-id: type=string, usage=GitHub Actions run ID for linking to workflow
package show

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowTestSummary)
}

// ShowTestSummary generates a pretty test summary.
func ShowTestSummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "test-summary"

	// Filter out flags to get positional args
	var positionalArgs []string
	for _, arg := range args {
		if arg != "" && arg[0] != '-' {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	runID := flags.GetFlagValue(args, "--run-id")

	// No args: show summary for all modules with test manifests
	if len(positionalArgs) == 0 {
		return generateAllModulesSummary(runID)
	}

	// One arg: show summary for that module (all suites)
	if len(positionalArgs) == 1 {
		return generateTestSummary(positionalArgs[0], "", runID)
	}

	// Two args: show summary for specific module and suite
	return generateTestSummary(positionalArgs[0], positionalArgs[1], runID)
}

// generateAllModulesSummary shows test summary for all modules with test manifests.
func generateAllModulesSummary(runID string) int {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Find all modules with test manifests
	testDir := filepath.Join(workspaceRoot, cfg.Repository.Paths.Out.Test)
	entries, err := os.ReadDir(testDir)
	if err != nil {
		log.Errorf("no test results found in %s", testDir)
		return 1
	}

	// Collect manifests
	var manifests []*implinternal.TestManifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		moduleName := entry.Name()
		moduleTestDir := filepath.Join(testDir, moduleName)
		manifest, err := implinternal.LoadTestManifest(moduleTestDir)
		if err != nil {
			continue // Skip modules without manifests
		}
		manifests = append(manifests, manifest)
	}

	if len(manifests) == 0 {
		log.Errorf("no test manifests found")
		return 1
	}

	// Sort by moniker for consistent output
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].Moniker < manifests[j].Moniker
	})

	// Generate combined summary
	summary := generateCombinedSummary(manifests, cfg)

	// Add footer with duration
	duration := time.Since(startTime)
	summary += fmt.Sprintf("\n---\n*Generated in %.2fs*\n", duration.Seconds())

	fmt.Print(summary)
	return 0
}

// testStats holds aggregated test statistics.
type testStats struct {
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Duration int64 // milliseconds
}

func (s *testStats) add(entry implinternal.TestEntry) {
	s.Total++
	switch entry.Status {
	case implinternal.TestStatusPassed:
		s.Passed++
	case implinternal.TestStatusFailed:
		s.Failed++
	case implinternal.TestStatusSkipped:
		s.Skipped++
	}
	s.Duration += entry.DurationMs
}

func (s *testStats) status() string {
	if s.Failed > 0 {
		return Emoji("failure")
	}
	return Emoji("success")
}

func (s *testStats) durationStr() string {
	if s.Duration == 0 {
		return "-"
	}
	secs := float64(s.Duration) / 1000.0
	if secs < 1 {
		return fmt.Sprintf("%dms", s.Duration)
	}
	return fmt.Sprintf("%.1fs", secs)
}

// generateCombinedSummary creates a summary with one table per module showing individual tests.
func generateCombinedSummary(manifests []*implinternal.TestManifest, cfg *config.EACConfig) string {
	var summary string

	// Overall status
	totalPassed := 0
	totalFailed := 0
	totalSkipped := 0
	totalTests := 0
	allPassed := true

	for _, m := range manifests {
		totalPassed += m.Summary.Passed
		totalFailed += m.Summary.Failed
		totalSkipped += m.Summary.Skipped
		totalTests += m.Summary.Total
		if m.Summary.Failed > 0 {
			allPassed = false
		}
	}

	// Header
	if allPassed {
		summary += fmt.Sprintf("# %s Test Summary\n\n", Emoji("success"))
		summary += fmt.Sprintf("**Modules:** %d | **Tests:** %d | **Passed:** %d | **Failed:** %d | **Skipped:** %d\n\n",
			len(manifests), totalTests, totalPassed, totalFailed, totalSkipped)
	} else {
		summary += fmt.Sprintf("# %s Test Summary\n\n", Emoji("failure"))
		summary += fmt.Sprintf("**Modules:** %d | **Tests:** %d | **Passed:** %d | **Failed:** %d | **Skipped:** %d\n\n",
			len(manifests), totalTests, totalPassed, totalFailed, totalSkipped)
	}
	summary += "---\n\n"

	// One table per module
	for _, m := range manifests {
		moduleStatus := Emoji("success")
		if m.Summary.Failed > 0 {
			moduleStatus = Emoji("failure")
		}

		// Module section header
		summary += fmt.Sprintf("## %s %s\n\n", moduleStatus, m.Moniker)
		summary += fmt.Sprintf("**Tests:** %d | **Passed:** %d | **Failed:** %d | **Skipped:** %d\n\n",
			m.Summary.Total, m.Summary.Passed, m.Summary.Failed, m.Summary.Skipped)

		if len(m.Tests) == 0 {
			summary += "_No tests recorded_\n\n---\n\n"
			continue
		}

		// Separate tests into unit tests and feature tests
		var unitTests []implinternal.TestEntry
		featureTests := make(map[string][]implinternal.TestEntry) // feature name -> tests

		for _, t := range m.Tests {
			if t.Type == "gotest" {
				unitTests = append(unitTests, t)
			} else {
				// Extract feature name from package or file path
				featureName := extractFeatureName(t.Package, t.FilePath)
				featureTests[featureName] = append(featureTests[featureName], t)
			}
		}

		// Unit Tests section
		if len(unitTests) > 0 {
			summary += "### Unit Tests\n\n"

			// Sort unit tests by name
			sort.Slice(unitTests, func(i, j int) bool {
				return unitTests[i].Name < unitTests[j].Name
			})

			summary += buildTestTable(unitTests, m.Moniker, false)
		}

		// Features section
		if len(featureTests) > 0 {
			summary += "### Features\n\n"

			// Sort feature names
			var featureNames []string
			for name := range featureTests {
				featureNames = append(featureNames, name)
			}
			sort.Strings(featureNames)

			for _, featureName := range featureNames {
				tests := featureTests[featureName]

				// Count passed/failed for this feature
				passed, failed := 0, 0
				for _, t := range tests {
					switch t.Status {
					case implinternal.TestStatusPassed:
						passed++
					case implinternal.TestStatusFailed:
						failed++
					}
				}

				featureStatus := Emoji("success")
				if failed > 0 {
					featureStatus = Emoji("failure")
				}

				summary += fmt.Sprintf("#### %s Feature: %s\n\n", featureStatus, featureName)
				summary += fmt.Sprintf("**Scenarios:** %d | **Passed:** %d | **Failed:** %d\n\n", len(tests), passed, failed)

				// Sort scenarios by name
				sort.Slice(tests, func(i, j int) bool {
					return tests[i].Name < tests[j].Name
				})

				summary += buildTestTable(tests, m.Moniker, true)
				summary += "\n\n"
			}
		}

		summary += "\n---\n\n"
	}

	return summary
}

// extractFeatureName extracts a human-readable feature name from package or file path.
func extractFeatureName(pkg, filePath string) string {
	// Try to extract from package path first (e.g., "specs/eac-commands/work-create")
	// The last path component is usually the feature name
	path := pkg
	if path == "" {
		path = filePath
	}

	// Normalize path separators
	path = strings.ReplaceAll(path, "\\", "/")

	// Remove file extension if present
	if idx := strings.LastIndex(path, "."); idx > 0 {
		lastSlash := strings.LastIndex(path, "/")
		if idx > lastSlash {
			path = path[:idx]
		}
	}

	// Get the last meaningful component
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		// Skip common non-descriptive names
		if part != "" && part != "specification" && part != "features" && part != "specs" {
			return part
		}
	}

	return "unknown"
}

// buildTestTable creates a test table from a list of test entries
// isScenario controls whether the first column is "Test" or "Scenario".
func buildTestTable(tests []implinternal.TestEntry, moduleName string, isScenario bool) string {
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
		case implinternal.TestStatusPassed:
			status = Emoji("success")
		case implinternal.TestStatusFailed:
			status = Emoji("failure")
		case implinternal.TestStatusSkipped:
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
	if suite != "" {
		summary += fmt.Sprintf("**Suite**: %s\n\n", Code(suite))
	} else {
		summary += "**Suite**: all\n\n"
	}

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

	// Read actual test log from the module's test output directory
	// Test logs are output to out/test/<module>/<subpackage>/test.log
	moduleDir := cfg.Repository.TestModuleDir(module.Moniker)

	// Try to find test.log files in the module directory and subdirectories
	logContent := findAndReadSubpackageLogs(moduleDir, 100)

	if logContent != "" {
		diagnostics += f.Section(Emoji("diagnostics")+" Test Log (last 100 lines)", f.CodeBlock("", logContent))
	} else {
		diagnostics += f.Section(Emoji("diagnostics")+" Diagnostics", fmt.Sprintf("Tests failed - no log file found in %s", moduleDir))
	}

	// Show test timing if available
	timingPath := cfg.Repository.TestModuleTimingPath(module.Moniker)
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
	if module.Components.HasComponent("go") {
		testFramework = "go test"
	} else if module.Components.HasComponent("node") {
		testFramework = "mocha"
	}
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Framework"), testFramework)

	// Output directory
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Output"), Code(cfg.Repository.TestModuleDir(module.Moniker)))

	return f.CollapsibleSection(Emoji("config")+" Test Configuration", configDetails)
}

// deriveTestStatus determines test status from manifest.
// Status is derived as:
// - "success" if manifest exists and AllPassed() returns true
// - "failure" if manifest is missing or has failures.
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
