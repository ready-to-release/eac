package lint

import (
	"slices"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/resolver"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ResolveLintUnitSpecs converts modules to lint component work items.
// Each module is expanded to its lintable components with weight info.
// Uses ComponentResolver for consistent component-to-tool mapping.
// Work items are created for each unique module:component:provider combination.
// Returns nil if no lintable components are found.
func ResolveLintUnitSpecs(ctx *cmdframework.ExecutionContext) []workunit.UnitSpec {
	cfg := config.Global()
	if cfg == nil || cfg.LintProviders == nil {
		return nil
	}

	// Get cached modules map from lint context (set during incremental detection)
	var cachedModules map[string]bool
	if lctx, ok := ctx.Config.LintCmdContext.(*lintContext); ok && lctx != nil {
		cachedModules = lctx.cachedModules
	}

	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		return nil
	}

	// Create component resolver for unified resolution
	compResolver := resolver.NewComponentResolver()

	var componentWork []workunit.UnitSpec
	globalIndex := 0

	for _, moniker := range monikers {
		// Get module contract
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Check if linting is disabled for this module
		if module.Linting != nil && slices.Contains(module.Linting.Disabled, "all") {
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
			handler := tool.GlobalLintBridge().GetHandlerForProvider(providerName)
			if handler == nil {
				continue
			}

			// Update with handler info
			weight := specs[i].Weight
			if weight == 0 {
				weight = 1
			}
			specs[i].PoolAllocation = core.AllocationForWeight(weight, handler.IsContainer())
			specs[i].Index = globalIndex
			globalIndex++

			componentWork = append(componentWork, specs[i])
		}
	}

	return componentWork
}

// CountLintComponents returns the total number of lintable component work items.
func CountLintComponents(units []workunit.UnitSpec) int {
	return len(units)
}
