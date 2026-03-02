package deploy

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ResolveUnitSpecs converts modules to deployable unit specs.
// Each deployable component in a module produces one unit spec.
func ResolveUnitSpecs(ctx *cmdframework.ExecutionContext) []workunit.UnitSpec {
	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		return nil
	}

	// Optional component filter from deploy config
	var componentFilter string
	if deployCfg, ok := ctx.Config.DeployCmdConfig.(*DeployConfig); ok {
		componentFilter = deployCfg.Component
	}

	var specs []workunit.UnitSpec
	globalIndex := 0

	for _, moniker := range monikers {
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Find deployable components via DeployBridge
		handlers := tool.GlobalDeployBridge().GetHandlersForModule(module)
		if len(handlers) == 0 {
			continue
		}

		for _, ch := range handlers {
			// Apply component filter if specified
			if componentFilter != "" && ch.Component != componentFilter {
				continue
			}

			compTypeName := module.Components.GetComponentType(ch.Component)

			specs = append(specs, workunit.UnitSpec{
				ID: workunit.UnitID{
					Action:        core.ActionDeploy,
					Module:        moniker,
					ComponentType: compTypeName,
					ComponentName: ch.Component,
					Tool:          ch.Handler.Name(),
				},
				ComponentType:  compTypeName,
				Weight:         1,
				PoolAllocation: core.HostOnlyAllocation(1),
				Metadata:       make(map[string]any),
				Index:          globalIndex,
			})
			globalIndex++
		}
	}

	return specs
}
