// Package ghexec provides tool-routed GitHub CLI execution.
// It routes gh commands through the tool registry and executor,
// implementing the github.CLIExecutor interface from go/core/github.
package ghexec

import (
	"context"
	"fmt"

	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/tool"
)

// Ensure interface compliance.
var _ github.CLIExecutor = (*ToolRoutedExecutor)(nil)

// ToolRoutedExecutor implements github.CLIExecutor by routing gh commands
// through the tool registry and executor.
type ToolRoutedExecutor struct {
	workDir string
}

// New creates a tool-routed gh CLI executor.
func New(workDir string) *ToolRoutedExecutor {
	return &ToolRoutedExecutor{workDir: workDir}
}

// Exec executes a gh CLI command through the tool registry.
func (e *ToolRoutedExecutor) Exec(args ...string) ([]byte, error) {
	return e.ExecContext(context.Background(), args...)
}

// ExecContext executes a gh CLI command with context through the tool registry.
func (e *ToolRoutedExecutor) ExecContext(ctx context.Context, args ...string) ([]byte, error) {
	registry := tool.GlobalRegistry()
	ghTool := registry.GetOrAdhoc("gh")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: e.workDir,
		ModuleRoot:    e.workDir,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, ghTool, execCtx)
	if err != nil {
		return nil, fmt.Errorf("gh execution failed: %w", err)
	}
	if result.ExitCode != 0 {
		return result.Stderr, fmt.Errorf("gh command failed (exit %d): %s",
			result.ExitCode, string(result.Stderr))
	}
	return result.Stdout, nil
}

// Run executes a gh CLI command through the tool registry.
// Convenience function for callers that don't need a persistent executor.
func Run(workDir string, args ...string) ([]byte, error) {
	return RunContext(context.Background(), workDir, args...)
}

// RunContext executes a gh CLI command with context through the tool registry.
func RunContext(ctx context.Context, workDir string, args ...string) ([]byte, error) {
	registry := tool.GlobalRegistry()
	ghTool := registry.GetOrAdhoc("gh")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workDir,
		ModuleRoot:    workDir,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, ghTool, execCtx)
	if err != nil {
		return nil, fmt.Errorf("gh execution failed: %w", err)
	}
	if result.ExitCode != 0 {
		return result.Stderr, fmt.Errorf("gh command failed (exit %d): %s",
			result.ExitCode, string(result.Stderr))
	}
	return result.Stdout, nil
}

// RunCombined executes a gh CLI command and returns stdout, exit code, and error.
// Does not treat non-zero exit codes as errors - useful for commands where
// non-zero exit has semantic meaning (e.g., gh release view for non-existent release).
func RunCombined(ctx context.Context, workDir string, args ...string) (stdout []byte, exitCode int, err error) {
	registry := tool.GlobalRegistry()
	ghTool := registry.GetOrAdhoc("gh")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workDir,
		ModuleRoot:    workDir,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, ghTool, execCtx)
	if err != nil {
		return nil, -1, fmt.Errorf("gh execution failed: %w", err)
	}
	return result.Stdout, result.ExitCode, nil
}
