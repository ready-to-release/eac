// Package reports provides report generation functions for CLI commands.
package reports

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// Framework represents a command phase (build, test, lint, scan).
type Framework string

const (
	FrameworkBuild Framework = "build"
	FrameworkTest  Framework = "test"
	FrameworkLint  Framework = "lint"
	FrameworkScan  Framework = "scan"
)

// ComponentInfo represents a resolved component with phase information.
type ComponentInfo struct {
	// Moniker is the module moniker
	Moniker string `json:"moniker" yaml:"moniker"`

	// Component is the component name (key in module.components map)
	Component string `json:"component" yaml:"component"`

	// Type is the component type (from component-types.yml)
	Type string `json:"type" yaml:"type"`

	// Root is the component root path
	Root string `json:"root" yaml:"root"`

	// Patterns contains file patterns for this component
	Patterns *ComponentPatterns `json:"patterns,omitempty" yaml:"patterns,omitempty"`

	// Phases contains phase support information
	Phases *ComponentPhases `json:"phases" yaml:"phases"`

	// Dependencies contains dependency information
	Dependencies *ComponentDependencies `json:"dependencies" yaml:"dependencies"`

	// Metadata contains additional component metadata
	Metadata map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ComponentPatterns contains file patterns for source, tests, and config.
type ComponentPatterns struct {
	Source []string `json:"source,omitempty" yaml:"source,omitempty"`
	Tests  []string `json:"tests,omitempty" yaml:"tests,omitempty"`
	Config []string `json:"config,omitempty" yaml:"config,omitempty"`
}

// ComponentPhases contains phase support information.
type ComponentPhases struct {
	Build *PhaseInfo `json:"build,omitempty" yaml:"build,omitempty"`
	Lint  *PhaseInfo `json:"lint,omitempty" yaml:"lint,omitempty"`
	Test  *PhaseInfo `json:"test,omitempty" yaml:"test,omitempty"`
	Scan  *PhaseInfo `json:"scan,omitempty" yaml:"scan,omitempty"`
}

// PhaseInfo contains information about a specific phase.
type PhaseInfo struct {
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Tool    string   `json:"tool,omitempty" yaml:"tool,omitempty"`
	Tools   []string `json:"tools,omitempty" yaml:"tools,omitempty"`
}

// ComponentDependencies contains bidirectional dependency information.
type ComponentDependencies struct {
	// DependsOn lists components this component depends on (format: module:component)
	DependsOn []string `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`

	// DependedBy lists components that depend on this component (format: module:component)
	DependedBy []string `json:"depended_by,omitempty" yaml:"depended_by,omitempty"`
}

// ComponentReport contains the full component report.
type ComponentReport struct {
	// Components is the list of all resolved components
	Components []*ComponentInfo `json:"components" yaml:"components"`

	// Summary contains aggregate statistics
	Summary *ComponentSummary `json:"summary" yaml:"summary"`
}

// ComponentSummary contains aggregate statistics about components.
type ComponentSummary struct {
	Total     int            `json:"total" yaml:"total"`
	ByModule  map[string]int `json:"by_module" yaml:"by_module"`
	ByType    map[string]int `json:"by_type" yaml:"by_type"`
	Buildable int            `json:"buildable" yaml:"buildable"`
	Lintable  int            `json:"lintable" yaml:"lintable"`
	Testable  int            `json:"testable" yaml:"testable"`
	Scannable int            `json:"scannable" yaml:"scannable"`
}

// Global process-level cache for component report
var (
	globalComponentCache     *ComponentCache
	globalComponentCacheOnce sync.Once
)

// ComponentCache provides cached component data.
type ComponentCache struct {
	mu            sync.RWMutex
	workspaceRoot string
	populated     bool
	report        *ComponentReport
}

// NewComponentCache creates a new empty cache.
func NewComponentCache() *ComponentCache {
	return &ComponentCache{}
}

// GetComponents loads and returns the component report.
// Uses global in-memory cache to avoid repeated computation.
func GetComponents(workspaceRoot string) (*ComponentReport, error) {
	globalComponentCacheOnce.Do(func() {
		globalComponentCache = NewComponentCache()
	})

	if err := globalComponentCache.EnsurePopulated(workspaceRoot); err != nil {
		return nil, err
	}

	return globalComponentCache.GetReport()
}

// EnsurePopulated ensures the cache is populated.
func (c *ComponentCache) EnsurePopulated(workspaceRoot string) error {
	c.mu.RLock()
	if c.populated && c.workspaceRoot == workspaceRoot {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.populated && c.workspaceRoot == workspaceRoot {
		return nil
	}

	if c.workspaceRoot != workspaceRoot {
		c.report = nil
		c.populated = false
		c.workspaceRoot = workspaceRoot
	}

	report, err := buildComponentReport(workspaceRoot)
	if err != nil {
		return err
	}

	c.report = report
	c.populated = true
	return nil
}

// GetReport returns the cached report.
// Returns an error if the cache has not been populated via EnsurePopulated().
func (c *ComponentCache) GetReport() (*ComponentReport, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.populated {
		return nil, fmt.Errorf("ComponentCache not populated - call EnsurePopulated() first")
	}

	return c.report, nil
}

// buildComponentReport builds the component report from module contracts.
func buildComponentReport(workspaceRoot string) (*ComponentReport, error) {
	// Load module contracts
	moduleReport, err := GetModuleContracts(workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Load config for phase determination
	cfg := config.Global()

	// Build component info list
	var components []*ComponentInfo
	componentIndex := make(map[string]*ComponentInfo) // module:component -> info

	for _, mod := range moduleReport.Modules {
		for compName := range mod.Components {
			entry := mod.Components[compName]
			if entry == nil {
				continue
			}

			compType := mod.Components.GetComponentType(compName)
			info := &ComponentInfo{
				Moniker:   mod.Moniker,
				Component: compName,
				Type:      compType,
				Root:      entry.Root,
				Phases:    resolvePhases(compType, cfg),
				Dependencies: &ComponentDependencies{
					DependsOn:  []string{},
					DependedBy: []string{},
				},
				Metadata: make(map[string]any),
			}

			// Copy patterns if present
			if entry.Patterns != nil {
				info.Patterns = &ComponentPatterns{
					Source: entry.Patterns.Source,
					Tests:  entry.Patterns.Tests,
					Config: entry.Patterns.Config,
				}
			}

			// Store amp values in metadata if present
			if entry.Amp != nil {
				if entry.Amp.Build > 0 {
					info.Metadata["amp_build"] = entry.Amp.Build
				}
				if entry.Amp.Lint > 0 {
					info.Metadata["amp_lint"] = entry.Amp.Lint
				}
				if entry.Amp.Test > 0 {
					info.Metadata["amp_test"] = entry.Amp.Test
				}
				if entry.Amp.Scan > 0 {
					info.Metadata["amp_scan"] = entry.Amp.Scan
				}
			}

			components = append(components, info)
			componentIndex[mod.Moniker+":"+compName] = info
		}
	}

	// Build dependency graph
	buildDependencyGraph(moduleReport, componentIndex)

	// Sort components by moniker, then by component name
	sort.Slice(components, func(i, j int) bool {
		if components[i].Moniker != components[j].Moniker {
			return components[i].Moniker < components[j].Moniker
		}
		return components[i].Component < components[j].Component
	})

	// Build summary
	summary := buildComponentSummary(components)

	return &ComponentReport{
		Components: components,
		Summary:    summary,
	}, nil
}

// resolvePhases determines phase support for a component type.
func resolvePhases(compType string, cfg *config.EACConfig) *ComponentPhases {
	phases := &ComponentPhases{}

	if cfg == nil {
		return phases
	}

	// Build phase
	if cfg.ComponentTypes != nil {
		typeConfig := cfg.ComponentTypes.Get(compType)
		if typeConfig != nil && typeConfig.IsBuildable() {
			builders := typeConfig.GetBuilders()
			phases.Build = &PhaseInfo{
				Enabled: true,
				Tool:    builders[0],
				Tools:   builders,
			}
		}

		// Scan phase
		if typeConfig != nil && typeConfig.IsScannable() {
			phases.Scan = &PhaseInfo{
				Enabled: true,
				Tools:   typeConfig.GetScanners(),
			}
		}
	}

	// Lint phase
	if cfg.LintProviders != nil {
		providers := cfg.LintProviders.GetProvidersForComponentType(compType)
		if len(providers) > 0 {
			phases.Lint = &PhaseInfo{
				Enabled: true,
				Tools:   providers,
			}
		}
	}

	// Test phase - determined by component type supporting tests
	// Go, gherkin, typescript are testable
	testableTypes := map[string]bool{
		"go":         true,
		"gherkin":    true,
		"typescript": true,
		"python":     true,
	}
	if testableTypes[compType] {
		phases.Test = &PhaseInfo{
			Enabled: true,
		}
	}

	return phases
}

// buildDependencyGraph builds bidirectional dependencies.
func buildDependencyGraph(moduleReport *ModuleContractReport, componentIndex map[string]*ComponentInfo) {
	// Module-level dependencies map to component dependencies
	// When module A depends on module B, all components of A depend on the "primary" component of B
	for _, mod := range moduleReport.Modules {
		for _, depMoniker := range mod.DependsOn {
			depMod, ok := moduleReport.GetModuleByMoniker(depMoniker)
			if !ok {
				continue
			}

			// Find primary component of dependency (typically "go" or first buildable)
			primaryComp := findPrimaryComponent(depMod)
			if primaryComp == "" {
				continue
			}

			depKey := depMoniker + ":" + primaryComp
			depInfo := componentIndex[depKey]

			// Add dependency from all components in this module to the primary component
			for compName := range mod.Components {
				srcKey := mod.Moniker + ":" + compName
				srcInfo := componentIndex[srcKey]
				if srcInfo == nil {
					continue
				}

				srcInfo.Dependencies.DependsOn = append(srcInfo.Dependencies.DependsOn, depKey)
				if depInfo != nil {
					depInfo.Dependencies.DependedBy = append(depInfo.Dependencies.DependedBy, srcKey)
				}
			}
		}
	}

	// Sort dependency lists for consistent output
	for _, info := range componentIndex {
		sort.Strings(info.Dependencies.DependsOn)
		sort.Strings(info.Dependencies.DependedBy)
	}
}

// findPrimaryComponent finds the primary component of a module.
func findPrimaryComponent(mod *modules.ModuleContract) string {
	// Prefer "go" component
	if mod.Components.HasComponent("go") {
		return "go"
	}
	// Then typescript
	if mod.Components.HasComponent("typescript") {
		return "typescript"
	}
	// Otherwise first component
	for name := range mod.Components {
		return name
	}
	return ""
}

// buildComponentSummary builds aggregate statistics.
func buildComponentSummary(components []*ComponentInfo) *ComponentSummary {
	summary := &ComponentSummary{
		Total:    len(components),
		ByModule: make(map[string]int),
		ByType:   make(map[string]int),
	}

	for _, comp := range components {
		summary.ByModule[comp.Moniker]++
		summary.ByType[comp.Type]++

		if comp.Phases != nil {
			if comp.Phases.Build != nil && comp.Phases.Build.Enabled {
				summary.Buildable++
			}
			if comp.Phases.Lint != nil && comp.Phases.Lint.Enabled {
				summary.Lintable++
			}
			if comp.Phases.Test != nil && comp.Phases.Test.Enabled {
				summary.Testable++
			}
			if comp.Phases.Scan != nil && comp.Phases.Scan.Enabled {
				summary.Scannable++
			}
		}
	}

	return summary
}

// FilterComponents applies filters to a component list.
func FilterComponents(components []*ComponentInfo, filters ComponentFilters) []*ComponentInfo {
	if !filters.HasFilters() {
		return components
	}

	var result []*ComponentInfo
	for _, comp := range components {
		if matchesFilters(comp, filters) {
			result = append(result, comp)
		}
	}
	return result
}

// ComponentFilters contains filter criteria for components.
type ComponentFilters struct {
	Module    string
	Type      string
	Buildable bool
	Lintable  bool
	Testable  bool
	Scannable bool
}

// HasFilters returns true if any filter is set.
func (f ComponentFilters) HasFilters() bool {
	return f.Module != "" || f.Type != "" || f.Buildable || f.Lintable || f.Testable || f.Scannable
}

// matchesFilters checks if a component matches the filters.
func matchesFilters(comp *ComponentInfo, filters ComponentFilters) bool {
	if filters.Module != "" && comp.Moniker != filters.Module {
		return false
	}
	if filters.Type != "" && comp.Type != filters.Type {
		return false
	}
	if filters.Buildable && (comp.Phases == nil || comp.Phases.Build == nil || !comp.Phases.Build.Enabled) {
		return false
	}
	if filters.Lintable && (comp.Phases == nil || comp.Phases.Lint == nil || !comp.Phases.Lint.Enabled) {
		return false
	}
	if filters.Testable && (comp.Phases == nil || comp.Phases.Test == nil || !comp.Phases.Test.Enabled) {
		return false
	}
	if filters.Scannable && (comp.Phases == nil || comp.Phases.Scan == nil || !comp.Phases.Scan.Enabled) {
		return false
	}
	return true
}
