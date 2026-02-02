package riskassess

import (
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/google/uuid"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
	"github.com/stretchr/testify/assert"
)

// TestBuildAIRiskAssessmentInput tests the input builder for AI risk assessment.
func TestBuildAIRiskAssessmentInput(t *testing.T) {
	// Create test profile with controls
	profile := &oscalTypes.Profile{
		UUID: "test-profile",
		Imports: []oscalTypes.Import{
			{
				IncludeControls: &[]oscalTypes.SelectControlById{
					{
						WithIds: &[]string{"ac-2", "ac-3", "ia-2", "ia-5"},
					},
				},
			},
		},
	}

	// Create mock module results with evidence (no security files for simplicity)
	results := []*ModuleAssessmentResult{
		{
			Module:       "module1",
			Satisfied:    3,
			NotSatisfied: 1,
			Evidence: &evidence.EvidenceCollection{
				Module: "module1",
			},
			AssessmentResults: createMockAssessmentResults("module1", []string{"ac-2", "ac-3", "ia-2"}, []string{"ia-5"}),
		},
		{
			Module:       "module2",
			Satisfied:    2,
			NotSatisfied: 0,
			Evidence: &evidence.EvidenceCollection{
				Module: "module2",
			},
			AssessmentResults: createMockAssessmentResults("module2", []string{"ac-2", "ac-3"}, []string{}),
		},
	}

	config := &AssessConfig{
		WorkspaceRoot: ".",
		ProfilePath:   "/path/to/profile.json",
	}

	// Build AI input
	input := buildAIRiskAssessmentInput(config, results, profile)

	// Verify input structure
	assert.NotNil(t, input)
	assert.Equal(t, "profile.json", input.ProfileName)
	assert.Equal(t, 4, input.TotalControls)
	assert.Equal(t, 5, input.SatisfiedControls)
	assert.Equal(t, 1, input.NotSatisfiedControls)
	assert.Len(t, input.Modules, 2)

	// Check first module input
	module1 := input.Modules[0]
	assert.Equal(t, "module1", module1.Module)
	assert.Equal(t, 3, module1.ControlsSatisfied)
	assert.Equal(t, 1, module1.ControlsNotSatisfied)

	// Check second module input
	module2 := input.Modules[1]
	assert.Equal(t, "module2", module2.Module)
	assert.Equal(t, 2, module2.ControlsSatisfied)
	assert.Equal(t, 0, module2.ControlsNotSatisfied)
}

// TestExtractSatisfiedControlIDs tests extracting satisfied control IDs.
func TestExtractSatisfiedControlIDs(t *testing.T) {
	result := &ModuleAssessmentResult{
		Module:            "test-module",
		AssessmentResults: createMockAssessmentResults("test-module", []string{"ac-2", "ac-3", "ia-2"}, []string{"ia-5"}),
	}

	controlIDs := extractSatisfiedControlIDs(result)

	assert.Len(t, controlIDs, 3)
	assert.Contains(t, controlIDs, "ac-2")
	assert.Contains(t, controlIDs, "ac-3")
	assert.Contains(t, controlIDs, "ia-2")
	assert.NotContains(t, controlIDs, "ia-5")
}

// TestExtractSatisfiedControlIDs_NilResults tests with nil assessment results.
func TestExtractSatisfiedControlIDs_NilResults(t *testing.T) {
	result := &ModuleAssessmentResult{
		Module:            "test-module",
		AssessmentResults: nil,
	}

	controlIDs := extractSatisfiedControlIDs(result)

	assert.Empty(t, controlIDs)
}

// TestApplyBasicRiskScoring tests the fallback basic scoring.
func TestApplyBasicRiskScoring(t *testing.T) {
	results := []*ModuleAssessmentResult{
		{
			Module:       "module1",
			Satisfied:    5,
			NotSatisfied: 2,
		},
		{
			Module:       "module2",
			Satisfied:    8,
			NotSatisfied: 0,
		},
	}

	applyBasicRiskScoring(results)

	// Check that risk scores were applied
	assert.NotNil(t, results[0].RiskScore)
	assert.Equal(t, 3, results[0].RiskScore.Likelihood)
	assert.Equal(t, 3, results[0].RiskScore.Impact)
	assert.Greater(t, results[0].RiskScore.Score, 0)
	assert.Contains(t, results[0].RiskScore.Reasoning, "Basic scoring applied")
	assert.Contains(t, results[0].RiskScore.Reasoning, "5 controls satisfied")

	assert.NotNil(t, results[1].RiskScore)
	assert.Equal(t, 3, results[1].RiskScore.Likelihood)
	assert.Equal(t, 3, results[1].RiskScore.Impact)
}

// TestApplyBasicRiskScoring_EmptyResults tests with empty results.
func TestApplyBasicRiskScoring_EmptyResults(t *testing.T) {
	var results []*ModuleAssessmentResult

	// Should not panic
	applyBasicRiskScoring(results)
}

// Helper function to create mock assessment results.
func createMockAssessmentResults(moduleName string, satisfiedControls, notSatisfiedControls []string) *oscalTypes.AssessmentResults {
	arUUID := uuid.New().String()
	ar := oscal.NewAssessmentResults(arUUID, "Test Assessment", "test-profile.json")

	// Create a result UUID and findings
	resultUUID := uuid.New().String()
	var findings []oscalTypes.Finding

	// Create findings for satisfied controls
	for _, controlID := range satisfiedControls {
		finding := oscalTypes.Finding{
			UUID:        "finding-" + controlID,
			Title:       "Control " + controlID + " Assessment",
			Description: "Assessment for control " + controlID,
			Target: oscalTypes.FindingTarget{
				Type:     "objective-id",
				TargetId: controlID,
				Status: oscalTypes.ObjectiveStatus{
					State: oscal.StateSatisfied,
				},
			},
		}
		findings = append(findings, finding)
	}

	// Create findings for not-satisfied controls
	for _, controlID := range notSatisfiedControls {
		finding := oscalTypes.Finding{
			UUID:        "finding-" + controlID,
			Title:       "Control " + controlID + " Assessment",
			Description: "Assessment for control " + controlID,
			Target: oscalTypes.FindingTarget{
				Type:     "objective-id",
				TargetId: controlID,
				Status: oscalTypes.ObjectiveStatus{
					State: oscal.StateNotSatisfied,
				},
			},
		}
		findings = append(findings, finding)
	}

	// Add result with findings
	result := oscalTypes.Result{
		UUID:     resultUUID,
		Title:    "Test Result",
		Findings: &findings,
	}

	// Manually add result (since we don't have AddResult helper)
	ar.Results = []oscalTypes.Result{result}

	return ar
}

// TestImpactCalculation tests that impact is calculated based on module criticality.
func TestImpactCalculation(t *testing.T) {
	profile := &oscalTypes.Profile{
		UUID:    "test-profile",
		Imports: []oscalTypes.Import{},
	}

	// Create test results with different module types
	// Note: Without module registry, all modules will have default criticality
	results := []*ModuleAssessmentResult{
		{
			Module:            "module1",
			Satisfied:         5,
			NotSatisfied:      2,
			Evidence:          &evidence.EvidenceCollection{Module: "module1"},
			AssessmentResults: createMockAssessmentResults("module1", []string{"ac-2"}, []string{}),
		},
	}

	config := &AssessConfig{
		WorkspaceRoot: ".",
		ProfilePath:   "/path/to/profile.json",
	}

	input := buildAIRiskAssessmentInput(config, results, profile)

	assert.Len(t, input.Modules, 1)
	// Default impact should be 3 (medium) when module registry is not available
	assert.Equal(t, 3, input.Modules[0].Impact)
}

// TestBuildAIRiskAssessmentInput_EmptyProfile tests with minimal profile.
func TestBuildAIRiskAssessmentInput_EmptyProfile(t *testing.T) {
	profile := &oscalTypes.Profile{
		UUID: "test-profile",
	}

	results := []*ModuleAssessmentResult{
		{
			Module:            "module1",
			Satisfied:         2,
			NotSatisfied:      1,
			Evidence:          &evidence.EvidenceCollection{Module: "module1"},
			AssessmentResults: createMockAssessmentResults("module1", []string{"ac-2", "ac-3"}, []string{"ia-2"}),
		},
	}

	config := &AssessConfig{
		WorkspaceRoot: ".",
		ProfilePath:   "/path/to/profile.json",
	}

	input := buildAIRiskAssessmentInput(config, results, profile)

	assert.NotNil(t, input)
	assert.Equal(t, 0, input.TotalControls) // No imports, so no total controls
	assert.Equal(t, 2, input.SatisfiedControls)
	assert.Equal(t, 1, input.NotSatisfiedControls)
}
