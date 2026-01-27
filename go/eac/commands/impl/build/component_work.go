package build

import (
	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

// FlattenModulesToComponentWork converts module layers to component work layers.
// Each module is expanded to its buildable components with weight and dependency info.
// Returns nil if no buildable components are found.
func FlattenModulesToComponentWork(ctx *cmdframework.ExecutionContext) [][]orchestrator.ComponentWork {
	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
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

	var componentLayers [][]orchestrator.ComponentWork
	globalIndex := 0

	for _, layerMonikers := range layers {
		var layerWork []orchestrator.ComponentWork

		for _, moniker := range layerMonikers {
			// Get module contract
			module, exists := ctx.ModuleRegistry.Get(moniker)
			if !exists {
				continue
			}

			// Get all handlers for module's buildable components
			compHandlers := builders.GetHandlersForModule(module)
			if len(compHandlers) == 0 {
				// Module has no buildable components - create a placeholder
				layerWork = append(layerWork, orchestrator.ComponentWork{
					Module:        moniker,
					Component:     "none",
					ComponentType: "none",
					Handler:       "",
					Weight:        1,
					BuildAfter:    nil,
					Index:         globalIndex,
				})
				globalIndex++
				continue
			}

			// Create work item for each component
			for _, ch := range compHandlers {
				componentName := ch.Component
				handlerName := ch.Handler.Name()

				// Get component type (may differ from name for named components)
				compTypeName := module.Components.GetComponentType(componentName)

				// Get weight and build_after from component type config
				weight := 1
				var buildAfter []string

				compType := cfg.ComponentTypes.Get(compTypeName)
				if compType != nil {
					weight = compType.GetBuildWeight()
					buildAfter = compType.GetBuildAfter()
				}

					work := orchestrator.ComponentWork{
					Module:        moniker,
					Component:     componentName,
					ComponentType: compTypeName,
					Handler:       handlerName,
					Weight:        weight,
					BuildAfter:    buildAfter,
					Index:         globalIndex,
				}

				layerWork = append(layerWork, work)
				globalIndex++
			}
		}

		if len(layerWork) > 0 {
			componentLayers = append(componentLayers, layerWork)
		}
	}

	return componentLayers
}

// FlattenComponentLayers flattens component work layers to a single slice.
func FlattenComponentLayers(layers [][]orchestrator.ComponentWork) []orchestrator.ComponentWork {
	var all []orchestrator.ComponentWork
	for _, layer := range layers {
		all = append(all, layer...)
	}
	return all
}

// CountComponents returns the total number of component work items.
func CountComponents(layers [][]orchestrator.ComponentWork) int {
	count := 0
	for _, layer := range layers {
		count += len(layer)
	}
	return count
}
