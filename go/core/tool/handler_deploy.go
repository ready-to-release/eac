package tool

import (
	"context"
	"io"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	deploy "github.com/ready-to-release/eac/contracts/runner/0.1.0/deploy"
)

// DeployToolHandlerAdapter wraps a ToolDefinition to implement deploy.DeployerPort.
// This allows YAML-defined tools to be used with the deploy handler system.
type DeployToolHandlerAdapter struct {
	BaseHandlerAdapter
}

// NewDeployToolHandlerAdapter creates a handler adapter for a tool definition.
func NewDeployToolHandlerAdapter(tool *ToolDefinition, executor Executor) *DeployToolHandlerAdapter {
	return &DeployToolHandlerAdapter{
		BaseHandlerAdapter: BaseHandlerAdapter{Tool: tool, Executor: executor},
	}
}

// Deploy executes the tool for a deploy operation.
func (a *DeployToolHandlerAdapter) Deploy(ctx context.Context, module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, rawOpts any) int {
	return a.execute(ctx, module, workspaceRoot, outputDir, logWriter, rawOpts, false)
}

// DryRun executes the tool for a dry-run (what-if) deploy operation.
func (a *DeployToolHandlerAdapter) DryRun(ctx context.Context, module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, rawOpts any) int {
	return a.execute(ctx, module, workspaceRoot, outputDir, logWriter, rawOpts, true)
}

// ValidateModule validates the tool can deploy — always returns nil for YAML tools.
func (a *DeployToolHandlerAdapter) ValidateModule(module core.ModuleContractPort, workspaceRoot, environment string) error {
	return nil
}

func (a *DeployToolHandlerAdapter) execute(ctx context.Context, module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, rawOpts any, dryRun bool) int {
	opts, _ := rawOpts.(DeployOptions)

	moduleRelPath := module.GetComponentRoot(opts.Component)

	dryRunFlag := ""
	if dryRun {
		dryRunFlag = "--what-if"
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRelPath,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     core.ActionDeploy,
		Placeholders: map[string]string{
			"{workspace}":    workspaceRoot,
			"{module}":       moduleRelPath,
			"{output}":       outputDir,
			"{component}":    opts.Component,
			"{dry-run-flag}": dryRunFlag,
		},
	}

	if dryRun {
		execCtx.ArgsOverrides = append(execCtx.ArgsOverrides, "--what-if")
	}

	result, err := a.Executor.Execute(ctx, a.Tool, execCtx)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, "Error executing deploy tool: "+err.Error()+"\n")
		}
		return 1
	}

	return result.ExitCode
}

var _ deploy.DeployerPort = (*DeployToolHandlerAdapter)(nil)
