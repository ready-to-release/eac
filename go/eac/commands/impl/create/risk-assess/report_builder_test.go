// File: go/eac/commands/impl/create/risk-assess/report_builder_test.go
// Purpose: Unit tests for report data builder functions

package riskassess

import (
	"path/filepath"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBasicExecutiveSummary(t *testing.T) {
	tests := []struct {
		name     string
		results  []*ModuleAssessmentResult
		expected ExecutiveSummary
	}{
		{
			name:    "empty results",
			results: []*ModuleAssessmentResult{},
			expected: ExecutiveSummary{
				ModulesAssessed:  0,
				TotalControls:    0,
				Satisfied:        0,
				NotSatisfied:     0,
				SatisfactionRate: 0,
				NotSatisfiedRate: 0,
				HasAISummary:     false,
			},
		},
		{
			name: "single module all satisfied",
			results: []*ModuleAssessmentResult{
				{
					Module:       "test-module",
					Satisfied:    5,
					NotSatisfied: 0,
				},
			},
			expected: ExecutiveSummary{
				ModulesAssessed:  1,
				TotalControls:    5,
				Satisfied:        5,
				NotSatisfied:     0,
				SatisfactionRate: 100.0,
				NotSatisfiedRate: 0.0,
				HasAISummary:     false,
			},
		},
		{
			name: "single module partial satisfaction",
			results: []*ModuleAssessmentResult{
				{
					Module:       "test-module",
					Satisfied:    3,
					NotSatisfied: 2,
				},
			},
			expected: ExecutiveSummary{
				ModulesAssessed:  1,
				TotalControls:    5,
				Satisfied:        3,
				NotSatisfied:     2,
				SatisfactionRate: 60.0,
				NotSatisfiedRate: 40.0,
				HasAISummary:     false,
			},
		},
		{
			name: "multiple modules",
			results: []*ModuleAssessmentResult{
				{
					Module:       "module-1",
					Satisfied:    10,
					NotSatisfied: 2,
				},
				{
					Module:       "module-2",
					Satisfied:    8,
					NotSatisfied: 4,
				},
				{
					Module:       "module-3",
					Satisfied:    15,
					NotSatisfied: 0,
				},
			},
			expected: ExecutiveSummary{
				ModulesAssessed:  3,
				TotalControls:    39,
				Satisfied:        33,
				NotSatisfied:     6,
				SatisfactionRate: 84.61538461538461,
				NotSatisfiedRate: 15.384615384615385,
				HasAISummary:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := buildBasicExecutiveSummary(tt.results)
			assert.Equal(t, tt.expected.ModulesAssessed, summary.ModulesAssessed)
			assert.Equal(t, tt.expected.TotalControls, summary.TotalControls)
			assert.Equal(t, tt.expected.Satisfied, summary.Satisfied)
			assert.Equal(t, tt.expected.NotSatisfied, summary.NotSatisfied)
			assert.InDelta(t, tt.expected.SatisfactionRate, summary.SatisfactionRate, 0.0001)
			assert.InDelta(t, tt.expected.NotSatisfiedRate, summary.NotSatisfiedRate, 0.0001)
			assert.Equal(t, tt.expected.HasAISummary, summary.HasAISummary)
		})
	}
}

func TestExtractEvidenceFormatted(t *testing.T) {
	tests := []struct {
		name             string
		result           *ModuleAssessmentResult
		expectedTest     string
		expectedSecurity string
	}{
		{
			name: "no assessment results",
			result: &ModuleAssessmentResult{
				Module:            "test-module",
				AssessmentResults: nil,
			},
			expectedTest:     "N/A",
			expectedSecurity: "N/A",
		},
		{
			name: "empty results",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{},
				},
			},
			expectedTest:     "N/A",
			expectedSecurity: "N/A",
		},
		{
			name: "test results only - all passed",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Observations: &[]oscalTypes.Observation{
								{
									Title: "Test Results",
									Props: &[]oscalTypes.Property{
										{Name: "total-tests", Value: "10"},
										{Name: "passed-tests", Value: "10"},
										{Name: "failed-tests", Value: "0"},
									},
								},
							},
						},
					},
				},
			},
			expectedTest:     "10/10 tests",
			expectedSecurity: "N/A",
		},
		{
			name: "test results - with failures",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Observations: &[]oscalTypes.Observation{
								{
									Title: "Test Results",
									Props: &[]oscalTypes.Property{
										{Name: "total-tests", Value: "10"},
										{Name: "passed-tests", Value: "7"},
										{Name: "failed-tests", Value: "3"},
									},
								},
							},
						},
					},
				},
			},
			expectedTest:     "7/10 tests (3 failed)",
			expectedSecurity: "N/A",
		},
		{
			name: "test results - with skipped",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Observations: &[]oscalTypes.Observation{
								{
									Title: "Test Results",
									Props: &[]oscalTypes.Property{
										{Name: "total-tests", Value: "10"},
										{Name: "passed-tests", Value: "7"},
										{Name: "failed-tests", Value: "0"},
									},
								},
							},
						},
					},
				},
			},
			expectedTest:     "7/10 tests (3 skipped)",
			expectedSecurity: "N/A",
		},
		{
			name: "security results only",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Observations: &[]oscalTypes.Observation{
								{
									Title: "Security Scan Results",
									Props: &[]oscalTypes.Property{
										{Name: "critical-vulns", Value: "2"},
										{Name: "high-vulns", Value: "5"},
										{Name: "medium-vulns", Value: "10"},
										{Name: "low-vulns", Value: "3"},
										{Name: "sbom-components", Value: "42"},
									},
								},
							},
						},
					},
				},
			},
			expectedTest:     "N/A",
			expectedSecurity: "Vulns: 2C/5H/10M/3L, SBOM: 42 pkgs",
		},
		{
			name: "security results without SBOM",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Observations: &[]oscalTypes.Observation{
								{
									Title: "Security Scan Results",
									Props: &[]oscalTypes.Property{
										{Name: "critical-vulns", Value: "0"},
										{Name: "high-vulns", Value: "0"},
										{Name: "medium-vulns", Value: "0"},
										{Name: "low-vulns", Value: "0"},
									},
								},
							},
						},
					},
				},
			},
			expectedTest:     "N/A",
			expectedSecurity: "Vulns: 0C/0H/0M/0L",
		},
		{
			name: "both test and security results",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Observations: &[]oscalTypes.Observation{
								{
									Title: "Test Results",
									Props: &[]oscalTypes.Property{
										{Name: "total-tests", Value: "10"},
										{Name: "passed-tests", Value: "10"},
										{Name: "failed-tests", Value: "0"},
									},
								},
								{
									Title: "Security Scan Results",
									Props: &[]oscalTypes.Property{
										{Name: "critical-vulns", Value: "0"},
										{Name: "high-vulns", Value: "1"},
										{Name: "medium-vulns", Value: "2"},
										{Name: "low-vulns", Value: "3"},
										{Name: "sbom-components", Value: "25"},
									},
								},
							},
						},
					},
				},
			},
			expectedTest:     "10/10 tests",
			expectedSecurity: "Vulns: 0C/1H/2M/3L, SBOM: 25 pkgs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEvidence, secEvidence := extractEvidenceFormatted(tt.result)
			assert.Equal(t, tt.expectedTest, testEvidence)
			assert.Equal(t, tt.expectedSecurity, secEvidence)
		})
	}
}

func TestExtractControlsAndFindings(t *testing.T) {
	tests := []struct {
		name                 string
		result               *ModuleAssessmentResult
		expectedSatisfied    []string
		expectedNotSatisfied []string
		expectedFindings     []FindingData
	}{
		{
			name: "no assessment results",
			result: &ModuleAssessmentResult{
				Module:            "test-module",
				AssessmentResults: nil,
			},
			expectedSatisfied:    nil,
			expectedNotSatisfied: nil,
			expectedFindings:     nil,
		},
		{
			name: "empty results",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{},
				},
			},
			expectedSatisfied:    nil,
			expectedNotSatisfied: nil,
			expectedFindings:     nil,
		},
		{
			name: "no findings",
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
			expectedSatisfied:    nil,
			expectedNotSatisfied: nil,
			expectedFindings:     nil,
		},
		{
			name: "all satisfied",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: &[]oscalTypes.Finding{
								{
									Title: "Control AC-1 is satisfied",
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
								{
									Title: "Control AC-2 is satisfied",
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
			expectedSatisfied:    []string{"ac-1", "ac-2"},
			expectedNotSatisfied: nil,
			expectedFindings:     nil,
		},
		{
			name: "all not satisfied",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: &[]oscalTypes.Finding{
								{
									Title: "Control AC-1 is not satisfied",
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateNotSatisfied,
										},
									},
								},
								{
									Title: "Control AC-2 is not satisfied",
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-2",
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
			expectedSatisfied:    nil,
			expectedNotSatisfied: []string{"ac-1", "ac-2"},
			expectedFindings: []FindingData{
				{ControlID: "ac-1", Title: "Control AC-1 is not satisfied"},
				{ControlID: "ac-2", Title: "Control AC-2 is not satisfied"},
			},
		},
		{
			name: "mixed satisfied and not satisfied",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				AssessmentResults: &oscalTypes.AssessmentResults{
					Results: []oscalTypes.Result{
						{
							Findings: &[]oscalTypes.Finding{
								{
									Title: "Control AC-1 is satisfied",
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-1",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateSatisfied,
										},
									},
								},
								{
									Title: "Control AC-2 is not satisfied",
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-2",
										Status: oscalTypes.ObjectiveStatus{
											State: oscal.StateNotSatisfied,
										},
									},
								},
								{
									Title: "Control AC-3 is satisfied",
									Target: oscalTypes.FindingTarget{
										TargetId: "ac-3",
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
			expectedSatisfied:    []string{"ac-1", "ac-3"},
			expectedNotSatisfied: []string{"ac-2"},
			expectedFindings: []FindingData{
				{ControlID: "ac-2", Title: "Control AC-2 is not satisfied"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			satisfied, notSatisfied, findings := extractControlsAndFindings(tt.result)
			assert.Equal(t, tt.expectedSatisfied, satisfied)
			assert.Equal(t, tt.expectedNotSatisfied, notSatisfied)
			assert.Equal(t, tt.expectedFindings, findings)
		})
	}
}

func TestBuildModuleReportData(t *testing.T) {
	tests := []struct {
		name     string
		config   *AssessConfig
		result   *ModuleAssessmentResult
		validate func(t *testing.T, data ModuleReportData)
	}{
		{
			name: "basic module data",
			config: &AssessConfig{
				WorkspaceRoot: "/workspace",
				OutputDir:     "/workspace/out",
			},
			result: &ModuleAssessmentResult{
				Module:       "test-module",
				Satisfied:    5,
				NotSatisfied: 2,
				RiskScore:    nil,
				OutputPath:   "/workspace/out/module-report.md",
			},
			validate: func(t *testing.T, data ModuleReportData) {
				assert.Equal(t, "test-module", data.Module)
				assert.Equal(t, 5, data.Satisfied)
				assert.Equal(t, 2, data.NotSatisfied)
				assert.Equal(t, "N/A", data.RiskScoreFormatted)
				// Normalize path separators for cross-platform testing
				assert.Equal(t, "out/module-report.md", filepath.ToSlash(data.ReportPath))
			},
		},
		{
			name: "module with risk score",
			config: &AssessConfig{
				WorkspaceRoot: "/workspace",
				OutputDir:     "/workspace/out",
			},
			result: &ModuleAssessmentResult{
				Module:       "test-module",
				Satisfied:    3,
				NotSatisfied: 5,
				RiskScore: &scoring.RiskScore{
					Band:       scoring.RiskHigh,
					Likelihood: 4, // Likely
					Impact:     4, // High
					Score:      16,
					Reasoning:  "Multiple critical controls not satisfied",
				},
				OutputPath: "/workspace/out/module-report.md",
			},
			validate: func(t *testing.T, data ModuleReportData) {
				assert.Equal(t, "test-module", data.Module)
				assert.Equal(t, 3, data.Satisfied)
				assert.Equal(t, 5, data.NotSatisfied)
				assert.NotNil(t, data.RiskScore)
				assert.Equal(t, scoring.RiskHigh, data.RiskScore.Band)
				assert.Contains(t, data.RiskScoreFormatted, "High")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildModuleReportData(tt.config, tt.result)
			tt.validate(t, data)
		})
	}
}

func TestBuildReportData(t *testing.T) {
	config := &AssessConfig{
		WorkspaceRoot: "/workspace",
		OutputDir:     "/workspace/out",
		ProfilePath:   "/workspace/specs/.risk-controls/risk-profile.json",
	}

	results := []*ModuleAssessmentResult{
		{
			Module:       "module-1",
			Satisfied:    10,
			NotSatisfied: 2,
			RiskScore: &scoring.RiskScore{
				Band:       scoring.RiskMedium,
				Likelihood: 3, // Possible
				Impact:     3, // Moderate
				Score:      9,
				Reasoning:  "Some controls not satisfied",
			},
			OutputPath: "/workspace/out/module-1.md",
		},
		{
			Module:       "module-2",
			Satisfied:    8,
			NotSatisfied: 0,
			RiskScore:    nil,
			OutputPath:   "/workspace/out/module-2.md",
		},
	}

	profile := &oscalTypes.Profile{
		Metadata: oscalTypes.Metadata{
			Title: "Test Profile",
		},
	}

	t.Run("full assessment", func(t *testing.T) {
		data := buildReportData(config, results, profile, false, 2, nil)

		assert.NotEmpty(t, data.GeneratedAt)
		assert.Equal(t, "Full Assessment (all 2 modules)", data.ScopeDescription)
		assert.Equal(t, "risk-profile.json", data.ProfileName)
		assert.Equal(t, "all", data.TestSuite)

		assert.Equal(t, 2, data.Summary.ModulesAssessed)
		assert.Equal(t, 20, data.Summary.TotalControls)
		assert.Equal(t, 18, data.Summary.Satisfied)
		assert.Equal(t, 2, data.Summary.NotSatisfied)
		assert.InDelta(t, 90.0, data.Summary.SatisfactionRate, 0.01)
		assert.InDelta(t, 10.0, data.Summary.NotSatisfiedRate, 0.01)
		assert.False(t, data.Summary.HasAISummary)

		require.Len(t, data.ModuleResults, 2)
		assert.Equal(t, "module-1", data.ModuleResults[0].Module)
		assert.Equal(t, "module-2", data.ModuleResults[1].Module)
	})

	t.Run("subset assessment", func(t *testing.T) {
		data := buildReportData(config, results, profile, true, 10, nil)

		assert.Equal(t, "Subset Assessment (2 of 10 modules)", data.ScopeDescription)
		assert.Equal(t, 2, data.Summary.ModulesAssessed)
		assert.False(t, data.Summary.HasAISummary)
	})
}
