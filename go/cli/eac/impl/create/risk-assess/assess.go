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
// Flag.sequential: type=bool, default=false, usage=Run assessments sequentially instead of parallel
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs to out/commands.log
// Args: modules
package riskassess

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/scoring"
	sharedTemplate "github.com/ready-to-release/eac/go/clibase/template"
	"github.com/ready-to-release/eac/go/clibase/registry"
	eacConfig "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
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
	OutputDir      string // Base output directory: out/risk/
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
		assessLog.Errorf("Error: %v", err)
		return 1
	}

	// Initialize output directory with timestamp for uniqueness
	timestamp := time.Now().UTC().Format("20060102-150405")
	baseRiskDir := paths.RiskOutputPath(config.WorkspaceRoot, "")
	config.OutputDir = filepath.Join(baseRiskDir, timestamp)

	// Create timestamped output directory
	if err := os.MkdirAll(config.OutputDir, 0o755); err != nil {
		assessLog.Errorf("Failed to create output directory: %v", err)
		return 1
	}

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(config.WorkspaceRoot, "commands", nil, config.Debug); err != nil {
		assessLog.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	assessLog.Info("Starting risk assess")
	assessLog.Debugf("Modules: modules=%v", config.Modules)
	assessLog.Debugf("Profile: %s", config.ProfilePath)
	assessLog.Debugf("Max evidence age: %v", config.MaxEvidenceAge)
	assessLog.Debugf("Sequential: sequential=%v", config.Sequential)
	assessLog.Debugf("Debug: debug=%v", config.Debug)

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

	// Generate AI-powered risk assessment (always generate executive summary)
	var aiOutput *AIRiskAssessmentOutput
	assessLog.Info("")
	assessLog.Info("🤖 Generating AI-powered risk assessment...")

	aiInput := buildAIRiskAssessmentInput(config, successfulResults, profile)

	aiResult, err := GenerateRiskAssessment(context.Background(), config, aiInput)
	if err != nil {
		assessLog.Warnf("⚠️  AI risk assessment failed: %v", err)
		if len(successfulResults) > 0 {
			assessLog.Info("Falling back to basic risk scoring...")
			// Apply basic fallback scoring for successful results
			applyBasicRiskScoring(successfulResults)
		}
		aiOutput = nil // No AI output available
	} else {
		assessLog.Info("✅ AI risk assessment completed")
		if len(successfulResults) > 0 {
			// Apply AI-generated risk scores to modules (if any)
			ApplyAIRiskAssessment(successfulResults, aiResult, assessLog)
		}
		aiOutput = aiResult // Store for report generation
	}

	// Create aggregated report file
	if len(successfulResults) > 0 {
		if err := writeAggregatedReport(config, successfulResults, profile, aiOutput); err != nil {
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
	} else if len(successfulResults) > 0 {
		// All modules succeeded
		assessLog.Info("")
		assessLog.Infof("✓ %d module(s) completed successfully", len(successfulResults))
	}

	assessLog.Info("Assessment completed")
	assessLog.Debugf("Assessment completed: successful=%d, failed=%d", len(successfulResults), len(failedModules))

	return 0
}

// loadModuleRegistry loads the module registry from the workspace.
func loadModuleRegistry(workspaceRoot string) (*modules.Registry, error) {
	return modules.LoadFromWorkspace(workspaceRoot)
}

// parseAssessConfig parses command line configuration.
func parseAssessConfig() (*AssessConfig, error) {
	args := os.Args[3:] // Skip program name, "create", and "risk-assess"

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return nil, err
	}

	config := &AssessConfig{
		MaxEvidenceAge: 24 * time.Hour,
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
  - View module contracts: cat .eac/repository.yml`,
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

// writeAggregatedReport creates a consolidated Markdown report combining all module results.
func writeAggregatedReport(config *AssessConfig, results []*ModuleAssessmentResult, profile *oscalTypes.Profile, aiOutput *AIRiskAssessmentOutput) error {
	// Determine if this is all modules or a subset
	allModules, err := loadModuleRegistry(config.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	totalModuleCount := len(allModules.All())
	assessedModuleCount := len(results)
	isSubset := assessedModuleCount < totalModuleCount

	// Build filename with subset indication
	var filename string
	if isSubset {
		filename = fmt.Sprintf("risk-assessment-subset-%dof%d.md", assessedModuleCount, totalModuleCount)
	} else {
		filename = "risk-assessment-all.md"
	}

	outputPath := filepath.Join(config.OutputDir, filename)

	// Generate Markdown report using template
	report, err := generateMarkdownReport(config, results, profile, isSubset, totalModuleCount, aiOutput)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(report), 0o644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	assessLog.Infof("Created aggregated report: %s", outputPath)
	return nil
}

// generateMarkdownReport creates the Markdown content for the aggregated report using templates.
func generateMarkdownReport(config *AssessConfig, results []*ModuleAssessmentResult, profile *oscalTypes.Profile, isSubset bool, totalModules int, aiOutput *AIRiskAssessmentOutput) (string, error) {
	// Build template data
	reportData := buildReportData(config, results, profile, isSubset, totalModules, aiOutput)

	// Load template with fallback (team override -> system default)
	templatePath, err := loadRiskAssessmentTemplate(config.WorkspaceRoot)
	if err != nil {
		return "", err
	}

	renderer := sharedTemplate.NewRenderer(templatePath)
	return renderer.RenderToString(reportData)
}

// loadRiskAssessmentTemplate loads the risk assessment template with fallback logic.
// Priority 1: Team override (.eac/templates/reports/risk/risk-assess.md)
// Priority 2: System default (templates/reports/risk/risk-assess.md)
// This follows the same pattern as AI prompt loading (see contracts/ai_loader.go:LoadPrompt).
func loadRiskAssessmentTemplate(workspaceRoot string) (string, error) {
	// Load EAC config for template directory paths and filenames
	cfg, err := eacConfig.Load(eacConfig.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Get template configuration from contracts (no hardcoded values)
	category := cfg.Repository.Conventions.RiskReportsCategory
	templateFilename := cfg.Repository.Conventions.RiskAssessmentTemplate
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

	return "", fmt.Errorf("risk assessment template not found: %s (checked team override: %s and system default: %s)",
		templateFilename, teamOverridePath, systemDefaultPath)
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
				for i := range *moduleResult.Observations {
					obs := &(*moduleResult.Observations)[i]
					// Add module name to observation description
					obs.Description = fmt.Sprintf("[%s] %s", result.Module, obs.Description)
					allObservations = append(allObservations, *obs)
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
				for i := range *moduleResult.Findings {
					finding := &(*moduleResult.Findings)[i]
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
						*controlFindings[controlID].RelatedObservations = append(*controlFindings[controlID].RelatedObservations, *finding.RelatedObservations...)
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
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := oscal.WriteAssessmentResults(outputPath, aggregateAR); err != nil {
		return fmt.Errorf("failed to write aggregated report: %w", err)
	}

	assessLog.Infof("Created aggregated report: %s", outputPath)
	return nil
}

// buildAIRiskAssessmentInput prepares input data for AI risk analysis.
func buildAIRiskAssessmentInput(
	config *AssessConfig,
	results []*ModuleAssessmentResult,
	profile *oscalTypes.Profile,
) *AIRiskAssessmentInput {
	// Load module registry for context
	registry, err := loadModuleRegistry(config.WorkspaceRoot)
	if err != nil {
		assessLog.Warnf("Failed to load module registry: %v", err)
		registry = nil
	}

	// Count total controls from profile
	totalControls := 0
	satisfiedControls := 0
	notSatisfiedControls := 0

	// Build per-module inputs
	var moduleInputs []ModuleAnalysisInput
	for _, result := range results {
		// Extract vulnerability findings from evidence
		vulnFindings := scoring.ExtractVulnerabilityFindings(result.Evidence)

		// Extract satisfied control IDs
		satisfiedControlIDs := extractSatisfiedControlIDs(result)

		// Build module context
		moduleContext := scoring.BuildModuleContext(result.Module, registry, satisfiedControlIDs)

		// Determine impact based on module criticality
		impact := 3 // Default medium
		switch moduleContext.Criticality {
		case "high":
			impact = 4
		case "low":
			impact = 2
		}

		moduleInputs = append(moduleInputs, ModuleAnalysisInput{
			Module:                result.Module,
			VulnerabilityFindings: vulnFindings,
			Context:               moduleContext,
			ControlsSatisfied:     result.Satisfied,
			ControlsNotSatisfied:  result.NotSatisfied,
			Impact:                impact,
		})

		// Aggregate control counts
		satisfiedControls += result.Satisfied
		notSatisfiedControls += result.NotSatisfied
	}

	// Calculate total unique controls from profile
	if profile != nil && profile.Imports != nil {
		for _, imp := range profile.Imports {
			if imp.IncludeControls != nil {
				for _, includeControl := range *imp.IncludeControls {
					if includeControl.WithIds != nil {
						totalControls += len(*includeControl.WithIds)
					}
				}
			}
		}
	}

	// Get profile name from path
	profileName := filepath.Base(config.ProfilePath)

	return &AIRiskAssessmentInput{
		Modules:              moduleInputs,
		ProfileName:          profileName,
		TotalControls:        totalControls,
		SatisfiedControls:    satisfiedControls,
		NotSatisfiedControls: notSatisfiedControls,
	}
}

// extractSatisfiedControlIDs extracts control IDs from findings that are satisfied.
func extractSatisfiedControlIDs(result *ModuleAssessmentResult) []string {
	var controlIDs []string

	if result.AssessmentResults == nil || result.AssessmentResults.Results == nil {
		return controlIDs
	}

	for i := range result.AssessmentResults.Results {
		assessmentResult := &result.AssessmentResults.Results[i]
		if assessmentResult.Findings != nil {
			for j := range *assessmentResult.Findings {
				finding := &(*assessmentResult.Findings)[j]
				// Check if finding is satisfied
				if finding.Target.Status.State == oscal.StateSatisfied {
					controlID := finding.Target.TargetId
					if controlID != "" {
						controlIDs = append(controlIDs, controlID)
					}
				}
			}
		}
	}

	return controlIDs
}

// applyBasicRiskScoring applies fallback basic scoring when AI generation fails.
func applyBasicRiskScoring(results []*ModuleAssessmentResult) {
	for _, result := range results {
		// Use fixed likelihood of 3 (moderate) as fallback
		likelihood := 3
		impact := 3 // Default medium impact

		// Simple reasoning for fallback
		reasoning := fmt.Sprintf("Basic scoring applied: %d controls satisfied, %d not satisfied",
			result.Satisfied, result.NotSatisfied)

		// Compute risk score using scoring package
		result.RiskScore = scoring.ComputeRiskScore(
			result.Module,
			likelihood,
			impact,
			reasoning,
		)

		assessLog.Debugf("Applied basic risk score: module=%s, likelihood=%d, impact=%d, score=%d",
			result.Module, likelihood, impact, result.RiskScore.Score)
	}
}
