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

	// GetWeight returns the scheduling weight for resource allocation
	GetWeight() int

	// IsContainer returns whether this runs in Docker
	IsContainer() bool

	// IsCached returns whether execution should be skipped
	IsCached() bool

	// GetDependsOn returns work units that must complete first (within module)
	GetDependsOn() []UnitIDPort
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

	// Count returns the total number of work units
	Count() int
}
