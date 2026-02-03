package riskassess

import (
	"testing"

	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/scoring"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBuildRiskAssessmentPrompt(t *testing.T) {
	input := &AIRiskAssessmentInput{
		ProfileName:          "test-profile.json",
		TotalControls:        10,
		SatisfiedControls:    7,
		NotSatisfiedControls: 3,
		Modules: []ModuleAnalysisInput{
			{
				Module: "test-module",
				VulnerabilityFindings: []scoring.VulnerabilityInput{
					{
						Scanner:       "trivy",
						Severity:      "HIGH",
						Vulnerability: "CVE-2024-1234",
						Description:   "Test vulnerability",
					},
				},
				Context: scoring.ModuleContext{
					ComponentType:    "service",
					Criticality:      "high",
					ExistingControls: []string{"ac-2", "ac-3"},
				},
				ControlsSatisfied:    5,
				ControlsNotSatisfied: 2,
				Impact:               4,
			},
		},
	}

	prompt, err := buildRiskAssessmentPrompt("", input)
	assert.NoError(t, err)
	assert.NotEmpty(t, prompt)

	// Check prompt contains expected sections
	assert.Contains(t, prompt, "executive_summary")
	assert.Contains(t, prompt, "Profile: test-profile.json")
	assert.Contains(t, prompt, "Modules assessed: 1")
	assert.Contains(t, prompt, "Total controls: 10")
	assert.Contains(t, prompt, "Controls satisfied: 7")
	assert.Contains(t, prompt, "Controls not satisfied: 3")
}

func TestApplyAIRiskAssessment(t *testing.T) {
	// Set up test logger
	cleanup := logging.SetTestLogger(zaptest.NewLogger(t))
	defer cleanup()
	logger := logging.C()

	results := []*ModuleAssessmentResult{
		{
			Module:       "module1",
			Satisfied:    5,
			NotSatisfied: 2,
		},
		{
			Module:       "module2",
			Satisfied:    8,
			NotSatisfied: 1,
		},
	}

	aiOutput := &AIRiskAssessmentOutput{
		ModuleAnalyses: []ModuleAnalysisData{
			{
				Module:             "module1",
				ComputedLikelihood: 4,
				Reasoning:          "High vulnerabilities present",
				RiskSummary:        "Elevated risk",
			},
			{
				Module:             "module2",
				ComputedLikelihood: 2,
				Reasoning:          "Low vulnerabilities, good controls",
				RiskSummary:        "Low risk",
			},
		},
		Confidence: 0.85,
	}

	ApplyAIRiskAssessment(results, aiOutput, logger)

	// Check that risk scores were applied
	assert.NotNil(t, results[0].RiskScore)
	assert.Equal(t, 4, results[0].RiskScore.Likelihood)
	assert.Equal(t, "High vulnerabilities present", results[0].RiskScore.Reasoning)

	assert.NotNil(t, results[1].RiskScore)
	assert.Equal(t, 2, results[1].RiskScore.Likelihood)
	assert.Equal(t, "Low vulnerabilities, good controls", results[1].RiskScore.Reasoning)
}

func TestApplyAIRiskAssessment_MissingModule(t *testing.T) {
	// Set up test logger
	cleanup := logging.SetTestLogger(zaptest.NewLogger(t))
	defer cleanup()
	logger := logging.C()

	results := []*ModuleAssessmentResult{
		{
			Module:       "module1",
			Satisfied:    5,
			NotSatisfied: 2,
		},
		{
			Module:       "module-missing",
			Satisfied:    3,
			NotSatisfied: 1,
		},
	}

	aiOutput := &AIRiskAssessmentOutput{
		ModuleAnalyses: []ModuleAnalysisData{
			{
				Module:             "module1",
				ComputedLikelihood: 4,
				Reasoning:          "Test reasoning",
				RiskSummary:        "Test summary",
			},
		},
		Confidence: 0.85,
	}

	// Should not panic when module is missing from AI output
	ApplyAIRiskAssessment(results, aiOutput, logger)

	// First module should have risk score
	assert.NotNil(t, results[0].RiskScore)

	// Second module should not have risk score (missing from AI output)
	assert.Nil(t, results[1].RiskScore)
}

func TestApplyAIRiskAssessment_NoModuleAnalyses(t *testing.T) {
	// Set up test logger
	cleanup := logging.SetTestLogger(zaptest.NewLogger(t))
	defer cleanup()
	logger := logging.C()

	results := []*ModuleAssessmentResult{
		{
			Module:       "module1",
			Satisfied:    5,
			NotSatisfied: 2,
		},
		{
			Module:       "module2",
			Satisfied:    3,
			NotSatisfied: 1,
		},
	}

	// AI output with no per-module analyses (only executive summary)
	aiOutput := &AIRiskAssessmentOutput{
		ExecutiveSummary: ExecutiveSummaryData{
			OverallRiskPosture: "moderate",
			SummaryNarrative:   "The assessment reveals a moderate risk posture...",
		},
		ModuleAnalyses: nil, // No per-module analysis
		Confidence:     0.85,
	}

	// Should not panic when module analyses are missing
	ApplyAIRiskAssessment(results, aiOutput, logger)

	// Modules should not have AI-generated risk scores applied
	assert.Nil(t, results[0].RiskScore)
	assert.Nil(t, results[1].RiskScore)
}

func TestBuildExecutiveSummaryWithAI(t *testing.T) {
	basicStats := ExecutiveSummary{
		TotalControls:    10,
		Satisfied:        7,
		NotSatisfied:     3,
		SatisfactionRate: 70.0,
		NotSatisfiedRate: 30.0,
		ModulesAssessed:  2,
	}

	aiOutput := &AIRiskAssessmentOutput{
		ExecutiveSummary: ExecutiveSummaryData{
			OverallRiskPosture: "moderate",
			SummaryNarrative:   "The assessment reveals a moderate risk posture...",
			KeyFindings: []string{
				"Finding 1: High vulnerability in module1",
				"Finding 2: Strong test coverage across modules",
			},
			CriticalModules: []CriticalModuleInfo{
				{
					Module: "module1",
					Reason: "Critical vulnerabilities detected",
				},
			},
			Trends: []string{
				"Dependency vulnerabilities are common",
			},
			StrategicRecommendations: []string{
				"Implement automated dependency scanning",
				"Enhance authentication controls",
			},
		},
		Confidence: 0.90,
	}

	summary := BuildExecutiveSummary(aiOutput, basicStats)

	// Check basic stats are preserved
	assert.Equal(t, 10, summary.TotalControls)
	assert.Equal(t, 7, summary.Satisfied)
	assert.Equal(t, 3, summary.NotSatisfied)
	assert.Equal(t, 70.0, summary.SatisfactionRate)
	assert.Equal(t, 2, summary.ModulesAssessed)

	// Check AI-generated content is added
	assert.Equal(t, "moderate", summary.OverallRiskPosture)
	assert.Contains(t, summary.SummaryNarrative, "moderate risk posture")
	assert.Len(t, summary.KeyFindings, 2)
	assert.Len(t, summary.CriticalModules, 1)
	assert.Equal(t, "module1", summary.CriticalModules[0].Module)
	assert.Len(t, summary.Trends, 1)
	assert.Len(t, summary.StrategicRecommendations, 2)
	assert.Equal(t, 0.90, summary.AIConfidence)
	assert.True(t, summary.HasAISummary)
}
