package eac

import (
	"context"
	"fmt"
)

// RunCommand is a convenience function that wraps EACPort.Execute().
// Returns stdout as a string, or an error if the command fails.
func RunCommand(ctx context.Context, port EACPort, workspaceRoot string, args []string) (string, error) {
	result, err := port.Execute(ctx, args, &ExecConfig{WorkspaceRoot: workspaceRoot})
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("command %v failed (exit %d): %s",
			args, result.ExitCode, string(result.Stderr))
	}
	return string(result.Stdout), nil
}

// CommandExecutorAdapter wraps EACPort to implement command executor
// interfaces that use the func(ctx, workspaceRoot, args) (string, error)
// signature. This bridges the new EACPort to existing CommandExecutor
// interfaces without requiring those packages to import this adapter.
type CommandExecutorAdapter struct {
	Port EACPort
}

// Run implements the CommandExecutor pattern used by docprep/content.
func (a *CommandExecutorAdapter) Run(ctx context.Context, workspaceRoot string, args []string) (string, error) {
	return RunCommand(ctx, a.Port, workspaceRoot, args)
}
