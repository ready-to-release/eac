package execution

import (
	"fmt"
	"sort"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// LayerPlanner computes execution plans from work units.
// It implements LayerPolicy and provides the domain's single source
// of truth for execution ordering.
//
// The planner organizes work units into nested layers:
//  1. Module layers: from inter-module dependencies (DependsOn cross-module)
//  2. Component layers: from build_after relationships (DependsOn within module)
//
// This supports three modes:
//   - Strict: enforce both module and component layers
//   - ModuleOnly: enforce only module layers
//   - None: pure dependency mode (only explicit DependsOn)
type LayerPlanner struct {
	mode LayerMode
}

// Ensure LayerPlanner implements LayerPolicy.
var _ LayerPolicy = (*LayerPlanner)(nil)

// NewLayerPlanner creates a planner with the specified mode.
func NewLayerPlanner(mode LayerMode) *LayerPlanner {
	return &LayerPlanner{mode: mode}
}

// Mode returns the current layer mode.
func (p *LayerPlanner) Mode() LayerMode {
	return p.mode
}

// ComputePlan implements LayerPolicy.ComputePlan.
// It creates a LayerPlan from work units by:
//  1. Grouping units by module
//  2. Computing module layers from inter-module dependencies
//  3. Computing component layers within each module from DependsOn relationships
//  4. Building the nested LayerPlan structure
func (p *LayerPlanner) ComputePlan(units []workunit.UnitSpec) (*LayerPlan, error) {
	if len(units) == 0 {
		return &LayerPlan{
			ModuleLayers: []ModuleLayer{},
			Stats:        LayerStats{},
			unitIndex:    make(map[string]unitPosition),
		}, nil
	}

	// Group units by module
	moduleUnits := make(map[string][]workunit.UnitSpec)
	for _, u := range units {
		moduleUnits[u.ID.Module] = append(moduleUnits[u.ID.Module], u)
	}

	// Get unique modules
	modules := make([]string, 0, len(moduleUnits))
	for m := range moduleUnits {
		modules = append(modules, m)
	}

	// Compute module layers using topological sort
	moduleLayers, err := p.computeModuleLayers(modules, units)
	if err != nil {
		return nil, err
	}

	// Build the plan structure
	plan := &LayerPlan{
		ModuleLayers: make([]ModuleLayer, len(moduleLayers)),
		unitIndex:    make(map[string]unitPosition),
	}

	totalComponentLayers := 0
	maxParallelism := 0

	for mlIdx, moduleLayer := range moduleLayers {
		ml := ModuleLayer{
			Index:       mlIdx,
			Modules:     moduleLayer,
			moduleUnits: make(map[string][]workunit.UnitSpec),
		}

		// Collect all units for this module layer
		var layerUnits []workunit.UnitSpec
		for _, module := range moduleLayer {
			if mUnits, ok := moduleUnits[module]; ok {
				layerUnits = append(layerUnits, mUnits...)
				ml.moduleUnits[module] = mUnits
			}
		}

		// Compute component layers within this module layer
		componentLayers, err := p.computeComponentLayers(layerUnits)
		if err != nil {
			return nil, err
		}

		ml.ComponentLayers = make([]ComponentLayer, len(componentLayers))
		for clIdx, clUnits := range componentLayers {
			ml.ComponentLayers[clIdx] = ComponentLayer{
				Index: clIdx,
				Units: clUnits,
			}

			// Track max parallelism
			if len(clUnits) > maxParallelism {
				maxParallelism = len(clUnits)
			}

			// Build unit index
			for uIdx, u := range clUnits {
				plan.unitIndex[u.ID.Longname()] = unitPosition{
					moduleLayerIdx:    mlIdx,
					module:            u.ID.Module,
					componentLayerIdx: clIdx,
					unitIdx:           uIdx,
				}
			}
		}

		totalComponentLayers += len(componentLayers)
		plan.ModuleLayers[mlIdx] = ml
	}

	// Compute stats
	plan.Stats = LayerStats{
		TotalModuleLayers:    len(moduleLayers),
		TotalComponentLayers: totalComponentLayers,
		TotalUoWs:            len(units),
		MaxParallelism:       maxParallelism,
		CriticalPathLength:   len(moduleLayers),
	}

	return plan, nil
}

// computeModuleLayers computes module layers using topological sort.
// Modules with no inter-module dependencies go in layer 0.
// Modules in layer N depend only on modules in layers 0..N-1.
func (p *LayerPlanner) computeModuleLayers(modules []string, units []workunit.UnitSpec) ([][]string, error) {
	if len(modules) == 0 {
		return nil, nil
	}

	// Build inter-module dependency graph
	// A module depends on another if any unit in it depends on a unit in another module
	moduleDeps := make(map[string]map[string]bool)
	for _, m := range modules {
		moduleDeps[m] = make(map[string]bool)
	}

	for _, u := range units {
		srcModule := u.ID.Module
		for _, dep := range u.DependsOn {
			depModule := dep.Module
			if depModule != srcModule && moduleDeps[srcModule] != nil {
				moduleDeps[srcModule][depModule] = true
			}
		}
	}

	// Kahn's algorithm for topological sort with layering
	inDegree := make(map[string]int)
	for _, m := range modules {
		inDegree[m] = 0
	}

	for m, deps := range moduleDeps {
		for dep := range deps {
			if _, ok := inDegree[dep]; ok {
				inDegree[m]++
			}
		}
	}

	var layers [][]string
	processed := make(map[string]bool)

	for len(processed) < len(modules) {
		// Find all modules with in-degree 0
		var layer []string
		for _, m := range modules {
			if !processed[m] && inDegree[m] == 0 {
				layer = append(layer, m)
			}
		}

		if len(layer) == 0 {
			// Circular dependency detected
			var remaining []string
			for _, m := range modules {
				if !processed[m] {
					remaining = append(remaining, m)
				}
			}
			return nil, fmt.Errorf("circular module dependency detected among: %v", remaining)
		}

		// Sort layer for deterministic output
		sort.Strings(layer)
		layers = append(layers, layer)

		// Mark layer as processed and update in-degrees
		for _, m := range layer {
			processed[m] = true
			for dependent, deps := range moduleDeps {
				if deps[m] {
					inDegree[dependent]--
				}
			}
		}
	}

	return layers, nil
}

// computeComponentLayers computes component layers within a module layer.
// Uses DependsOn relationships to determine ordering.
// Components with no intra-layer dependencies go in layer 0.
func (p *LayerPlanner) computeComponentLayers(units []workunit.UnitSpec) ([][]workunit.UnitSpec, error) {
	if len(units) == 0 {
		return nil, nil
	}

	// Build unit dependency graph (only for units within this set)
	// First, detect and report duplicates - UoW longnames must be unique
	unitSet := make(map[string]bool)
	unitByLongname := make(map[string]workunit.UnitSpec)
	var duplicates []string
	for _, u := range units {
		key := u.ID.Longname()
		if unitSet[key] {
			duplicates = append(duplicates, key)
		}
		unitSet[key] = true
		unitByLongname[key] = u
	}
	if len(duplicates) > 0 {
		return nil, fmt.Errorf("duplicate unit longnames detected (UoW IDs must be unique): %v", duplicates)
	}

	// Compute in-degrees based on dependencies within this unit set
	inDegree := make(map[string]int)
	for _, u := range units {
		key := u.ID.Longname()
		inDegree[key] = 0
		for _, dep := range u.DependsOn {
			depKey := dep.Longname()
			if unitSet[depKey] {
				inDegree[key]++
			}
		}
	}

	// Kahn's algorithm with layering
	var layers [][]workunit.UnitSpec
	processed := make(map[string]bool)

	for len(processed) < len(units) {
		var layer []workunit.UnitSpec
		var layerKeys []string

		// Find all units with in-degree 0
		for _, u := range units {
			key := u.ID.Longname()
			if !processed[key] && inDegree[key] == 0 {
				layerKeys = append(layerKeys, key)
			}
		}

		if len(layerKeys) == 0 && len(processed) < len(units) {
			// Circular dependency detected
			var remaining []string
			for _, u := range units {
				key := u.ID.Longname()
				if !processed[key] {
					remaining = append(remaining, key)
				}
			}
			return nil, fmt.Errorf("circular dependency detected among units: %v", remaining)
		}

		// Sort for deterministic output
		sort.Strings(layerKeys)
		for _, key := range layerKeys {
			layer = append(layer, unitByLongname[key])
		}

		layers = append(layers, layer)

		// Mark layer as processed and update in-degrees
		for _, key := range layerKeys {
			processed[key] = true
			// Update in-degrees of dependents
			for _, u := range units {
				uKey := u.ID.Longname()
				if processed[uKey] {
					continue
				}
				for _, dep := range u.DependsOn {
					if dep.Longname() == key {
						inDegree[uKey]--
					}
				}
			}
		}
	}

	return layers, nil
}

// IsReady implements LayerPolicy.IsReady.
// Checks if a unit can execute given the current completion state.
//
// The check depends on the LayerMode:
//   - Strict: module layers + component layers + DependsOn (all three)
//   - None: component layers + DependsOn (skip module layer enforcement)
func (p *LayerPlanner) IsReady(plan *LayerPlan, unitID workunit.UnitID, completed map[string]bool) bool {
	if plan == nil {
		return false
	}

	// Find the unit's position in the plan
	pos, ok := plan.unitIndex[unitID.Longname()]
	if !ok {
		// Unit not in plan - can't be ready
		return false
	}

	switch p.mode {
	case LayerModeStrict:
		return p.isReadyStrict(plan, pos, unitID, completed)
	case LayerModeNone:
		return p.isReadyComponentAndDependsOn(plan, pos, unitID, completed)
	default:
		return p.isReadyStrict(plan, pos, unitID, completed)
	}
}

// isReadyStrict checks readiness with full layer constraints.
// A unit is ready when:
//  1. All previous module layers are complete
//  2. All previous component layers (within current module layer) are complete
//  3. All DependsOn units are complete
func (p *LayerPlanner) isReadyStrict(plan *LayerPlan, pos unitPosition, unitID workunit.UnitID, completed map[string]bool) bool {
	// Check all previous module layers are complete
	for mlIdx := 0; mlIdx < pos.moduleLayerIdx; mlIdx++ {
		if !p.isModuleLayerComplete(plan, mlIdx, completed) {
			return false
		}
	}

	// Check all previous component layers in current module layer are complete
	ml := &plan.ModuleLayers[pos.moduleLayerIdx]
	for clIdx := 0; clIdx < pos.componentLayerIdx; clIdx++ {
		if !p.isComponentLayerComplete(ml, clIdx, completed) {
			return false
		}
	}

	// Check explicit DependsOn dependencies
	return p.areDependsOnComplete(unitID, plan, completed)
}

// isReadyComponentAndDependsOn checks readiness with component layers and DependsOn.
// Module layer ordering is NOT enforced.
// A unit is ready when:
//  1. All previous component layers (within current module layer) are complete
//  2. All DependsOn units are complete
func (p *LayerPlanner) isReadyComponentAndDependsOn(plan *LayerPlan, pos unitPosition, unitID workunit.UnitID, completed map[string]bool) bool {
	// Check all previous component layers in current module layer are complete
	// (Module layer ordering is NOT checked - that's the difference from Strict)
	ml := &plan.ModuleLayers[pos.moduleLayerIdx]
	for clIdx := 0; clIdx < pos.componentLayerIdx; clIdx++ {
		if !p.isComponentLayerComplete(ml, clIdx, completed) {
			return false
		}
	}

	// Check explicit DependsOn dependencies
	return p.areDependsOnComplete(unitID, plan, completed)
}

// isModuleLayerComplete returns true if all units in the module layer are complete.
func (p *LayerPlanner) isModuleLayerComplete(plan *LayerPlan, mlIdx int, completed map[string]bool) bool {
	if mlIdx < 0 || mlIdx >= len(plan.ModuleLayers) {
		return true
	}
	ml := &plan.ModuleLayers[mlIdx]
	for _, cl := range ml.ComponentLayers {
		for _, u := range cl.Units {
			if !completed[u.ID.Longname()] {
				return false
			}
		}
	}
	return true
}

// isComponentLayerComplete returns true if all units in the component layer are complete.
func (p *LayerPlanner) isComponentLayerComplete(ml *ModuleLayer, clIdx int, completed map[string]bool) bool {
	if clIdx < 0 || clIdx >= len(ml.ComponentLayers) {
		return true
	}
	cl := &ml.ComponentLayers[clIdx]
	for _, u := range cl.Units {
		if !completed[u.ID.Longname()] {
			return false
		}
	}
	return true
}

// areDependsOnComplete returns true if all DependsOn units are complete.
func (p *LayerPlanner) areDependsOnComplete(unitID workunit.UnitID, plan *LayerPlan, completed map[string]bool) bool {
	// Find the unit in the plan to get its DependsOn
	pos, ok := plan.unitIndex[unitID.Longname()]
	if !ok {
		return false
	}

	unit := plan.GetUnit(pos.moduleLayerIdx, pos.componentLayerIdx, pos.unitIdx)
	if unit == nil {
		return false
	}

	for _, dep := range unit.DependsOn {
		depKey := dep.Longname()
		if !completed[depKey] {
			return false
		}
	}
	return true
}

// GetReadyUnits implements LayerPolicy.GetReadyUnits.
// Returns all units that can execute now, sorted by weight descending (LPT).
func (p *LayerPlanner) GetReadyUnits(plan *LayerPlan, completed map[string]bool) []workunit.UnitSpec {
	if plan == nil {
		return nil
	}

	var ready []workunit.UnitSpec
	for _, ml := range plan.ModuleLayers {
		for _, cl := range ml.ComponentLayers {
			for _, u := range cl.Units {
				if completed[u.ID.Longname()] {
					continue // Already complete
				}
				if p.IsReady(plan, u.ID, completed) {
					ready = append(ready, u)
				}
			}
		}
	}

	// Sort by weight descending (LPT - Longest Processing Time First)
	sort.Slice(ready, func(i, j int) bool {
		return ready[i].Weight > ready[j].Weight
	})

	return ready
}

// SimpleCompletionTracker is a basic implementation of CompletionTracker.
type SimpleCompletionTracker struct {
	completed map[string]bool
	failed    map[string]bool
}

// NewSimpleCompletionTracker creates a new completion tracker.
func NewSimpleCompletionTracker() *SimpleCompletionTracker {
	return &SimpleCompletionTracker{
		completed: make(map[string]bool),
		failed:    make(map[string]bool),
	}
}

// MarkComplete marks a unit as completed successfully.
func (t *SimpleCompletionTracker) MarkComplete(id workunit.UnitID) {
	key := id.Longname()
	t.completed[key] = true
}

// MarkFailed marks a unit as failed.
func (t *SimpleCompletionTracker) MarkFailed(id workunit.UnitID) {
	key := id.Longname()
	t.completed[key] = true // Failed units are "complete" for dependency resolution
	t.failed[key] = true
}

// IsComplete returns true if the unit has completed (success or failure).
func (t *SimpleCompletionTracker) IsComplete(id workunit.UnitID) bool {
	return t.completed[id.Longname()]
}

// IsFailed returns true if the unit failed.
func (t *SimpleCompletionTracker) IsFailed(id workunit.UnitID) bool {
	return t.failed[id.Longname()]
}

// Completed returns a snapshot of all completed units.
func (t *SimpleCompletionTracker) Completed() map[string]bool {
	snapshot := make(map[string]bool, len(t.completed))
	for k, v := range t.completed {
		snapshot[k] = v
	}
	return snapshot
}
