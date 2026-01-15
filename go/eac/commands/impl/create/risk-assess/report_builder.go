// File: go/eac/commands/impl/create/risk-assess/report_builder.go
// Purpose: Build template data from assessment results

package riskassess

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"
)

// buildReportData converts assessment results into template data structure
func buildReportData(
	config *AssessConfig,
	results []*ModuleAssessmentResult,
	profile *oscalTypes.Profile,
	isSubset bool,
	totalModules int,
	aiOutput *AIRiskAssessmentOutput,
) *RiskAssessmentReportData {
	data := &RiskAssessmentReportData{
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05 MST"),
		ProfileName:   filepath.Base(config.ProfilePath),
		TestSuite:     "all", // Always uses all tests
		ModuleResults: make([]ModuleReportData, len(results)),
	}

	// Build scope description
	if isSubset {
		data.ScopeDescription = fmt.Sprintf("Subset Assessment (%d of %d modules)", len(results), totalModules)
	} else {
		data.ScopeDescription = fmt.Sprintf("Full Assessment (all %d modules)", len(results))
	}

	// Build summary statistics (basic)
	basicSummary := buildBasicExecutiveSummary(results)

	// Enhance with AI-generated content if available
	if aiOutput != nil {
		data.Summary = BuildExecutiveSummary(aiOutput, basicSummary)
	} else {
		data.Summary = basicSummary
	}

	// Build per-module data
	for i, result := range results {
		data.ModuleResults[i] = buildModuleReportData(config, result)
	}

	// Set aggregated report path
	data.AggregatedReportPath = filepath.Join(config.OutputDir, "*.md")

	return data
}

// buildBasicExecutiveSummary calculates basic summary statistics
func buildBasicExecutiveSummary(results []*ModuleAssessmentResult) ExecutiveSummary {
	summary := ExecutiveSummary{
		ModulesAssessed: len(results),
		HasAISummary:    false, // Will be set to true if AI summary is added
	}

	for _, result := range results {
		summary.Satisfied += result.Satisfied
		summary.NotSatisfied += result.NotSatisfied
	}

	summary.TotalControls = summary.Satisfied + summary.NotSatisfied
	if summary.TotalControls > 0 {
		summary.SatisfactionRate = float64(summary.Satisfied) / float64(summary.TotalControls) * 100
		summary.NotSatisfiedRate = 100 - summary.SatisfactionRate
	}

	return summary
}

// buildModuleReportData builds report data for a single module
func buildModuleReportData(config *AssessConfig, result *ModuleAssessmentResult) ModuleReportData {
	data := ModuleReportData{
		Module:       result.Module,
		Satisfied:    result.Satisfied,
		NotSatisfied: result.NotSatisfied,
		RiskScore:    result.RiskScore,
		Warnings:     result.Warnings,
	}

	// Format risk score
	if result.RiskScore != nil {
		data.RiskScoreFormatted = scoring.FormatRiskScorePlain(result.RiskScore)
	} else {
		data.RiskScoreFormatted = "N/A"
	}

	// Extract test and security evidence
	data.TestEvidenceFormatted, data.SecurityEvidenceFormatted = extractEvidenceFormatted(result)

	// Extract test type breakdown and suite summaries from manifest
	data.TestTypeBreakdown, data.SuiteSummary = extractManifestDetails(result)

	// Extract control lists and findings
	data.SatisfiedControls, data.NotSatisfiedControls, data.NotSatisfiedFindings = extractControlsAndFindings(result)

	// Set report path (relative to workspace root)
	if result.OutputPath != "" {
		relPath, err := filepath.Rel(config.WorkspaceRoot, result.OutputPath)
		if err == nil {
			data.ReportPath = relPath
		} else {
			data.ReportPath = result.OutputPath
		}
	}

	return data
}

// extractEvidenceFormatted extracts formatted test and security evidence strings
func extractEvidenceFormatted(result *ModuleAssessmentResult) (testEvidence, secEvidence string) {
	testEvidence = "N/A"
	secEvidence = "N/A"

	if result.AssessmentResults == nil || len(result.AssessmentResults.Results) == 0 {
		return
	}

	moduleResult := result.AssessmentResults.Results[0]
	if moduleResult.Observations == nil {
		return
	}

	for _, obs := range *moduleResult.Observations {
		if obs.Title == "Test Results" && obs.Props != nil {
			var totalTests, passedTests, failedTests string
			for _, prop := range *obs.Props {
				switch prop.Name {
				case "total-tests":
					totalTests = prop.Value
				case "passed-tests":
					passedTests = prop.Value
				case "failed-tests":
					failedTests = prop.Value
				}
			}
			if totalTests != "" {
				// Calculate skipped tests
				total, _ := strconv.Atoi(totalTests)
				passed, _ := strconv.Atoi(passedTests)
				failed, _ := strconv.Atoi(failedTests)
				skipped := total - passed - failed

				if failed > 0 {
					testEvidence = fmt.Sprintf("%s/%s tests (%d failed)", passedTests, totalTests, failed)
				} else if skipped > 0 {
					testEvidence = fmt.Sprintf("%s/%s tests (%d skipped)", passedTests, totalTests, skipped)
				} else {
					testEvidence = fmt.Sprintf("%s/%s tests", passedTests, totalTests)
				}
			}
		} else if obs.Title == "Security Scan Results" && obs.Props != nil {
			var critVulns, highVulns, medVulns, lowVulns, sbomComponents string
			for _, prop := range *obs.Props {
				switch prop.Name {
				case "critical-vulns":
					critVulns = prop.Value
				case "high-vulns":
					highVulns = prop.Value
				case "medium-vulns":
					medVulns = prop.Value
				case "low-vulns":
					lowVulns = prop.Value
				case "sbom-components":
					sbomComponents = prop.Value
				}
			}
			if critVulns != "" {
				secEvidence = fmt.Sprintf("Vulns: %sC/%sH/%sM/%sL", critVulns, highVulns, medVulns, lowVulns)
				if sbomComponents != "" {
					secEvidence += fmt.Sprintf(", SBOM: %s pkgs", sbomComponents)
				}
			}
		}
	}

	return
}

// extractControlsAndFindings extracts control IDs and findings from assessment results
func extractControlsAndFindings(result *ModuleAssessmentResult) (satisfied, notSatisfied []string, findings []FindingData) {
	if result.AssessmentResults == nil || len(result.AssessmentResults.Results) == 0 {
		return
	}

	moduleResult := result.AssessmentResults.Results[0]
	if moduleResult.Findings == nil {
		return
	}

	for _, finding := range *moduleResult.Findings {
		controlID := finding.Target.TargetId
		if finding.Target.Status.State == oscal.StateSatisfied {
			satisfied = append(satisfied, controlID)
		} else {
			notSatisfied = append(notSatisfied, controlID)
			findings = append(findings, FindingData{
				ControlID: controlID,
				Title:     finding.Title,
			})
		}
	}

	return
}

// extractManifestDetails extracts test type breakdown and suite summaries from manifest metadata
func extractManifestDetails(result *ModuleAssessmentResult) (testTypeBreakdown, suiteSummary string) {
	testTypeBreakdown = "N/A"
	suiteSummary = "N/A"

	// Check if we have test manifest in evidence collection
	if result.Evidence == nil || result.Evidence.TestManifestData == nil {
		return
	}

	manifest := result.Evidence.TestManifestData

	// Build test type breakdown
	typeBreakdown := make(map[string]int)
	for _, test := range manifest.Tests {
		typeBreakdown[test.Type]++
	}

	if len(typeBreakdown) > 0 {
		var typeParts []string
		for testType, count := range typeBreakdown {
			typeParts = append(typeParts, fmt.Sprintf("%d %s", count, testType))
		}
		testTypeBreakdown = strings.Join(typeParts, ", ")
	}

	// Build suite-level summaries
	if len(manifest.Suites) > 0 {
		var suiteLines []string
		for suiteName, suiteResult := range manifest.Suites {
			suiteLines = append(suiteLines, fmt.Sprintf(
				"  - %s: %d/%d passed",
				suiteName,
				suiteResult.Passed,
				suiteResult.Total,
			))
		}
		suiteSummary = strings.Join(suiteLines, "\n")
	}

	return
}

// formatPercentage formats a percentage value with one decimal place
func formatPercentage(value, total int) string {
	if total == 0 {
		return "0.0"
	}
	return fmt.Sprintf("%.1f", float64(value)/float64(total)*100)
}
