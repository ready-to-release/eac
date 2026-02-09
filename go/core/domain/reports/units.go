package reports

import (
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/workunit"
)

var log = logging.C()

// UnitInfo represents a resolved unit of work with status information.
type UnitInfo struct {
	// ID is the full unit identifier
	ID string `json:"id" yaml:"id"`

	// Module is the module moniker
	Module string `json:"module" yaml:"module"`

	// Component is the component name
	Component string `json:"component" yaml:"component"`

	// Tool is the tool/handler name
	Tool string `json:"tool" yaml:"tool"`

	// ComponentType is the component type from component-types.yml
	ComponentType string `json:"component_type" yaml:"component_type"`

	// Weight is the scheduling weight
	Weight int `json:"weight" yaml:"weight"`

	// Container indicates if this runs in Docker
	Container bool `json:"container" yaml:"container"`

	// HostInstalled indicates if this runs on host system
	HostInstalled bool `json:"host_installed" yaml:"host_installed"`

	// CacheStatus contains cache state information
	CacheStatus *CacheStatus `json:"cache_status" yaml:"cache_status"`

	// Dependencies contains intra-module dependencies
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	// Metadata contains additional unit metadata
	Metadata map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// CacheStatus represents the cache state of a unit.
type CacheStatus struct {
	// State is the cache state: "up_to_date", "stale", "new"
	State string `json:"state" yaml:"state"`

	// SourceHash is the hash of source files
	SourceHash string `json:"source_hash,omitempty" yaml:"source_hash,omitempty"`

	// ExecutedAt is when this unit was last executed
	ExecutedAt time.Time `json:"executed_at,omitempty" yaml:"executed_at,omitempty"`

	// Reason explains why the unit needs execution (if stale)
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// SkippedComponent represents a component that was skipped (no handler configured).
type SkippedComponent struct {
	// Module is the module moniker
	Module string `json:"module" yaml:"module"`

	// Component is the component name
	Component string `json:"component" yaml:"component"`

	// ComponentType is the component type from module.yml
	ComponentType string `json:"component_type" yaml:"component_type"`

	// Reason explains why this component was skipped
	Reason string `json:"reason" yaml:"reason"`

	// ConfigHint suggests what config to add
	ConfigHint string `json:"config_hint,omitempty" yaml:"config_hint,omitempty"`
}

// UnitReport contains the full unit report for a framework.
type UnitReport struct {
	// Framework is the framework type (build, test, lint, scan)
	Framework Framework `json:"framework" yaml:"framework"`

	// Units is the list of all resolved units
	Units []*UnitInfo `json:"units" yaml:"units"`

	// Skipped lists components without handlers for this framework
	Skipped []*SkippedComponent `json:"skipped,omitempty" yaml:"skipped,omitempty"`

	// Summary contains aggregate statistics
	Summary *UnitSummary `json:"summary" yaml:"summary"`
}

// UnitSummary contains aggregate statistics about units.
type UnitSummary struct {
	Total         int            `json:"total" yaml:"total"`
	Cached        int            `json:"cached" yaml:"cached"`
	Stale         int            `json:"stale" yaml:"stale"`
	New           int            `json:"new" yaml:"new"`
	Container     int            `json:"container" yaml:"container"`
	HostInstalled int            `json:"host_installed" yaml:"host_installed"`
	ByModule      map[string]int `json:"by_module" yaml:"by_module"`
	ByTool        map[string]int `json:"by_tool" yaml:"by_tool"`
}

// GetUnits returns the unit report for a specific framework.
func GetUnits(workspaceRoot string, framework Framework) (*UnitReport, error) {
	// Load module contracts
	moduleReport, err := GetModuleContracts(workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Load config
	cfg := config.Global()

	// Collect all units and skipped components
	var units []*UnitInfo
	var skipped []*SkippedComponent
	stateMgr := workunit.NewStateManager(workspaceRoot)

	for _, mod := range moduleReport.Modules {
		modUnits, modSkipped := resolveUnitsFromConfig(mod, framework, cfg, workspaceRoot, stateMgr)
		units = append(units, modUnits...)
		skipped = append(skipped, modSkipped...)
	}

	// Sort units by module, then component, then tool
	sort.Slice(units, func(i, j int) bool {
		if units[i].Module != units[j].Module {
			return units[i].Module < units[j].Module
		}
		if units[i].Component != units[j].Component {
			return units[i].Component < units[j].Component
		}
		return units[i].Tool < units[j].Tool
	})

	// Sort skipped by module, then component
	sort.Slice(skipped, func(i, j int) bool {
		if skipped[i].Module != skipped[j].Module {
			return skipped[i].Module < skipped[j].Module
		}
		return skipped[i].Component < skipped[j].Component
	})

	// Build summary
	summary := buildUnitSummary(units)

	return &UnitReport{
		Framework: framework,
		Units:     units,
		Skipped:   skipped,
		Summary:   summary,
	}, nil
}

// resolveUnitsFromConfig resolves units using config information directly.
// This doesn't require build handlers to be registered.
// Returns both resolved units and skipped components (those without handlers).
func resolveUnitsFromConfig(mod *modules.ModuleContract, framework Framework, cfg *config.EACConfig, workspaceRoot string, stateMgr *workunit.StateManager) ([]*UnitInfo, []*SkippedComponent) {
	var units []*UnitInfo
	var skipped []*SkippedComponent

	if cfg == nil {
		return units, skipped
	}

	for compName := range mod.Components {
		entry := mod.Components[compName]
		if entry == nil {
			continue
		}

		compType := mod.Components.GetComponentType(compName)
		typeConfig := cfg.ComponentKinds.Get(compType)

		switch framework {
		case FrameworkBuild:
			if typeConfig != nil && typeConfig.IsBuildable() {
				unit := createUnitInfo(mod, compName, compType, typeConfig.GetBuilders()[0], core.ActionBuild, workspaceRoot, stateMgr)
				units = append(units, unit)
			} else {
				skipped = append(skipped, &SkippedComponent{
					Module:        mod.Moniker,
					Component:     compName,
					ComponentType: compType,
					Reason:        "no builder configured",
					ConfigHint:    "add 'builders' to component-types.yml for type '" + compType + "'",
				})
			}

		case FrameworkLint:
			if typeConfig != nil && typeConfig.IsLintable() {
				providers := cfg.LintProviders.GetProvidersForComponentType(compType)
				if len(providers) > 0 {
					for _, provider := range providers {
						unit := createUnitInfo(mod, compName, compType, provider, core.ActionLint, workspaceRoot, stateMgr)
						units = append(units, unit)
					}
				} else {
					skipped = append(skipped, &SkippedComponent{
						Module:        mod.Moniker,
						Component:     compName,
						ComponentType: compType,
						Reason:        "lintable but no providers configured",
						ConfigHint:    "add lint provider mapping in lint-providers.yml for type '" + compType + "'",
					})
				}
			} else {
				skipped = append(skipped, &SkippedComponent{
					Module:        mod.Moniker,
					Component:     compName,
					ComponentType: compType,
					Reason:        "component type not lintable",
					ConfigHint:    "add 'lintable: true' to component-types.yml for type '" + compType + "'",
				})
			}

		case FrameworkScan:
			if typeConfig != nil && typeConfig.IsScannable() {
				for _, scanner := range typeConfig.GetScanners() {
					unit := createUnitInfo(mod, compName, compType, scanner, core.ActionScan, workspaceRoot, stateMgr)
					unit.Container = true // All scanners run in containers
					unit.HostInstalled = false
					units = append(units, unit)
				}
			} else {
				skipped = append(skipped, &SkippedComponent{
					Module:        mod.Moniker,
					Component:     compName,
					ComponentType: compType,
					Reason:        "no scanners configured",
					ConfigHint:    "add 'scanners' to component-types.yml for type '" + compType + "'",
				})
			}

		case FrameworkTest:
			// Check if component type is testable
			if typeConfig != nil && typeConfig.IsTestable() {
				unit := createUnitInfo(mod, compName, compType, "test", core.ActionTest, workspaceRoot, stateMgr)
				units = append(units, unit)
			}
			// Note: non-testable components are not reported as skipped since
			// most component types (markdown, yaml, etc.) are not expected to have tests
		}
	}

	return units, skipped
}

// createUnitInfo creates a UnitInfo for a component.
func createUnitInfo(mod *modules.ModuleContract, compName, compType, tool string, action core.ActionType, workspaceRoot string, stateMgr *workunit.StateManager) *UnitInfo {
	unitID := workunit.UnitID{
		Action:        action,
		Module:        mod.Moniker,
		ComponentType: compType,
		ComponentName: compName,
		Tool:          tool,
	}

	info := &UnitInfo{
		ID:            unitID.Longname(),
		Module:        mod.Moniker,
		Component:     compName,
		Tool:          tool,
		ComponentType: compType,
		Weight:        1,
		Container:     false,
		HostInstalled: true,
		CacheStatus:   getCacheStatusForUnit(unitID, workspaceRoot, stateMgr, mod),
		Metadata:      make(map[string]any),
	}

	return info
}

// getCacheStatusForUnit determines the cache status of a unit.
func getCacheStatusForUnit(unitID workunit.UnitID, workspaceRoot string, stateMgr *workunit.StateManager, mod *modules.ModuleContract) *CacheStatus {
	status := &CacheStatus{
		State: "new",
	}

	// Try to load existing state
	state, err := stateMgr.Load(unitID)
	if err != nil {
		// No state file - unit is new
		status.Reason = "no prior state"
		return status
	}

	status.SourceHash = state.SourceHash
	status.ExecutedAt = state.ExecutedAt

	// Compute current source hash
	patterns := mod.GetGlobPatterns()
	files, err := hash.ExpandGlobPatterns(workspaceRoot, patterns)
	if err != nil {
		status.State = "stale"
		status.Reason = "failed to get source files"
		return status
	}

	currentHash, err := hash.Files(workspaceRoot, files)
	if err != nil {
		status.State = "stale"
		status.Reason = "failed to compute hash"
		return status
	}

	// Compare hashes
	if state.SourceHash != currentHash {
		status.State = "stale"
		status.Reason = "source changed"
		return status
	}

	// Check if previous execution failed
	if !state.Passed {
		status.State = "stale"
		status.Reason = "previous failure"
		return status
	}

	status.State = "up_to_date"
	return status
}

// buildUnitSummary builds aggregate statistics for units.
func buildUnitSummary(units []*UnitInfo) *UnitSummary {
	summary := &UnitSummary{
		Total:    len(units),
		ByModule: make(map[string]int),
		ByTool:   make(map[string]int),
	}

	for _, unit := range units {
		summary.ByModule[unit.Module]++
		if unit.Tool != "" {
			summary.ByTool[unit.Tool]++
		}

		if unit.Container {
			summary.Container++
		}
		if unit.HostInstalled {
			summary.HostInstalled++
		}

		if unit.CacheStatus != nil {
			switch unit.CacheStatus.State {
			case "up_to_date":
				summary.Cached++
			case "stale":
				summary.Stale++
			case "new":
				summary.New++
			}
		}
	}

	return summary
}

// FilterUnits applies filters to a unit list.
func FilterUnits(units []*UnitInfo, filters UnitFilters) []*UnitInfo {
	if !filters.HasFilters() {
		return units
	}

	var result []*UnitInfo
	for _, unit := range units {
		if matchesUnitFilters(unit, filters) {
			result = append(result, unit)
		}
	}
	return result
}

// UnitFilters contains filter criteria for units.
type UnitFilters struct {
	Module    string
	Component string
	Cached    bool
	Stale     bool
	Container bool
	Host      bool
}

// HasFilters returns true if any filter is set.
func (f UnitFilters) HasFilters() bool {
	return f.Module != "" || f.Component != "" || f.Cached || f.Stale || f.Container || f.Host
}

// matchesUnitFilters checks if a unit matches the filters.
func matchesUnitFilters(unit *UnitInfo, filters UnitFilters) bool {
	if filters.Module != "" && unit.Module != filters.Module {
		return false
	}
	if filters.Component != "" && unit.Component != filters.Component {
		return false
	}
	if filters.Cached && (unit.CacheStatus == nil || unit.CacheStatus.State != "up_to_date") {
		return false
	}
	if filters.Stale && (unit.CacheStatus == nil || unit.CacheStatus.State != "stale") {
		return false
	}
	if filters.Container && !unit.Container {
		return false
	}
	if filters.Host && !unit.HostInstalled {
		return false
	}
	return true
}

// GroupUnitsByModule groups units by their module.
func GroupUnitsByModule(units []*UnitInfo) map[string][]*UnitInfo {
	result := make(map[string][]*UnitInfo)
	for _, unit := range units {
		result[unit.Module] = append(result[unit.Module], unit)
	}
	return result
}

// GetStaleUnits returns only stale units.
func GetStaleUnits(units []*UnitInfo) []*UnitInfo {
	var result []*UnitInfo
	for _, unit := range units {
		if unit.CacheStatus != nil && unit.CacheStatus.State == "stale" {
			result = append(result, unit)
		}
	}
	return result
}
