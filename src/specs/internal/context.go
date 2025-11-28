// Package internal provides shared helpers for godog BDD tests.
//
// This package contains common test context, step definitions, and utilities
// that are shared across all spec implementations in src/specs/impl/.
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// TestContext wraps the core SharedTestContext with additional spec-specific state.
type TestContext struct {
	*coretesting.SharedTestContext

	// OriginalRepoRoot is the actual repository root (for running go commands)
	OriginalRepoRoot string

	// IsolatedDir is the temp directory for isolated tests
	IsolatedDir string

	// Isolation infrastructure
	Isolation *coretesting.TestIsolation
}

// NewTestContext creates a new test context.
func NewTestContext() *TestContext {
	return &TestContext{
		SharedTestContext: coretesting.NewSharedTestContext(),
	}
}

// Reset clears all fields for a new scenario.
func (c *TestContext) Reset() {
	c.SharedTestContext.Reset()
	// Don't reset OriginalRepoRoot - it's set once at init
}

// SetupIsolation creates an isolated test environment.
func (c *TestContext) SetupIsolation() error {
	c.Isolation = coretesting.NewTestIsolation().
		WithOriginalRepoRoot(c.OriginalRepoRoot).
		WithCopyContracts(true).
		WithCopyAIContracts(true).
		WithMockAIConfig(true)

	if err := c.Isolation.Setup(); err != nil {
		return fmt.Errorf("failed to setup test isolation: %w", err)
	}

	c.IsolatedDir = c.Isolation.IsolatedDir()
	c.SharedTestContext.SetIsolation(c.OriginalRepoRoot, c.IsolatedDir)
	c.SharedTestContext.Isolation = c.Isolation

	return nil
}

// CleanupIsolation tears down the isolated test environment.
func (c *TestContext) CleanupIsolation() {
	if c.Isolation != nil {
		c.Isolation.Cleanup()
		c.Isolation = nil
		c.IsolatedDir = ""
	}
	c.SharedTestContext.ClearIsolation()
}

// RunCommand executes a command and captures output.
// This is the core command execution used by step definitions.
func (c *TestContext) RunCommand(cmdLine string) error {
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Build go run command
	cmdArgs := append([]string{"run", "."}, parts...)
	cmd := exec.Command("go", cmdArgs...)

	// Determine working directory
	if c.IsolatedDir != "" {
		// For isolated tests, run from source but use isolated env
		cmd.Dir = filepath.Join(c.OriginalRepoRoot, "src", "commands")
	} else {
		cmd.Dir = filepath.Join(c.OriginalRepoRoot, "src", "commands")
	}

	// Build environment
	env := os.Environ()
	if c.IsolatedDir != "" {
		env = append(env, fmt.Sprintf("R2R_PWD=%s", c.IsolatedDir))
		env = append(env, fmt.Sprintf("R2R_REPO_ROOT=%s", c.IsolatedDir))
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	c.CommandOutput = string(output)
	c.CommandError = err

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.ExitCode = exitErr.ExitCode()
		} else {
			c.ExitCode = 1
		}
	} else {
		c.ExitCode = 0
	}

	return nil
}
