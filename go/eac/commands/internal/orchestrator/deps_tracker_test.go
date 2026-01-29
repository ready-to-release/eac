package orchestrator

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/workunit"
	"github.com/stretchr/testify/assert"
)

func TestNewDepsTracker(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}},
	}

	dt := NewDepsTracker(work)

	assert.NotNil(t, dt)
	assert.NotNil(t, dt.completed)
	assert.NotNil(t, dt.depsOf)
	assert.NotNil(t, dt.moduleOf)
}

func TestDepsTracker_IsReady_NoDeps(t *testing.T) {
	// Component with no dependencies should always be ready
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
	}

	dt := NewDepsTracker(work)

	assert.True(t, dt.IsReady(work[0].ID), "component with no deps should be ready")
}

func TestDepsTracker_IsReady_WithUnmetDeps(t *testing.T) {
	// Component depends on another that hasn't completed
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDepsTracker(work)

	assert.True(t, dt.IsReady(work[0].ID), "comp1 has no deps, should be ready")
	assert.False(t, dt.IsReady(work[1].ID), "comp2 depends on comp1, should not be ready")
}

func TestDepsTracker_IsReady_AfterDepCompletes(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDepsTracker(work)

	// Initially comp2 not ready
	assert.False(t, dt.IsReady(work[1].ID))

	// Mark comp1 complete
	dt.MarkComplete(work[0].ID)

	// Now comp2 should be ready
	assert.True(t, dt.IsReady(work[1].ID), "comp2 should be ready after comp1 completes")
}

func TestDepsTracker_IsReady_MultipleDeps(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}},
		{
			ID: workunit.UnitID{Module: "mod1", Component: "comp3"},
			DependsOn: []workunit.UnitID{
				{Module: "mod1", Component: "comp1"},
				{Module: "mod1", Component: "comp2"},
			},
		},
	}

	dt := NewDepsTracker(work)

	// comp3 needs both comp1 and comp2
	assert.False(t, dt.IsReady(work[2].ID))

	// Complete comp1 only
	dt.MarkComplete(work[0].ID)
	assert.False(t, dt.IsReady(work[2].ID), "still waiting on comp2")

	// Complete comp2
	dt.MarkComplete(work[1].ID)
	assert.True(t, dt.IsReady(work[2].ID), "both deps complete, should be ready")
}

func TestDepsTracker_IsReady_CrossModuleDepsIgnored(t *testing.T) {
	// Dependencies are intra-module only
	// A component in mod1 depending on "comp1" only looks at mod1:comp1
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{ID: workunit.UnitID{Module: "mod2", Component: "comp1"}}, // same component name, different module
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDepsTracker(work)

	// comp2 in mod1 depends on comp1 in mod1
	assert.False(t, dt.IsReady(work[2].ID))

	// Complete mod2:comp1 (wrong module)
	dt.MarkComplete(work[1].ID)
	assert.False(t, dt.IsReady(work[2].ID), "mod2:comp1 doesn't satisfy mod1:comp2's dep")

	// Complete mod1:comp1
	dt.MarkComplete(work[0].ID)
	assert.True(t, dt.IsReady(work[2].ID), "mod1:comp1 completes satisfies the dep")
}

func TestDepsTracker_MarkComplete_Idempotent(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
	}

	dt := NewDepsTracker(work)

	// Multiple marks should not panic or cause issues
	dt.MarkComplete(work[0].ID)
	dt.MarkComplete(work[0].ID)
	dt.MarkComplete(work[0].ID)

	// Should still work
	assert.True(t, dt.completed["mod1:comp1"])
}

func TestDepsTracker_WithTool(t *testing.T) {
	// Units with tools should still resolve deps correctly
	// Dependencies are by component, not by tool
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1", Tool: "tool1"}},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1", Tool: "tool2"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2", Tool: "tool1"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDepsTracker(work)

	// comp2:tool1 depends on comp1 (any tool)
	assert.False(t, dt.IsReady(work[2].ID))

	// Complete comp1:tool1
	dt.MarkComplete(work[0].ID)
	assert.True(t, dt.IsReady(work[2].ID), "comp1:tool1 completing should satisfy dep on comp1")
}

func TestDepsTracker_ChainedDeps(t *testing.T) {
	// comp1 -> comp2 -> comp3
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp3"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp2"}},
		},
	}

	dt := NewDepsTracker(work)

	assert.True(t, dt.IsReady(work[0].ID))
	assert.False(t, dt.IsReady(work[1].ID))
	assert.False(t, dt.IsReady(work[2].ID))

	dt.MarkComplete(work[0].ID)
	assert.True(t, dt.IsReady(work[1].ID))
	assert.False(t, dt.IsReady(work[2].ID))

	dt.MarkComplete(work[1].ID)
	assert.True(t, dt.IsReady(work[2].ID))
}

func TestDepsTracker_BlockedCount(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp3"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDepsTracker(work)

	// Count blocked items
	blocked := 0
	for _, w := range work {
		if !dt.IsReady(w.ID) {
			blocked++
		}
	}
	assert.Equal(t, 2, blocked, "2 items blocked on comp1")

	// After comp1 completes
	dt.MarkComplete(work[0].ID)

	blocked = 0
	for _, w := range work {
		if !dt.IsReady(w.ID) {
			blocked++
		}
	}
	assert.Equal(t, 0, blocked, "no items blocked after comp1 completes")
}
