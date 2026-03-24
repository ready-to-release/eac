package scan

import (
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

// scanDepsVerifier verifies system dependencies for scanning.
// Resolves scanner tools for each module's component types and collects their requirements.
func scanDepsVerifier(ctx *cmdframework.ExecutionContext) *initsummary.DepsStatus {
	// Defensive guard: ModuleRegistry should be populated (async check starts after phaseResolve).
	if ctx.ModuleRegistry == nil {
		return &initsummary.DepsStatus{Verified: true}
	}

	depsMap := make(map[string]bool)
	cfg := config.Global()
	bridge := tool.GlobalScanBridge()

	for _, moniker := range ctx.GetExecutionMonikers() {
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		tools := bridge.GetScannerToolsForModule(module, cfg.ComponentKinds)
		for _, td := range tools {
			for _, req := range td.Requirements {
				depsMap[req] = true
			}
		}
	}

	return cmdframework.VerifyDeps(depsMap)
}
