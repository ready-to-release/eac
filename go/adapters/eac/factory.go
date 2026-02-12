package eac

import (
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

// New creates the appropriate EAC adapter based on registry binding resolution.
// This is the canonical entry point for production code.
//
// Dependency injection pattern: Callers should call New() once and pass the
// returned EACPort to functions that need it (constructor/parameter injection).
// Do NOT call New() inside the function that needs EAC execution.
func New(workspaceRoot string, registry *tool.DefaultRegistry, executor *tool.DefaultExecutor) EACPort {
	td, ok := registry.Get("eac")
	if !ok {
		binaryPath := paths.CommandsBinaryPath(workspaceRoot)
		return NewNative(binaryPath, workspaceRoot, executor)
	}

	switch td.Type {
	case tool.ToolTypeContainer:
		return NewContainer(workspaceRoot, executor, td)
	default:
		binaryPath := paths.CommandsBinaryPath(workspaceRoot)
		return NewNative(binaryPath, workspaceRoot, executor)
	}
}
