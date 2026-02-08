package core

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// UnitID uniquely identifies a unit of work.
// Format: context:module:component:tool[:extra]
type UnitID struct {
	Action    ActionType        `json:"context"` // build, test, lint, scan (JSON tag kept as "context" for cache compat)
	Module    string            // module moniker (e.g., "core")
	Component string            // component name (e.g., "go", "gherkin")
	Tool      string            // handler/provider/scanner (e.g., "go", "gotest", "golangci-lint")
	Extra     map[string]string // context-specific (e.g., testset: "unit")
	Spec      string            // Spec name for BDD tests (godog, tscucumber), e.g., "build-module"
}

// Path returns module:component for component-level identification.
// Note: This is NOT a unique work unit identifier - use Longname() for that.
func (u UnitID) Path() string {
	return u.Module + ":" + u.Component
}

// ComponentName returns just the component name.
func (u UnitID) ComponentName() string {
	return u.Component
}

// DisplayName returns a context-aware compact display name.
// This is the primary method for TUI tabs and progress indicators.
func (u UnitID) DisplayName() string {
	switch u.Action {
	case ActionTest:
		if u.Spec != "" {
			return u.Spec + ": " + u.Tool
		}
		testname := u.Extra["testname"]
		if testname == "" {
			testname = u.Component
		}
		return testname + ": unit"

	case ActionBuild:
		if u.Component == u.Tool || u.Tool == "" {
			return u.Module + ": " + u.Component
		}
		return u.Module + ": " + u.Component + ": " + u.Tool

	case ActionLint:
		return "lint:" + u.Component + ":" + u.Tool

	case ActionScan:
		category := u.Extra["category"]
		if category == "" {
			category = u.Tool
		}
		return "scan:" + u.Component + ":" + category

	default:
		return u.Component
	}
}

// NormalDisplay returns a human-readable sentence describing the work unit.
func (u UnitID) NormalDisplay() string {
	switch u.Action {
	case ActionBuild:
		if u.Tool == "" || u.Tool == u.Component {
			return "Building " + u.Component + " in " + u.Module
		}
		return "Building " + u.Component + " in " + u.Module + " with " + u.Tool

	case ActionTest:
		if u.Spec != "" {
			return "Testing spec " + u.Spec + " in " + u.Module
		}
		testname := u.Extra["testname"]
		if testname == "" {
			testname = u.Component
		}
		return "Testing " + testname + " in " + u.Module

	case ActionLint:
		return "Linting " + u.Component + " in " + u.Module + " with " + u.Tool

	case ActionScan:
		category := u.Extra["category"]
		if category == "" {
			category = u.Tool
		}
		return "Scanning " + u.Component + " in " + u.Module + " for " + category

	default:
		return string(u.Action) + " " + u.Component + " in " + u.Module
	}
}

// DisplayKey returns the key used for disambiguation checking.
func (u UnitID) DisplayKey() string {
	if u.Action == ActionTest {
		if u.Spec != "" {
			return u.Spec
		}
		if testname := u.Extra["testname"]; testname != "" {
			return testname
		}
	}
	return u.Component
}

// TabLabel returns truncated name for TUI tabs (max width).
func (u UnitID) TabLabel(maxWidth int) string {
	name := u.Component
	if u.Spec != "" {
		name = u.Spec
	}
	if maxWidth <= 3 {
		if len(name) > maxWidth {
			return name[:maxWidth]
		}
		return name
	}
	if len(name) > maxWidth {
		return name[:maxWidth-3] + "..."
	}
	return name
}

// Longname returns full ID: context:module:component:tool[:extra...]
func (u UnitID) Longname() string {
	base := fmt.Sprintf("%s:%s:%s:%s", u.Action, u.Module, u.Component, u.Tool)

	if len(u.Extra) > 0 {
		keys := make([]string, 0, len(u.Extra))
		for k := range u.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if v := u.Extra[k]; v != "" {
				base += ":" + v
			}
		}
	}
	return base
}

// String returns Longname.
func (u UnitID) String() string {
	return u.Longname()
}

// DirName returns the unique directory name for this unit.
func (u UnitID) DirName() string {
	dirName := u.Component
	if u.Tool != "" {
		dirName += "-" + u.Tool
	}

	if len(u.Extra) > 0 {
		keys := make([]string, 0, len(u.Extra))
		for k := range u.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if v := u.Extra[k]; v != "" {
				dirName += "-" + v
			}
		}
	}
	return dirName
}

// OutDir returns the unique output directory for this unit.
func (u UnitID) OutDir() string {
	return filepath.Join("out", string(u.Action), u.Module, u.DirName())
}

// LockFile returns the path to the execution lock.
func (u UnitID) LockFile() string {
	return filepath.Join(u.OutDir(), ".lock")
}

// StateCacheDir returns the cache directory for incremental state.
func (u UnitID) StateCacheDir() string {
	return filepath.Join(".cache", "eac", "incremental", string(u.Action), u.Module, u.DirName())
}

// StateFile returns the path to the cache state file.
func (u UnitID) StateFile() string {
	return filepath.Join(u.StateCacheDir(), "state.json")
}

// LogFile returns the path to the execution log.
func (u UnitID) LogFile() string {
	return filepath.Join(u.OutDir(), "execution.log")
}

// ResultsFile returns the path to results (test/lint/scan).
func (u UnitID) ResultsFile() string {
	return filepath.Join(u.OutDir(), "results.json")
}

// GetAction implements UnitIDPort.
func (u UnitID) GetAction() string { return string(u.Action) }

// GetModule implements UnitIDPort.
func (u UnitID) GetModule() string { return u.Module }

// GetComponent implements UnitIDPort.
func (u UnitID) GetComponent() string { return u.Component }

// GetTool implements UnitIDPort.
func (u UnitID) GetTool() string { return u.Tool }

// GetSpec implements UnitIDPort.
func (u UnitID) GetSpec() string { return u.Spec }

// GetExtra implements UnitIDPort.
func (u UnitID) GetExtra() map[string]string { return u.Extra }

// ============================================================================
// TagSummary
// ============================================================================

// TagSummary carries classified tag data for a work unit.
// Value type — always present on test UoWs (zero value is valid empty).
type TagSummary struct {
	All          []string `json:"all"`
	Levels       []string `json:"levels"`
	Verification []string `json:"verification"`
	RiskControls []string `json:"risk_controls"`
	SystemDeps   []string `json:"system_deps"`
	ModuleDeps   []string `json:"module_deps"`
	IsGxP        bool     `json:"is_gxp"`
	HasManual    bool     `json:"has_manual"`
}

// IsEmpty returns true if no tags are present.
func (ts TagSummary) IsEmpty() bool { return len(ts.All) == 0 }

// Summary returns a compact human-readable summary of the tag data.
func (ts TagSummary) Summary() string {
	if ts.IsEmpty() {
		return ""
	}

	var parts []string

	if len(ts.Levels) > 0 {
		parts = append(parts, trimPrefixAndJoin(ts.Levels, "@", ","))
	}

	if len(ts.Verification) > 0 {
		parts = append(parts, trimPrefixAndJoin(ts.Verification, "@", ","))
	}

	if ts.IsGxP {
		parts = append(parts, "gxp")
	}

	if ts.HasManual {
		parts = append(parts, "manual")
	}

	if n := len(ts.RiskControls); n == 1 {
		parts = append(parts, "1 control")
	} else if n > 0 {
		parts = append(parts, fmt.Sprintf("%d controls", n))
	}

	return strings.Join(parts, " ")
}

// Merge returns a new TagSummary that is the union of ts and other.
func (ts TagSummary) Merge(other TagSummary) TagSummary {
	return TagSummary{
		All:          mergeAndSort(ts.All, other.All),
		Levels:       mergeAndSort(ts.Levels, other.Levels),
		Verification: mergeAndSort(ts.Verification, other.Verification),
		RiskControls: mergeAndSort(ts.RiskControls, other.RiskControls),
		SystemDeps:   mergeAndSort(ts.SystemDeps, other.SystemDeps),
		ModuleDeps:   mergeAndSort(ts.ModuleDeps, other.ModuleDeps),
		IsGxP:        ts.IsGxP || other.IsGxP,
		HasManual:    ts.HasManual || other.HasManual,
	}
}

func mergeAndSort(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	for _, s := range a {
		seen[s] = true
	}
	for _, s := range b {
		seen[s] = true
	}
	if len(seen) == 0 {
		return nil
	}
	result := make([]string, 0, len(seen))
	for s := range seen {
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}

func trimPrefixAndJoin(tags []string, prefix, sep string) string {
	stripped := make([]string, len(tags))
	for i, t := range tags {
		stripped[i] = strings.TrimPrefix(t, prefix)
	}
	return strings.Join(stripped, sep)
}

// ============================================================================
// PoolAllocation (value type for resource scheduling)
// ============================================================================

// PoolType identifies which resource pool a work unit uses.
type PoolType string

const (
	// PoolHost represents the host system resource pool.
	PoolHost PoolType = "host"

	// PoolDocker represents the Docker/container resource pool.
	PoolDocker PoolType = "docker"
)

// PoolAllocation describes the resource pools a work unit needs.
type PoolAllocation struct {
	// HostWeight is the weight to acquire from the host pool.
	HostWeight int

	// DockerWeight is the weight to acquire from the docker pool.
	DockerWeight int
}

// GetHostWeight returns the weight to acquire from the host pool.
func (a PoolAllocation) GetHostWeight() int { return a.HostWeight }

// GetDockerWeight returns the weight to acquire from the docker pool.
func (a PoolAllocation) GetDockerWeight() int { return a.DockerWeight }

// IsContainer returns true if this allocation requires docker resources.
func (a PoolAllocation) IsContainer() bool { return a.DockerWeight > 0 }

// TotalWeight returns the effective scheduling weight.
func (a PoolAllocation) TotalWeight() int {
	if a.DockerWeight > a.HostWeight {
		return a.DockerWeight
	}
	return a.HostWeight
}

// HostOnlyAllocation creates an allocation for host-only work.
func HostOnlyAllocation(weight int) PoolAllocation {
	return PoolAllocation{HostWeight: weight, DockerWeight: 0}
}

// ContainerAllocation creates an allocation for container work.
func ContainerAllocation(hostWeight, dockerWeight int) PoolAllocation {
	return PoolAllocation{HostWeight: hostWeight, DockerWeight: dockerWeight}
}

// ============================================================================
// UnitSpec (moved from go/core/workunit)
// ============================================================================

// UnitSpec represents the input specification for a unit of work.
// It describes what to execute and how to schedule it.
type UnitSpec struct {
	ID             UnitID         // Unique identifier for this work unit
	ComponentType  string         // From component-types.yml (e.g., "go", "gherkin")
	Weight         int            // Scheduling weight for resource allocation (host pool)
	Container      bool           // Whether this runs in Docker (DEPRECATED: use PoolAllocation)
	HostInstalled  bool           // Whether this runs on host system (opposite of Container)
	PoolAllocation PoolAllocation // Dual pool allocation for scheduling
	DependsOn      []UnitID       // Work units that must complete first (within module)
	Cached         bool           // Skip execution if up-to-date
	Metadata       map[string]any // Context-specific configuration
	Index          int            // Position in input slice for result ordering
	Tags           TagSummary     // Classified tag data (test UoWs only)
}

// DependsOnComponents returns the component names from DependsOn.
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
func (s UnitSpec) DisplayName() string { return s.ID.DisplayName() }

// Longname returns the full ID: context:module:component:tool[:extra]
func (s UnitSpec) Longname() string { return s.ID.Longname() }

// OutDir returns the output directory for this unit.
func (s UnitSpec) OutDir() string { return s.ID.OutDir() }

// SpecDir returns the specification directory path (for test specs).
func (s UnitSpec) SpecDir() string {
	if _, ok := s.ID.Extra["testset"]; ok {
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
			Action:    ActionBuild,
			Module:    module,
			Component: component,
			Tool:      tool,
		},
		ComponentType: component,
		Weight:        1,
		Container:     false,
		HostInstalled: true,
		DependsOn:     []UnitID{},
		Cached:        false,
		Metadata:      make(map[string]any),
	}
}

// NewTestSpec creates a UnitSpec for a test operation.
func NewTestSpec(module, component, tool, testset string) UnitSpec {
	return UnitSpec{
		ID: UnitID{
			Action:    ActionTest,
			Module:    module,
			Component: component,
			Tool:      tool,
			Extra:     map[string]string{"testset": testset},
		},
		ComponentType: component,
		Weight:        1,
		Container:     false,
		HostInstalled: true,
		DependsOn:     []UnitID{},
		Cached:        false,
		Metadata:      make(map[string]any),
	}
}

// NewLintSpec creates a UnitSpec for a lint operation.
func NewLintSpec(module, component, provider string) UnitSpec {
	return UnitSpec{
		ID: UnitID{
			Action:    ActionLint,
			Module:    module,
			Component: component,
			Tool:      provider,
		},
		ComponentType: component,
		Weight:        1,
		Container:     false,
		HostInstalled: true,
		DependsOn:     []UnitID{},
		Cached:        false,
		Metadata:      make(map[string]any),
	}
}

// NewScanSpec creates a UnitSpec for a scan operation.
func NewScanSpec(module, component, scanner string) UnitSpec {
	return UnitSpec{
		ID: UnitID{
			Action:    ActionScan,
			Module:    module,
			Component: component,
			Tool:      scanner,
		},
		ComponentType: component,
		Weight:        1,
		Container:     false,
		HostInstalled: true,
		DependsOn:     []UnitID{},
		Cached:        false,
		Metadata:      make(map[string]any),
	}
}

// GetID implements UnitSpecPort.
func (s UnitSpec) GetID() UnitIDPort { return s.ID }

// GetComponentType implements UnitSpecPort.
func (s UnitSpec) GetComponentType() string { return s.ComponentType }

// GetWeight implements UnitSpecPort.
func (s UnitSpec) GetWeight() int { return s.Weight }

// IsContainerSpec implements UnitSpecPort.
func (s UnitSpec) IsContainerSpec() bool { return s.Container }

// IsCached implements UnitSpecPort.
func (s UnitSpec) IsCached() bool { return s.Cached }

// GetDependsOn implements UnitSpecPort.
func (s UnitSpec) GetDependsOn() []UnitIDPort {
	result := make([]UnitIDPort, len(s.DependsOn))
	for i, dep := range s.DependsOn {
		result[i] = dep
	}
	return result
}

// GetPoolAllocation implements UnitSpecPort.
func (s UnitSpec) GetPoolAllocation() PoolAllocationPort {
	if s.PoolAllocation.HostWeight != 0 || s.PoolAllocation.DockerWeight != 0 {
		return s.PoolAllocation
	}
	alloc := PoolAllocation{HostWeight: s.Weight}
	if s.Container {
		alloc.DockerWeight = s.Weight
	}
	return alloc
}

// ============================================================================
// DisplayNameResolver
// ============================================================================

// DisplayNameResolver computes shortest unique names within a set.
type DisplayNameResolver struct {
	keyCounts map[string]int
}

// NewDisplayNameResolver creates a resolver for the given set of units.
func NewDisplayNameResolver(units []UnitID) *DisplayNameResolver {
	r := &DisplayNameResolver{keyCounts: make(map[string]int)}
	for _, u := range units {
		r.keyCounts[u.DisplayKey()]++
	}
	return r
}

// Resolve returns shortest unique name for unit.
func (r *DisplayNameResolver) Resolve(u UnitID) string {
	key := u.DisplayKey()
	if r.keyCounts[key] == 1 {
		return key
	}
	return u.Module + ":" + key
}

// ResolveTabLabel returns tab-constrained name.
func (r *DisplayNameResolver) ResolveTabLabel(u UnitID, maxWidth int) string {
	name := r.Resolve(u)
	if len(name) <= maxWidth {
		return name
	}
	if maxWidth > 3 {
		return name[:maxWidth-3] + "..."
	}
	return name[:maxWidth]
}

// NeedsDisambiguation returns true if the display key appears more than once.
func (r *DisplayNameResolver) NeedsDisambiguation(key string) bool {
	return r.keyCounts[key] > 1
}

// Count returns how many units have the given display key.
func (r *DisplayNameResolver) Count(key string) int {
	return r.keyCounts[key]
}
