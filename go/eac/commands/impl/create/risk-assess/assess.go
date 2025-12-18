// Command: create risk-assess
// Short: Create OSCAL assessment-results from existing test and security evidence
// Long: The create risk-assess command creates OSCAL assessment-results for modules
// Long: by reading existing test results and security scan evidence. It maps @control tags in
// Long: feature files to OSCAL control IDs and determines satisfied/not-satisfied status.
// Long:
// Long: This command does NOT run tests or scans. It only reads existing evidence.
// Long: The command will warn (but continue) if:
// Long: - Evidence is missing
// Long: - Evidence is older than max-evidence-age (default: 24h)
// Long:
// Long: Evidence is collected from:
// Long: - Test results: out/test/<module>/*.json
// Long: - Security scans: out/scan/<module>/**/*.json
// Long:
// Long: Expected Output:
// Long: - OSCAL assessment-results JSON file
// Long: - Control status (satisfied/not-satisfied) based on test results
// Long: - Risk assessment reports in Markdown format
// Flag.profile: type=string, shorthand=p, required=true, usage=Path to OSCAL profile JSON file
// Flag.max-evidence-age: type=string, default=24h, usage=Maximum age for evidence before warning (e.g., 24h, 7d)
// Flag.suites: type=[]string, default=all, usage=Test suites to check for evidence (e.g., all, integration, acceptance)
// Flag.report-template: type=string, default=risk-assessment-detailed, usage=Report template variant (risk-assessment, risk-assessment-summary, risk-assessment-detailed)
// Flag.sequential: type=bool, default=false, usage=Run assessments sequentially instead of parallel
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs to out/logs/risk/
// Args: modules
package riskassess

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"
	sharedTemplate "github.com/ready-to-release/eac/go/eac/commands/internal/template"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	eacConfig "github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

var assessLog = logging.C()

func init() {
	registry.Register(CreateRiskAssess)
}

// AssessConfig holds configuration for risk assess command.
type AssessConfig struct {
	Modules        []string // Module names to assess
	ProfilePath    string
	MaxEvidenceAge time.Duration
	Sequential     bool // Disable parallel execution
	Debug          bool
	WorkspaceRoot  string
	Logger         *logging.Logger
	Timestamp      string // Timestamp for this assessment run (format: 2006-01-02T15-04-05)
	OutputDir      string // Base output directory for this run: out/risk/<timestamp>/
	ReportTemplate string // Report template variant (default: risk-assessment-detailed)
}

// ModuleAssessmentResult holds the results of assessing a single module.
type ModuleAssessmentResult struct {
	Module            string
	AssessmentResults *oscalTypes.AssessmentResults
	Evidence          *evidence.EvidenceCollection // Evidence collection for reporting
	OutputPath        string
	Satisfied         int
	NotSatisfied      int
	RiskScore         *scoring.RiskScore
	Error             error
	Warnings          []string // Evidence collection warnings
}

// CreateRiskAssess is the entry point for the create risk-assess command.
func CreateRiskAssess() int {
	config, err := parseAssessConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showAssessHelp()
			return 0
		}
		assessLog.Errorf("Error: %v", err)
		return 1
	}

	// Initialize timestamp and output directory for this assessment run
	config.Timestamp = time.Now().Format("2006-01-02T15-04-05")
	config.OutputDir = filepath.Join(paths.RiskOutputPath(config.WorkspaceRoot, config.Timestamp))

	// Ensure output directory exists
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		assessLog.Errorf("Failed to create output directory: %v", err)
		return 1
	}

	// Initialize logger
	var logger *logging.Logger
	if config.Debug {
		logger, err = logging.NewWithDebug("risk", config.WorkspaceRoot)
	} else {
		logger, err = logging.NewDefault("risk", config.WorkspaceRoot)
	}
	if err != nil {
		assessLog.Errorf("Warning: Failed to initialize logger: %v", err)
	} else {
		config.Logger = logger
		defer logger.Sync()
	}

	if config.Logger != nil {
		config.Logger.Info("Starting risk assess",
			zap.Strings("modules", config.Modules),
			zap.String("profile", config.ProfilePath),
			zap.Duration("max_evidence_age", config.MaxEvidenceAge),
			zap.Bool("sequential", config.Sequential),
			zap.Bool("debug", config.Debug))
	}

	// Load profile from specified path
	profile, err := oscal.LoadProfile(config.ProfilePath)
	if err != nil {
		assessLog.Errorf("Error: Failed to load profile: %s", config.ProfilePath)
		assessLog.Errorf("Error details: %v", err)
		assessLog.Error("")
		assessLog.Error("Create a profile first:")
		assessLog.Error("  create risk-profile <assessment.md>")
		return 1
	}

	controlIDs := oscal.GetProfileControlIDs(profile)
	assessLog.Infof("Loaded profile with %d controls: %s", len(controlIDs), strings.Join(controlIDs, ", "))
	assessLog.Info("")

	// Run assessments (parallel or sequential)
	var results []*ModuleAssessmentResult

	if config.Sequential || len(config.Modules) == 1 {
		// Sequential: single module or --sequential flag
		if len(config.Modules) == 1 {
			assessLog.Infof("Assessing module: %s", config.Modules[0])
		} else {
			assessLog.Infof("Assessing %d modules sequentially...", len(config.Modules))
		}
		results = assessModulesSequential(config, profile)
	} else {
		// Parallel: multiple modules (default)
		assessLog.Infof("Assessing %d modules in parallel...", len(config.Modules))
		results = assessModulesParallel(config, profile)
	}

	// Check for errors and separate successful results
	var successfulResults []*ModuleAssessmentResult
	var failedModules []string

	for _, result := range results {
		if result.Error != nil {
			failedModules = append(failedModules, result.Module)
			assessLog.Errorf("Module %s failed: %v", result.Module, result.Error)
		} else {
			successfulResults = append(successfulResults, result)
			// Display warnings for this module if any
			if len(result.Warnings) > 0 {
				for _, warning := range result.Warnings {
					assessLog.Warnf("[%s] %s", result.Module, warning)
				}
			}
		}
	}

	// Create aggregated report file
	if len(successfulResults) > 0 {
		if err := writeAggregatedReport(config, successfulResults, profile); err != nil {
			assessLog.Errorf("⚠️  Failed to write aggregated report: %v", err)
			assessLog.Error("Individual module reports are still available in module subdirectories")
		}
	}

	// Report aggregate results
	if len(successfulResults) > 0 {
		reportAggregateResults(config, successfulResults)
	}

	// Handle partial or total failure
	if len(failedModules) > 0 {
		assessLog.Error("")
		assessLog.Errorf("⚠️  %d module(s) failed: %s", len(failedModules), strings.Join(failedModules, ", "))

		if len(successfulResults) == 0 {
			assessLog.Error("All assessments failed")
			return 1
		}

		assessLog.Infof("✓ %d module(s) completed successfully", len(successfulResults))
		// Return success if at least some modules passed
	}

	if config.Logger != nil {
		config.Logger.Info("Assessment completed",
			zap.Int("successful", len(successfulResults)),
			zap.Int("failed", len(failedModules)))
	}

	return 0
}

// loadModuleRegistry loads the module registry from the workspace.
func loadModuleRegistry(workspaceRoot string) (*modules.Registry, error) {
	return modules.LoadFromWorkspace(workspaceRoot)
}

// parseAssessConfig parses command line configuration.
func parseAssessConfig() (*AssessConfig, error) {
	args := os.Args[3:] // Skip program name, "create", and "risk-assess"

	config := &AssessConfig{
		MaxEvidenceAge: 24 * time.Hour,
		ReportTemplate: "risk-assessment-detailed", // Default to detailed template
	}

	// Parse debug flag using shared package
	config.Debug = flags.ParseDebugFlag(args)
	config.Sequential = flags.HasFlag(args, "--sequential", "")

	// Get workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace root: %w", err)
	}

	config.WorkspaceRoot = workspaceRoot

	// Collect positional arguments (module names) before flags
	seen := make(map[string]bool)
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Stop at first flag
		if strings.HasPrefix(arg, "-") {
			break
		}

		// Check for duplicates
		if seen[arg] {
			return nil, fmt.Errorf("duplicate module specified: %s", arg)
		}
		seen[arg] = true

		// Collect module name
		config.Modules = append(config.Modules, arg)
	}

	// Parse flags starting from where modules ended
	i := len(config.Modules)
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "--profile" || arg == "-p":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--profile requires a value")
			}
			config.ProfilePath = args[i+1]
			// Make path absolute if relative
			if !filepath.IsAbs(config.ProfilePath) {
				config.ProfilePath = filepath.Join(config.WorkspaceRoot, config.ProfilePath)
			}
			i += 2

		case arg == "--max-evidence-age":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--max-evidence-age requires a value")
			}
			duration, err := time.ParseDuration(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration: %s", args[i+1])
			}

			// Validate duration is positive and reasonable
			if duration <= 0 {
				return nil, fmt.Errorf("--max-evidence-age must be positive")
			}

			const maxDuration = 30 * 24 * time.Hour // 30 days
			if duration > maxDuration {
				return nil, fmt.Errorf("--max-evidence-age too large (max: 30d)")
			}

			config.MaxEvidenceAge = duration
			i += 2

		case arg == "--report-template":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--report-template requires a value")
			}
			template := args[i+1]
			// Validate template variant
			validTemplates := []string{"risk-assessment", "risk-assessment-summary", "risk-assessment-detailed"}
			valid := false
			for _, vt := range validTemplates {
				if template == vt {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("invalid template variant: %s (must be one of: %s)", template, strings.Join(validTemplates, ", "))
			}
			config.ReportTemplate = template
			i += 2

		case arg == "--debug" || arg == "-d" || arg == "--sequential":
			// Already handled by shared flags package
			i++

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			return nil, fmt.Errorf("unexpected positional argument after flags: %s", arg)
		}
	}

	// Load registry to validate modules or discover all
	registry, err := loadModuleRegistry(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load modules: %w", err)
	}

	// If no modules specified, discover all modules
	if len(config.Modules) == 0 {
		allModules := registry.All()
		for _, mod := range allModules {
			config.Modules = append(config.Modules, mod.Moniker)
		}

		if len(config.Modules) == 0 {
			return nil, fmt.Errorf("no modules found in workspace")
		}
	} else {
		// Validate that all specified modules exist
		allModules := make(map[string]bool)
		for _, mod := range registry.All() {
			allModules[mod.Moniker] = true
		}

		var invalidModules []string
		for _, moduleName := range config.Modules {
			if !allModules[moduleName] {
				invalidModules = append(invalidModules, moduleName)
			}
		}

		if len(invalidModules) > 0 {
			availableModules := make([]string, 0, len(allModules))
			for mod := range allModules {
				availableModules = append(availableModules, mod)
			}
			return nil, fmt.Errorf(`unknown module(s): %s

Available modules:
  %s

Try:
  - Check module name spelling
  - List all modules: show modules
  - View module contracts: cat .r2r/eac/repository.yml`,
				strings.Join(invalidModules, ", "),
				strings.Join(availableModules, ", "))
		}
	}

	// Validate required flags
	if config.ProfilePath == "" {
		return nil, fmt.Errorf("--profile flag is required")
	}

	return config, nil
}

// showAssessHelp displays help information.
func showAssessHelp() {
	help := `Usage: create risk-assess [module...] --profile <path> [flags]

Create OSCAL assessment-results from existing test and security evidence

NOTE: This command does NOT run tests or scans. It only reads existing evidence.
      The command will warn (but continue) if evidence is missing or too old.

Arguments:
  module...            Module name(s) to assess (space-separated)
                       If omitted, all modules are assessed

Required Flags:
  -p, --profile <path>               Path to OSCAL profile JSON file

Optional Flags:
      --suites <name...>             Test suites to check for evidence (default: all)
                                     Can specify multiple: --suites integration acceptance
                                     Options: all, integration, acceptance, commit, production-verification
      --max-evidence-age <duration>  Maximum age for evidence before warning (default: 24h)
                                     Evidence older than this will be flagged in the report
      --sequential                   Run assessments sequentially (default: parallel)
  -d, --debug                        Save intermediate outputs to out/logs/risk/

Output:
  out/risk/<timestamp>/
    ├── risk-assessment-<timestamp>-all.md         # Aggregated Markdown report (all modules)
    ├── risk-assessment-<timestamp>-subset-NofM.md # Aggregated Markdown report (subset)
    ├── <module>/assessment-results.json           # Per-module OSCAL JSON
    └── ...

Examples:
  # Assess all modules with all test suites (default)
  create risk-assess --profile specs/.risk-controls/risk-profile.json

  # Assess specific modules (parallel, space-separated)
  create risk-assess billing api auth --profile specs/.risk-controls/risk-profile.json

  # Check specific test suites
  create risk-assess --profile specs/.risk-controls/risk-profile.json --suites integration acceptance

  # Use single test suite
  create risk-assess --profile specs/.risk-controls/risk-profile.json --suites unit

  # Assess single module with production tests
  create risk-assess billing --profile specs/.risk-controls/risk-profile.json --suites production-verification

  # Sequential execution for debugging
  create risk-assess --profile specs/.risk-controls/risk-profile.json --sequential

  # Set custom evidence freshness threshold
  create risk-assess data --profile specs/.risk-controls/risk-profile.json --max-evidence-age 1h
`
	assessLog.Info(help)
}

// writeAggregatedReport creates a consolidated Markdown report combining all module results.
func writeAggregatedReport(config *AssessConfig, results []*ModuleAssessmentResult, profile *oscalTypes.Profile) error {
	// Determine if this is all modules or a subset
	allModules, err := loadModuleRegistry(config.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	totalModuleCount := len(allModules.All())
	assessedModuleCount := len(results)
	isSubset := assessedModuleCount < totalModuleCount

	// Build filename with timestamp and subset indication
	var filename string
	if isSubset {
		filename = fmt.Sprintf("risk-assessment-%s-subset-%dof%d.md",
			config.Timestamp, assessedModuleCount, totalModuleCount)
	} else {
		filename = fmt.Sprintf("risk-assessment-%s-all.md", config.Timestamp)
	}

	outputPath := filepath.Join(config.OutputDir, filename)

	// Generate Markdown report using template
	report, err := generateMarkdownReport(config, results, profile, isSubset, totalModuleCount)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(report), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	assessLog.Infof("Created aggregated report: %s", outputPath)
	return nil
}

// generateMarkdownReport creates the Markdown content for the aggregated report using templates.
func generateMarkdownReport(config *AssessConfig, results []*ModuleAssessmentResult, profile *oscalTypes.Profile, isSubset bool, totalModules int) (string, error) {
	// Build template data
	reportData := buildReportData(config, results, profile, isSubset, totalModules)

	// Load EAC config for template path
	cfg, err := eacConfig.Load(eacConfig.LoadOptions{RepoRoot: config.WorkspaceRoot})
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Use selected template variant
	templateName := config.ReportTemplate + ".md"
	templatePath := filepath.Join(config.WorkspaceRoot, cfg.Repository.Paths.Templates, "reports", "risk", templateName)
	renderer := sharedTemplate.NewRenderer(templatePath)

	return renderer.RenderToString(reportData)
}

// writeAggregatedOSCALReport creates a consolidated OSCAL JSON file (kept for compatibility).
func writeAggregatedOSCALReport(config *AssessConfig, results []*ModuleAssessmentResult, profile *oscalTypes.Profile) error {
	// Create aggregate assessment-results document
	aggregateAR := oscal.NewAssessmentResults(
		uuid.New().String(),
		"Risk Assessment - All Modules",
		config.ProfilePath,
	)

	// Collect all control IDs from profile
	controlIDs := oscal.GetProfileControlIDs(profile)

	// Build a combined result with all observations and findings
	props := []oscalTypes.Property{
		{Name: "assessment-type", Value: "automated"},
		{Name: "modules-assessed", Value: fmt.Sprintf("%d", len(results))},
	}

	controlRefs := make([]oscalTypes.AssessedControlsSelectControlById, len(controlIDs))
	for i, id := range controlIDs {
		controlRefs[i] = oscalTypes.AssessedControlsSelectControlById{
			ControlId: id,
		}
	}

	controlSelections := []oscalTypes.AssessedControls{
		{
			IncludeControls: &controlRefs,
		},
	}

	aggregateResult := oscalTypes.Result{
		UUID:        uuid.New().String(),
		Title:       "Aggregate Assessment Results",
		Description: fmt.Sprintf("Combined assessment results from %d modules", len(results)),
		Start:       time.Now().UTC(),
		Props:       &props,
		ReviewedControls: oscalTypes.ReviewedControls{
			ControlSelections: controlSelections,
		},
	}

	// Collect all observations from all modules
	var allObservations []oscalTypes.Observation
	for _, result := range results {
		if result.AssessmentResults != nil && len(result.AssessmentResults.Results) > 0 {
			moduleResult := result.AssessmentResults.Results[0]
			if moduleResult.Observations != nil {
				for _, obs := range *moduleResult.Observations {
					// Add module name to observation description
					obs.Description = fmt.Sprintf("[%s] %s", result.Module, obs.Description)
					allObservations = append(allObservations, obs)
				}
			}
		}
	}
	aggregateResult.Observations = &allObservations

	// Aggregate findings by control ID across all modules
	controlFindings := make(map[string]*oscalTypes.Finding)

	for _, result := range results {
		if result.AssessmentResults != nil && len(result.AssessmentResults.Results) > 0 {
			moduleResult := result.AssessmentResults.Results[0]
			if moduleResult.Findings != nil {
				for _, finding := range *moduleResult.Findings {
					controlID := finding.Target.TargetId

					// If we don't have a finding for this control yet, create one
					if controlFindings[controlID] == nil {
						controlFindings[controlID] = &oscalTypes.Finding{
							UUID:                uuid.New().String(),
							Title:               fmt.Sprintf("Control %s Assessment", strings.ToUpper(controlID)),
							Description:         fmt.Sprintf("Aggregate finding for control %s across %d modules", controlID, len(results)),
							Target:              finding.Target,
							RelatedObservations: &[]oscalTypes.RelatedObservation{},
							Props:               &[]oscalTypes.Property{},
							Remarks:             "",
						}
					}

					// Aggregate the status - if any module has not-satisfied, the control is not-satisfied
					if finding.Target.Status.State == oscal.StateNotSatisfied {
						controlFindings[controlID].Target.Status.State = oscal.StateNotSatisfied
					}

					// Add module-specific remarks
					if finding.Remarks != "" {
						if controlFindings[controlID].Remarks != "" {
							controlFindings[controlID].Remarks += "\n\n"
						}
						controlFindings[controlID].Remarks += fmt.Sprintf("=== %s ===\n%s", result.Module, finding.Remarks)
					}

					// Collect observation references
					if finding.RelatedObservations != nil {
						for _, relObs := range *finding.RelatedObservations {
							*controlFindings[controlID].RelatedObservations = append(
								*controlFindings[controlID].RelatedObservations,
								relObs,
							)
						}
					}
				}
			}
		}
	}

	// Convert map to slice
	var aggregatedFindings []oscalTypes.Finding
	for _, finding := range controlFindings {
		aggregatedFindings = append(aggregatedFindings, *finding)
	}
	aggregateResult.Findings = &aggregatedFindings

	// Add the aggregate result to the assessment-results
	oscal.AddResult(aggregateAR, aggregateResult)

	// Write to file
	outputPath := filepath.Join(paths.RiskOutputPath(config.WorkspaceRoot, ""), "assessment-results-aggregate.json")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := oscal.WriteAssessmentResults(outputPath, aggregateAR); err != nil {
		return fmt.Errorf("failed to write aggregated report: %w", err)
	}

	assessLog.Infof("Created aggregated report: %s", outputPath)
	return nil
}
