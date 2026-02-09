package testview

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCompleteTestData_Empty(t *testing.T) {
	data := BuildCompleteTestData(nil)
	assert.Equal(t, 0, data.ModulesTested)
	assert.Equal(t, 0, data.TotalTests)
	assert.Empty(t, data.Tests)
}

func TestBuildCompleteTestData_SingleModule(t *testing.T) {
	views := []*TestModuleView{
		{
			Module:     "core",
			ExecutedAt: time.Now(),
			Duration:   10 * time.Second,
			Summary:    TestSummary{Total: 3, Passed: 2, Failed: 1},
			Tests: []TestEntry{
				{Name: "TestA", Module: "core", Type: "gotest", Suite: "unit", Status: StatusPassed},
				{Name: "TestB", Module: "core", Type: "gotest", Suite: "unit", Status: StatusPassed},
				{Name: "TestC", Module: "core", Type: "gotest", Suite: "unit", Status: StatusFailed},
			},
			Suites: map[string]*SuiteSummary{
				"unit": {Total: 3, Passed: 2, Failed: 1},
			},
		},
	}

	data := BuildCompleteTestData(views)

	assert.Equal(t, 1, data.ModulesTested)
	assert.Equal(t, 3, data.TotalTests)
	assert.Equal(t, 2, data.TotalPassed)
	assert.Equal(t, 1, data.TotalFailed)
	assert.Len(t, data.Tests, 3)
	assert.Len(t, data.ModuleStats, 1)
	assert.Equal(t, "core", data.ModuleStats[0].Module)
}

func TestBuildCompleteTestData_ControlSummary(t *testing.T) {
	views := []*TestModuleView{
		{
			Module: "core",
			Tests: []TestEntry{
				{Name: "Test auth", Module: "core", Type: "godog", Tags: []string{"@control:auth-1"}, Status: StatusPassed},
				{Name: "Test auth 2", Module: "core", Type: "godog", Tags: []string{"@control:auth-1"}, Status: StatusFailed},
				{Name: "Test sec", Module: "core", Type: "godog", Tags: []string{"@control:sec-1"}, Status: StatusPassed},
			},
			Summary: TestSummary{Total: 3, Passed: 2, Failed: 1},
			Suites:  map[string]*SuiteSummary{},
		},
	}

	data := BuildCompleteTestData(views)

	require.Len(t, data.ControlSummary, 2)
	// Sorted by control ID
	assert.Equal(t, "auth-1", data.ControlSummary[0].ControlID)
	assert.Equal(t, 2, data.ControlSummary[0].TestCount)
	assert.Equal(t, 1, data.ControlSummary[0].PassedCount)
	assert.Equal(t, 1, data.ControlSummary[0].FailedCount)
	assert.Equal(t, "sec-1", data.ControlSummary[1].ControlID)
}

func TestBuildCompleteTestData_SpecCoverage(t *testing.T) {
	views := []*TestModuleView{
		{
			Module: "mymod",
			Tests: []TestEntry{
				{
					Name:     "Scenario 1",
					Module:   "mymod",
					Type:     "godog",
					Status:   StatusPassed,
					FilePath: "specs/mymod/login/specification.feature",
					Tags:     []string{"@control:auth-1"},
				},
				{
					Name:     "Scenario 2",
					Module:   "mymod",
					Type:     "godog",
					Status:   StatusFailed,
					FilePath: "specs/mymod/login/specification.feature",
				},
			},
			Summary: TestSummary{Total: 2, Passed: 1, Failed: 1},
			Suites:  map[string]*SuiteSummary{},
		},
	}

	data := BuildCompleteTestData(views)

	require.Len(t, data.SpecCoverage, 1)
	assert.Equal(t, "login", data.SpecCoverage[0].FeatureName)
	assert.Equal(t, 2, data.SpecCoverage[0].ScenarioCount)
	assert.Equal(t, 1, data.SpecCoverage[0].PassedCount)
	assert.Equal(t, 1, data.SpecCoverage[0].FailedCount)
	assert.Contains(t, data.SpecCoverage[0].Controls, "auth-1")
}

func TestBuildCompleteTestData_TypeAndSuiteSummaries(t *testing.T) {
	views := []*TestModuleView{
		{
			Module: "core",
			Tests: []TestEntry{
				{Name: "T1", Type: "gotest", Suite: "unit", Status: StatusPassed},
				{Name: "T2", Type: "gotest", Suite: "unit", Status: StatusFailed},
				{Name: "S1", Type: "godog", Suite: "integration", Status: StatusPassed},
			},
			Summary: TestSummary{Total: 3, Passed: 2, Failed: 1},
			Suites:  map[string]*SuiteSummary{},
		},
	}

	data := BuildCompleteTestData(views)

	// Type summary
	require.Len(t, data.SummaryByType, 2)
	// Sorted alphabetically
	assert.Equal(t, "godog", data.SummaryByType[0].Type)
	assert.Equal(t, 1, data.SummaryByType[0].Count)
	assert.Equal(t, "gotest", data.SummaryByType[1].Type)
	assert.Equal(t, 2, data.SummaryByType[1].Count)

	// Suite summary
	require.Len(t, data.SummaryBySuite, 2)
	assert.Equal(t, "integration", data.SummaryBySuite[0].Suite)
	assert.Equal(t, "unit", data.SummaryBySuite[1].Suite)
}

func TestExtractControlTags(t *testing.T) {
	tags := []string{"@L0", "@control:auth-1", "@control:sec-2", "@other"}
	controls := ExtractControlTags(tags)
	assert.Equal(t, []string{"auth-1", "sec-2"}, controls)
}

func TestStripTagPrefixes(t *testing.T) {
	tags := []string{"@L0", "@control:auth-1", "no-prefix"}
	stripped := StripTagPrefixes(tags)
	assert.Equal(t, []string{"L0", "control:auth-1", "no-prefix"}, stripped)
}

func TestExtractFeatureName(t *testing.T) {
	assert.Equal(t, "login", ExtractFeatureName("specs/mymod/login/specification.feature"))
	assert.Equal(t, "create-design", ExtractFeatureName("specs/eac/create-design/specification.feature"))
	assert.Equal(t, "", ExtractFeatureName(""))
	assert.Equal(t, "", ExtractFeatureName("no/specs/here"))
}
