package execution

import "github.com/ready-to-release/eac/go/core/workunit"

// LayerPlan is the complete execution plan with nested layers.
// This is the domain's answer to "how should these UoWs be executed?"
//
// Structure:
//
//	LayerPlan
//	├── ModuleLayer[0]  (modules with no inter-module dependencies)
//	│   ├── ComponentLayer[0]  (components with no build_after deps)
//	│   │   └── UnitSpec[...]
//	│   └── ComponentLayer[1]  (components that depend on layer 0)
//	│       └── UnitSpec[...]
//	├── ModuleLayer[1]  (modules that depend on layer 0 modules)
//	│   └── ...
//	└── Stats
type LayerPlan struct {
	// ModuleLayers groups modules by inter-module dependency order.
	// All modules in layer N must complete before layer N+1 starts.
	ModuleLayers []ModuleLayer

	// Stats provides execution plan metrics.
	Stats LayerStats

	// unitIndex maps UnitID.Longname() to its position in the plan.
	// Used for O(1) lookup in IsReady checks.
	unitIndex map[string]unitPosition
}

// unitPosition tracks where a unit lives in the plan.
type unitPosition struct {
	moduleLayerIdx    int
	module            string
	componentLayerIdx int
	unitIdx           int
}

// ModuleLayer groups modules that can execute in parallel.
// All modules in layer N must complete before layer N+1 starts.
type ModuleLayer struct {
	// Index is the layer number (0-based).
	Index int

	// Modules contains the module monikers in this layer.
	// These modules have no dependencies on each other and can run in parallel.
	Modules []string

	// ComponentLayers groups components within this module layer.
	// This is a flattened view - components are ordered by build_after constraints.
	ComponentLayers []ComponentLayer

	// moduleUnits maps module moniker to its units for quick lookup.
	moduleUnits map[string][]workunit.UnitSpec
}

// ComponentLayer groups components that can execute in parallel.
// Components in the same layer have no build_after dependencies on each other.
type ComponentLayer struct {
	// Index is the layer number within the module layer (0-based).
	Index int

	// Units contains the work units in this component layer.
	// These units can run in parallel.
	Units []workunit.UnitSpec
}

// LayerStats provides execution plan metrics.
type LayerStats struct {
	// TotalModuleLayers is the number of module layers.
	TotalModuleLayers int

	// TotalComponentLayers is the total number of component layers across all module layers.
	TotalComponentLayers int

	// TotalUoWs is the total number of work units in the plan.
	TotalUoWs int

	// MaxParallelism is the maximum number of concurrent UoWs possible.
	// This is the size of the largest component layer.
	MaxParallelism int

	// CriticalPathLength is the minimum number of sequential steps required.
	// This equals TotalModuleLayers when component layers are collapsed.
	CriticalPathLength int
}

// LayerMode controls enforcement strictness for layer constraints.
//
// Three layers exist: Module → Component → UoW
//   - Module layer: inter-module dependency ordering (OPTIONAL)
//   - Component layer: intra-module build_after ordering (MANDATORY)
//   - UoW layer: explicit DependsOn constraints (MANDATORY)
type LayerMode int

const (
	// LayerModeStrict enforces all three layers.
	// A unit is ready only when:
	// 1. All previous module layers are complete
	// 2. All previous component layers (within module layer) are complete
	// 3. All DependsOn units are complete
	//
	// Use for build commands where module dependencies must be respected.
	LayerModeStrict LayerMode = iota

	// LayerModeNone skips module layer enforcement.
	// A unit is ready when:
	// 1. All previous component layers (within module layer) are complete
	// 2. All DependsOn units are complete
	//
	// Module layer ordering is NOT enforced - units from different modules
	// can run in parallel regardless of inter-module dependencies.
	//
	// Use for lint/scan/test commands that can run in parallel across modules.
	LayerModeNone
)

// String returns the string representation of a LayerMode.
func (m LayerMode) String() string {
	switch m {
	case LayerModeStrict:
		return "strict"
	case LayerModeNone:
		return "none"
	default:
		return "unknown"
	}
}

// ParseLayerMode converts a string to LayerMode.
// Returns LayerModeStrict for unrecognized values.
func ParseLayerMode(s string) LayerMode {
	switch s {
	case "strict", "layered":
		return LayerModeStrict
	case "none", "unlayered":
		return LayerModeNone
	default:
		return LayerModeStrict
	}
}

// GetUnit returns the UnitSpec at the given position, or nil if not found.
func (p *LayerPlan) GetUnit(moduleLayerIdx, componentLayerIdx, unitIdx int) *workunit.UnitSpec {
	if p == nil || moduleLayerIdx < 0 || moduleLayerIdx >= len(p.ModuleLayers) {
		return nil
	}
	ml := &p.ModuleLayers[moduleLayerIdx]
	if componentLayerIdx < 0 || componentLayerIdx >= len(ml.ComponentLayers) {
		return nil
	}
	cl := &ml.ComponentLayers[componentLayerIdx]
	if unitIdx < 0 || unitIdx >= len(cl.Units) {
		return nil
	}
	return &cl.Units[unitIdx]
}

// FindUnit returns the position of a unit by its ID, or (-1,-1,-1) if not found.
func (p *LayerPlan) FindUnit(id workunit.UnitID) (moduleLayerIdx, componentLayerIdx, unitIdx int) {
	if p == nil || p.unitIndex == nil {
		return -1, -1, -1
	}
	pos, ok := p.unitIndex[id.Longname()]
	if !ok {
		return -1, -1, -1
	}
	return pos.moduleLayerIdx, pos.componentLayerIdx, pos.unitIdx
}

// AllUnits returns all units in the plan in layer order.
func (p *LayerPlan) AllUnits() []workunit.UnitSpec {
	if p == nil {
		return nil
	}
	units := make([]workunit.UnitSpec, 0, p.Stats.TotalUoWs)
	for _, ml := range p.ModuleLayers {
		for _, cl := range ml.ComponentLayers {
			units = append(units, cl.Units...)
		}
	}
	return units
}

// ModuleLayerForUnit returns the module layer index for a unit.
// Returns -1 if the unit is not found.
func (p *LayerPlan) ModuleLayerForUnit(id workunit.UnitID) int {
	mlIdx, _, _ := p.FindUnit(id)
	return mlIdx
}

// ComponentLayerForUnit returns the component layer index for a unit (within its module layer).
// Returns -1 if the unit is not found.
func (p *LayerPlan) ComponentLayerForUnit(id workunit.UnitID) int {
	_, clIdx, _ := p.FindUnit(id)
	return clIdx
}

// UnitsInModule returns all units for a given module across all layers.
func (p *LayerPlan) UnitsInModule(module string) []workunit.UnitSpec {
	if p == nil {
		return nil
	}
	var units []workunit.UnitSpec
	for _, ml := range p.ModuleLayers {
		if moduleUnits, ok := ml.moduleUnits[module]; ok {
			units = append(units, moduleUnits...)
		}
	}
	return units
}

// ModulesInLayer returns the modules in a given layer.
func (p *LayerPlan) ModulesInLayer(layerIdx int) []string {
	if p == nil || layerIdx < 0 || layerIdx >= len(p.ModuleLayers) {
		return nil
	}
	return p.ModuleLayers[layerIdx].Modules
}
