package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSuite(t *testing.T) {
	suite, err := GetSuite("commit")
	require.NoError(t, err)
	assert.Equal(t, "commit", suite.Moniker)
	assert.Equal(t, "Commit Tests", suite.Name)
}

func TestGetSuite_NotFound(t *testing.T) {
	_, err := GetSuite("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suite not found")
}

func TestListSuites(t *testing.T) {
	suites := ListSuites()

	assert.Contains(t, suites, "commit")
	assert.Contains(t, suites, "acceptance")
	assert.Contains(t, suites, "production-verification")
	assert.Len(t, suites, 3)
}

func TestSelectTests_PreCommit(t *testing.T) {
	suite, _ := GetSuite("commit")

	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L0", "@ov"}},
		{TestName: "Test B", Tags: []string{"@L1", "@ov"}},
		{TestName: "Test C", Tags: []string{"@L2", "@ov"}},
		{TestName: "Test D", Tags: []string{"@L3", "@iv"}},
	}

	selected := suite.SelectTests(tests)

	assert.Len(t, selected, 3)

	names := []string{selected[0].TestName, selected[1].TestName, selected[2].TestName}
	assert.Contains(t, names, "Test A")
	assert.Contains(t, names, "Test B")
	assert.Contains(t, names, "Test C")
	assert.NotContains(t, names, "Test D")
}

func TestSelectTests_Acceptance(t *testing.T) {
	suite, _ := GetSuite("acceptance")

	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L3", "@iv"}},
		{TestName: "Test B", Tags: []string{"@L3", "@ov"}},
		{TestName: "Test C", Tags: []string{"@L3", "@pv"}},
		{TestName: "Test D", Tags: []string{"@L4", "@piv"}},
	}

	selected := suite.SelectTests(tests)

	assert.Len(t, selected, 3)

	names := []string{selected[0].TestName, selected[1].TestName, selected[2].TestName}
	assert.Contains(t, names, "Test A")
	assert.Contains(t, names, "Test B")
	assert.Contains(t, names, "Test C")
	assert.NotContains(t, names, "Test D")
}

func TestSelectTests_ProductionVerification(t *testing.T) {
	suite, _ := GetSuite("production-verification")

	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L4", "@piv"}},
		{TestName: "Test B", Tags: []string{"@L4", "@ov"}},
		{TestName: "Test C", Tags: []string{"@L3", "@piv"}},
	}

	selected := suite.SelectTests(tests)

	// Only Test A has both @L4 AND @piv
	assert.Len(t, selected, 1)
	assert.Equal(t, "Test A", selected[0].TestName)
}

func TestMatchesSelector_AnyOfTags(t *testing.T) {
	selector := TagSelector{
		AnyOfTags: []string{"@L0", "@L1", "@L2"},
	}

	assert.True(t, matchesSelector([]string{"@L0", "@ov"}, selector))
	assert.True(t, matchesSelector([]string{"@L1", "@ov"}, selector))
	assert.True(t, matchesSelector([]string{"@L2", "@ov"}, selector))
	assert.False(t, matchesSelector([]string{"@L3", "@iv"}, selector))
}

func TestMatchesSelector_RequireTags(t *testing.T) {
	selector := TagSelector{
		RequireTags: []string{"@L4", "@piv"},
	}

	assert.True(t, matchesSelector([]string{"@L4", "@piv"}, selector))
	assert.False(t, matchesSelector([]string{"@L4", "@ov"}, selector))
	assert.False(t, matchesSelector([]string{"@L3", "@piv"}, selector))
}

func TestMatchesSelector_ExcludeTags(t *testing.T) {
	selector := TagSelector{
		AnyOfTags:   []string{"@L1"},
		ExcludeTags: []string{"@deps:docker"},
	}

	assert.True(t, matchesSelector([]string{"@L1", "@ov"}, selector))
	assert.False(t, matchesSelector([]string{"@L1", "@deps:docker"}, selector))
}

func TestGetSystemDependencies(t *testing.T) {
	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L1", "@deps:go"}},
		{TestName: "Test B", Tags: []string{"@L2", "@deps:docker"}},
		{TestName: "Test C", Tags: []string{"@L3", "@deps:git"}},
	}

	deps := GetSystemDependencies(tests)

	assert.Len(t, deps, 3)
	assert.Contains(t, deps, "@deps:go")
	assert.Contains(t, deps, "@deps:docker")
	assert.Contains(t, deps, "@deps:git")
}

func TestGetSystemDependencies_NoDuplicates(t *testing.T) {
	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L1", "@deps:go"}},
		{TestName: "Test B", Tags: []string{"@L1", "@deps:go", "@deps:docker"}},
		{TestName: "Test C", Tags: []string{"@L2", "@deps:docker"}},
	}

	deps := GetSystemDependencies(tests)

	assert.Len(t, deps, 2)
	assert.Contains(t, deps, "@deps:go")
	assert.Contains(t, deps, "@deps:docker")
}

func TestGetSystemDependencies_NoDependencies(t *testing.T) {
	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L1", "@ov"}},
		{TestName: "Test B", Tags: []string{"@L2", "@ov"}},
	}

	deps := GetSystemDependencies(tests)

	assert.Len(t, deps, 0)
}

// Test @ignore filtering
func TestSelectTests_IgnoredTestsExcluded(t *testing.T) {
	suite, _ := GetSuite("commit")

	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L1", "@ov"}, IsIgnored: false},
		{TestName: "Test B", Tags: []string{"@L1", "@ov"}, IsIgnored: true},
		{TestName: "Test C", Tags: []string{"@L2", "@ov"}, IsIgnored: false},
	}

	selected := suite.SelectTests(tests)

	assert.Len(t, selected, 2)
	names := []string{selected[0].TestName, selected[1].TestName}
	assert.Contains(t, names, "Test A")
	assert.Contains(t, names, "Test C")
	assert.NotContains(t, names, "Test B")
}

func TestSelectTests_IgnoredBeforeOtherSelection(t *testing.T) {
	suite, _ := GetSuite("commit")

	tests := []TestReference{
		// This test matches the suite criteria (@L1) but is ignored
		{TestName: "Ignored L1", Tags: []string{"@L1", "@ov"}, IsIgnored: true},
		// This test doesn't match suite criteria (@L3) and is NOT ignored
		{TestName: "Non-matching L3", Tags: []string{"@L3", "@iv"}, IsIgnored: false},
		// This test matches suite criteria (@L2) and is NOT ignored
		{TestName: "Matching L2", Tags: []string{"@L2", "@ov"}, IsIgnored: false},
	}

	selected := suite.SelectTests(tests)

	// Only "Matching L2" should be selected
	// "Ignored L1" is filtered out FIRST (even though it matches suite criteria)
	// "Non-matching L3" doesn't match suite criteria
	assert.Len(t, selected, 1)
	assert.Equal(t, "Matching L2", selected[0].TestName)
}

// Test GetManualTests helper
func TestGetManualTests_FiltersCorrectly(t *testing.T) {
	tests := []TestReference{
		{TestName: "Manual test 1", IsManual: true},
		{TestName: "Automated test", IsManual: false},
		{TestName: "Manual test 2", IsManual: true},
		{TestName: "Another automated", IsManual: false},
	}

	manualTests := GetManualTests(tests)

	assert.Len(t, manualTests, 2)
	assert.Equal(t, "Manual test 1", manualTests[0].TestName)
	assert.Equal(t, "Manual test 2", manualTests[1].TestName)
}

// Test GetGxPTests helper
func TestGetGxPTests_FiltersCorrectly(t *testing.T) {
	tests := []TestReference{
		{TestName: "GxP test 1", IsGxP: true},
		{TestName: "Regular test", IsGxP: false},
		{TestName: "GxP test 2", IsGxP: true},
	}

	gxpTests := GetGxPTests(tests)

	assert.Len(t, gxpTests, 2)
	assert.Equal(t, "GxP test 1", gxpTests[0].TestName)
	assert.Equal(t, "GxP test 2", gxpTests[1].TestName)
}
