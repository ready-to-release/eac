package scheduling

import "github.com/ready-to-release/eac/go/core/workunit"

// DependencyTracker provides bidirectional dependency tracking.
// Uses two maps for efficient scheduling:
//   - depsOf: "who do I depend on?" (forward) - used for IsReady checks
//   - blocks: "who depends on me?" (reverse) - used for targeted wake-ups
//
// A unit is ready when all its depsOf entries are complete.
type DependencyTracker struct {
	completed map[string]bool              // longname -> done (success or failure)
	failed    map[string]bool              // longname -> failed
	depsOf    map[string][]workunit.UnitID // longname -> what I depend on (forward)
	blocks    map[string][]workunit.UnitID // longname -> who depends on me (reverse)
}

// NewDependencyTracker creates a tracker with bidirectional dependency mapping.
func NewDependencyTracker(work []workunit.UnitSpec) *DependencyTracker {
	dt := &DependencyTracker{
		completed: make(map[string]bool),
		failed:    make(map[string]bool),
		depsOf:    make(map[string][]workunit.UnitID),
		blocks:    make(map[string][]workunit.UnitID),
	}

	// Build forward dependency map
	for _, w := range work {
		dt.depsOf[w.ID.Longname()] = w.DependsOn
	}

	// Build reverse dependency map (who depends on me?)
	for _, w := range work {
		for _, dep := range w.DependsOn {
			depKey := dep.Longname()
			dt.blocks[depKey] = append(dt.blocks[depKey], w.ID)
		}
	}

	return dt
}

// IsReady returns true if all dependencies for the unit are complete.
// A unit with no dependencies is always ready.
func (dt *DependencyTracker) IsReady(id workunit.UnitID) bool {
	deps := dt.depsOf[id.Longname()]
	for _, dep := range deps {
		if !dt.completed[dep.Longname()] {
			return false
		}
	}
	return true
}

// HasFailedDependency returns true if any dependency of the unit has failed.
func (dt *DependencyTracker) HasFailedDependency(id workunit.UnitID) bool {
	for _, dep := range dt.depsOf[id.Longname()] {
		if dt.failed[dep.Longname()] {
			return true
		}
	}
	return false
}

// MarkComplete marks a unit as done successfully.
func (dt *DependencyTracker) MarkComplete(id workunit.UnitID) {
	dt.completed[id.Longname()] = true
}

// MarkFailed marks a unit as failed.
// Failed units are also marked complete for dependency resolution,
// allowing dependents to become ready and fail via HasFailedDependency.
func (dt *DependencyTracker) MarkFailed(id workunit.UnitID) {
	longname := id.Longname()
	dt.failed[longname] = true
	dt.completed[longname] = true
}

// CascadeFail marks all transitive dependents of the given ID as failed.
// Returns the list of cascade-failed IDs (NOT including the original ID,
// which should already be marked failed by the caller).
func (dt *DependencyTracker) CascadeFail(id workunit.UnitID) []workunit.UnitID {
	var cascaded []workunit.UnitID
	dt.cascadeFailRecursive(id, &cascaded)
	return cascaded
}

func (dt *DependencyTracker) cascadeFailRecursive(id workunit.UnitID, cascaded *[]workunit.UnitID) {
	for _, dependent := range dt.blocks[id.Longname()] {
		depKey := dependent.Longname()
		if dt.failed[depKey] {
			continue // Already failed — avoid infinite loops on diamond deps
		}
		dt.failed[depKey] = true
		dt.completed[depKey] = true
		*cascaded = append(*cascaded, dependent)
		dt.cascadeFailRecursive(dependent, cascaded)
	}
}

// GetDependsOn returns the units that this unit depends on (forward deps).
func (dt *DependencyTracker) GetDependsOn(id workunit.UnitID) []workunit.UnitID {
	return dt.depsOf[id.Longname()]
}

// GetBlocks returns the units that depend on this unit (reverse deps).
// Useful for targeted wake-ups when a unit completes.
func (dt *DependencyTracker) GetBlocks(id workunit.UnitID) []workunit.UnitID {
	return dt.blocks[id.Longname()]
}

// CompletedCount returns the number of completed units.
func (dt *DependencyTracker) CompletedCount() int {
	return len(dt.completed)
}

// FailedCount returns the number of failed units.
func (dt *DependencyTracker) FailedCount() int {
	return len(dt.failed)
}
