package lint

import (
	"math"

	"github.com/ready-to-release/eac/go/eac/commands/impl/update/lint/linters"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// FlattenModulesToLintComponentWork converts modules to lint component work items.
// Each module is expanded to its lintable components with weight info.
// Work items are created for each unique module:component:provider combination.
// Returns nil if no lintable components are found.
func FlattenModulesToLintComponentWork(ctx *cmdframework.ExecutionContext) [][]orchestrator.ComponentWork {
	cfg := config.Global()
	if cfg == nil || cfg.LintProviders == nil {
		return nil
	}

	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		return nil
	}

	// Lint runs in parallel (no layers), so treat all modules as single layer
	var componentWork []orchestrator.ComponentWork
	globalIndex := 0

	for _, moniker := range monikers {
		// Get module contract
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Check if linting is disabled for this module
		if module.Linting != nil && containsString(module.Linting.Disabled, "all") {
			continue
		}

		// Get all lintable components
		if module.Components == nil || len(module.Components.GetEnabled()) == 0 {
			continue
		}

		for _, compName := range module.Components.GetEnabled() {
			compType := module.Components.GetComponentType(compName)

			providerNames := cfg.LintProviders.GetProvidersForComponentType(compType)
			for _, providerName := range providerNames {
				// Check module-level linting overrides
				if !isProviderEnabledForModule(providerName, module.Linting) {
					continue
				}

				provider := cfg.LintProviders.Get(providerName)
				if provider == nil {
					continue
				}

				// Check if handler exists
				handler := linters.GetHandlerForProvider(providerName)
				if handler == nil {
					continue
				}

				// Component work item: component name includes provider for unique identification
				// Format: "component:provider" so display becomes "module:component:provider"
				componentWithProvider := compName + ":" + providerName

				// Get weight (base weight × amp, calculated internally)
				weight := getComponentWeight(moniker, compName, compType, tool.OperationLint)

				work := orchestrator.ComponentWork{
					Module:        moniker,
					Component:     componentWithProvider,
					ComponentType: compType,
					Handler:       providerName,
					Weight:        weight,
					BuildAfter:    nil,
					Index:         globalIndex,
				}

				componentWork = append(componentWork, work)
				globalIndex++
			}
		}
	}

	if len(componentWork) == 0 {
		return nil
	}

	// Return as single layer (lint runs in parallel)
	return [][]orchestrator.ComponentWork{componentWork}
}

// CountLintComponents returns the total number of lintable component work items.
func CountLintComponents(layers [][]orchestrator.ComponentWork) int {
	count := 0
	for _, layer := range layers {
		count += len(layer)
	}
	return count
}

// getLintToolWeight returns the scheduling weight for a lint operation.
// Weight is derived from tool.Resources.CPUs. Defaults to 1.
func getLintToolWeight(componentType string) int {
	bridge := tool.GlobalLintBridge()
	if bridge == nil {
		return 1
	}

	t := bridge.ResolveTool(componentType, tool.OperationLint)
	if t == nil {
		return 1
	}

	return t.Resources.Weight()
}

// getComponentWeight returns the scheduling weight for a component.
// Weight = base tool weight × component amp (from config).
func getComponentWeight(moniker, componentName, componentType string, operation tool.OperationType) int {
	baseWeight := getLintToolWeight(componentType)

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
