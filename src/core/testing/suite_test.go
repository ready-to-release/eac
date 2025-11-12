package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSuite(t *testing.T) {
	suite, err := GetSuite("pre-commit")
	require.NoError(t, err)
	assert.Equal(t, "pre-commit", suite.Moniker)
	assert.Equal(t, "Pre-Commit Tests", suite.Name)
}

func TestGetSuite_NotFound(t *testing.T) {
	_, err := GetSuite("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suite not found")
}

func TestListSuites(t *testing.T) {
	suites := ListSuites()

	assert.Contains(t, suites, "pre-commit")
	assert.Contains(t, suites, "acceptance")
	assert.Contains(t, suites, "production-verification")
	assert.Len(t, suites, 3)
}

func TestSelectTests_PreCommit(t *testing.T) {
	suite, _ := GetSuite("pre-commit")

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
		ExcludeTags: []string{"@dep:docker"},
	}

	assert.True(t, matchesSelector([]string{"@L1", "@ov"}, selector))
	assert.False(t, matchesSelector([]string{"@L1", "@dep:docker"}, selector))
}

func TestGetSystemDependencies(t *testing.T) {
	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L1", "@dep:go"}},
		{TestName: "Test B", Tags: []string{"@L2", "@dep:docker"}},
		{TestName: "Test C", Tags: []string{"@L3", "@dep:git"}},
	}

	deps := GetSystemDependencies(tests)

	assert.Len(t, deps, 3)
	assert.Contains(t, deps, "@dep:go")
	assert.Contains(t, deps, "@dep:docker")
	assert.Contains(t, deps, "@dep:git")
}

func TestGetSystemDependencies_NoDuplicates(t *testing.T) {
	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L1", "@dep:go"}},
		{TestName: "Test B", Tags: []string{"@L1", "@dep:go", "@dep:docker"}},
		{TestName: "Test C", Tags: []string{"@L2", "@dep:docker"}},
	}

	deps := GetSystemDependencies(tests)

	assert.Len(t, deps, 2)
	assert.Contains(t, deps, "@dep:go")
	assert.Contains(t, deps, "@dep:docker")
}

func TestGetSystemDependencies_NoDependencies(t *testing.T) {
	tests := []TestReference{
		{TestName: "Test A", Tags: []string{"@L1", "@ov"}},
		{TestName: "Test B", Tags: []string{"@L2", "@ov"}},
	}

	deps := GetSystemDependencies(tests)

	assert.Len(t, deps, 0)
}
