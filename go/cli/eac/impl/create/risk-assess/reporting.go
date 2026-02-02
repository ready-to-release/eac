// File: go/cli/eac/impl/create/risk-assess/reporting.go
// Purpose: Report generation for risk assessments
//
// This file contains all functions related to displaying assessment results
// and summaries to the user.

package riskassess

import (
	"path/filepath"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/scoring"
)

// reportAggregateResults displays summary for multiple modules with test evidence.
func reportAggregateResults(config *AssessConfig, results []*ModuleAssessmentResult) {
	assessLog.Info("")
	assessLog.Info("═══════════════════════════════════════════════════════════")
	assessLog.Infof("  Risk Assessment Summary")
	assessLog.Info("═══════════════════════════════════════════════════════════")
	assessLog.Info("")
	assessLog.Infof("  Modules assessed: %d", len(results))
	assessLog.Info("")

	// Per-module summary with detailed evidence
	totalSatisfied := 0
	totalNotSatisfied := 0

	for _, result := range results {
		status := "✓"
		if result.NotSatisfied > result.Satisfied {
			status = "⚠"
		}

		riskStr := "N/A"
		if result.RiskScore != nil {
			riskStr = scoring.FormatRiskScore(result.RiskScore)
		}

		assessLog.Infof("  %s %-20s  %2d satisfied, %2d not-satisfied  (Risk: %s)",
			status,
			result.Module,
			result.Satisfied,
			result.NotSatisfied,
			riskStr)

		// Show test evidence summary for this module
		if result.AssessmentResults != nil && len(result.AssessmentResults.Results) > 0 {
			moduleResult := result.AssessmentResults.Results[0]
			if moduleResult.Observations != nil && len(*moduleResult.Observations) > 0 {
				for i := range *moduleResult.Observations {
					obs := &(*moduleResult.Observations)[i]
					if obs.Title == "Test Results" && obs.Props != nil {
						for _, prop := range *obs.Props {
							if prop.Name == "total-tests" || prop.Name == "passed-tests" || prop.Name == "failed-tests" {
								assessLog.Infof("      • %s: %s", prop.Name, prop.Value)
							}
						}
					} else if obs.Title == "Security Scan Results" && obs.Props != nil {
						for _, prop := range *obs.Props {
							if prop.Name == "critical-vulns" || prop.Name == "high-vulns" {
								assessLog.Infof("      • %s: %s", prop.Name, prop.Value)
							}
						}
					}
				}
			}
		}

		totalSatisfied += result.Satisfied
		totalNotSatisfied += result.NotSatisfied
	}

	assessLog.Info("")
	assessLog.Infof("  Total Controls: %d satisfied, %d not-satisfied",
		totalSatisfied, totalNotSatisfied)
	assessLog.Info("")
	assessLog.Info("  Output:")
	for _, result := range results {
		assessLog.Infof("    %s → %s", result.Module, result.OutputPath)
	}
	assessLog.Info("")
	assessLog.Info("  Aggregated Report:")
	assessLog.Infof("    %s/risk-assessment-*.md", config.OutputDir)
	assessLog.Info("")
}

// reportResults displays detailed results for a single module.
func reportResults(config *AssessConfig, ar *oscalTypes.AssessmentResults, arPath string) {
	assessLog.Info("")
	assessLog.Info("═══════════════════════════════════════════════════════════")
	assessLog.Infof("  Assessment Results: %s", config.Modules[0])
	assessLog.Info("═══════════════════════════════════════════════════════════")
	assessLog.Info("")

	if len(ar.Results) > 0 {
		result := ar.Results[0]

		satisfied := 0
		notSatisfied := 0
		if result.Findings != nil {
			for i := range *result.Findings {
				finding := &(*result.Findings)[i]
				if finding.Target.Status.State == oscal.StateSatisfied {
					satisfied++
				} else {
					notSatisfied++
				}
			}
		}

		total := 0
		if result.Findings != nil {
			total = len(*result.Findings)
		}
		assessLog.Infof("  Controls assessed: %d", total)
		assessLog.Infof("  ✓ Satisfied:       %d", satisfied)
		assessLog.Infof("  ✗ Not satisfied:   %d", notSatisfied)

		if notSatisfied > 0 && result.Findings != nil {
			assessLog.Info("")
			assessLog.Info("  Controls needing attention:")
			for i := range *result.Findings {
				finding := &(*result.Findings)[i]
				if finding.Target.Status.State == oscal.StateNotSatisfied {
					assessLog.Infof("    - %s", finding.Target.TargetId)
				}
			}
		}

		// Calculate risk score
		likelihood := scoring.CalculateBaseLikelihood(0, 0, 0, notSatisfied)
		impact := scoring.GetDefaultImpact("service")
		riskScore := scoring.ComputeRiskScore(config.Modules[0], likelihood, impact, "Based on control satisfaction")

		assessLog.Info("")
		assessLog.Infof("  Risk Score: %s", scoring.FormatRiskScore(riskScore))
	}

	assessLog.Info("")
	assessLog.Infof("  Output: %s", arPath)
	assessLog.Info("")
	assessLog.Info("  Next steps:")
	assessLog.Infof("    validate risk %s  # Validate OSCAL", filepath.Base(arPath))
	assessLog.Info("")
}
