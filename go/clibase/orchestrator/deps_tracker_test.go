package orchestrator

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newDepsTrackerWithPolicy creates a DepsTracker with LayerPolicy configured.
// Uses LayerModeNone for testing (component layers + DependsOn enforced).
func newDepsTrackerWithPolicy(t *testing.T, work []workunit.UnitSpec) *DepsTracker {
	dt := NewDepsTracker(work)
	planner := execution.NewLayerPlanner(execution.LayerModeNone)
	plan, err := planner.ComputePlan(work)
	require.NoError(t, err, "failed to compute plan")
	dt.SetLayerPolicy(planner, plan)
	return dt
}

func TestNewDepsTracker(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}},
	}

	dt := NewDepsTracker(work)

	assert.NotNil(t, dt)
	assert.NotNil(t, dt.completed)
	assert.NotNil(t, dt.depsOf)
}

func TestDepsTracker_IsReady_NoDeps(t *testing.T) {
	// Component with no dependencies should always be ready
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
	}

	dt := newDepsTrackerWithPolicy(t, work)

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

	dt := newDepsTrackerWithPolicy(t, work)

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

	dt := newDepsTrackerWithPolicy(t, work)

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

	dt := newDepsTrackerWithPolicy(t, work)

	// comp3 needs both comp1 and comp2
	assert.False(t, dt.IsReady(work[2].ID))

	// Complete comp1 only
	dt.MarkComplete(work[0].ID)
	assert.False(t, dt.IsReady(work[2].ID), "still waiting on comp2")

	// Complete comp2
	dt.MarkComplete(work[1].ID)
	assert.True(t, dt.IsReady(work[2].ID), "both deps complete, should be ready")
}

func TestDepsTracker_IsReady_CrossModuleDeps(t *testing.T) {
	// Dependencies can be cross-module via DependsOn
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{ID: workunit.UnitID{Module: "mod2", Component: "comp1"}}, // same component name, different module
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := newDepsTrackerWithPolicy(t, work)

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

	assert.True(t, dt.completed[work[0].ID.Longname()])
}

func TestDepsTracker_WithTool(t *testing.T) {
	// Units with tools should resolve deps via Longname matching
	// Component layers are enforced: all units in layer 0 must complete before layer 1
	ctx := workunit.ContextBuild
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Context: ctx, Module: "mod1", Component: "comp1", Tool: "tool1"}},
		{ID: workunit.UnitID{Context: ctx, Module: "mod1", Component: "comp1", Tool: "tool2"}},
		{
			ID:        workunit.UnitID{Context: ctx, Module: "mod1", Component: "comp2", Tool: "tool1"},
			DependsOn: []workunit.UnitID{{Context: ctx, Module: "mod1", Component: "comp1", Tool: "tool1"}},
		},
	}

	dt := newDepsTrackerWithPolicy(t, work)

	// Layer 0: comp1:tool1, comp1:tool2 (no deps)
	// Layer 1: comp2:tool1 (depends on comp1:tool1)

	// Initially layer 0 is ready
	assert.True(t, dt.IsReady(work[0].ID), "comp1:tool1 should be ready (layer 0)")
	assert.True(t, dt.IsReady(work[1].ID), "comp1:tool2 should be ready (layer 0)")

	// comp2:tool1 is NOT ready (layer 0 not complete)
	assert.False(t, dt.IsReady(work[2].ID), "comp2:tool1 should not be ready (layer 0 incomplete)")

	// Complete comp1:tool1 only - layer 0 still not complete
	dt.MarkComplete(work[0].ID)
	assert.False(t, dt.IsReady(work[2].ID), "comp2:tool1 still not ready (comp1:tool2 not done)")

	// Complete comp1:tool2 - now layer 0 is complete
	dt.MarkComplete(work[1].ID)
	assert.True(t, dt.IsReady(work[2].ID), "comp2:tool1 should be ready after layer 0 completes")
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

	dt := newDepsTrackerWithPolicy(t, work)

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

	dt := newDepsTrackerWithPolicy(t, work)

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

func TestDepsTracker_MarkFailed_UnblocksDependents(t *testing.T) {
	// This test verifies the deadlock fix: when a dependency fails,
	// dependents become ready (so they can fail via HasFailedDependency).
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := newDepsTrackerWithPolicy(t, work)

	// comp2 blocked on comp1
	assert.False(t, dt.IsReady(work[1].ID))

	// Mark comp1 as FAILED (not complete)
	dt.MarkFailed(work[0].ID)

	// comp2 should now be ready (unblocked)
	assert.True(t, dt.IsReady(work[1].ID), "dependent should be ready after dep fails")

	// But HasFailedDependency should return true
	assert.True(t, dt.HasFailedDependency(work[1].ID), "should detect failed dependency")
}

func TestDepsTracker_MarkFailed_ChainedUnblock(t *testing.T) {
	// comp1 -> comp2 -> comp3: failing comp1 should unblock the chain
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

	dt := newDepsTrackerWithPolicy(t, work)

	// Initially only comp1 is ready
	assert.True(t, dt.IsReady(work[0].ID))
	assert.False(t, dt.IsReady(work[1].ID))
	assert.False(t, dt.IsReady(work[2].ID))

	// Fail comp1
	dt.MarkFailed(work[0].ID)

	// comp2 should be ready now (to fail)
	assert.True(t, dt.IsReady(work[1].ID))
	assert.True(t, dt.HasFailedDependency(work[1].ID))

	// comp3 still blocked on comp2
	assert.False(t, dt.IsReady(work[2].ID))

	// Fail comp2 (propagating the failure)
	dt.MarkFailed(work[1].ID)

	// Now comp3 is ready (to fail)
	assert.True(t, dt.IsReady(work[2].ID))
	assert.True(t, dt.HasFailedDependency(work[2].ID))
}
