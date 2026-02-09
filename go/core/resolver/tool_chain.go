package resolver

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// toolChainDep represents an external dependency for a tool chain.
// This is used to pass build_after or depends_on dependencies to the first tool.
type toolChainDep struct {
	Module        string
	ComponentType string
	Component     string
	Tool          string
}

// toolChainSpec is an intermediate representation for a tool in a chain.
// This is returned by expandToolChain and later converted to full UnitSpecs
// by the ComponentResolver which adds weight, container info, etc.
type toolChainSpec struct {
	Component string
	Tool      string
	DependsOn []workunit.UnitID
}

// expandToolChain expands a list of tools into tool chain specs with proper dependencies.
// Each tool depends on the previous tool in the chain. External dependencies from
// build_after or component depends_on are added only to the first tool.
//
// Example:
//
//	tools: ["preprocess", "mkdocs-pdf"]
//	externalDeps: [{Component: "structurizr", Tool: "structurizr-cli"}]
//
// Returns:
//
//	[0] preprocess -> depends on structurizr
//	[1] mkdocs-pdf -> depends on preprocess
func expandToolChain(module, component, compType string, tools []string, externalDeps []toolChainDep) []toolChainSpec {
	if len(tools) == 0 {
		return nil
	}

	specs := make([]toolChainSpec, len(tools))

	for i, tool := range tools {
		spec := toolChainSpec{
			Component: component,
			Tool:      tool,
			DependsOn: []workunit.UnitID{},
		}

		if i == 0 {
			// First tool gets external dependencies (build_after, component depends_on)
			for _, dep := range externalDeps {
				spec.DependsOn = append(spec.DependsOn, workunit.UnitID{
					Action:        core.ActionBuild,
					Module:        dep.Module,
					ComponentType: dep.ComponentType,
					ComponentName: dep.Component,
					Tool:          dep.Tool,
				})
			}
		} else {
			// Subsequent tools depend on the previous tool in the chain
			prevTool := tools[i-1]
			spec.DependsOn = append(spec.DependsOn, workunit.UnitID{
				Action:        core.ActionBuild,
				Module:        module,
				ComponentType: compType,
				ComponentName: component,
				Tool:          prevTool,
			})
		}

		specs[i] = spec
	}

	return specs
}
