// File: go/cli/eac/impl/create/risk-assess/assess_evidence.go
// Purpose: Evidence-related functions for risk assessment input building and scoring
//
// This file contains functions that bridge evidence data with AI risk assessment input,
// extract control satisfaction data from assessment results, and provide fallback scoring.

package riskassess

import (
	"fmt"
	"path/filepath"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/commands/repository/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/commands/repository/internal/risk/scoring"
)

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
