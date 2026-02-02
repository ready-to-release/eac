// Command: show test-results
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

	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/manifests"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/clibase/flags"
	sharedTemplate "github.com/ready-to-release/eac/go/clibase/template"
	"github.com/ready-to-release/eac/go/clibase/registry"
	eacConfig "github.com/ready-to-release/eac/go/core/config"
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

	ModuleStats    []manifests.ModuleStats
	SpecCoverage   []manifests.SpecCoverage
	ControlSummary []manifests.ControlSummary
	Tests          []manifests.TestResult
	SummaryByType  []manifests.TypeSummary
	SummaryBySuite []manifests.SuiteSummary
}

// buildTemplateData converts CompleteTestData to template-ready data with formatted times.
func buildTemplateData(data *manifests.CompleteTestData) *TestResultsData {
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
		log.Errorf("%v", err)
		return 1
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
		log.Errorf("failed to get current directory: %v", err)
		return 1
	}

	repoRoot, err := testdata.FindRepoRoot(cwd)
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Load test manifests for specified modules (or all if no modules specified)
	var manifestList []*implinternal.TestManifest
	if len(monikers) == 0 {
		// No modules specified - scan all modules with test manifests
		manifestList, err = manifests.LoadAllTestManifests(repoRoot)
		if err != nil {
			log.Errorf("failed to load manifests: %v", err)
			return 1
		}
	} else {
		// Specific modules requested
		manifestList, err = manifests.LoadTestManifestsForModules(repoRoot, monikers)
		if err != nil {
			log.Errorf("failed to load manifests: %v", err)
			return 1
		}
	}

	if len(manifestList) == 0 {
		if len(monikers) == 0 {
			log.Errorf("no test manifests found")
			log.Errorf("Run tests first: test <module>")
		} else {
			log.Errorf("no test manifests found for modules: %s", strings.Join(monikers, ", "))
			log.Errorf("Run tests first: test %s", strings.Join(monikers, " "))
		}
		return 1
	}

	// Use shared function to build complete test data
	completeData := manifests.BuildCompleteTestData(manifestList)

	// Convert to template-ready format with formatted times
	data := buildTemplateData(completeData)

	// Load template with fallback (team override -> system default)
	templatePath, err := loadTestResultsTemplate(repoRoot)
	if err != nil {
		log.Errorf("failed to load template: %v", err)
		return 1
	}

	// Render using template
	renderer := sharedTemplate.NewRenderer(templatePath).
		WithFuncs(testResultsTemplateFuncs())

	output, err := renderer.RenderToString(data)
	if err != nil {
		log.Errorf("failed to render: %v", err)
		return 1
	}

	fmt.Println(output)
	return 0
}

// loadTestResultsTemplate loads the test results template with fallback logic.
// Priority 1: Team override (.r2r/eac/templates/reports/<category>/<template>)
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

	// Priority 1: Team override (.r2r/eac/templates/reports/<category>/<template>)
	teamOverridePath := filepath.Join(workspaceRoot, paths.R2RDir, paths.EACDir, paths.TemplatesDir, reportsDir, category, templateFilename)
	if _, err := os.Stat(teamOverridePath); err == nil {
		return teamOverridePath, nil
	}

	// Priority 2: System default (templates/reports/<category>/<template>)
	// In container: uses R2R_CONTAINER_ROOT (/app where Dockerfile copies templates)
	// In local dev: uses workspaceRoot (repo root where templates/ exists)
	distRoot := workspaceRoot
	if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
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
			case manifests.StatusPassed:
				return "✓"
			case manifests.StatusFailed:
				return "✗"
			case "skipped":
				return "⊘"
			default:
				return "-"
			}
		},
	}
}
