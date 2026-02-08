package eac

import (
	"context"

	"github.com/ready-to-release/eac/go/core/tool"
)

// nativeAdapter executes EAC commands using the local binary via tool.Executor.
type nativeAdapter struct {
	workspaceRoot string
	executor      *tool.DefaultExecutor
	toolDef       *tool.ToolDefinition
}

// NewNative creates an EAC adapter that uses the local binary.
// binaryPath is typically resolved from paths.CommandsBinaryPath().
func NewNative(binaryPath, workspaceRoot string, executor *tool.DefaultExecutor) EACPort {
	td := &tool.ToolDefinition{
		ID:          "eac",
		Description: "EAC CLI - native mode",
		Type:        tool.ToolTypeSystem,
		Binary:      binaryPath,
	}
	return &nativeAdapter{
		workspaceRoot: workspaceRoot,
		executor:      executor,
		toolDef:       td,
	}
}

func (a *nativeAdapter) Execute(ctx context.Context, args []string, cfg *ExecConfig) (*Result, error) {
	if cfg == nil {
		cfg = &ExecConfig{}
	}
	ws := cfg.WorkspaceRoot
	if ws == "" {
		ws = a.workspaceRoot
	}

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: ws,
		ModuleRoot:    cfg.ModuleRoot,
		OutputDir:     cfg.OutputDir,
		StdoutWriter:  cfg.StdoutWriter,
		StderrWriter:  cfg.StderrWriter,
		StdinReader:   cfg.StdinReader,
		EnvOverrides:  cfg.Env,
		FullEnv:       cfg.FullEnv,
		ArgsOverrides: args,
	}
	toolResult, err := a.executor.Execute(ctx, a.toolDef, execCtx)
	if err != nil {
		return nil, err
	}
	return toResult(toolResult), nil
}

// ResolvedBinaryPath returns the binary path for diagnostics.
// This is NOT part of the EACPort interface.
func (a *nativeAdapter) ResolvedBinaryPath() string { return a.toolDef.Binary }
