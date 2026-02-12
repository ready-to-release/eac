// File: go/cli/eac/impl/create/risk-assess/assess_oscal.go
// Purpose: OSCAL report writing and template loading for risk assessments
//
// This file contains functions for writing aggregated OSCAL assessment results,
// generating Markdown reports, and loading risk assessment templates.

package riskassess

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	sharedTemplate "github.com/ready-to-release/eac/go/clibase/template"
	eacConfig "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
)

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
	teamOverridePath := filepath.Join(workspaceRoot, paths.CLIEDir, paths.EACDir, paths.TemplatesDir, reportsDir, category, templateFilename)
	if _, err := os.Stat(teamOverridePath); err == nil {
		return teamOverridePath, nil
	}

	// Priority 2: System default (templates/reports/<category>/<template>)
	// In container: uses CLIE_CONTAINER_ROOT (/app where Dockerfile copies templates)
	// In local dev: uses workspaceRoot (repo root where templates/ exists)
	distRoot := workspaceRoot
	if containerRoot := os.Getenv(environments.EnvCLIEContainerRoot); containerRoot != "" {
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
