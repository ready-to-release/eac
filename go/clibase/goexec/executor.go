// Package goexec provides tool-routed Go CLI execution.
// It routes go commands through the tool registry and executor,
// enabling consistent command routing via executor-mode configuration.
package goexec

import (
	"context"
	"fmt"

	"github.com/ready-to-release/eac/go/core/tool"
)

// Run executes a go CLI command through the tool registry.
// Convenience function for callers that don't need context cancellation.
func Run(workDir string, args ...string) ([]byte, error) {
	return RunContext(context.Background(), workDir, args...)
}

// RunContext executes a go CLI command with context through the tool registry.
func RunContext(ctx context.Context, workDir string, args ...string) ([]byte, error) {
	registry := tool.GlobalRegistry()
	goTool := registry.GetOrAdhoc("go")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workDir,
		ModuleRoot:    workDir,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, goTool, execCtx)
	if err != nil {
		return nil, fmt.Errorf("go execution failed: %w", err)
	}
	if result.ExitCode != 0 {
		return result.Stderr, fmt.Errorf("go command failed (exit %d): %s",
			result.ExitCode, string(result.Stderr))
	}
	return result.Stdout, nil
}

// RunCombined executes a go CLI command and returns stdout, exit code, and error.
// Does not treat non-zero exit codes as errors - useful for commands where
// non-zero exit has semantic meaning (e.g., go mod tidy -diff detecting changes).
func RunCombined(ctx context.Context, workDir string, args ...string) (stdout []byte, exitCode int, err error) {
	registry := tool.GlobalRegistry()
	goTool := registry.GetOrAdhoc("go")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workDir,
		ModuleRoot:    workDir,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(ctx, goTool, execCtx)
	if err != nil {
		return nil, -1, fmt.Errorf("go execution failed: %w", err)
	}
	return result.Stdout, result.ExitCode, nil
}

// RunSilent executes a go command and returns only the error (discards output).
// Useful for commands like go mod tidy where only success/failure matters.
func RunSilent(ctx context.Context, workDir string, args ...string) error {
	_, err := RunContext(ctx, workDir, args...)
	return err
}
