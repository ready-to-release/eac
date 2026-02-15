package show

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	testview "github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/repository"
)

type showTestSummaryCommand struct{}

var _ core.SimpleCommandPort = (*showTestSummaryCommand)(nil)

func (c *showTestSummaryCommand) Name() string { return "show test-summary" }

func (c *showTestSummaryCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-test-summary",
		Short:         "Generate pretty test summary for a module",
		Long:          "The show test-summary command generates a formatted test summary with test results, metrics, and diagnostics.\nThis command is designed to be used in GitHub Actions workflows to create consistent, attractive test summaries.\nThe output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.\n\nWhen run without arguments, shows a summary for all modules that have test manifests.\nWhen run with just a module, shows summary for that module across all suites.\n\nExpected Output:\n- Markdown-formatted test summary with emojis and styling, suitable for GitHub Actions $GITHUB_STEP_SUMMARY\n- Success: includes status section, test metrics table (components, tests, passed/failed/skipped, duration), component breakdown\n- Failure: includes status section, diagnostics with last 100 lines of test log, timing data, and configuration",
		Flags: []core.FlagSpec{
			{Name: "status", Type: "string", Usage: "Test status override (success or failure)"},
			{Name: "run-id", Type: "string", Usage: "GitHub Actions run ID for linking to workflow"},
		},
	}
}

func (c *showTestSummaryCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowTestSummary()
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

	// Load test views from UoW manifests
	views, err := testview.LoadAllTestViews(workspaceRoot)
	if err != nil {
		log.Errorf("no test results found: %v", err)
		return 1
	}

	if len(views) == 0 {
		log.Errorf("no test manifests found")
		return 1
	}

	// Generate combined summary
	summary := generateCombinedSummary(views, cfg)

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

func (s *testStats) add(entry testview.TestEntry) {
	s.Total++
	switch entry.Status {
	case testview.StatusPassed:
		s.Passed++
	case testview.StatusFailed:
		s.Failed++
	case testview.StatusSkipped:
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
func generateCombinedSummary(views []*testview.TestModuleView, cfg *config.EACConfig) string {
	var summary string

	// Overall status
	totalPassed := 0
	totalFailed := 0
	totalSkipped := 0
	totalTests := 0
	allPassed := true

	for _, v := range views {
		totalPassed += v.Summary.Passed
		totalFailed += v.Summary.Failed
		totalSkipped += v.Summary.Skipped
		totalTests += v.Summary.Total
		if v.Summary.Failed > 0 {
			allPassed = false
		}
	}

	// Header
	if allPassed {
		summary += fmt.Sprintf("# %s Test Summary\n\n", Emoji("success"))
		summary += fmt.Sprintf("**Modules:** %d | **Tests:** %d | **Passed:** %d | **Failed:** %d | **Skipped:** %d\n\n",
			len(views), totalTests, totalPassed, totalFailed, totalSkipped)
	} else {
		summary += fmt.Sprintf("# %s Test Summary\n\n", Emoji("failure"))
		summary += fmt.Sprintf("**Modules:** %d | **Tests:** %d | **Passed:** %d | **Failed:** %d | **Skipped:** %d\n\n",
			len(views), totalTests, totalPassed, totalFailed, totalSkipped)
	}
	summary += "---\n\n"

	// One table per module
	for _, v := range views {
		moduleStatus := Emoji("success")
		if v.Summary.Failed > 0 {
			moduleStatus = Emoji("failure")
		}

		// Module section header
		summary += fmt.Sprintf("## %s %s\n\n", moduleStatus, v.Module)
		summary += fmt.Sprintf("**Tests:** %d | **Passed:** %d | **Failed:** %d | **Skipped:** %d\n\n",
			v.Summary.Total, v.Summary.Passed, v.Summary.Failed, v.Summary.Skipped)

		if len(v.Tests) == 0 {
			summary += "_No tests recorded_\n\n---\n\n"
			continue
		}

		// Separate tests into unit tests and feature tests
		var unitTests []testview.TestEntry
		featureTests := make(map[string][]testview.TestEntry) // feature name -> tests

		for _, t := range v.Tests {
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

			summary += buildTestTable(unitTests, v.Module, false)
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
					case testview.StatusPassed:
						passed++
					case testview.StatusFailed:
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

				summary += buildTestTable(tests, v.Module, true)
				summary += "\n\n"
			}
		}

		summary += "\n---\n\n"
	}

	return summary
}

// extractFeatureName extracts a human-readable feature name from package or file path.
func extractFeatureName(pkg, filePath string) string {
	// Try to extract from package path first (e.g., "specs/eac/work-create")
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

	// Load test view from UoW manifests
	view, err := testview.LoadModuleTestView(workspaceRoot, moduleName)
	if err != nil || view == nil {
		// No test data = failure
		return "failure"
	}

	// Check if all tests passed
	if view.AllPassed() {
		return "success"
	}

	return "failure"
}
