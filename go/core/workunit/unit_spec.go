package workunit

import "github.com/ready-to-release/eac/go/core/resource"

// UnitSpec represents the input specification for a unit of work.
// It describes what to execute and how to schedule it.
// Implements interfaces.UnitSpecPort for port-based access.
type UnitSpec struct {
	ID              UnitID                  // Unique identifier for this work unit
	ComponentType   string                  // From component-types.yml (e.g., "go", "gherkin")
	Weight          int                     // Scheduling weight for resource allocation (host pool)
	Container       bool                    // Whether this runs in Docker (DEPRECATED: use PoolAllocation)
	HostInstalled   bool                    // Whether this runs on host system (opposite of Container)
	PoolAllocation  resource.PoolAllocation // Dual pool allocation for scheduling
	DependsOn       []UnitID                // Work units that must complete first (within module)
	Cached          bool                    // Skip execution if up-to-date
	Metadata        map[string]any          // Context-specific configuration
	Index           int                     // Position in input slice for result ordering
	Tags            TagSummary              // Classified tag data (test UoWs only)
}

// DependsOnComponents returns the component names from DependsOn.
// This is used by the scheduler for intra-module dependency tracking.
func (s UnitSpec) DependsOnComponents() []string {
	if len(s.DependsOn) == 0 {
		return nil
	}
	components := make([]string, len(s.DependsOn))
	for i, dep := range s.DependsOn {
		components[i] = dep.Component
	}
	return components
}

// DisplayName returns context-aware compact display name.
func (s UnitSpec) DisplayName() string {
	return s.ID.DisplayName()
}

// Longname returns the full ID: context:module:component:tool[:extra]
func (s UnitSpec) Longname() string {
	return s.ID.Longname()
}

// OutDir returns the output directory for this unit.
func (s UnitSpec) OutDir() string {
	return s.ID.OutDir()
}

// SpecDir returns the specification directory path (for test specs).
func (s UnitSpec) SpecDir() string {
	if testset, ok := s.ID.Extra["testset"]; ok && testset != "" {
		return s.ID.OutDir() + "/specs"
	}
	return ""
}

// ImplDir returns the implementation directory path.
func (s UnitSpec) ImplDir() string {
	return s.ID.OutDir() + "/impl"
}

// NewBuildSpec creates a UnitSpec for a build operation.
func NewBuildSpec(module, component, tool string) UnitSpec {
	return UnitSpec{
		ID: UnitID{
			Context:   ContextBuild,
			Module:    module,
			Component: component,
			Tool:      tool,
		},
		ComponentType:   component,
		Weight:          1,
		Container:     false,
		HostInstalled: true, // Opposite of Container
		DependsOn:       []UnitID{},
		Cached:          false,
		Metadata:        make(map[string]any),
	}
}

// NewTestSpec creates a UnitSpec for a test operation.
func NewTestSpec(module, component, tool, testset string) UnitSpec {
	return UnitSpec{
		ID: UnitID{
			Context:   ContextTest,
			Module:    module,
			Component: component,
			Tool:      tool,
			Extra:     map[string]string{"testset": testset},
		},
		ComponentType:   component,
		Weight:          1,
		Container:     false,
		HostInstalled: true, // Opposite of Container
		DependsOn:       []UnitID{},
		Cached:          false,
		Metadata:        make(map[string]any),
	}
}

// NewLintSpec creates a UnitSpec for a lint operation.
func NewLintSpec(module, component, provider string) UnitSpec {
	return UnitSpec{
		ID: UnitID{
			Context:   ContextLint,
			Module:    module,
			Component: component,
			Tool:      provider,
		},
		ComponentType:   component,
		Weight:          1,
		Container:     false,
		HostInstalled: true, // Opposite of Container
		DependsOn:       []UnitID{},
		Cached:          false,
		Metadata:        make(map[string]any),
	}
}

// NewScanSpec creates a UnitSpec for a scan operation.
func NewScanSpec(module, component, scanner string) UnitSpec {
	return UnitSpec{
		ID: UnitID{
			Context:   ContextScan,
			Module:    module,
			Component: component,
			Tool:      scanner,
		},
		ComponentType:   component,
		Weight:          1,
		Container:     false,
		HostInstalled: true, // Opposite of Container
		DependsOn:       []UnitID{},
		Cached:          false,
		Metadata:        make(map[string]any),
	}
}

// ============================================================================
// UnitSpecPort interface implementation
// ============================================================================

// GetID returns the unique identifier for this work unit.
// Implements interfaces.UnitSpecPort.
func (s UnitSpec) GetID() UnitID {
	return s.ID
}

// GetComponentType returns the component type from component-types.yml.
// Implements interfaces.UnitSpecPort.
func (s UnitSpec) GetComponentType() string {
	return s.ComponentType
}

// GetWeight returns the scheduling weight for resource allocation.
// Implements interfaces.UnitSpecPort.
func (s UnitSpec) GetWeight() int {
	return s.Weight
}

// IsContainer returns whether this runs in Docker.
// Implements interfaces.UnitSpecPort.
func (s UnitSpec) IsContainer() bool {
	return s.Container
}

// IsCached returns whether execution should be skipped.
// Implements interfaces.UnitSpecPort.
func (s UnitSpec) IsCached() bool {
	return s.Cached
}

// GetDependsOn returns work units that must complete first (within module).
// Implements interfaces.UnitSpecPort.
func (s UnitSpec) GetDependsOn() []UnitID {
	return s.DependsOn
}

// GetPoolAllocation returns the resource pool allocation for this spec.
// If PoolAllocation is explicitly set (non-zero weights), returns it directly.
// Otherwise, derives allocation from the legacy Container/Weight fields:
//   - Host-only (Container=false): HostWeight=Weight, DockerWeight=0
//   - Container (Container=true): HostWeight=Weight, DockerWeight=Weight
func (s UnitSpec) GetPoolAllocation() resource.PoolAllocation {
	// If explicitly set, use it
	if s.PoolAllocation.HostWeight > 0 || s.PoolAllocation.DockerWeight > 0 {
		return s.PoolAllocation
	}

	// Legacy derivation from Container/Weight fields
	alloc := resource.PoolAllocation{
		HostWeight: s.Weight,
	}
	if s.Container {
		alloc.DockerWeight = s.Weight // Container uses both pools with same weight
	}
	return alloc
}
