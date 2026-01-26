// Package toolcontainer provides a unified framework for executing container-based tools.
// It supports two execution modes:
// - Build mode: Builds from Dockerfile (for local development)
// - Container mode: Pulls from GHCR (for r2r/eac container mode)
package toolcontainer

import (
	"context"
	"io"
	"time"
)

// Mode represents the execution mode for a container tool.
type Mode string

const (
	// ModeLocal builds containers from Dockerfile for local development.
	ModeLocal Mode = "local"

	// ModeContainer pulls containers from registry when running in r2r/eac container.
	ModeContainer Mode = "container"
)

// RunConfig contains the configuration for running a container command.
type RunConfig struct {
	// Command is the command to execute (overrides container ENTRYPOINT).
	Command string

	// Args are the arguments to pass to the command.
	Args []string

	// WorkDir is the host directory to mount as the container's working directory.
	// This directory will be mounted at the container's configured workdir.
	WorkDir string

	// Stdin is the standard input reader (optional).
	Stdin io.Reader

	// Stdout is the standard output writer (optional, defaults to discard).
	Stdout io.Writer

	// Stderr is the standard error writer (optional, defaults to discard).
	Stderr io.Writer

	// Timeout is the maximum execution time (default: 10 minutes).
	Timeout time.Duration

	// Env is additional environment variables to set in the container.
	Env []string
}

// RunResult contains the result of a container execution.
type RunResult struct {
	// ExitCode is the container's exit code.
	ExitCode int

	// Error is any error that occurred during execution.
	Error error
}

// Runner defines the interface for executing container-based tools.
type Runner interface {
	// Run executes the configured command in the container.
	Run(ctx context.Context, cfg *RunConfig) *RunResult

	// Mode returns the execution mode (local or container).
	Mode() Mode

	// Close releases any resources held by the runner.
	Close() error
}

// DefaultTimeout is the default execution timeout for container commands.
const DefaultTimeout = 10 * time.Minute
