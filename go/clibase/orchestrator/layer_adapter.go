package orchestrator

import (
	"sync"

	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// LayerSchedulingAdapter adapts core LayerPolicy to orchestrator needs.
// It wraps a LayerPolicy and manages completion state, providing a simple
// interface for the work queue to check unit readiness.
//
// This adapter is the bridge between the domain (go/core/execution) and
// the CLI infrastructure (go/clibase/orchestrator). It ensures that
// scheduling policy lives in the domain while the orchestrator remains
// a thin adapter.
//
// Thread-safety: All methods are safe for concurrent use.
type LayerSchedulingAdapter struct {
	policy execution.LayerPolicy
	plan   *execution.LayerPlan
	mode   execution.LayerMode

	mu        sync.RWMutex
	completed map[string]bool
	failed    map[string]bool
}

// NewLayerSchedulingAdapter creates an adapter that wraps a LayerPolicy.
// The plan must be computed before creating the adapter.
func NewLayerSchedulingAdapter(policy execution.LayerPolicy, plan *execution.LayerPlan, mode execution.LayerMode) *LayerSchedulingAdapter {
	return &LayerSchedulingAdapter{
		policy:    policy,
		plan:      plan,
		mode:      mode,
		completed: make(map[string]bool),
		failed:    make(map[string]bool),
	}
}

// NewLayerSchedulingAdapterFromUnits creates an adapter by computing a plan from units.
// This is a convenience method that combines plan computation and adapter creation.
func NewLayerSchedulingAdapterFromUnits(units []workunit.UnitSpec, mode execution.LayerMode) (*LayerSchedulingAdapter, error) {
	policy := execution.NewLayerPlanner(mode)
	plan, err := policy.ComputePlan(units)
	if err != nil {
		return nil, err
	}
	return NewLayerSchedulingAdapter(policy, plan, mode), nil
}

// Mode returns the current LayerMode.
func (a *LayerSchedulingAdapter) Mode() execution.LayerMode {
	return a.mode
}

// Plan returns the computed execution plan.
func (a *LayerSchedulingAdapter) Plan() *execution.LayerPlan {
	return a.plan
}

// IsReady delegates to core domain policy.
// Returns true if the unit can execute given current completion state.
func (a *LayerSchedulingAdapter) IsReady(id workunit.UnitID) bool {
	if a.policy == nil || a.plan == nil {
		return false
	}

	a.mu.RLock()
	completed := a.snapshotCompletedLocked()
	a.mu.RUnlock()

	return a.policy.IsReady(a.plan, id, completed)
}

// GetReadyUnits returns all units that can execute now.
// Units are sorted by weight descending (LPT scheduling).
func (a *LayerSchedulingAdapter) GetReadyUnits() []workunit.UnitSpec {
	if a.policy == nil || a.plan == nil {
		return nil
	}

	a.mu.RLock()
	completed := a.snapshotCompletedLocked()
	a.mu.RUnlock()

	return a.policy.GetReadyUnits(a.plan, completed)
}

// MarkComplete marks a unit as completed successfully.
// This updates the adapter's completion state used for IsReady checks.
func (a *LayerSchedulingAdapter) MarkComplete(id workunit.UnitID) {
	key := id.Longname()
	a.mu.Lock()
	a.completed[key] = true
	a.mu.Unlock()
}

// MarkFailed marks a unit as failed.
// Failed units are considered "complete" for dependency resolution.
func (a *LayerSchedulingAdapter) MarkFailed(id workunit.UnitID) {
	key := id.Longname()
	a.mu.Lock()
	a.completed[key] = true
	a.failed[key] = true
	a.mu.Unlock()
}

// IsFailed returns true if the unit failed.
func (a *LayerSchedulingAdapter) IsFailed(id workunit.UnitID) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.failed[id.Longname()]
}

// HasFailedDependency returns true if any dependency of the unit has failed.
// This is used to fail dependents immediately without scheduling them.
func (a *LayerSchedulingAdapter) HasFailedDependency(id workunit.UnitID) bool {
	if a.plan == nil {
		return false
	}

	// Find the unit in the plan to get its dependencies
	mlIdx, clIdx, uIdx := a.plan.FindUnit(id)
	if mlIdx < 0 {
		return false
	}

	unit := a.plan.GetUnit(mlIdx, clIdx, uIdx)
	if unit == nil {
		return false
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, dep := range unit.DependsOn {
		if a.failed[dep.Longname()] {
			return true
		}
	}
	return false
}

// Completed returns a snapshot of all completed unit IDs.
func (a *LayerSchedulingAdapter) Completed() map[string]bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.snapshotCompletedLocked()
}

// snapshotCompletedLocked creates a copy of the completed map.
// Must be called with a.mu held (read or write).
func (a *LayerSchedulingAdapter) snapshotCompletedLocked() map[string]bool {
	snapshot := make(map[string]bool, len(a.completed))
	for k, v := range a.completed {
		snapshot[k] = v
	}
	return snapshot
}

// Stats returns execution plan statistics.
func (a *LayerSchedulingAdapter) Stats() execution.LayerStats {
	if a.plan == nil {
		return execution.LayerStats{}
	}
	return a.plan.Stats
}

// ModuleLayerForUnit returns the module layer index for a unit.
// Returns -1 if the unit is not found.
func (a *LayerSchedulingAdapter) ModuleLayerForUnit(id workunit.UnitID) int {
	if a.plan == nil {
		return -1
	}
	return a.plan.ModuleLayerForUnit(id)
}

// ModulesInLayer returns the modules in a given layer.
func (a *LayerSchedulingAdapter) ModulesInLayer(layerIdx int) []string {
	if a.plan == nil {
		return nil
	}
	return a.plan.ModulesInLayer(layerIdx)
}
