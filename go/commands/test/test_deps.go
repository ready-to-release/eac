package test

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/tool"
)

// testDepsVerifier verifies system dependencies for testing.
// Resolves test tools for each module's component types and collects their requirements.
func testDepsVerifier(ctx *cmdframework.ExecutionContext) *initsummary.DepsStatus {
	// Defensive guard: ModuleRegistry should be populated (async check starts after phaseResolve).
	if ctx.ModuleRegistry == nil {
		return &initsummary.DepsStatus{Verified: true}
	}

	depsMap := make(map[string]bool)
	bridge := tool.GlobalTestBridge()

	for _, moniker := range ctx.GetExecutionMonikers() {
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		for _, compName := range module.GetEnabledComponents() {
			compType := module.Components.GetComponentType(compName)
			td := bridge.ResolveTool(compType, core.ActionTest)
			if td != nil {
				for _, req := range td.Requirements {
					depsMap[req] = true
				}
			}
		}
	}

	return cmdframework.VerifyDeps(depsMap)
}
