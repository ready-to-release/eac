// File: go/eac/commands/impl/create/risk-assess/oscal.go
// Purpose: OSCAL assessment-results document building
//
// This file contains all functions related to building OSCAL assessment-results
// documents from evidence.

package riskassess

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
)

// buildAssessmentResultsForModule creates OSCAL assessment-results for a specific module.
func buildAssessmentResultsForModule(config *AssessConfig, moduleName string, profile *oscalTypes.Profile, ec *evidence.EvidenceCollection) (*oscalTypes.AssessmentResults, error) {
	controlIDs := oscal.GetProfileControlIDs(profile)

	// Get control evidence directly from test manifest
	// This extracts @control: tags and test status from the manifest test entries
	controlTestEvidence := make(map[string]*evidence.ControlTestEvidence)

	if ec.TestManifestData != nil {
		// Search all tests regardless of suite
		controlTestEvidence = evidence.GetControlTestEvidenceFromManifest(ec.TestManifestData.Tests, "")
	}

	if len(controlTestEvidence) == 0 {
		assessLog.Debugf("No control evidence found for %s in any suite", moduleName)
	}

	// Create assessment-results
	arUUID := uuid.New().String()
	profilePath := oscal.GetProfilePath(config.WorkspaceRoot, moduleName)
	relativeProfilePath, relErr := filepath.Rel(config.WorkspaceRoot, profilePath)
	if relErr != nil {
		relativeProfilePath = profilePath // Fallback to absolute path
	}

	ar := oscal.NewAssessmentResults(
		arUUID,
		fmt.Sprintf("Risk Assessment for %s", moduleName),
		relativeProfilePath,
	)

	// Create result with observations and findings
	props := []oscalTypes.Property{
		{Name: "assessment-type", Value: "automated"},
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
	resultData := oscalTypes.Result{
		UUID:        uuid.New().String(),
		Title:       fmt.Sprintf("%s Assessment", moduleName),
		Description: "Assessment-results generated from test and security evidence",
		Start:       time.Now().UTC(),
		Props:       &props,
		ReviewedControls: oscalTypes.ReviewedControls{
			ControlSelections: controlSelections,
		},
	}

	// Create observations from evidence
	observations := createObservationsForModule(config, moduleName, ec)
	resultData.Observations = observations

	// Create findings for each control
	findings := createFindingsForModule(config, moduleName, controlIDs, controlTestEvidence, ec, observations)
	resultData.Findings = findings

	oscal.AddResult(ar, resultData)
	return ar, nil
}

// createObservationsForModule creates OSCAL observations from evidence for a specific module.
func createObservationsForModule(config *AssessConfig, moduleName string, ec *evidence.EvidenceCollection) *[]oscalTypes.Observation {
	var observations []oscalTypes.Observation
	now := time.Now().UTC()

	// Create observation for test evidence from manifest
	if ec.TestManifestData != nil {
		var relevantEvidence []oscalTypes.RelevantEvidence

		// Add artifact references from manifest
		for _, artifact := range ec.TestManifestData.Artifacts {
			if artifact.Type == "report" || artifact.Type == "coverage" || artifact.Type == "log" {
				relevantEvidence = append(relevantEvidence, oscalTypes.RelevantEvidence{
					Href:        artifact.Path,
					Description: fmt.Sprintf("%s artifact: %s", artifact.Type, artifact.Name),
				})
			}
		}

		var props []oscalTypes.Property

		// Add manifest metadata properties
		props = append(props,
			oscalTypes.Property{Name: "test-agent", Value: ec.TestManifestData.TestAgent},
			oscalTypes.Property{Name: "test-id", Value: ec.TestManifestData.TestID},
		)

		if ec.TestManifestData.GitCommit != "" {
			props = append(props, oscalTypes.Property{Name: "git-commit", Value: ec.TestManifestData.GitCommit})
		}

		if ec.TestManifestData.BuildID != "" {
			props = append(props, oscalTypes.Property{Name: "build-id", Value: ec.TestManifestData.BuildID})
		}

		// Add summary props
		if ec.TestSummary != nil {
			props = append(props,
				oscalTypes.Property{Name: "total-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Total)},
				oscalTypes.Property{Name: "passed-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Passed)},
				oscalTypes.Property{Name: "failed-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Failed)},
				oscalTypes.Property{Name: "skipped-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Skipped)},
			)
		}

		// Add suite-level breakdown
		for suiteName, suiteResult := range ec.TestManifestData.Suites {
			props = append(props,
				oscalTypes.Property{
					Name:  fmt.Sprintf("suite-%s-total", suiteName),
					Value: fmt.Sprintf("%d", suiteResult.Total),
				},
				oscalTypes.Property{
					Name:  fmt.Sprintf("suite-%s-passed", suiteName),
					Value: fmt.Sprintf("%d", suiteResult.Passed),
				},
				oscalTypes.Property{
					Name:  fmt.Sprintf("suite-%s-failed", suiteName),
					Value: fmt.Sprintf("%d", suiteResult.Failed),
				},
				oscalTypes.Property{
					Name:  fmt.Sprintf("suite-%s-skipped", suiteName),
					Value: fmt.Sprintf("%d", suiteResult.Skipped),
				},
			)
		}

		testObs := oscalTypes.Observation{
			UUID:             uuid.New().String(),
			Title:            "Test Results",
			Description:      fmt.Sprintf("Test evidence for module %s from test manifest", moduleName),
			Methods:          []string{oscal.MethodTestAutomated},
			Collected:        now,
			RelevantEvidence: &relevantEvidence,
			Props:            &props,
		}

		observations = append(observations, testObs)
	}

	// Create observation for security evidence
	if ec.SecurityResults != nil {
		var relevantEvidence []oscalTypes.RelevantEvidence

		if ec.SecurityResults.VulnFile != "" {
			relPath, relErr := filepath.Rel(config.WorkspaceRoot, ec.SecurityResults.VulnFile)
			if relErr != nil {
				relPath = ec.SecurityResults.VulnFile // Fallback to absolute path
			}
			relevantEvidence = append(relevantEvidence, oscalTypes.RelevantEvidence{
				Href:        relPath,
				Description: "Vulnerability scan results",
			})
		}

		if ec.SecurityResults.SBOMFile != "" {
			relPath, relErr := filepath.Rel(config.WorkspaceRoot, ec.SecurityResults.SBOMFile)
			if relErr != nil {
				relPath = ec.SecurityResults.SBOMFile // Fallback to absolute path
			}
			relevantEvidence = append(relevantEvidence, oscalTypes.RelevantEvidence{
				Href:        relPath,
				Description: "Software Bill of Materials (SBOM)",
			})
		}

		var props []oscalTypes.Property
		// Add vulnerability summary props
		if ec.VulnSummary != nil {
			props = append(props,
				oscalTypes.Property{Name: "critical-vulns", Value: fmt.Sprintf("%d", ec.VulnSummary.Critical)},
				oscalTypes.Property{Name: "high-vulns", Value: fmt.Sprintf("%d", ec.VulnSummary.High)},
				oscalTypes.Property{Name: "medium-vulns", Value: fmt.Sprintf("%d", ec.VulnSummary.Medium)},
				oscalTypes.Property{Name: "low-vulns", Value: fmt.Sprintf("%d", ec.VulnSummary.Low)},
			)
		}

		// Add SBOM summary props
		if ec.SBOMSummary != nil {
			props = append(props,
				oscalTypes.Property{Name: "sbom-components", Value: fmt.Sprintf("%d", ec.SBOMSummary.TotalComponents)},
			)
		}

		secObs := oscalTypes.Observation{
			UUID:             uuid.New().String(),
			Title:            "Security Scan Results",
			Description:      fmt.Sprintf("Security evidence for module %s", moduleName),
			Methods:          []string{oscal.MethodTestAutomated},
			Collected:        now,
			RelevantEvidence: &relevantEvidence,
			Props:            &props,
		}

		observations = append(observations, secObs)
	}

	return &observations
}

// createFindingsForModule creates OSCAL findings for each control for a specific module.
func createFindingsForModule(config *AssessConfig, moduleName string, controlIDs []string, controlTestEvidence map[string]*evidence.ControlTestEvidence, ec *evidence.EvidenceCollection, observations *[]oscalTypes.Observation) *[]oscalTypes.Finding {
	var findings []oscalTypes.Finding

	for _, controlID := range controlIDs {
		// Get test evidence for this control
		testEv := controlTestEvidence[controlID]

		// Determine status based on evidence
		state := determineControlStatusFromEvidence(controlID, testEv, ec)

		target := oscalTypes.FindingTarget{
			Type:     oscal.TargetTypeControlID,
			TargetId: controlID,
			Status: oscalTypes.ObjectiveStatus{
				State: state,
			},
		}

		var relatedObs []oscalTypes.RelatedObservation
		// Link to observations
		if observations != nil {
			for _, obs := range *observations {
				relatedObs = append(relatedObs, oscalTypes.RelatedObservation{
					ObservationUuid: obs.UUID,
				})
			}
		}

		var props []oscalTypes.Property
		// Add test coverage info with detailed status
		if testEv != nil && testEv.HasEvidence {
			props = append(props, oscalTypes.Property{
				Name: "test-coverage",
				Value: fmt.Sprintf("%d tests (%d passed, %d failed, %d skipped/not-run)",
					testEv.TotalTests, testEv.PassedTests, testEv.FailedTests,
					testEv.SkippedTests+(testEv.TotalTests-testEv.PassedTests-testEv.FailedTests-testEv.SkippedTests)),
			})
		} else {
			props = append(props, oscalTypes.Property{
				Name:  "test-coverage",
				Value: "no tests",
			})
		}

		// Build detailed remarks with test evidence
		remarks := buildRemarksFromEvidence(controlID, testEv)

		finding := oscalTypes.Finding{
			UUID:                uuid.New().String(),
			Title:               fmt.Sprintf("Control %s Assessment", strings.ToUpper(controlID)),
			Description:         fmt.Sprintf("Assessment finding for control %s", controlID),
			Target:              target,
			RelatedObservations: &relatedObs,
			Props:               &props,
			Remarks:             remarks,
		}

		findings = append(findings, finding)
	}

	return &findings
}

// determineControlStatusFromEvidence determines the satisfied/not-satisfied status using comprehensive evidence.
func determineControlStatusFromEvidence(controlID string, testEv *evidence.ControlTestEvidence, ec *evidence.EvidenceCollection) string {
	// If no tests exist for this control, not-satisfied
	if testEv == nil || !testEv.HasEvidence {
		return oscal.StateNotSatisfied
	}

	// If any tests failed, not-satisfied
	if testEv.FailedTests > 0 {
		return oscal.StateNotSatisfied
	}

	// If critical/high vulnerabilities exist, not-satisfied
	if ec.VulnSummary != nil && (ec.VulnSummary.Critical > 0 || ec.VulnSummary.High > 0) {
		return oscal.StateNotSatisfied
	}

	// If tests passed and no critical/high vulns, satisfied
	if testEv.PassedTests > 0 {
		return oscal.StateSatisfied
	}

	// If tests exist but haven't run yet or were skipped, not-satisfied
	return oscal.StateNotSatisfied
}

// buildRemarksFromEvidence builds detailed remarks string from test evidence.
func buildRemarksFromEvidence(controlID string, testEv *evidence.ControlTestEvidence) string {
	if testEv == nil || !testEv.HasEvidence {
		return "No test coverage found for this control."
	}

	var remarks strings.Builder
	remarks.WriteString(fmt.Sprintf("Control %s is covered by %d test(s):\n\n", controlID, testEv.TotalTests))

	// Group tests by status
	passed := []string{}
	failed := []string{}
	notRun := []string{}

	for _, test := range testEv.Tests {
		testDesc := fmt.Sprintf("  - %s: %s", test.ScenarioName, test.FeaturePath)
		switch test.Status {
		case "passed":
			passed = append(passed, testDesc)
		case "failed":
			failed = append(failed, testDesc)
		default:
			notRun = append(notRun, testDesc)
		}
	}

	if len(passed) > 0 {
		remarks.WriteString("✓ Passed:\n")
		remarks.WriteString(strings.Join(passed, "\n"))
		remarks.WriteString("\n\n")
	}

	if len(failed) > 0 {
		remarks.WriteString("✗ Failed:\n")
		remarks.WriteString(strings.Join(failed, "\n"))
		remarks.WriteString("\n\n")
	}

	if len(notRun) > 0 {
		remarks.WriteString("○ Not Run/Skipped:\n")
		remarks.WriteString(strings.Join(notRun, "\n"))
		remarks.WriteString("\n")
	}

	return remarks.String()
}

// determineControlStatus determines the satisfied/not-satisfied status for a control.
// Deprecated: Use determineControlStatusFromEvidence instead.
func determineControlStatus(controlID string, controlTestMap map[string][]string, ec *evidence.EvidenceCollection) string {
	// Check if we have tests for this control
	tests, hasTests := controlTestMap[controlID]

	// If no tests and no security evidence, status is not-satisfied
	if !hasTests && (ec.VulnSummary == nil || ec.VulnSummary.Total == 0) {
		return oscal.StateNotSatisfied
	}

	// If tests exist, check if they pass
	if hasTests && len(tests) > 0 {
		// If all tests pass and no critical/high vulnerabilities, satisfied
		if ec.TestSummary != nil && ec.TestSummary.Failed == 0 {
			if ec.VulnSummary == nil || (ec.VulnSummary.Critical == 0 && ec.VulnSummary.High == 0) {
				return oscal.StateSatisfied
			}
		}
	}

	// Default to not-satisfied if any concerns
	if ec.TestSummary != nil && ec.TestSummary.Failed > 0 {
		return oscal.StateNotSatisfied
	}

	if ec.VulnSummary != nil && (ec.VulnSummary.Critical > 0 || ec.VulnSummary.High > 0) {
		return oscal.StateNotSatisfied
	}

	// If we have passing tests, satisfied
	if ec.TestSummary != nil && ec.TestSummary.Passed > 0 {
		return oscal.StateSatisfied
	}

	return oscal.StateNotSatisfied
}
