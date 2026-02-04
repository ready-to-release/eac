package build

import (
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
	if bctx, ok := ctx.Config.Extra["buildContext"].(*buildContext); ok && bctx != nil {
		cachedModules = bctx.cachedModules
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
					Context:   workunit.ContextBuild,
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

	return specs
}

// CountUnits returns the total number of component work items.
func CountUnits(specs []workunit.UnitSpec) int {
	return len(specs)
}
