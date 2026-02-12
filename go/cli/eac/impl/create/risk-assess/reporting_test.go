package riskassess

import (
	"testing"

	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/evidence"
	"github.com/stretchr/testify/assert"
)

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		total    int
		expected string
	}{
		{
			name:     "zero total returns 0.0",
			value:    0,
			total:    0,
			expected: "0.0",
		},
		{
			name:     "zero value returns 0.0",
			value:    0,
			total:    10,
			expected: "0.0",
		},
		{
			name:     "full value returns 100.0",
			value:    10,
			total:    10,
			expected: "100.0",
		},
		{
			name:     "half value returns 50.0",
			value:    5,
			total:    10,
			expected: "50.0",
		},
		{
			name:     "one third",
			value:    1,
			total:    3,
			expected: "33.3",
		},
		{
			name:     "two thirds",
			value:    2,
			total:    3,
			expected: "66.7",
		},
		{
			name:     "small fraction",
			value:    1,
			total:    100,
			expected: "1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPercentage(tt.value, tt.total)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractManifestDetails(t *testing.T) {
	tests := []struct {
		name                  string
		result                *ModuleAssessmentResult
		expectedTypeBreakdown string
		validateSuiteSummary  func(t *testing.T, summary string)
	}{
		{
			name: "nil evidence returns N/A",
			result: &ModuleAssessmentResult{
				Module:   "test-module",
				Evidence: nil,
			},
			expectedTypeBreakdown: "N/A",
			validateSuiteSummary: func(t *testing.T, summary string) {
				assert.Equal(t, "N/A", summary)
			},
		},
		{
			name: "nil test manifest returns N/A",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				Evidence: &evidence.EvidenceCollection{
					Module:           "test-module",
					TestManifestData: nil,
				},
			},
			expectedTypeBreakdown: "N/A",
			validateSuiteSummary: func(t *testing.T, summary string) {
				assert.Equal(t, "N/A", summary)
			},
		},
		{
			name: "empty tests and suites returns N/A",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				Evidence: &evidence.EvidenceCollection{
					Module: "test-module",
					TestManifestData: &evidence.TestManifestData{
						Tests:  []evidence.TestEntryData{},
						Suites: map[string]evidence.SuiteResultData{},
					},
				},
			},
			expectedTypeBreakdown: "N/A",
			validateSuiteSummary: func(t *testing.T, summary string) {
				assert.Equal(t, "N/A", summary)
			},
		},
		{
			name: "single test type",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				Evidence: &evidence.EvidenceCollection{
					Module: "test-module",
					TestManifestData: &evidence.TestManifestData{
						Tests: []evidence.TestEntryData{
							{Name: "Test1", Type: "gotest"},
							{Name: "Test2", Type: "gotest"},
							{Name: "Test3", Type: "gotest"},
						},
					},
				},
			},
			expectedTypeBreakdown: "3 gotest",
			validateSuiteSummary: func(t *testing.T, summary string) {
				assert.Equal(t, "N/A", summary)
			},
		},
		{
			name: "multiple test types",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				Evidence: &evidence.EvidenceCollection{
					Module: "test-module",
					TestManifestData: &evidence.TestManifestData{
						Tests: []evidence.TestEntryData{
							{Name: "Test1", Type: "gotest"},
							{Name: "Test2", Type: "gotest"},
							{Name: "Test3", Type: "godog"},
							{Name: "Test4", Type: "mocha"},
						},
					},
				},
			},
			// Cannot check exact string because map ordering is non-deterministic
			expectedTypeBreakdown: "",
			validateSuiteSummary: func(t *testing.T, summary string) {
				assert.Equal(t, "N/A", summary)
			},
		},
		{
			name: "suites with results",
			result: &ModuleAssessmentResult{
				Module: "test-module",
				Evidence: &evidence.EvidenceCollection{
					Module: "test-module",
					TestManifestData: &evidence.TestManifestData{
						Tests: []evidence.TestEntryData{},
						Suites: map[string]evidence.SuiteResultData{
							"unit": {
								Total:  20,
								Passed: 18,
								Failed: 2,
							},
						},
					},
				},
			},
			expectedTypeBreakdown: "N/A",
			validateSuiteSummary: func(t *testing.T, summary string) {
				assert.Contains(t, summary, "unit")
				assert.Contains(t, summary, "18/20 passed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeBreakdown, suiteSummary := extractManifestDetails(tt.result)

			if tt.expectedTypeBreakdown != "" {
				assert.Equal(t, tt.expectedTypeBreakdown, typeBreakdown)
			} else {
				// For non-deterministic map ordering, verify content
				assert.NotEqual(t, "N/A", typeBreakdown)
				assert.Contains(t, typeBreakdown, "gotest")
				assert.Contains(t, typeBreakdown, "godog")
				assert.Contains(t, typeBreakdown, "mocha")
			}

			tt.validateSuiteSummary(t, suiteSummary)
		})
	}
}

func TestExtractManifestDetails_MultipleTestTypes_Content(t *testing.T) {
	// Separate test for verifying content of multi-type breakdown
	// since map iteration order is non-deterministic
	result := &ModuleAssessmentResult{
		Module: "test-module",
		Evidence: &evidence.EvidenceCollection{
			Module: "test-module",
			TestManifestData: &evidence.TestManifestData{
				Tests: []evidence.TestEntryData{
					{Name: "Test1", Type: "gotest"},
					{Name: "Test2", Type: "gotest"},
					{Name: "Test3", Type: "gotest"},
					{Name: "Test4", Type: "godog"},
					{Name: "Test5", Type: "godog"},
				},
			},
		},
	}

	typeBreakdown, _ := extractManifestDetails(result)

	assert.Contains(t, typeBreakdown, "3 gotest")
	assert.Contains(t, typeBreakdown, "2 godog")
	assert.Contains(t, typeBreakdown, ", ")
}

func TestExtractManifestDetails_MultipleSuites(t *testing.T) {
	result := &ModuleAssessmentResult{
		Module: "test-module",
		Evidence: &evidence.EvidenceCollection{
			Module: "test-module",
			TestManifestData: &evidence.TestManifestData{
				Tests: []evidence.TestEntryData{},
				Suites: map[string]evidence.SuiteResultData{
					"unit": {
						Total:  10,
						Passed: 10,
						Failed: 0,
					},
					"integration": {
						Total:  5,
						Passed: 3,
						Failed: 2,
					},
				},
			},
		},
	}

	_, suiteSummary := extractManifestDetails(result)

	assert.Contains(t, suiteSummary, "unit")
	assert.Contains(t, suiteSummary, "10/10 passed")
	assert.Contains(t, suiteSummary, "integration")
	assert.Contains(t, suiteSummary, "3/5 passed")
}
