package lint

import (
	"github.com/ready-to-release/eac/go/cli/eac/impl/update/lint/linters"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/resolver"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ResolveLintUnitSpecs converts modules to lint component work items.
// Each module is expanded to its lintable components with weight info.
// Uses ComponentResolver for consistent component-to-tool mapping.
// Work items are created for each unique module:component:provider combination.
// Returns nil if no lintable components are found.
func ResolveLintUnitSpecs(ctx *cmdframework.ExecutionContext) [][]workunit.UnitSpec {
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

	// Create component resolver for unified resolution
	compResolver := resolver.NewComponentResolver()

	// Lint runs in parallel (no layers), so treat all modules as single layer
	var componentWork []workunit.UnitSpec
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

		// Use ComponentResolver to get lint specs for this module
		specs := compResolver.ResolveForLint(module, cachedModules)

		// Filter specs based on module-level linting overrides
		for i := range specs {
			providerName := specs[i].ID.Tool

			// Check module-level linting overrides
			if !isProviderEnabledForModule(providerName, module.Linting) {
				continue
			}

			// Check if handler exists
			handler := linters.GetHandlerForProvider(providerName)
			if handler == nil {
				continue
			}

			// Update with handler info
			specs[i].Container = handler.IsContainer()
			specs[i].HostInstalled = !handler.IsContainer()
			specs[i].Index = globalIndex
			globalIndex++

			componentWork = append(componentWork, specs[i])
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
