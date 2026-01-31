package adapters

import (
	"time"

	"github.com/ready-to-release/eac/contracts/eac-core-interfaces"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

// Compile-time interface checks.
var (
	_ interfaces.UnitIDPort     = (*UnitIDAdapter)(nil)
	_ interfaces.UnitSpecPort   = (*UnitSpecAdapter)(nil)
	_ interfaces.UnitResultPort = (*UnitResultAdapter)(nil)
)

// UnitIDAdapter wraps a workunit.UnitID to implement interfaces.UnitIDPort.
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

func (a *UnitIDAdapter) GetContext() string   { return string(a.id.Context) }
func (a *UnitIDAdapter) GetModule() string    { return a.id.Module }
func (a *UnitIDAdapter) GetComponent() string { return a.id.Component }
func (a *UnitIDAdapter) GetTool() string      { return a.id.Tool }
func (a *UnitIDAdapter) GetSpec() string      { return a.id.Spec }
func (a *UnitIDAdapter) Shortname() string    { return a.id.Shortname() }
func (a *UnitIDAdapter) Longname() string     { return a.id.Longname() }
func (a *UnitIDAdapter) String() string       { return a.id.String() }
func (a *UnitIDAdapter) OutDir() string       { return a.id.OutDir() }

// AdaptUnitID is a convenience function to wrap a unit ID.
func AdaptUnitID(id workunit.UnitID) interfaces.UnitIDPort {
	return NewUnitIDAdapter(id)
}

// UnitSpecAdapter wraps a workunit.UnitSpec to implement interfaces.UnitSpecPort.
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

func (a *UnitSpecAdapter) GetID() interfaces.UnitIDPort {
	return AdaptUnitID(a.spec.ID)
}

func (a *UnitSpecAdapter) GetComponentType() string { return a.spec.ComponentType }
func (a *UnitSpecAdapter) GetWeight() int           { return a.spec.Weight }
func (a *UnitSpecAdapter) IsContainer() bool        { return a.spec.IsContainer }
func (a *UnitSpecAdapter) IsCached() bool           { return a.spec.Cached }

func (a *UnitSpecAdapter) GetDependsOn() []interfaces.UnitIDPort {
	if a.spec.DependsOn == nil {
		return nil
	}
	result := make([]interfaces.UnitIDPort, len(a.spec.DependsOn))
	for i, dep := range a.spec.DependsOn {
		result[i] = AdaptUnitID(dep)
	}
	return result
}

// AdaptUnitSpec is a convenience function to wrap a unit spec.
func AdaptUnitSpec(spec workunit.UnitSpec) interfaces.UnitSpecPort {
	return NewUnitSpecAdapter(spec)
}

// AdaptUnitSpecs wraps a slice of unit specs.
func AdaptUnitSpecs(specs []workunit.UnitSpec) []interfaces.UnitSpecPort {
	if specs == nil {
		return nil
	}
	result := make([]interfaces.UnitSpecPort, len(specs))
	for i, s := range specs {
		result[i] = AdaptUnitSpec(s)
	}
	return result
}

// UnitResultAdapter wraps a workunit.UnitResult to implement interfaces.UnitResultPort.
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

func (a *UnitResultAdapter) GetID() interfaces.UnitIDPort {
	return AdaptUnitID(a.result.ID)
}

func (a *UnitResultAdapter) GetExitCode() int           { return a.result.ExitCode }
func (a *UnitResultAdapter) GetDuration() time.Duration { return a.result.Duration }
func (a *UnitResultAdapter) GetLogPath() string         { return a.result.LogPath }
func (a *UnitResultAdapter) Success() bool              { return a.result.Success() }
func (a *UnitResultAdapter) Cached() bool               { return a.result.Cached() }
func (a *UnitResultAdapter) Failed() bool               { return a.result.Failed() }

// AdaptUnitResult is a convenience function to wrap a unit result.
func AdaptUnitResult(result workunit.UnitResult) interfaces.UnitResultPort {
	return NewUnitResultAdapter(result)
}

// AdaptUnitResults wraps a slice of unit results.
func AdaptUnitResults(results []workunit.UnitResult) []interfaces.UnitResultPort {
	if results == nil {
		return nil
	}
	result := make([]interfaces.UnitResultPort, len(results))
	for i, r := range results {
		result[i] = AdaptUnitResult(r)
	}
	return result
}
