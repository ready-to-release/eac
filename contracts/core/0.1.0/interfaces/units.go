package interfaces

import "time"

// UnitIDPort uniquely identifies a unit of work.
// Format: context:module:component:tool[:extra]
type UnitIDPort interface {
	// GetContext returns the operation type: "build", "test", "lint", "scan"
	GetContext() string

	// GetModule returns the module moniker (e.g., "eac-core")
	GetModule() string

	// GetComponent returns the component name (e.g., "go", "gherkin")
	GetComponent() string

	// GetTool returns the handler/provider/scanner (e.g., "go", "gotest", "golangci-lint")
	GetTool() string

	// GetSpec returns the spec name for BDD tests (e.g., "build-module")
	GetSpec() string

	// Shortname returns display name: module:component, or just spec name for BDD tests
	Shortname() string

	// Longname returns full ID: context:module:component:tool[:extra]
	Longname() string

	// String returns the string representation (same as Longname)
	String() string

	// OutDir returns the output directory path for this unit
	OutDir() string
}

// UnitSpecPort represents the input specification for a unit of work.
// It describes what to execute and how to schedule it.
type UnitSpecPort interface {
	// GetID returns the unique identifier for this work unit
	GetID() UnitIDPort

	// GetComponentType returns the component type from component-types.yml
	GetComponentType() string

	// GetWeight returns the scheduling weight for resource allocation.
	// Deprecated: Use GetPoolAllocation().GetHostWeight() for scheduling.
	GetWeight() int

	// IsContainer returns whether this runs in Docker.
	// Deprecated: Use GetPoolAllocation().IsContainer() instead.
	IsContainer() bool

	// IsCached returns whether execution should be skipped
	IsCached() bool

	// GetDependsOn returns work units that must complete first (within module)
	GetDependsOn() []UnitIDPort

	// GetPoolAllocation returns the resource pool allocation for this spec.
	// Returns HostWeight and DockerWeight for dual-pool scheduling.
	GetPoolAllocation() PoolAllocationPort
}

// PoolAllocationPort describes the resource pools a work unit needs.
// This interface provides read-only access to pool allocation data.
type PoolAllocationPort interface {
	// GetHostWeight returns the weight to acquire from the host pool.
	// All work units consume host resources.
	GetHostWeight() int

	// GetDockerWeight returns the weight to acquire from the docker pool.
	// Only container tools consume docker resources.
	// Zero means host-only execution.
	GetDockerWeight() int

	// IsContainer returns true if this allocation requires docker resources.
	IsContainer() bool

	// TotalWeight returns the effective scheduling weight.
	// For bin-packing, this is the maximum of HostWeight and DockerWeight
	// since both must be satisfied simultaneously.
	TotalWeight() int
}

// UnitResultPort represents the outcome of executing a unit of work.
type UnitResultPort interface {
	// GetID returns the work unit that was executed
	GetID() UnitIDPort

	// GetExitCode returns 0=success, -1=cached/skipped, >0=failure
	GetExitCode() int

	// GetDuration returns how long execution took
	GetDuration() time.Duration

	// GetLogPath returns the path to the execution log
	GetLogPath() string

	// Success returns true if the unit executed successfully
	Success() bool

	// Cached returns true if the unit was skipped due to cache hit
	Cached() bool

	// Failed returns true if the unit execution failed
	Failed() bool
}

// UnitRegistryPort provides access to work units.
type UnitRegistryPort interface {
	// All returns all work units
	All() []UnitSpecPort

	// ByModule returns work units for a specific module
	ByModule(module string) []UnitSpecPort

	// ByComponent returns work units for a specific module:component
	ByComponent(module, component string) []UnitSpecPort

	// ByID returns the work unit with the given ID, or nil if not found.
	// ID should be the Longname() format: context:module:component:tool[:extra]
	ByID(id string) UnitSpecPort

	// Count returns the total number of work units
	Count() int
}

// UnitResolverPort resolves modules to work unit specifications.
// This is the single source of truth for component → tool → work unit mapping.
// Core implementation lives in core/resolver; commands use this port interface.
type UnitResolverPort interface {
	// ResolveForBuild returns work units for buildable components in a module.
	// Returns one UnitSpec per buildable component.
	// Respects build_after dependencies defined in component-types.yml.
	ResolveForBuild(module ModuleContractPort, cached map[string]bool) []UnitSpecPort

	// ResolveForLint returns work units for lintable components in a module.
	// May return multiple UnitSpecs per component (one per lint provider).
	ResolveForLint(module ModuleContractPort, cached map[string]bool) []UnitSpecPort

	// ResolveForScan returns work units for scannable components in a module.
	// Returns multiple UnitSpecs per component (one per scanner category).
	// If categories is empty, uses defaults from component-types.yml.
	ResolveForScan(module ModuleContractPort, categories []string, cached map[string]bool) []UnitSpecPort
}
