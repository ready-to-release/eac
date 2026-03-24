package lint

import (
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/config"
)

// lintDepsVerifier verifies system dependencies for linting.
func lintDepsVerifier(ctx *cmdframework.ExecutionContext) *initsummary.DepsStatus {
	// Defensive guard: ModuleRegistry should be populated (async check starts after phaseResolve).
	if ctx.ModuleRegistry == nil {
		return &initsummary.DepsStatus{Verified: true}
	}

	depsMap := make(map[string]bool)
	cfg := config.Global()

	for _, moniker := range ctx.GetExecutionMonikers() {
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

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

	return cmdframework.VerifyDeps(depsMap)
}
