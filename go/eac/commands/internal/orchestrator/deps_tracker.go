package orchestrator

import (
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

// DepsTracker tracks component completion for dependency resolution.
// Used by the dispatcher to determine which work items are ready to execute.
type DepsTracker struct {
	// completed tracks which components have finished
	// key: "module:component"
	completed map[string]bool

	// depsOf maps each unit to its dependencies
	// key: unit longname, value: list of dep component names
	depsOf map[string][]string

	// moduleOf maps each unit to its module (for resolving dep component names)
	moduleOf map[string]string
}

// NewDepsTracker creates a tracker for the given work units.
func NewDepsTracker(work []workunit.UnitSpec) *DepsTracker {
	dt := &DepsTracker{
		completed: make(map[string]bool),
		depsOf:    make(map[string][]string),
		moduleOf:  make(map[string]string),
	}

	for _, w := range work {
		key := w.ID.Longname()
		dt.depsOf[key] = w.DependsOnComponents()
		dt.moduleOf[key] = w.ID.Module
	}

	return dt
}

// IsReady returns true if all dependencies for the unit are completed.
func (dt *DepsTracker) IsReady(id workunit.UnitID) bool {
	key := id.Longname()
	deps := dt.depsOf[key]
	module := dt.moduleOf[key]

	for _, depComp := range deps {
		depKey := module + ":" + depComp
		if !dt.completed[depKey] {
			return false
		}
	}
	return true
}

// MarkComplete marks a component as done.
func (dt *DepsTracker) MarkComplete(id workunit.UnitID) {
	// Mark by module:component (not full longname with tool)
	key := id.Module + ":" + id.Component
	dt.completed[key] = true
}
