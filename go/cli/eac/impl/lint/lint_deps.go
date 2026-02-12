package lint

import (
	"sort"

	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

// lintDepsVerifier verifies system dependencies for linting.
func lintDepsVerifier(ctx *cmdframework.ExecutionContext) *initsummary.DepsStatus {
	status := &initsummary.DepsStatus{Verified: true}

	// ModuleRegistry may not be populated yet (async deps check starts before phaseResolve).
	// If unavailable, skip provider-based dep detection — return verified with no requirements.
	if ctx.ModuleRegistry == nil {
		return status
	}

	// Collect unique requirements from all lint providers that will be used
	depsMap := make(map[string]bool)
	cfg := config.Global()

	for _, moniker := range ctx.GetExecutionMonikers() {
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Get requirements from lint providers for each component type
		if module.Components != nil && len(module.Components.GetEnabled()) > 0 && cfg != nil && cfg.LintProviders != nil {
			for _, compName := range module.Components.GetEnabled() {
				compType := module.Components.GetComponentType(compName)
				providerNames := cfg.LintProviders.GetProvidersForComponentType(compType)
				for _, providerName := range providerNames {
					provider := cfg.LintProviders.Get(providerName)
					if provider != nil && provider.SystemDependency != "" {
						depsMap[provider.SystemDependency] = true
					}
				}
			}
		}
	}

	// No dependencies to verify
	if len(depsMap) == 0 {
		return status
	}

	// Convert to sorted slice for consistent output
	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)

	// Filter out platform-incompatible tools before verification
	deps = tool.FilterPlatformSupported(deps)
	status.Required = deps

	// Verify dependencies using tool registry
	registry := tool.GlobalRegistry()
	results := registry.VerifyAll(deps)

	for _, result := range results {
		status.Available = append(status.Available, initsummary.DepsResult{
			Name:      result.ToolID,
			Available: result.Available,
			Version:   result.Version,
		})
		if !result.Available {
			status.Missing = append(status.Missing, result.ToolID)
		}
	}

	return status
}
