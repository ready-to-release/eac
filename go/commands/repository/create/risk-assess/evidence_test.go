package riskassess

import (
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/commands/repository/internal/risk/evidence"
	testview "github.com/ready-to-release/eac/go/core/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"seconds only", 30 * time.Second, "30 seconds"},
		{"one minute", 1 * time.Minute, "1 minutes"},
		{"minutes and seconds", 2*time.Minute + 30*time.Second, "2 minutes 30 seconds"},
		{"one hour", 1 * time.Hour, "1 hours"},
		{"hours and minutes", 3*time.Hour + 15*time.Minute, "3 hours 15 minutes"},
		{"one day", 24 * time.Hour, "1 days"},
		{"days and hours", 48*time.Hour + 6*time.Hour, "2 days 6 hours"},
		{"many days", 7 * 24 * time.Hour, "7 days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertViewToManifestData(t *testing.T) {
	t.Run("nil view returns nil", func(t *testing.T) {
		result := convertViewToManifestData(nil)
		assert.Nil(t, result)
	})

	t.Run("converts tests and suites", func(t *testing.T) {
		execTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
		view := &testview.TestModuleView{
			Module:     "my-module",
			ExecutedAt: execTime,
			Duration:   45 * time.Second,
			Summary: testview.TestSummary{
				Total:   3,
				Passed:  2,
				Failed:  1,
				Skipped: 0,
			},
			Tests: []testview.TestEntry{
				{
					Name:     "TestFoo",
					Package:  "pkg/foo",
					Type:     "gotest",
					Suite:    "L1",
					Status:   "passed",
					Tags:     []string{"@control:ac-1"},
					FilePath: "foo_test.go",
				},
				{
					Name:   "TestBar",
					Status: "failed",
				},
			},
			Suites: map[string]*testview.SuiteSummary{
				"L1": {
					RunTime: execTime,
					Total:   2,
					Passed:  1,
					Failed:  1,
				},
			},
			Artifacts: []testview.ArtifactRef{
				{Type: "report", Path: "/out/test/my-module/report.json"},
			},
		}

		result := convertViewToManifestData(view)
		require.NotNil(t, result)

		assert.Equal(t, execTime, result.TestTime)
		assert.Len(t, result.Tests, 2)

		// Verify first test entry fields
		assert.Equal(t, "TestFoo", result.Tests[0].Name)
		assert.Equal(t, "pkg/foo", result.Tests[0].Package)
		assert.Equal(t, "gotest", result.Tests[0].Type)
		assert.Equal(t, "L1", result.Tests[0].Suite)
		assert.Equal(t, "passed", result.Tests[0].Status)
		assert.Equal(t, []string{"@control:ac-1"}, result.Tests[0].Tags)

		// Verify suite conversion
		require.Contains(t, result.Suites, "L1")
		assert.Equal(t, 2, result.Suites["L1"].Total)
		assert.Equal(t, 1, result.Suites["L1"].Passed)

		// Verify artifacts
		assert.Len(t, result.Artifacts, 1)
		assert.Equal(t, "report", result.Artifacts[0].Type)
		assert.Equal(t, "report.json", result.Artifacts[0].Name)
	})

	t.Run("handles empty collections", func(t *testing.T) {
		view := &testview.TestModuleView{
			Module:    "empty-module",
			Tests:     []testview.TestEntry{},
			Suites:    map[string]*testview.SuiteSummary{},
			Artifacts: []testview.ArtifactRef{},
		}

		result := convertViewToManifestData(view)
		require.NotNil(t, result)
		assert.Len(t, result.Tests, 0)
		assert.Len(t, result.Suites, 0)
		assert.Len(t, result.Artifacts, 0)
	})
}

func TestModuleAssessmentResult_EvidenceAccessors(t *testing.T) {
	t.Run("evidence nil", func(t *testing.T) {
		r := &ModuleAssessmentResult{
			Module:   "test",
			Evidence: nil,
		}
		assert.Nil(t, r.Evidence)
	})

	t.Run("evidence with test summary", func(t *testing.T) {
		r := &ModuleAssessmentResult{
			Module: "test",
			Evidence: &evidence.EvidenceCollection{
				Module:      "test",
				CollectedAt: time.Now(),
				TestSummary: &evidence.TestSummary{
					Total:  10,
					Passed: 8,
					Failed: 2,
				},
			},
		}
		require.NotNil(t, r.Evidence)
		require.NotNil(t, r.Evidence.TestSummary)
		assert.Equal(t, 10, r.Evidence.TestSummary.Total)
		assert.Equal(t, 8, r.Evidence.TestSummary.Passed)
		assert.Equal(t, 2, r.Evidence.TestSummary.Failed)
	})
}
