package ghost

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildReport_EmptyRepository(t *testing.T) {
	source := newMockFileSource([]string{})

	report, err := BuildReport(source, nil, "ghost")

	require.NoError(t, err)
	assert.Len(t, report.Ghosts, 0)
	assert.Equal(t, 0, report.Summary.Total)
	assert.Equal(t, 0, report.Summary.Files)
	assert.Equal(t, 0, report.Summary.Directories)
	assert.Equal(t, 0, report.Summary.Unowned)
	assert.Equal(t, "ghost", report.Config.Alias)
}

func TestBuildReport_WithGhosts(t *testing.T) {
	source := newMockFileSource([]string{
		"ghost-feature.go",
		"ghost-monitoring/probe.go",
		"src/ghost-api/handler.go",
		"src/normal.go",
	})

	report, err := BuildReport(source, nil, "ghost")

	require.NoError(t, err)
	assert.Equal(t, 3, report.Summary.Total)
	assert.Equal(t, 1, report.Summary.Files)
	assert.Equal(t, 2, report.Summary.Directories)
	assert.Equal(t, 3, report.Summary.Unowned) // No registry, so all unowned
}

func TestBuildReport_CustomAlias(t *testing.T) {
	source := newMockFileSource([]string{
		"hidden-feature.go",
		"ghost-other.go", // Should NOT match
	})

	report, err := BuildReport(source, nil, "hidden")

	require.NoError(t, err)
	assert.Len(t, report.Ghosts, 1)
	assert.Equal(t, "hidden", report.Config.Alias)
	assert.Equal(t, []string{"hidden-*", "hidden.*", "hidden"}, report.Config.Patterns)
}

func TestBuildReport_DefaultAlias(t *testing.T) {
	source := newMockFileSource([]string{
		"ghost-feature.go",
	})

	// Empty alias should default to "ghost"
	report, err := BuildReport(source, nil, "")

	require.NoError(t, err)
	assert.Equal(t, "ghost", report.Config.Alias)
}

func TestBuildReport_SummaryByModule(t *testing.T) {
	// Note: Without a module registry, all ghosts are unowned
	// This test validates the structure is correct
	source := newMockFileSource([]string{
		"ghost-feature.go",
		"ghost-other.go",
	})

	report, err := BuildReport(source, nil, "ghost")

	require.NoError(t, err)
	assert.NotNil(t, report.Summary.ByModule)
	assert.Equal(t, 2, report.Summary.Unowned)
}

func TestFilter_ByType(t *testing.T) {
	ghosts := []Ghost{
		{Path: "ghost-file.go", Type: GhostTypeFile},
		{Path: "ghost-dir", Type: GhostTypeDirectory},
		{Path: "ghost-other.go", Type: GhostTypeFile},
	}

	filtered := Filter(ghosts, FilterOptions{Type: "file"})

	assert.Len(t, filtered, 2)
	for _, g := range filtered {
		assert.Equal(t, GhostTypeFile, g.Type)
	}
}

func TestFilter_ByModule(t *testing.T) {
	ghosts := []Ghost{
		{Path: "a.go", Module: "core"},
		{Path: "b.go", Module: "docs"},
		{Path: "c.go", Module: "core"},
	}

	filtered := Filter(ghosts, FilterOptions{Module: "core"})

	assert.Len(t, filtered, 2)
	for _, g := range filtered {
		assert.Equal(t, "core", g.Module)
	}
}

func TestFilter_Unowned(t *testing.T) {
	ghosts := []Ghost{
		{Path: "a.go", Module: "core"},
		{Path: "b.go", Module: ""},
		{Path: "c.go", Module: ""},
	}

	filtered := Filter(ghosts, FilterOptions{Unowned: true})

	assert.Len(t, filtered, 2)
	for _, g := range filtered {
		assert.Empty(t, g.Module)
	}
}

func TestFilter_NoOptions(t *testing.T) {
	ghosts := []Ghost{
		{Path: "a.go"},
		{Path: "b.go"},
	}

	filtered := Filter(ghosts, FilterOptions{})

	assert.Equal(t, ghosts, filtered)
}
