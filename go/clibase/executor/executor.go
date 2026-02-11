// Package executor dispatches CLI commands via the CommandRegistryPort.
package executor

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Executor looks up commands from a registry and executes them.
type Executor struct {
	registry core.CommandRegistryPort
}

// New creates an Executor backed by the given command registry.
func New(registry core.CommandRegistryPort) *Executor {
	return &Executor{registry: registry}
}

// Execute dispatches to the named command with the given args.
// Returns the command's exit code, or 1 if the command is not found
// or is a parent-only command (not executable).
func (e *Executor) Execute(ctx context.Context, name string, args []string) int {
	cmd, ok := e.registry.Get(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", name)
		return 1
	}

	simple, ok := cmd.(core.SimpleCommandPort)
	if !ok {
		fmt.Fprintf(os.Stderr, "command %q is a parent command and cannot be executed directly\n", name)
		return 1
	}

	req := &core.CommandRequest{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return simple.Execute(ctx, req)
}

// Registry returns the underlying command registry.
func (e *Executor) Registry() core.CommandRegistryPort {
	return e.registry
}
