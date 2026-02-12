package eac

import (
	"context"

	"github.com/ready-to-release/eac/go/core/tool"
)

// containerAdapter executes EAC commands through the eac-ext container.
type containerAdapter struct {
	workspaceRoot string
	executor      *tool.DefaultExecutor
	toolDef       *tool.ToolDefinition
}

// NewContainer creates an EAC adapter that uses the eac-ext container.
func NewContainer(workspaceRoot string, executor *tool.DefaultExecutor, toolDef *tool.ToolDefinition) EACPort {
	return &containerAdapter{
		workspaceRoot: workspaceRoot,
		executor:      executor,
		toolDef:       toolDef,
	}
}

func (a *containerAdapter) Execute(ctx context.Context, args []string, cfg *ExecConfig) (*Result, error) {
	execCtx, ws := buildExecContext(a.workspaceRoot, args, cfg)

	// Container overrides: ModuleRoot is always the workspace root,
	// and placeholders map {workspace} for mount path resolution.
	execCtx.ModuleRoot = ws
	execCtx.Placeholders = map[string]string{
		"{workspace}": ws,
	}

	toolResult, err := a.executor.Execute(ctx, a.toolDef, execCtx)
	if err != nil {
		return nil, err
	}
	return toResult(toolResult), nil
}
