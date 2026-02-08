package scheduling

import "github.com/ready-to-release/eac/go/core/workunit"

// DependencyTracker provides bidirectional dependency tracking.
// Uses two maps for efficient scheduling:
//   - depsOf: "who do I depend on?" (forward) - built eagerly for IsReady checks
//   - blocks: "who depends on me?" (reverse) - built lazily on first CascadeFail/GetBlocks
//
// A unit is ready when all its depsOf entries are complete.
type DependencyTracker struct {
	completed map[string]bool              // longname -> done (success or failure)
	failed    map[string]bool              // longname -> failed
	depsOf    map[string][]workunit.UnitID // longname -> what I depend on (forward)
	blocks    map[string][]workunit.UnitID // longname -> who depends on me (reverse, lazy)
	work      []workunit.UnitSpec          // retained for lazy reverse map construction
}

// NewDependencyTracker creates a tracker with forward dependency mapping.
// The reverse map (blocks) is built lazily on first use by CascadeFail or GetBlocks.
func NewDependencyTracker(work []workunit.UnitSpec) *DependencyTracker {
	dt := &DependencyTracker{
		completed: make(map[string]bool),
		failed:    make(map[string]bool),
		depsOf:    make(map[string][]workunit.UnitID),
		work:      work,
	}

	// Build forward dependency map (needed immediately for IsReady)
	for _, w := range work {
		dt.depsOf[w.ID.Longname()] = w.DependsOn
	}

	return dt
}

// ensureBlocks lazily builds the reverse dependency map on first access.
func (dt *DependencyTracker) ensureBlocks() {
	if dt.blocks != nil {
		return
	}
	dt.blocks = make(map[string][]workunit.UnitID)
	for _, w := range dt.work {
		for _, dep := range w.DependsOn {
			depKey := dep.Longname()
			dt.blocks[depKey] = append(dt.blocks[depKey], w.ID)
		}
	}
	dt.work = nil // Release reference after building
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
	dt.ensureBlocks()
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

// FailedDependencyModules returns the unique module names of failed dependencies for the given unit.
func (dt *DependencyTracker) FailedDependencyModules(id workunit.UnitID) []string {
	seen := make(map[string]struct{})
	var modules []string
	for _, dep := range dt.depsOf[id.Longname()] {
		if dt.failed[dep.Longname()] {
			if _, exists := seen[dep.Module]; !exists {
				seen[dep.Module] = struct{}{}
				modules = append(modules, dep.Module)
			}
		}
	}
	return modules
}

// GetDependsOn returns the units that this unit depends on (forward deps).
func (dt *DependencyTracker) GetDependsOn(id workunit.UnitID) []workunit.UnitID {
	return dt.depsOf[id.Longname()]
}

// GetBlocks returns the units that depend on this unit (reverse deps).
// Useful for targeted wake-ups when a unit completes.
// Lazily builds the reverse map on first call.
func (dt *DependencyTracker) GetBlocks(id workunit.UnitID) []workunit.UnitID {
	dt.ensureBlocks()
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
