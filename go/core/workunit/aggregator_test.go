package workunit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUoWAggregator(t *testing.T) {
	t.Run("counts UoWs per module", func(t *testing.T) {
		expectedUoWs := []UnitID{
			{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"},
			{Action: ActionTest, Module: "cli", Component: "go", Tool: "gotest"},
		}

		agg := NewUoWAggregator(expectedUoWs)

		total, cached := agg.Stats("core")
		assert.Equal(t, 2, total, "core should have 2 UoWs")
		assert.Equal(t, 0, cached, "core should have 0 cached initially")

		total, cached = agg.Stats("cli")
		assert.Equal(t, 1, total, "cli should have 1 UoW")
		assert.Equal(t, 0, cached, "cli should have 0 cached initially")
	})

	t.Run("handles empty slice", func(t *testing.T) {
		agg := NewUoWAggregator([]UnitID{})

		total, cached := agg.Stats("nonexistent")
		assert.Equal(t, 0, total)
		assert.Equal(t, 0, cached)
	})
}

func TestUoWAggregator_MarkCached(t *testing.T) {
	t.Run("increments cached count for module", func(t *testing.T) {
		expectedUoWs := []UnitID{
			{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"},
		}

		agg := NewUoWAggregator(expectedUoWs)

		// Mark first UoW as cached
		agg.MarkCached(UnitID{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"})

		total, cached := agg.Stats("core")
		assert.Equal(t, 2, total)
		assert.Equal(t, 1, cached)

		// Mark second UoW as cached
		agg.MarkCached(UnitID{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"})

		total, cached = agg.Stats("core")
		assert.Equal(t, 2, total)
		assert.Equal(t, 2, cached)
	})
}

func TestUoWAggregator_IsModuleCached(t *testing.T) {
	t.Run("returns false when no UoWs cached", func(t *testing.T) {
		expectedUoWs := []UnitID{
			{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"},
		}

		agg := NewUoWAggregator(expectedUoWs)
		assert.False(t, agg.IsModuleCached("core"))
	})

	t.Run("returns false when only some UoWs cached", func(t *testing.T) {
		expectedUoWs := []UnitID{
			{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"},
		}

		agg := NewUoWAggregator(expectedUoWs)
		agg.MarkCached(UnitID{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"})

		assert.False(t, agg.IsModuleCached("core"), "module with partial UoWs cached should not be cached")
	})

	t.Run("returns true when all UoWs cached", func(t *testing.T) {
		expectedUoWs := []UnitID{
			{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"},
		}

		agg := NewUoWAggregator(expectedUoWs)
		agg.MarkCached(UnitID{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"})
		agg.MarkCached(UnitID{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"})

		assert.True(t, agg.IsModuleCached("core"), "module with all UoWs cached should be cached")
	})

	t.Run("returns false for unknown module", func(t *testing.T) {
		agg := NewUoWAggregator([]UnitID{})
		assert.False(t, agg.IsModuleCached("nonexistent"))
	})
}

func TestUoWAggregator_GetModuleLists(t *testing.T) {
	t.Run("separates modules into changed and up-to-date", func(t *testing.T) {
		expectedUoWs := []UnitID{
			{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"},
			{Action: ActionTest, Module: "cli", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "api", Component: "go", Tool: "gotest"},
		}

		agg := NewUoWAggregator(expectedUoWs)

		// Mark all core UoWs as cached
		agg.MarkCached(UnitID{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"})
		agg.MarkCached(UnitID{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"})

		// Mark cli UoW as cached (only one, so module is cached)
		agg.MarkCached(UnitID{Action: ActionTest, Module: "cli", Component: "go", Tool: "gotest"})

		// api has no cached UoWs

		changed, upToDate := agg.GetModuleLists(expectedUoWs)

		assert.ElementsMatch(t, []string{"api"}, changed)
		assert.ElementsMatch(t, []string{"core", "cli"}, upToDate)
	})

	t.Run("preserves order from expectedUoWs", func(t *testing.T) {
		expectedUoWs := []UnitID{
			{Action: ActionTest, Module: "zzz", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "aaa", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "mmm", Component: "go", Tool: "gotest"},
		}

		agg := NewUoWAggregator(expectedUoWs)

		changed, upToDate := agg.GetModuleLists(expectedUoWs)

		// All are changed (none cached), order should match input
		assert.Equal(t, []string{"zzz", "aaa", "mmm"}, changed)
		assert.Empty(t, upToDate)
	})

	t.Run("handles empty input", func(t *testing.T) {
		agg := NewUoWAggregator([]UnitID{})

		changed, upToDate := agg.GetModuleLists([]UnitID{})

		assert.Empty(t, changed)
		assert.Empty(t, upToDate)
	})

	t.Run("deduplicates modules", func(t *testing.T) {
		expectedUoWs := []UnitID{
			{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest"},
			{Action: ActionTest, Module: "core", Component: "gherkin", Tool: "godog"},
			{Action: ActionTest, Module: "core", Component: "ts", Tool: "jest"},
		}

		agg := NewUoWAggregator(expectedUoWs)

		changed, upToDate := agg.GetModuleLists(expectedUoWs)

		// Should only have "core" once in changed list
		assert.Equal(t, []string{"core"}, changed)
		assert.Empty(t, upToDate)
	})
}
