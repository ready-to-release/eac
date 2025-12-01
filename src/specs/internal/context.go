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

	// MockOverrides holds per-scenario mock environment variable overrides.
	// Keys are env var names (e.g., "R2R_MOCK_AI_SPECS"), values are mock file names.
	MockOverrides map[string]string
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
	c.MockOverrides = nil // Clear per-scenario mock overrides
	// Don't reset OriginalRepoRoot - it's set once at init
}

// SetMockOverride sets a mock environment variable override for this scenario.
// The key should be the env var name (e.g., "R2R_MOCK_AI_SPECS"),
// and value is the mock file name (e.g., "mock-response-conflict.txt").
func (c *TestContext) SetMockOverride(key, value string) {
	if c.MockOverrides == nil {
		c.MockOverrides = make(map[string]string)
	}
	c.MockOverrides[key] = value
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
	parts := parseCommandLine(cmdLine)
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

	// Set mock AI directory for subprocess commands
	// This enables commands to use mock responses instead of real AI calls
	if c.OriginalRepoRoot != "" {
		assetsDir := filepath.Join(c.OriginalRepoRoot, "src", "specs", "impl", "src-commands", "assets")
		if _, err := os.Stat(assetsDir); err == nil {
			env = append(env, fmt.Sprintf("R2R_MOCK_AI_DIR=%s", assetsDir))
		}
	}

	// Apply per-scenario mock overrides
	for key, value := range c.MockOverrides {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
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

// parseCommandLine parses a command line string, respecting quoted arguments.
// Handles both single quotes and double quotes.
// Examples:
//
//	"create spec -o 'custom/path' 'Test feature'" -> ["create", "spec", "-o", "custom/path", "Test feature"]
//	'hello "world test"' -> ["hello", "world test"]
func parseCommandLine(cmdLine string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range cmdLine {
		switch {
		case (r == '\'' || r == '"') && !inQuote:
			// Start of quoted string
			inQuote = true
			quoteChar = r
		case r == quoteChar && inQuote:
			// End of quoted string
			inQuote = false
			quoteChar = 0
		case r == ' ' && !inQuote:
			// Space outside quotes - end of token
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	// Don't forget the last token
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
