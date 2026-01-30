package build

import (
	"math"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/tool"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

var compWorkLog = logging.C()

// FlattenModulesToComponentWork converts module layers to component work layers.
// Each module is expanded to its buildable components with weight and dependency info.
// Returns nil if no buildable components are found.
func FlattenModulesToComponentWork(ctx *cmdframework.ExecutionContext) [][]workunit.UnitSpec {
	flattenStart := time.Now()
	defer func() {
		compWorkLog.Debugf("[FLATTEN] Total FlattenModulesToComponentWork: %v", time.Since(flattenStart))
	}()

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

	var componentLayers [][]workunit.UnitSpec
	globalIndex := 0

	var handlersTime, weightTime time.Duration
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

			// Get all handlers for module's buildable components
			handlerStart := time.Now()
			compHandlers := builders.GetHandlersForModule(module)
			handlersTime += time.Since(handlerStart)
			if len(compHandlers) == 0 {
				// Module has no buildable components - create a placeholder
				layerWork = append(layerWork, workunit.UnitSpec{
					ID: workunit.UnitID{
						Context:   workunit.ContextBuild,
						Module:    moniker,
						Component: "none",
						Tool:      "",
					},
					ComponentType:   "none",
					Weight:          1,
					IsContainer:     false,
					IsHostInstalled: true,
					DependsOn:       nil,
					Cached:          isCached,
					Metadata:      make(map[string]any),
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

				// Get build_after from component type config and convert to DependsOn
				var dependsOn []workunit.UnitID
				compType := cfg.ComponentTypes.Get(compTypeName)
				if compType != nil {
					buildAfter := compType.GetBuildAfter()
					for _, depComponent := range buildAfter {
						dependsOn = append(dependsOn, workunit.UnitID{
							Context:   workunit.ContextBuild,
							Module:    moniker, // Same module - intra-module dependency
							Component: depComponent,
						})
					}
				}

				// Get weight (base weight × amp, calculated internally)
				weightStart := time.Now()
				weight := getComponentWeight(moniker, componentName, compTypeName, tool.OperationBuild)
				weightTime += time.Since(weightStart)

				work := workunit.UnitSpec{
					ID: workunit.UnitID{
						Context:   workunit.ContextBuild,
						Module:    moniker,
						Component: componentName,
						Tool:      handlerName,
					},
					ComponentType:   compTypeName,
					Weight:          weight,
					IsContainer:     ch.Handler.IsContainer(),
					IsHostInstalled: !ch.Handler.IsContainer(),
					DependsOn:       dependsOn,
					Cached:          isCached,
					Metadata:      make(map[string]any),
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

	compWorkLog.Debugf("[FLATTEN] GetHandlersForModule total: %v, getComponentWeight total: %v", handlersTime, weightTime)
	return componentLayers
}

// FlattenComponentLayers flattens component work layers to a single slice.
func FlattenComponentLayers(layers [][]workunit.UnitSpec) []workunit.UnitSpec {
	var all []workunit.UnitSpec
	for _, layer := range layers {
		all = append(all, layer...)
	}
	return all
}

// CountComponents returns the total number of component work items.
func CountComponents(layers [][]workunit.UnitSpec) int {
	count := 0
	for _, layer := range layers {
		count += len(layer)
	}
	return count
}

// getToolWeight returns the scheduling weight for a component type and operation.
// Weight is derived from tool.Resources.CPUs. Defaults to 1.
func getToolWeight(componentType string, operation tool.OperationType) int {
	bridgeStart := time.Now()
	bridge := tool.GlobalBuildBridge()
	bridgeTime := time.Since(bridgeStart)
	if bridgeTime > 10*time.Millisecond {
		compWorkLog.Debugf("[WEIGHT] GlobalBuildBridge took %v", bridgeTime)
	}
	if bridge == nil {
		return 1
	}

	resolveStart := time.Now()
	t := bridge.ResolveTool(componentType, operation)
	resolveTime := time.Since(resolveStart)
	if resolveTime > 10*time.Millisecond {
		compWorkLog.Debugf("[WEIGHT] ResolveTool(%s) took %v", componentType, resolveTime)
	}
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
