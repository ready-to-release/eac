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

	// Get control to test mapping
	controlTestMap, _ := evidence.GetControlToTestMapping(config.WorkspaceRoot, moduleName)

	// Create assessment-results
	arUUID := uuid.New().String()
	profilePath := oscal.GetProfilePath(config.WorkspaceRoot, moduleName)
	relativeProfilePath, _ := filepath.Rel(config.WorkspaceRoot, profilePath)

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
	findings := createFindingsForModule(config, moduleName, controlIDs, controlTestMap, ec, observations)
	resultData.Findings = findings

	oscal.AddResult(ar, resultData)
	return ar, nil
}

// createObservationsForModule creates OSCAL observations from evidence for a specific module.
func createObservationsForModule(config *AssessConfig, moduleName string, ec *evidence.EvidenceCollection) *[]oscalTypes.Observation {
	var observations []oscalTypes.Observation
	now := time.Now().UTC()

	// Create observation for test evidence
	if ec.TestResults != nil {
		var relevantEvidence []oscalTypes.RelevantEvidence

		// Add evidence references
		for _, file := range ec.TestResults.AcceptanceFiles {
			relPath, _ := filepath.Rel(config.WorkspaceRoot, file)
			relevantEvidence = append(relevantEvidence, oscalTypes.RelevantEvidence{
				Href:        relPath,
				Description: "Acceptance test results (Cucumber JSON)",
			})
		}

		for _, file := range ec.TestResults.UnitTestFiles {
			relPath, _ := filepath.Rel(config.WorkspaceRoot, file)
			relevantEvidence = append(relevantEvidence, oscalTypes.RelevantEvidence{
				Href:        relPath,
				Description: "Unit test results",
			})
		}

		var props []oscalTypes.Property
		// Add summary props
		if ec.TestSummary != nil {
			props = append(props,
				oscalTypes.Property{Name: "total-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Total)},
				oscalTypes.Property{Name: "passed-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Passed)},
				oscalTypes.Property{Name: "failed-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Failed)},
			)
		}

		testObs := oscalTypes.Observation{
			UUID:             uuid.New().String(),
			Title:            "Test Results",
			Description:      fmt.Sprintf("Test evidence for module %s", moduleName),
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
			relPath, _ := filepath.Rel(config.WorkspaceRoot, ec.SecurityResults.VulnFile)
			relevantEvidence = append(relevantEvidence, oscalTypes.RelevantEvidence{
				Href:        relPath,
				Description: "Vulnerability scan results",
			})
		}

		if ec.SecurityResults.SBOMFile != "" {
			relPath, _ := filepath.Rel(config.WorkspaceRoot, ec.SecurityResults.SBOMFile)
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
func createFindingsForModule(config *AssessConfig, moduleName string, controlIDs []string, controlTestMap map[string][]string, ec *evidence.EvidenceCollection, observations *[]oscalTypes.Observation) *[]oscalTypes.Finding {
	var findings []oscalTypes.Finding

	for _, controlID := range controlIDs {
		// Determine status based on evidence
		state := determineControlStatus(controlID, controlTestMap, ec)

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
		// Add test coverage info
		if tests, ok := controlTestMap[controlID]; ok {
			props = append(props, oscalTypes.Property{
				Name:  "test-coverage",
				Value: fmt.Sprintf("%d tests", len(tests)),
			})
		} else {
			props = append(props, oscalTypes.Property{
				Name:  "test-coverage",
				Value: "no tests",
			})
		}

		remarks := ""
		if tests, ok := controlTestMap[controlID]; ok {
			remarks = fmt.Sprintf("Tested by: %s", strings.Join(tests, ", "))
		}

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

// determineControlStatus determines the satisfied/not-satisfied status for a control.
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
