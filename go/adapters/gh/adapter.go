// Package gh provides a tool-system-backed implementation of github.CLIExecutor.
// It routes gh CLI commands through the injected ToolSystem, enabling
// configurable binary paths, container execution, and testability.
package gh

import (
	"context"

	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/tool"
)

// Ensure interface compliance.
var _ github.CLIExecutor = (*Adapter)(nil)

// Adapter implements github.CLIExecutor by executing gh commands
// through an injected ToolSystem.
type Adapter struct {
	ts      *tool.ToolSystem
	workDir string
}

// New creates a gh CLI adapter backed by the given ToolSystem.
func New(ts *tool.ToolSystem, workDir string) *Adapter {
	return &Adapter{ts: ts, workDir: workDir}
}

// Exec implements github.CLIExecutor.
func (a *Adapter) Exec(args ...string) ([]byte, error) {
	return a.ExecContext(context.Background(), args...)
}

// ExecContext implements github.CLIExecutor.
func (a *Adapter) ExecContext(ctx context.Context, args ...string) ([]byte, error) {
	return a.ts.RunTool(ctx, "gh", a.workDir, args...)
}
