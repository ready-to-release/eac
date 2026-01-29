package lint

import (
	"math"

	"github.com/ready-to-release/eac/go/eac/commands/impl/update/lint/linters"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/tool"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

// FlattenModulesToLintComponentWork converts modules to lint component work items.
// Each module is expanded to its lintable components with weight info.
// Work items are created for each unique module:component:provider combination.
// Returns nil if no lintable components are found.
func FlattenModulesToLintComponentWork(ctx *cmdframework.ExecutionContext) [][]workunit.UnitSpec {
	cfg := config.Global()
	if cfg == nil || cfg.LintProviders == nil {
		return nil
	}

	// Get cached modules map from lint context (set during incremental detection)
	var cachedModules map[string]bool
	if lctx, ok := ctx.Config.Extra["lintContext"].(*lintContext); ok && lctx != nil {
		cachedModules = lctx.cachedModules
	}

	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		return nil
	}

	// Lint runs in parallel (no layers), so treat all modules as single layer
	var componentWork []workunit.UnitSpec
	globalIndex := 0

	for _, moniker := range monikers {
		// Check if module is cached
		isCached := cachedModules != nil && cachedModules[moniker]
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

				// Get weight (base weight × amp, calculated internally)
				weight := getComponentWeight(moniker, compName, providerName, tool.OperationLint)

				work := workunit.UnitSpec{
					ID: workunit.UnitID{
						Context:   workunit.ContextLint,
						Module:    moniker,
						Component: compName,
						Tool:      providerName,
					},
					ComponentType:   compType,
					Weight:          weight,
					IsContainer:     handler.IsContainer(),
					IsHostInstalled: !handler.IsContainer(),
					DependsOn:       nil,
					Cached:          isCached,
					Metadata:      make(map[string]any),
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
	return [][]workunit.UnitSpec{componentWork}
}

// CountLintComponents returns the total number of lintable component work items.
func CountLintComponents(layers [][]workunit.UnitSpec) int {
	count := 0
	for _, layer := range layers {
		count += len(layer)
	}
	return count
}

// getLintToolWeight returns the scheduling weight for a lint provider.
// Weight is derived from tool.Resources.CPUs. Defaults to 1.
func getLintToolWeight(providerName string) int {
	bridge := tool.GlobalLintBridge()
	if bridge == nil {
		return 1
	}
	return bridge.GetProviderWeight(providerName)
}

// getComponentWeight returns the scheduling weight for a component.
// Weight = base tool weight × component amp (from config).
func getComponentWeight(moniker, componentName, providerName string, operation tool.OperationType) int {
	baseWeight := getLintToolWeight(providerName)

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
