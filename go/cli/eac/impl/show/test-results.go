// Command: show test-results
// Short: Show test execution results from test manifests
// Long: Shows test execution results from test manifests with:
// Long:   - Module overview with pass/fail counts
// Long:   - Specification coverage for godog tests
// Long:   - Control tag summaries
// Long:   - Detailed test results table
// Long:
// Long: Example:
// Long:   show test-results
// Long:   show test-results ext-eac
// Long:   show test-results ext-eac r2r-cli
// Args: [module...]
package show

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/manifests/testview"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	sharedTemplate "github.com/ready-to-release/eac/go/clibase/template"
	eacConfig "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
)

func init() {
	registry.Register(ShowTestResults)
}

// TestResultsData holds data for the template with formatted time strings.
type TestResultsData struct {
	GeneratedAt   string
	LastRun       string
	ModulesTested int
	TotalTests    int
	TotalPassed   int
	TotalFailed   int

	ModuleStats    []testview.ModuleStats
	SpecCoverage   []testview.SpecCoverage
	ControlSummary []testview.ControlSummary
	Tests          []testview.TestResult
	SummaryByType  []testview.TypeSummary
	SummaryBySuite []testview.SuiteSummaryEntry
}

// buildTemplateData converts CompleteTestData to template-ready data with formatted times.
func buildTemplateData(data *testview.CompleteTestData) *TestResultsData {
	return &TestResultsData{
		GeneratedAt:    time.Now().Format(time.RFC3339),
		LastRun:        data.LastRun.Format(time.RFC3339),
		ModulesTested:  data.ModulesTested,
		TotalTests:     data.TotalTests,
		TotalPassed:    data.TotalPassed,
		TotalFailed:    data.TotalFailed,
		ModuleStats:    data.ModuleStats,
		SpecCoverage:   data.SpecCoverage,
		ControlSummary: data.ControlSummary,
		Tests:          data.Tests,
		SummaryByType:  data.SummaryByType,
		SummaryBySuite: data.SummaryBySuite,
	}
}

func ShowTestResults() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Check for help flag before doing any work
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			fmt.Println("Show test execution results from test manifests")
			fmt.Println("\nUsage: show test-results [module...]")
			fmt.Println("\nArguments:")
			fmt.Println("  module    Optional module moniker(s) to filter results")
			fmt.Println("\nFlags:")
			fmt.Println("  -h, --help   Show this help message")
			fmt.Println("\nOutput:")
			fmt.Println("  Module overview with pass/fail counts")
			fmt.Println("  Specification coverage for godog tests")
			fmt.Println("  Control tag summaries")
			fmt.Println("  Detailed test results table")
			fmt.Println("\nExamples:")
			fmt.Println("  show test-results")
			fmt.Println("  show test-results ext-eac")
			fmt.Println("  show test-results ext-eac r2r-cli")
			return 0
		}
	}

	// Parse module arguments
	args := os.Args[1:]
	var monikers []string

	// Skip command name and collect module arguments
	inCommand := false
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip until we find "test-results"
		if !inCommand {
			if arg == "test-results" {
				inCommand = true
			}
			continue
		}

		// Skip flags
		if strings.HasPrefix(arg, "--") {
			continue
		}

		// Collect module monikers
		monikers = append(monikers, arg)
	}

	// Get repository root
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get current directory: %v\n", err)
		return 1
	}

	repoRoot, err := testdata.FindRepoRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find repository root: %v\n", err)
		return 1
	}

	// Load test views from UoW manifests
	var views []*testview.TestModuleView
	if len(monikers) == 0 {
		views, err = testview.LoadAllTestViews(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load test data: %v\n", err)
			return 1
		}
	} else {
		views, err = testview.LoadTestViewsForModules(repoRoot, monikers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load test data: %v\n", err)
			return 1
		}
	}

	if len(views) == 0 {
		if len(monikers) == 0 {
			fmt.Fprintf(os.Stderr, "no test manifests found\n")
			fmt.Fprintf(os.Stderr, "Run tests first: test <module>\n")
		} else {
			fmt.Fprintf(os.Stderr, "no test manifests found for modules: %s\n", strings.Join(monikers, ", "))
			fmt.Fprintf(os.Stderr, "Run tests first: test %s\n", strings.Join(monikers, " "))
		}
		return 1
	}

	// Build complete test data from UoW-based views
	completeData := testview.BuildCompleteTestData(views)

	// Convert to template-ready format with formatted times
	data := buildTemplateData(completeData)

	// Load template with fallback (team override -> system default)
	templatePath, err := loadTestResultsTemplate(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load template: %v\n", err)
		return 1
	}

	// Render using template
	renderer := sharedTemplate.NewRenderer(templatePath).
		WithFuncs(testResultsTemplateFuncs())

	output, err := renderer.RenderToString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to render: %v\n", err)
		return 1
	}

	fmt.Println(output)
	return 0
}

// loadTestResultsTemplate loads the test results template with fallback logic.
// Priority 1: Team override (.eac/templates/reports/<category>/<template>)
// Priority 2: System default (templates/reports/<category>/<template>).
func loadTestResultsTemplate(workspaceRoot string) (string, error) {
	// Load EAC config for template directory paths and filenames
	// Skip workflow validation for test environments where workflow files may not exist
	// We only need template path configuration, not full workflow information
	cfg, err := eacConfig.Load(eacConfig.LoadOptions{
		RepoRoot:               workspaceRoot,
		SkipWorkflowValidation: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Get template configuration from contracts (no hardcoded values)
	category := cfg.Repository.Conventions.TestReportsCategory
	templateFilename := cfg.Repository.Conventions.TestResultsTemplate
	reportsDir := cfg.Repository.Conventions.TemplateReportsDir

	// Priority 1: Team override (.eac/templates/reports/<category>/<template>)
	teamOverridePath := filepath.Join(workspaceRoot, paths.R2RDir, paths.EACDir, paths.TemplatesDir, reportsDir, category, templateFilename)
	if _, err := os.Stat(teamOverridePath); err == nil {
		return teamOverridePath, nil
	}

	// Priority 2: System default (templates/reports/<category>/<template>)
	// In container: uses R2R_CONTAINER_ROOT (/app where Dockerfile copies templates)
	// In local dev: uses workspaceRoot (repo root where templates/ exists)
	distRoot := workspaceRoot
	if containerRoot := os.Getenv(environments.EnvR2RContainerRoot); containerRoot != "" {
		distRoot = containerRoot
	}
	systemDefaultPath := filepath.Join(distRoot, paths.TemplatesDir, reportsDir, category, templateFilename)
	if _, err := os.Stat(systemDefaultPath); err == nil {
		return systemDefaultPath, nil
	}

	return "", fmt.Errorf("test results template not found: %s (checked team override: %s and system default: %s)",
		templateFilename, teamOverridePath, systemDefaultPath)
}

// testResultsTemplateFuncs returns test-specific template functions
// Common functions like truncate, join, add, sub are already in defaultFuncMap.
func testResultsTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatStatus": func(passed, failed, skipped int) string {
			total := passed + failed + skipped
			if failed > 0 {
				return fmt.Sprintf("✗ %d/%d", passed, total)
			}
			if total == 0 {
				return "- 0/0"
			}
			return fmt.Sprintf("✓ %d/%d", passed, total)
		},
		"formatDurationSec": func(seconds float64) string {
			if seconds < 1.0 {
				return fmt.Sprintf("%.0fms", seconds*1000)
			}
			return fmt.Sprintf("%.1fs", seconds)
		},
		"statusIcon": func(status string) string {
			switch status {
			case testview.StatusPassed:
				return "✓"
			case testview.StatusFailed:
				return "✗"
			case testview.StatusSkipped:
				return "⊘"
			default:
				return "-"
			}
		},
	}
}
