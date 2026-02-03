package orchestrator

import (
	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// DepsTracker tracks unit completion for dependency resolution.
// Used by the dispatcher to determine which work items are ready to execute.
//
// Delegates IsReady checks to the configured LayerPolicy for layer-based
// execution ordering.
type DepsTracker struct {
	completed   map[string]bool              // longname -> done (success or failure)
	failed      map[string]bool              // longname -> failed
	depsOf      map[string][]workunit.UnitID // longname -> dependencies
	layerPolicy execution.LayerPolicy
	layerPlan   *execution.LayerPlan
}

// NewDepsTracker creates a tracker for the given work units.
func NewDepsTracker(work []workunit.UnitSpec) *DepsTracker {
	dt := &DepsTracker{
		completed: make(map[string]bool),
		failed:    make(map[string]bool),
		depsOf:    make(map[string][]workunit.UnitID),
	}

	for _, w := range work {
		dt.depsOf[w.ID.Longname()] = w.DependsOn
	}

	return dt
}

// IsReady returns true if all dependencies for the unit are resolved (completed or failed).
// Returns false only if any dependency is still pending.
// Use HasFailedDependency to check if the unit should be skipped due to failure.
//
// Requires LayerPolicy to be configured via SetLayerPolicy().
// Panics if LayerPolicy is not set.
func (dt *DepsTracker) IsReady(id workunit.UnitID) bool {
	if dt.layerPolicy == nil || dt.layerPlan == nil {
		panic("DepsTracker.IsReady: LayerPolicy not configured - use SetLayerPolicy()")
	}
	return dt.layerPolicy.IsReady(dt.layerPlan, id, dt.completedSnapshot())
}

// HasFailedDependency returns true if any dependency of the unit has failed.
func (dt *DepsTracker) HasFailedDependency(id workunit.UnitID) bool {
	for _, dep := range dt.depsOf[id.Longname()] {
		if dt.failed[dep.Longname()] {
			return true
		}
	}
	return false
}

// MarkComplete marks a unit as done successfully.
func (dt *DepsTracker) MarkComplete(id workunit.UnitID) {
	dt.completed[id.Longname()] = true
}

// MarkFailed marks a unit as failed.
// Failed units are also marked complete for dependency resolution,
// allowing dependents to become ready and fail via HasFailedDependency.
func (dt *DepsTracker) MarkFailed(id workunit.UnitID) {
	longname := id.Longname()
	dt.failed[longname] = true
	dt.completed[longname] = true
}

// SetLayerPolicy configures layer-based execution ordering.
func (dt *DepsTracker) SetLayerPolicy(policy execution.LayerPolicy, plan *execution.LayerPlan) {
	dt.layerPolicy = policy
	dt.layerPlan = plan
}

// HasLayerPolicy returns true if a LayerPolicy is configured.
func (dt *DepsTracker) HasLayerPolicy() bool {
	return dt.layerPolicy != nil && dt.layerPlan != nil
}

// GetLayerPlan returns the configured LayerPlan, or nil if not set.
func (dt *DepsTracker) GetLayerPlan() *execution.LayerPlan {
	return dt.layerPlan
}

// completedSnapshot returns a snapshot of completed units for policy queries.
func (dt *DepsTracker) completedSnapshot() map[string]bool {
	snapshot := make(map[string]bool, len(dt.completed))
	for k, v := range dt.completed {
		snapshot[k] = v
	}
	return snapshot
}
