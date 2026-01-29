package build

import (
	"math"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// FlattenModulesToComponentWork converts module layers to component work layers.
// Each module is expanded to its buildable components with weight and dependency info.
// Returns nil if no buildable components are found.
func FlattenModulesToComponentWork(ctx *cmdframework.ExecutionContext) [][]orchestrator.ComponentWork {
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

	var componentLayers [][]orchestrator.ComponentWork
	globalIndex := 0

	for _, layerMonikers := range layers {
		var layerWork []orchestrator.ComponentWork

		for _, moniker := range layerMonikers {
			// Check if module is cached
			isCached := cachedModules != nil && cachedModules[moniker]
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
					Cached:        isCached,
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

				// Get build_after from component type config
				var buildAfter []string
				compType := cfg.ComponentTypes.Get(compTypeName)
				if compType != nil {
					buildAfter = compType.GetBuildAfter()
				}

				// Get weight (base weight × amp, calculated internally)
				weight := getComponentWeight(moniker, componentName, compTypeName, tool.OperationBuild)

				// Component work item: component name includes builder for unique identification
				// Format: "component:builder" so display becomes "module:component:builder"
				// This matches the lint pattern (e.g., "go:go-lint")
				componentWithBuilder := componentName + ":" + handlerName

				work := orchestrator.ComponentWork{
					Module:        moniker,
					Component:     componentWithBuilder,
					ComponentType: compTypeName,
					Handler:       handlerName,
					IsContainer:   ch.Handler.IsContainer(),
					Weight:        weight,
					BuildAfter:    buildAfter,
					Index:         globalIndex,
					Cached:        isCached,
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

// getToolWeight returns the scheduling weight for a component type and operation.
// Weight is derived from tool.Resources.CPUs. Defaults to 1.
func getToolWeight(componentType string, operation tool.OperationType) int {
	bridge := tool.GlobalBuildBridge()
	if bridge == nil {
		return 1
	}

	t := bridge.ResolveTool(componentType, operation)
	if t == nil {
		return 1
	}

	return t.Resources.Weight()
}

// getComponentWeight returns the scheduling weight for a component.
// Weight = base tool weight × component amp (from config).
func getComponentWeight(moniker, componentName, componentType string, operation tool.OperationType) int {
	baseWeight := getToolWeight(componentType, operation)

	// Get amp from config (the source of truth)
	amp := 1.0
	cfg := config.Global()
	if cfg != nil && cfg.Repository != nil {
		if module, ok := cfg.Repository.GetModule(moniker); ok && module != nil {
			amp = module.GetComponentAmp(componentName, string(operation))
		}
	}

	// Apply amp to weight (ceil to ensure at least 1)
	weight := int(math.Ceil(float64(baseWeight) * amp))
	if weight < 1 {
		weight = 1
	}

	return weight
}
