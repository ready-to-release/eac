package mocks

import (
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Compile-time interface checks.
var (
	_ core.UnitIDPort     = (*MockUnitID)(nil)
	_ core.UnitSpecPort   = (*MockUnitSpec)(nil)
	_ core.UnitResultPort = (*MockUnitResult)(nil)
	_ core.UnitRegistryPort = (*MockUnitRegistry)(nil)
)

// MockUnitID implements core.UnitIDPort for testing.
type MockUnitID struct {
	ContextVal   string
	ModuleVal    string
	ComponentVal string
	ToolVal      string
	SpecVal      string
	ExtraVal     map[string]string
}

func (m *MockUnitID) GetAction() string            { return m.ContextVal }
func (m *MockUnitID) GetModule() string           { return m.ModuleVal }
func (m *MockUnitID) GetComponent() string        { return m.ComponentVal }
func (m *MockUnitID) GetTool() string             { return m.ToolVal }
func (m *MockUnitID) GetSpec() string             { return m.SpecVal }
func (m *MockUnitID) GetExtra() map[string]string { return m.ExtraVal }

func (m *MockUnitID) DisplayName() string {
	switch m.ContextVal {
	case "test":
		if m.SpecVal != "" {
			return m.SpecVal + ": " + m.ToolVal
		}
		testname := m.ExtraVal["testname"]
		if testname == "" {
			testname = m.ComponentVal
		}
		return testname + ": unit"
	case "build":
		if m.ComponentVal == m.ToolVal || m.ToolVal == "" {
			return m.ModuleVal + ": " + m.ComponentVal
		}
		return m.ModuleVal + ": " + m.ComponentVal + ": " + m.ToolVal
	case "lint":
		return "lint:" + m.ComponentVal + ":" + m.ToolVal
	case "scan":
		category := m.ExtraVal["category"]
		if category == "" {
			category = m.ToolVal
		}
		return "scan:" + m.ComponentVal + ":" + category
	default:
		return m.ComponentVal
	}
}

func (m *MockUnitID) Longname() string {
	if m.SpecVal != "" {
		return m.ModuleVal + ":spec:" + m.SpecVal
	}
	return m.ContextVal + ":" + m.ModuleVal + ":" + m.ComponentVal + ":" + m.ToolVal
}

func (m *MockUnitID) String() string { return m.Longname() }

func (m *MockUnitID) OutDir() string {
	dirName := m.ComponentVal
	if m.ToolVal != "" {
		dirName += "-" + m.ToolVal
	}
	if len(m.ExtraVal) > 0 {
		keys := make([]string, 0, len(m.ExtraVal))
		for k := range m.ExtraVal {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if v := m.ExtraVal[k]; v != "" {
				dirName += "-" + v
			}
		}
	}
	return "out/" + m.ContextVal + "/" + m.ModuleVal + "/" + dirName
}

// NewMockUnitID creates a MockUnitID with a fluent builder pattern.
func NewMockUnitID() *MockUnitID {
	return &MockUnitID{
		ContextVal: "build",
		ToolVal:    "go",
	}
}

func (m *MockUnitID) WithContext(ctx string) *MockUnitID {
	m.ContextVal = ctx
	return m
}

func (m *MockUnitID) WithModule(mod string) *MockUnitID {
	m.ModuleVal = mod
	return m
}

func (m *MockUnitID) WithComponent(comp string) *MockUnitID {
	m.ComponentVal = comp
	return m
}

func (m *MockUnitID) WithTool(tool string) *MockUnitID {
	m.ToolVal = tool
	return m
}

func (m *MockUnitID) WithSpec(spec string) *MockUnitID {
	m.SpecVal = spec
	return m
}

func (m *MockUnitID) WithExtra(extra map[string]string) *MockUnitID {
	m.ExtraVal = extra
	return m
}

// MockUnitSpec implements core.UnitSpecPort for testing.
type MockUnitSpec struct {
	IDVal            core.UnitIDPort
	ComponentTypeVal string
	WeightVal        int
	ContainerVal     bool
	CachedVal        bool
	DependsOnVal     []core.UnitIDPort
	HostWeightVal    int
	DockerWeightVal  int
}

func (m *MockUnitSpec) GetID() core.UnitIDPort         { return m.IDVal }
func (m *MockUnitSpec) GetComponentType() string        { return m.ComponentTypeVal }
func (m *MockUnitSpec) GetWeight() int                  { return m.WeightVal }
func (m *MockUnitSpec) IsContainer() bool               { return m.ContainerVal }
func (m *MockUnitSpec) IsCached() bool                  { return m.CachedVal }
func (m *MockUnitSpec) GetDependsOn() []core.UnitIDPort { return m.DependsOnVal }
func (m *MockUnitSpec) GetPoolAllocation() core.PoolAllocationPort {
	return &MockPoolAllocation{
		hostWeight:   m.HostWeightVal,
		dockerWeight: m.DockerWeightVal,
	}
}

// MockPoolAllocation implements core.PoolAllocationPort for testing.
type MockPoolAllocation struct {
	hostWeight   int
	dockerWeight int
}

func (m *MockPoolAllocation) GetHostWeight() int   { return m.hostWeight }
func (m *MockPoolAllocation) GetDockerWeight() int { return m.dockerWeight }
func (m *MockPoolAllocation) IsContainer() bool    { return m.dockerWeight > 0 }
func (m *MockPoolAllocation) TotalWeight() int {
	if m.dockerWeight > m.hostWeight {
		return m.dockerWeight
	}
	return m.hostWeight
}

// NewMockUnitSpec creates a MockUnitSpec with a fluent builder pattern.
func NewMockUnitSpec(module, component string) *MockUnitSpec {
	return &MockUnitSpec{
		IDVal: NewMockUnitID().WithModule(module).WithComponent(component),
		ComponentTypeVal: component,
		WeightVal:        1,
		DependsOnVal:     []core.UnitIDPort{},
	}
}

func (m *MockUnitSpec) WithID(id core.UnitIDPort) *MockUnitSpec {
	m.IDVal = id
	return m
}

func (m *MockUnitSpec) WithComponentType(ct string) *MockUnitSpec {
	m.ComponentTypeVal = ct
	return m
}

func (m *MockUnitSpec) WithWeight(w int) *MockUnitSpec {
	m.WeightVal = w
	return m
}

func (m *MockUnitSpec) WithContainer(c bool) *MockUnitSpec {
	m.ContainerVal = c
	return m
}

func (m *MockUnitSpec) WithCached(c bool) *MockUnitSpec {
	m.CachedVal = c
	return m
}

func (m *MockUnitSpec) WithDependsOn(deps ...core.UnitIDPort) *MockUnitSpec {
	m.DependsOnVal = deps
	return m
}

func (m *MockUnitSpec) WithHostWeight(w int) *MockUnitSpec {
	m.HostWeightVal = w
	return m
}

func (m *MockUnitSpec) WithDockerWeight(w int) *MockUnitSpec {
	m.DockerWeightVal = w
	return m
}

// MockUnitResult implements core.UnitResultPort for testing.
type MockUnitResult struct {
	IDVal       core.UnitIDPort
	ExitCodeVal int
	DurationVal time.Duration
	LogPathVal  string
}

func (m *MockUnitResult) GetID() core.UnitIDPort      { return m.IDVal }
func (m *MockUnitResult) GetExitCode() int             { return m.ExitCodeVal }
func (m *MockUnitResult) GetDuration() time.Duration   { return m.DurationVal }
func (m *MockUnitResult) GetLogPath() string           { return m.LogPathVal }
func (m *MockUnitResult) Success() bool                { return m.ExitCodeVal == 0 }
func (m *MockUnitResult) Cached() bool                 { return m.ExitCodeVal == -1 }
func (m *MockUnitResult) Failed() bool                 { return m.ExitCodeVal > 0 }

// NewMockUnitResult creates a MockUnitResult with a fluent builder pattern.
func NewMockUnitResult(module, component string) *MockUnitResult {
	return &MockUnitResult{
		IDVal:       NewMockUnitID().WithModule(module).WithComponent(component),
		ExitCodeVal: 0,
		DurationVal: time.Second,
	}
}

func (m *MockUnitResult) WithID(id core.UnitIDPort) *MockUnitResult {
	m.IDVal = id
	return m
}

func (m *MockUnitResult) WithExitCode(code int) *MockUnitResult {
	m.ExitCodeVal = code
	return m
}

func (m *MockUnitResult) WithDuration(d time.Duration) *MockUnitResult {
	m.DurationVal = d
	return m
}

func (m *MockUnitResult) WithLogPath(path string) *MockUnitResult {
	m.LogPathVal = path
	return m
}

// MockUnitRegistry implements core.UnitRegistryPort for testing.
type MockUnitRegistry struct {
	units []core.UnitSpecPort
}

func NewMockUnitRegistry() *MockUnitRegistry {
	return &MockUnitRegistry{
		units: []core.UnitSpecPort{},
	}
}

func (m *MockUnitRegistry) Add(units ...core.UnitSpecPort) *MockUnitRegistry {
	m.units = append(m.units, units...)
	return m
}

func (m *MockUnitRegistry) All() []core.UnitSpecPort {
	return m.units
}

func (m *MockUnitRegistry) ByModule(module string) []core.UnitSpecPort {
	var result []core.UnitSpecPort
	for _, u := range m.units {
		if u.GetID().GetModule() == module {
			result = append(result, u)
		}
	}
	return result
}

func (m *MockUnitRegistry) ByComponent(module, component string) []core.UnitSpecPort {
	var result []core.UnitSpecPort
	for _, u := range m.units {
		id := u.GetID()
		if id.GetModule() == module && id.GetComponent() == component {
			result = append(result, u)
		}
	}
	return result
}

func (m *MockUnitRegistry) ByID(id string) core.UnitSpecPort {
	for _, u := range m.units {
		if u.GetID().Longname() == id {
			return u
		}
	}
	return nil
}

func (m *MockUnitRegistry) Count() int {
	return len(m.units)
}
