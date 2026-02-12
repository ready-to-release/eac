package adapters

import (
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// Compile-time interface checks.
var (
	_ core.UnitIDPort         = (*UnitIDAdapter)(nil)
	_ core.UnitSpecPort       = (*UnitSpecAdapter)(nil)
	_ core.UnitResultPort     = (*UnitResultAdapter)(nil)
	_ core.PoolAllocationPort = (*PoolAllocationAdapter)(nil)
)

// UnitIDAdapter wraps a workunit.UnitID to implement core.UnitIDPort.
type UnitIDAdapter struct {
	id workunit.UnitID
}

// NewUnitIDAdapter creates a new adapter wrapping a unit ID.
func NewUnitIDAdapter(id workunit.UnitID) *UnitIDAdapter {
	return &UnitIDAdapter{id: id}
}

// Unwrap returns the underlying concrete UnitID.
func (a *UnitIDAdapter) Unwrap() workunit.UnitID {
	return a.id
}

func (a *UnitIDAdapter) GetAction() string             { return string(a.id.Action) }
func (a *UnitIDAdapter) GetModule() string             { return a.id.Module }
func (a *UnitIDAdapter) GetComponentType() string      { return a.id.ComponentType }
func (a *UnitIDAdapter) GetComponentName() string      { return a.id.ComponentName }
func (a *UnitIDAdapter) GetComponent() string          { return a.id.ComponentName } // Deprecated: use GetComponentName()
func (a *UnitIDAdapter) GetTool() string               { return a.id.Tool }
func (a *UnitIDAdapter) GetSpec() string               { return a.id.Spec }
func (a *UnitIDAdapter) GetExtra() map[string]string   { return a.id.Extra }
func (a *UnitIDAdapter) DisplayName() string { return a.id.DisplayName() }
func (a *UnitIDAdapter) Longname() string              { return a.id.Longname() }
func (a *UnitIDAdapter) String() string                { return a.id.String() }
func (a *UnitIDAdapter) OutDir() string                { return a.id.OutDir() }

// AdaptUnitID is a convenience function to wrap a unit ID.
func AdaptUnitID(id workunit.UnitID) core.UnitIDPort {
	return NewUnitIDAdapter(id)
}

// UnitSpecAdapter wraps a workunit.UnitSpec to implement core.UnitSpecPort.
type UnitSpecAdapter struct {
	spec workunit.UnitSpec
}

// NewUnitSpecAdapter creates a new adapter wrapping a unit spec.
func NewUnitSpecAdapter(spec workunit.UnitSpec) *UnitSpecAdapter {
	return &UnitSpecAdapter{spec: spec}
}

// Unwrap returns the underlying concrete UnitSpec.
func (a *UnitSpecAdapter) Unwrap() workunit.UnitSpec {
	return a.spec
}

func (a *UnitSpecAdapter) GetID() core.UnitIDPort {
	return AdaptUnitID(a.spec.ID)
}

func (a *UnitSpecAdapter) GetComponentType() string { return a.spec.ComponentType }
func (a *UnitSpecAdapter) IsCached() bool           { return a.spec.Cached }

func (a *UnitSpecAdapter) GetDependsOn() []core.UnitIDPort {
	if a.spec.DependsOn == nil {
		return nil
	}
	result := make([]core.UnitIDPort, len(a.spec.DependsOn))
	for i, dep := range a.spec.DependsOn {
		result[i] = AdaptUnitID(dep)
	}
	return result
}

func (a *UnitSpecAdapter) GetPoolAllocation() core.PoolAllocationPort {
	alloc := a.spec.GetPoolAllocation()
	return &PoolAllocationAdapter{alloc: alloc}
}

// PoolAllocationAdapter wraps a core.PoolAllocationPort value (typically core.PoolAllocation)
// to satisfy core.PoolAllocationPort across the adapter boundary.
type PoolAllocationAdapter struct {
	alloc core.PoolAllocationPort
}

func (a *PoolAllocationAdapter) GetHostWeight() int   { return a.alloc.GetHostWeight() }
func (a *PoolAllocationAdapter) GetDockerWeight() int { return a.alloc.GetDockerWeight() }
func (a *PoolAllocationAdapter) IsContainer() bool    { return a.alloc.IsContainer() }
func (a *PoolAllocationAdapter) TotalWeight() int     { return a.alloc.TotalWeight() }

// AdaptUnitSpec is a convenience function to wrap a unit spec.
func AdaptUnitSpec(spec workunit.UnitSpec) core.UnitSpecPort {
	return NewUnitSpecAdapter(spec)
}

// AdaptUnitSpecs wraps a slice of unit specs.
func AdaptUnitSpecs(specs []workunit.UnitSpec) []core.UnitSpecPort {
	return AdaptSlice(specs, AdaptUnitSpec)
}

// UnitResultAdapter wraps a workunit.UnitResult to implement core.UnitResultPort.
type UnitResultAdapter struct {
	result workunit.UnitResult
}

// NewUnitResultAdapter creates a new adapter wrapping a unit result.
func NewUnitResultAdapter(result workunit.UnitResult) *UnitResultAdapter {
	return &UnitResultAdapter{result: result}
}

// Unwrap returns the underlying concrete UnitResult.
func (a *UnitResultAdapter) Unwrap() workunit.UnitResult {
	return a.result
}

func (a *UnitResultAdapter) GetID() core.UnitIDPort {
	return AdaptUnitID(a.result.ID)
}

func (a *UnitResultAdapter) GetExitCode() int           { return a.result.ExitCode }
func (a *UnitResultAdapter) GetDuration() time.Duration { return a.result.Duration }
func (a *UnitResultAdapter) GetLogPath() string         { return a.result.LogPath }
func (a *UnitResultAdapter) Success() bool              { return a.result.Success() }
func (a *UnitResultAdapter) Cached() bool               { return a.result.Cached() }
func (a *UnitResultAdapter) Failed() bool               { return a.result.Failed() }

// AdaptUnitResult is a convenience function to wrap a unit result.
func AdaptUnitResult(result workunit.UnitResult) core.UnitResultPort {
	return NewUnitResultAdapter(result)
}

// AdaptUnitResults wraps a slice of unit results.
func AdaptUnitResults(results []workunit.UnitResult) []core.UnitResultPort {
	return AdaptSlice(results, AdaptUnitResult)
}
