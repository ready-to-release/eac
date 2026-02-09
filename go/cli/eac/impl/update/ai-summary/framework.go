// Package aisummary provides framework integration for the ai-summary command.
package aisummary

import (
	"context"
	"fmt"
	"io"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/ai/toolhandler"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register component-level execution support
	cmdframework.RegisterUnitProvider(core.ActionAISummary, ResolveUnitSpecs)
	cmdframework.RegisterUnitWorker(core.ActionAISummary, aiSummaryUnitWorker)
}

// RunAISummaryWithFramework executes the ai-summary command using the cmdframework.
func RunAISummaryWithFramework(cmdCfg *cmdframework.CommandConfig) int {
	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:    aiSummaryAfterInit,
		AfterResolve: aiSummaryAfterResolve,
	}

	return cmdframework.Run(cmdCfg, nil, hooks)
}

// aiSummaryAfterInit handles initialization after framework init.
func aiSummaryAfterInit(ctx *cmdframework.ExecutionContext) error {
	summary := initsummary.New("ai-summary").
		SetRequest(ctx.Config.Monikers, ctx.GetExecutionMonikers()).
		SetFlags(initsummary.Flags{
			DebugMode: ctx.Config.DebugMode,
			UseTUI:    ctx.Config.UseTUI,
		}).
		SetOutputDir(paths.OutAISummaryRelPath)

	ctx.InitSummary = summary
	return nil
}

// aiSummaryAfterResolve handles post-resolution setup.
func aiSummaryAfterResolve(ctx *cmdframework.ExecutionContext) error {
	// Nothing special needed for ai-summary
	return nil
}

// ResolveUnitSpecs generates UnitSpecs for AI analysis components.
func ResolveUnitSpecs(ctx *cmdframework.ExecutionContext) []workunit.UnitSpec {
	// Get analysis type filter from config
	analysisType := ctx.Config.AnalysisType

	// Determine which analysis types to run
	analysisTypes := []string{"dsl", "specs", "docs"}
	if analysisType != "" {
		analysisTypes = []string{analysisType}
	}

	// Build specs list (source depends on dsl and specs via DependsOn)
	var specs []workunit.UnitSpec

	for _, moniker := range ctx.GetExecutionMonikers() {
		for _, aType := range analysisTypes {
			toolID := "ai-" + aType + "-analyzer"

			spec := workunit.UnitSpec{
				ID: workunit.UnitID{
					Action:        core.ActionAISummary,
					Module:        moniker,
					ComponentType: "ai-" + aType,
					ComponentName: "ai-" + aType,
					Tool:          toolID,
				},
				ComponentType:  "ai-" + aType,
				Weight:         getAnalysisWeight(aType),
				PoolAllocation: core.HostOnlyAllocation(getAnalysisWeight(aType)),
				DependsOn:      []workunit.UnitID{},
				Cached:        false,
				Metadata:      make(map[string]any),
			}

			specs = append(specs, spec)
		}

		// Add source analysis if not filtered out
		if analysisType == "" || analysisType == "source" {
			sourceSpec := workunit.UnitSpec{
				ID: workunit.UnitID{
					Action:        core.ActionAISummary,
					Module:        moniker,
					ComponentType: "ai-source",
					ComponentName: "ai-source",
					Tool:          "ai-source-analyzer",
				},
				ComponentType:  "ai-source",
				Weight:         1, // AI tasks use weight 1 (API-bound, not CPU-bound)
				PoolAllocation: core.HostOnlyAllocation(1),
				DependsOn: []workunit.UnitID{
					{Action: core.ActionAISummary, Module: moniker, ComponentType: "ai-dsl", ComponentName: "ai-dsl", Tool: "ai-dsl-analyzer"},
					{Action: core.ActionAISummary, Module: moniker, ComponentType: "ai-specs", ComponentName: "ai-specs", Tool: "ai-specs-analyzer"},
				},
				Cached:   false,
				Metadata: make(map[string]any),
			}

			specs = append(specs, sourceSpec)
		}
	}

	// Set indices
	for i := range specs {
		specs[i].Index = i
	}

	return specs
}

// getAnalysisWeight returns the scheduling weight for an analysis type.
// AI tasks use weight 1 since they're API-bound, not CPU/RAM-bound.
// Parallelism is controlled via MaxConcurrency in the command config.
func getAnalysisWeight(analysisType string) int {
	// All AI tasks have weight 1 - parallelism is token-limited, not resource-limited
	return 1
}

// aiSummaryUnitWorker executes a single AI analysis unit.
func aiSummaryUnitWorker(goCtx context.Context, ctx *cmdframework.ExecutionContext, spec core.UnitSpec, logWriter io.Writer) int {
	module := spec.ID.Module
	component := spec.ID.ComponentName

	// Get module contract
	moduleContract, exists := ctx.ModuleRegistry.Get(module)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", module)
		return 1
	}

	// Extract analysis type from component name (e.g., "ai-dsl" -> "dsl")
	analysisType := extractAnalysisType(component)

	// Get the tool definition
	toolID := "ai-" + analysisType + "-analyzer"
	bridge := tool.GlobalBuildBridge()
	handler := bridge.GetHandler(toolID)

	// If no handler found, create AI handler directly
	if handler == nil {
		// Create tool definition
		toolDef := &tool.ToolDefinition{
			ID:          toolID,
			Description: fmt.Sprintf("AI-powered %s analysis", analysisType),
			Type:        tool.ToolTypeSystem,
			Binary:      "eac",
			Args:        []string{"ai", "analyze", "--type=" + analysisType},
			Resources: &tool.ToolResources{
				CPUs: 1, // AI tasks use weight 1 (API-bound, not CPU-bound)
			},
		}
		handler = toolhandler.NewAIToolHandler(toolDef)
	}

	if handler == nil {
		output.Writeln(logWriter, "Error: no handler found for %s", toolID)
		return 1
	}

	// Build output directory
	outputDir := paths.AISummaryOutputPath(ctx.WorkspaceRoot, module)

	// Adapt module to port interface
	modulePort := adapters.AdaptModule(moduleContract)

	// Build options
	opts := tool.BuildOptions{
		Component: analysisType,
		DryRun:    ctx.Config.DryRun,
	}

	// In dry-run mode, just report what would happen
	if ctx.Config.DryRun {
		output.Writeln(logWriter, "Would analyze %s for module %s", analysisType, module)
		output.Writeln(logWriter, "Output: %s/%s-status.md", outputDir, analysisType)
		return 0
	}

	// Execute the handler
	return handler.Build(modulePort, ctx.WorkspaceRoot, outputDir, logWriter, opts)
}

// extractAnalysisType extracts the analysis type from a component name.
// Example: "ai-dsl" -> "dsl", "ai-specs:ai-specs-analyzer" -> "specs"
func extractAnalysisType(component string) string {
	// Handle "ai-dsl:ai-dsl-analyzer" format - strip tool suffix
	for i, c := range component {
		if c == ':' {
			component = component[:i]
			break
		}
	}

	// Strip "ai-" prefix
	if len(component) > 3 && component[:3] == "ai-" {
		return component[3:]
	}
	return component
}
