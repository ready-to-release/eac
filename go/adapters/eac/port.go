package eac

import (
	"context"
	"io"
	"time"

	"github.com/ready-to-release/eac/go/core/tool"
)

// EACPort defines the interface for executing EAC commands.
// The interface is intentionally minimal to maximize mockability.
type EACPort interface {
	// Execute runs an EAC command and returns the result.
	// Pass nil for cfg to use defaults.
	Execute(ctx context.Context, args []string, cfg *ExecConfig) (*Result, error)
}

// ExecConfig configures EAC command execution.
// All fields are optional. Zero values use sensible defaults.
type ExecConfig struct {
	WorkspaceRoot string            // Repository root (defaults to adapter's root)
	ModuleRoot    string            // Module root (relative to workspace)
	OutputDir     string            // Output directory for artifacts
	StdoutWriter  io.Writer         // nil = capture to Result.Stdout
	StderrWriter  io.Writer         // nil = capture to Result.Stderr
	StdinReader   io.Reader         // nil = no stdin
	Env           map[string]string // Additive env overrides
	FullEnv       []string          // Replaces entire env (nil = use default merge)
}

// Result holds the execution result.
type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Duration time.Duration
}

// Success returns true if the command exited with code 0.
func (r *Result) Success() bool {
	return r.ExitCode == 0
}

// toResult converts a tool.ExecutionResult to a Result.
func toResult(tr *tool.ExecutionResult) *Result {
	return &Result{
		ExitCode: tr.ExitCode,
		Stdout:   tr.Stdout,
		Stderr:   tr.Stderr,
		Duration: tr.Duration,
	}
}
