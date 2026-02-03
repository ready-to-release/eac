package build

import (
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/resolver"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ResolveUnitSpecs converts module layers to component work layers.
// Each module is expanded to its buildable components with weight and dependency info.
// Uses ComponentResolver for consistent component-to-tool mapping.
// Returns nil if no buildable components are found.
func ResolveUnitSpecs(ctx *cmdframework.ExecutionContext) [][]workunit.UnitSpec {
	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
	}

	// Get cached modules map from build context (set during incremental detection)
	var cachedModules map[string]bool
	if bctx, ok := ctx.Config.Extra["buildContext"].(*buildContext); ok && bctx != nil {
		cachedModules = bctx.cachedModules
	}

	layers := ctx.GetLayers()
	if len(layers) == 0 {
		// Not using layers - treat all modules as single layer
		monikers := ctx.GetExecutionMonikers()
		if len(monikers) == 0 {
			return nil
		}
		layers = [][]string{monikers}
	}

	// Create component resolver for unified resolution
	compResolver := resolver.NewComponentResolver()

	var componentLayers [][]workunit.UnitSpec
	globalIndex := 0

	for _, layerMonikers := range layers {
		var layerWork []workunit.UnitSpec

		for _, moniker := range layerMonikers {
			// Check if module is cached
			isCached := cachedModules != nil && cachedModules[moniker]

			// Get module contract
			module, exists := ctx.ModuleRegistry.Get(moniker)
			if !exists {
				continue
			}

			// Use ComponentResolver to get build specs for this module
			specs := compResolver.ResolveForBuild(module, cachedModules)

			if len(specs) == 0 {
				// Module has no buildable components - create a placeholder
				layerWork = append(layerWork, workunit.UnitSpec{
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
			for i := range specs {
				specs[i].Index = globalIndex
				globalIndex++
			}

			layerWork = append(layerWork, specs...)
		}

		if len(layerWork) > 0 {
			componentLayers = append(componentLayers, layerWork)
		}
	}

	return componentLayers
}

// FlattenUnitLayers flattens component work layers to a single slice.
func FlattenUnitLayers(layers [][]workunit.UnitSpec) []workunit.UnitSpec {
	var all []workunit.UnitSpec
	for _, layer := range layers {
		all = append(all, layer...)
	}
	return all
}

// CountUnits returns the total number of component work items.
func CountUnits(layers [][]workunit.UnitSpec) int {
	count := 0
	for _, layer := range layers {
		count += len(layer)
	}
	return count
}
