package build

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/resolver"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ResolveUnitSpecs converts modules to component work specs.
// Each module is expanded to its buildable components with weight and dependency info.
// Uses ComponentResolver for consistent component-to-tool mapping.
// Returns nil if no buildable components are found.
func ResolveUnitSpecs(ctx *cmdframework.ExecutionContext) []workunit.UnitSpec {
	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
	}

	// Get cached modules map from build context (set during incremental detection)
	var cachedModules map[string]bool
	var componentFilter []string
	if bctx, ok := ctx.Config.BuildCmdContext.(*buildContext); ok && bctx != nil {
		cachedModules = bctx.cachedModules
		if buildCfg, ok := ctx.Config.BuildCmdConfig.(*BuildConfig); ok && buildCfg != nil {
			componentFilter = buildCfg.Components
		}
	}

	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		return nil
	}

	// Create component resolver for unified resolution
	compResolver := resolver.NewComponentResolver()

	var specs []workunit.UnitSpec
	globalIndex := 0

	for _, moniker := range monikers {
		// Check if module is cached
		isCached := cachedModules != nil && cachedModules[moniker]

		// Get module contract
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Use ComponentResolver to get build specs for this module
		moduleSpecs := compResolver.ResolveForBuild(module, cachedModules)

		if len(moduleSpecs) == 0 {
			// Module has no buildable components - create a placeholder
			specs = append(specs, workunit.UnitSpec{
				ID: workunit.UnitID{
					Action:    core.ActionBuild,
					Module:    moniker,
					Component: "none",
					Tool:      "",
				},
				ComponentType: "none",
				Weight:        1,
				Container:     false,
				HostInstalled: true,
				DependsOn:     nil,
				Cached:        isCached,
				Metadata:      make(map[string]any),
				Index:         globalIndex,
			})
			globalIndex++
			continue
		}

		// Add specs with correct index
		for i := range moduleSpecs {
			moduleSpecs[i].Index = globalIndex
			globalIndex++
		}

		specs = append(specs, moduleSpecs...)
	}

	// Apply component filter if specified
	if len(componentFilter) > 0 {
		specs = filterByComponents(specs, componentFilter)
	}

	return specs
}

// filterByComponents filters specs to only those whose component name matches the filter.
// When filter is empty, all specs are returned unchanged.
// Filtered-out DependsOn references are removed and specs are re-indexed.
func filterByComponents(specs []workunit.UnitSpec, filter []string) []workunit.UnitSpec {
	if len(filter) == 0 {
		return specs
	}

	// Build lookup set for filter
	allowed := make(map[string]bool, len(filter))
	for _, c := range filter {
		allowed[c] = true
	}

	// Collect matching specs and build a set of kept longnames
	var filtered []workunit.UnitSpec
	kept := make(map[string]bool)
	for _, spec := range specs {
		if allowed[spec.ID.Component] {
			filtered = append(filtered, spec)
			kept[spec.ID.Longname()] = true
		}
	}

	// Clean up DependsOn references and re-index
	for i := range filtered {
		if len(filtered[i].DependsOn) > 0 {
			var validDeps []workunit.UnitID
			for _, dep := range filtered[i].DependsOn {
				if kept[dep.Longname()] {
					validDeps = append(validDeps, dep)
				}
			}
			filtered[i].DependsOn = validDeps
		}
		filtered[i].Index = i
	}

	return filtered
}

// CountUnits returns the total number of component work items.
func CountUnits(specs []workunit.UnitSpec) int {
	return len(specs)
}
