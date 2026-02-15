package riskassess

import (
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/commands/repository/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestExtractSatisfiedControlIDs_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		result   *ModuleAssessmentResult
		expected []string
	}{
		{
			name: "nil assessment results",
			result: &ModuleAssessmentResult{
				Module:            "test-module",
				AssessmentResults: nil,
			},
			expected: nil,
		},
		{
			name: "nil results slice",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: nil,
				},
			},
			expected: nil,
		},
		{
			name: "empty results slice",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{},
				},
			},
			expected: nil,
		},
		{
			name: "result with nil findings",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: nil,
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "all satisfied controls",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: &[]oscalTypes.Finding{
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-2",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"ac-1", "ac-2"},
		},
		{
			name: "only not-satisfied controls",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: &[]oscalTypes.Finding{
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateNotSatisfied,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "mixed satisfied and not-satisfied",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: &[]oscalTypes.Finding{
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-2",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateNotSatisfied,
										},
									},
								},
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "au-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"ac-1", "au-1"},
		},
		{
			name: "multiple result entries",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: &[]oscalTypes.Finding{
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
							},
						},
						{
							Findings: &[]oscalTypes.Finding{
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "sc-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "sc-2",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateNotSatisfied,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"ac-1", "sc-1"},
		},
		{
			name: "finding with empty target ID is skipped",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: &[]oscalTypes.Finding{
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
								{
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"ac-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSatisfiedControlIDs(tt.result)
			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestApplyBasicRiskScoring_Comprehensive(t *testing.T) {
	// Set up test logger to suppress log output
	cleanup := logging.SetTestLogger(zaptest.NewLogger(t))
	defer cleanup()

	t.Run("applies fixed likelihood and impact to all results", func(t *testing.T) {
		results := []*ModuleAssessmentResult{
			{
				Module:       "module-1",
				Satisfied:    5,
				NotSatisfied: 2,
			},
			{
				Module:       "module-2",
				Satisfied:    8,
				NotSatisfied: 0,
			},
			{
				Module:       "module-3",
				Satisfied:    0,
				NotSatisfied: 10,
			},
		}

		applyBasicRiskScoring(results)

		for _, result := range results {
			require.NotNil(t, result.RiskScore, "module %s should have a risk score", result.Module)
			assert.Equal(t, 3, result.RiskScore.Likelihood, "fallback likelihood should be 3")
			assert.Equal(t, 3, result.RiskScore.Impact, "fallback impact should be 3")
			assert.Equal(t, 9, result.RiskScore.Score, "score should be 3*3=9")
			assert.Contains(t, result.RiskScore.Reasoning, "satisfied")
			assert.Contains(t, result.RiskScore.Reasoning, "not satisfied")
		}
	})

	t.Run("reasoning contains actual counts", func(t *testing.T) {
		results := []*ModuleAssessmentResult{
			{
				Module:       "test-module",
				Satisfied:    7,
				NotSatisfied: 3,
			},
		}

		applyBasicRiskScoring(results)

		require.NotNil(t, results[0].RiskScore)
		assert.Contains(t, results[0].RiskScore.Reasoning, "7 controls satisfied")
		assert.Contains(t, results[0].RiskScore.Reasoning, "3 not satisfied")
	})

	t.Run("empty results does not panic", func(t *testing.T) {
		results := []*ModuleAssessmentResult{}
		assert.NotPanics(t, func() {
			applyBasicRiskScoring(results)
		})
	})

	t.Run("nil result slice does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			applyBasicRiskScoring(nil)
		})
	})
}
