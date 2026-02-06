// Package gitexec provides tool-routed git CLI execution.
// It routes git commands through the tool registry and executor,
// enabling consistent command routing via executor-mode configuration.
package gitexec

import (
	"context"
	"fmt"

	"github.com/ready-to-release/eac/go/core/tool"
)

// Run executes a git CLI command through the tool registry.
// Convenience function for callers that don't need context cancellation.
func Run(workDir string, args ...string) ([]byte, error) {
	return RunContext(context.Background(), workDir, args...)
}

// RunContext executes a git CLI command with context through the tool registry.
func RunContext(ctx context.Context, workDir string, args ...string) ([]byte, error) {
	registry := tool.GlobalRegistry()
	gitTool := registry.GetOrAdhoc("git")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workDir,
		ModuleRoot:    workDir,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, gitTool, execCtx)
	if err != nil {
		return nil, fmt.Errorf("git execution failed: %w", err)
	}
	if result.ExitCode != 0 {
		return result.Stderr, fmt.Errorf("git command failed (exit %d): %s",
			result.ExitCode, string(result.Stderr))
	}
	return result.Stdout, nil
}

// RunCombined executes a git CLI command and returns stdout, exit code, and error.
// Does not treat non-zero exit codes as errors - useful for commands where
// non-zero exit has semantic meaning (e.g., git show-ref --verify for non-existent ref).
func RunCombined(ctx context.Context, workDir string, args ...string) (stdout []byte, exitCode int, err error) {
	registry := tool.GlobalRegistry()
	gitTool := registry.GetOrAdhoc("git")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workDir,
		ModuleRoot:    workDir,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, gitTool, execCtx)
	if err != nil {
		return nil, -1, fmt.Errorf("git execution failed: %w", err)
	}
	return result.Stdout, result.ExitCode, nil
}

// RunSilent executes a git command and returns only the error (discards output).
// Useful for commands like git checkout where only success/failure matters.
func RunSilent(ctx context.Context, workDir string, args ...string) error {
	_, err := RunContext(ctx, workDir, args...)
	return err
}
