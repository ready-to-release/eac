// Package internal provides shared helpers for godog BDD tests.
//
// This package contains common test context, step definitions, and utilities
// that are shared across all spec implementations in go/eac/specs/impl/.
package internal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	coretesting "github.com/ready-to-release/eac/go/eac/core/testing"
)

// Environment variable names used in test contexts.
const (
	EnvR2RTestLoggingActive = "R2R_TEST_LOGGING_ACTIVE"
	EnvR2RPWD               = "R2R_PWD"
	EnvR2RRepoRoot          = "R2R_REPO_ROOT"
	EnvR2RContainerRoot     = "R2R_CONTAINER_ROOT"
	EnvR2RMockAIDir         = "R2R_MOCK_AI_DIR"
	EnvR2RMockSecurity      = "R2R_MOCK_SECURITY"
	EnvR2RMockStructurizr   = "R2R_MOCK_STRUCTURIZR"
)

// TestContext wraps the core SharedTestContext with additional spec-specific state.
type TestContext struct {
	*coretesting.SharedTestContext

	// OriginalRepoRoot is the actual repository root (for running go commands)
	OriginalRepoRoot string

	// IsolatedDir is the temp directory for isolated tests (main repository root)
	// This should NEVER be changed after initial setup - it's the main isolated directory
	IsolatedDir string

	// CurrentWorkDir is the current working directory within the isolated environment
	// This changes when switching between main workspace and feature worktrees
	CurrentWorkDir string

	// Isolation infrastructure
	Isolation *coretesting.TestIsolation

	// FixturePool manages reusable test environment templates (per feature)
	FixturePool *coretesting.FixturePool

	// MockOverrides holds per-scenario mock environment variable overrides.
	// Keys are env var names (e.g., "R2R_MOCK_AI_SPECS"), values are mock file names.
	MockOverrides map[string]string

	// OriginalRepoCache provides cached data from the ORIGINAL repository root.
	// This is NOT for isolated/mocked test repositories - only the real repo.
	// This dramatically improves performance by avoiding repeated git/file operations.
	OriginalRepoCache *TestCache
}

// NewTestContext creates a new test context.
func NewTestContext() *TestContext {
	return &TestContext{
		SharedTestContext: coretesting.NewSharedTestContext(),
		// OriginalRepoCache is set by runner.go to the global cache
		// Don't create a new one here - it's wasteful and gets overwritten
	}
}

// Reset clears all fields for a new scenario.
func (c *TestContext) Reset() {
	c.SharedTestContext.Reset()
	c.MockOverrides = nil            // Clear per-scenario mock overrides
	c.CurrentWorkDir = c.IsolatedDir // Reset to main isolated directory
	// Don't reset OriginalRepoRoot, IsolatedDir, or Cache - they're set once at init
}

// EnsureOriginalRepoCache ensures the original repo cache is populated.
// This should be called at the start of any scenario that needs cached data
// from the ORIGINAL repository (not isolated/mocked repos).
func (c *TestContext) EnsureOriginalRepoCache() error {
	// OriginalRepoCache MUST be set by runner.go to globalRepoCache
	// If it's nil, that's a programmer error - panic to catch it early
	if c.OriginalRepoCache == nil {
		panic("OriginalRepoCache is nil - runner.go must set it to globalRepoCache")
	}
	return c.OriginalRepoCache.EnsurePopulated(c.OriginalRepoRoot)
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

// SetupIsolation creates an isolated test environment using fixture pool.
// Requires fixture pool to be initialized (fails fast if not available).
// Creates test isolation by fast-copying pre-built template (~50ms).
func (c *TestContext) SetupIsolation() error {
	if c.FixturePool == nil {
		return fmt.Errorf("fixture pool not initialized - required for test isolation")
	}

	template := c.FixturePool.GetTemplate(c.OriginalRepoRoot)
	if template == nil {
		return fmt.Errorf("fixture template not created - call CreateTemplate() before SetupIsolation()")
	}

	// Fast-copy from template (~50ms)
	isolation, err := c.FixturePool.NewIsolationFromTemplate(template)
	if err != nil {
		return fmt.Errorf("failed to create isolation from template: %w", err)
	}

	c.Isolation = isolation
	c.IsolatedDir = isolation.IsolatedDir()
	c.CurrentWorkDir = c.IsolatedDir
	c.SetIsolation(c.OriginalRepoRoot, c.IsolatedDir)
	c.SharedTestContext.Isolation = c.Isolation

	// Create specs/ directory for design output
	specsDir := filepath.Join(c.IsolatedDir, "specs")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create specs directory in isolation: %w", err)
	}

	return nil
}

// CleanupIsolation tears down the isolated test environment.
func (c *TestContext) CleanupIsolation() {
	if c.Isolation != nil {
		c.Isolation.Cleanup()
		c.Isolation = nil
		c.IsolatedDir = ""
	}
	c.ClearIsolation()
}

// getEffectiveWorkDir returns the working directory for command execution.
// Returns CurrentWorkDir if set, otherwise falls back to IsolatedDir.
func (c *TestContext) getEffectiveWorkDir() string {
	if c.CurrentWorkDir != "" {
		return c.CurrentWorkDir
	}
	return c.IsolatedDir
}

// buildBaseEnvironment creates base environment, filtering out test-specific vars.
// For isolated tests, filters out R2R_TEST_LOGGING_ACTIVE so unified log works.
func (c *TestContext) buildBaseEnvironment() []string {
	env := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		if c.IsolatedDir != "" && strings.HasPrefix(e, EnvR2RTestLoggingActive+"=") {
			continue // Don't inherit - isolated tests need unified log to work
		}
		env = append(env, e)
	}
	return env
}

// buildIsolationEnvironment adds isolation-specific environment variables.
// Sets R2R_PWD and R2R_REPO_ROOT for isolated test execution.
func (c *TestContext) buildIsolationEnvironment(env []string) []string {
	if c.IsolatedDir == "" {
		return env
	}

	// R2R_PWD: current working directory (may be worktree)
	// R2R_REPO_ROOT: main repository root (never changes)
	workDir := c.getEffectiveWorkDir()
	env = append(env, fmt.Sprintf("%s=%s", EnvR2RPWD, workDir))
	env = append(env, fmt.Sprintf("%s=%s", EnvR2RRepoRoot, c.IsolatedDir))
	return env
}

// buildMockingEnvironment adds mocking environment variables for tests.
// Sets R2R_CONTAINER_ROOT, R2R_MOCK_AI_DIR, R2R_MOCK_SECURITY, and R2R_MOCK_STRUCTURIZR.
func (c *TestContext) buildMockingEnvironment(env []string) []string {
	// Set distribution root for template loading
	// Templates are NOT copied to isolated test directories - they live in the original repo
	// This allows AI commands to load system default templates from templates/ai/
	env = append(env, fmt.Sprintf("%s=%s", EnvR2RContainerRoot, c.OriginalRepoRoot))

	// Set mock AI directory for subprocess commands
	// This enables commands to use mock responses instead of real AI calls
	// Use container root if in container, otherwise repo root
	assetsRoot := repository.GetDistRoot(c.OriginalRepoRoot)
	if assetsRoot != "" {
		assetsDir := filepath.Join(assetsRoot, "go", "eac", "specs", "impl", "eac-commands", "assets")
		if _, err := os.Stat(assetsDir); err == nil {
			env = append(env, fmt.Sprintf("%s=%s", EnvR2RMockAIDir, assetsDir))
		}
	}

	// Enable security tool mocking for subprocess commands
	// This enables security commands to use mock responses instead of real Docker tools
	env = append(env, fmt.Sprintf("%s=true", EnvR2RMockSecurity))

	// Enable Structurizr CLI mocking for design validation tests
	// This allows validation tests to run without Docker
	env = append(env, fmt.Sprintf("%s=true", EnvR2RMockStructurizr))

	return env
}

// applyMockOverrides applies per-scenario mock overrides to the environment.
func (c *TestContext) applyMockOverrides(env []string) []string {
	for key, value := range c.MockOverrides {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

// RunCommand executes a command and captures output.
// This is the core command execution used by step definitions.
//
// Returns nil if command succeeds with exit code 0.
// Returns nil (not error) if command fails with non-zero exit code - check ExitCode field.
// Returns error only for execution failures (binary not found, permission denied, etc).
func (c *TestContext) RunCommand(cmdLine string) error {
	parts := parseCommandLine(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Try to use pre-built binary for better performance
	cmd := c.createCommand(parts)

	// Set working directory for isolated tests
	if c.IsolatedDir != "" {
		cmd.Dir = c.getEffectiveWorkDir()
	}

	// Build environment with all necessary variables
	env := c.buildBaseEnvironment()
	env = c.buildIsolationEnvironment(env)
	env = c.buildMockingEnvironment(env)
	env = c.applyMockOverrides(env)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	c.CommandOutput = string(output)
	c.CommandError = err

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			c.ExitCode = exitErr.ExitCode()
		} else {
			// Non-exit error (e.g., binary not found, permission denied)
			// Include comprehensive diagnostic info so tests can diagnose CI issues
			c.ExitCode = 1
			c.CommandOutput = c.formatCommandExecutionError(cmd, cmdLine, err, string(output))
		}
	} else {
		c.ExitCode = 0
	}

	return nil
}

// ============================================================================
// State Validation Methods
// ============================================================================

// IsIsolated returns true if the test context is running in an isolated environment.
func (c *TestContext) IsIsolated() bool {
	return c.IsolatedDir != ""
}

// ValidateState checks if the context is in a valid state for test execution.
// Returns an error if required fields are not set.
func (c *TestContext) ValidateState() error {
	if c.OriginalRepoRoot == "" {
		return fmt.Errorf("OriginalRepoRoot not set - test context not properly initialized")
	}
	return nil
}

// MustBeIsolated panics if the test context is not running in isolation.
// Use this in step definitions that require isolation to catch setup errors early.
func (c *TestContext) MustBeIsolated() {
	if !c.IsIsolated() {
		panic("test must run in isolated environment (missing @env:isolated-test-project tag?)")
	}
}

// ============================================================================
// Command Execution Helper Methods
// ============================================================================

// CommandSucceeded returns true if the last command succeeded (exit code 0).
func (c *TestContext) CommandSucceeded() bool {
	return c.ExitCode == 0
}

// CommandFailed returns true if the last command failed (non-zero exit code).
func (c *TestContext) CommandFailed() bool {
	return c.ExitCode != 0
}

// MustSucceed returns an error if the last command failed.
// Use this in step definitions that expect command success.
//
// Example:
//
//	err := ctx.RunCommand("build eac-core")
//	if err != nil {
//	    return err
//	}
//	return ctx.MustSucceed() // Fail test if build failed
func (c *TestContext) MustSucceed() error {
	if c.CommandFailed() {
		return fmt.Errorf("command failed with exit code %d: %s", c.ExitCode, c.CommandOutput)
	}
	return nil
}

// formatCommandExecutionError creates a detailed error message for command execution failures.
// This helps diagnose issues in CI where we can't easily inspect the environment.
func (c *TestContext) formatCommandExecutionError(cmd *exec.Cmd, cmdLine string, err error, output string) string {
	var sb strings.Builder

	sb.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║  DIAGNOSTIC: Command execution failed                            ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("Command:     %s\n", cmdLine))
	sb.WriteString(fmt.Sprintf("Binary:      %s\n", cmd.Path))
	sb.WriteString(fmt.Sprintf("Error:       %v\n", err))
	sb.WriteString(fmt.Sprintf("GOOS:        %s\n", runtime.GOOS))
	sb.WriteString(fmt.Sprintf("GOARCH:      %s\n", runtime.GOARCH))
	sb.WriteString("\n")

	// Check if binary exists and its permissions
	if info, statErr := os.Stat(cmd.Path); statErr == nil {
		sb.WriteString("Binary exists: yes\n")
		sb.WriteString(fmt.Sprintf("Binary size:   %d bytes\n", info.Size()))
		sb.WriteString(fmt.Sprintf("Binary mode:   %s\n", info.Mode()))
	} else if os.IsNotExist(statErr) {
		sb.WriteString("Binary exists: NO - file not found\n")
	} else {
		sb.WriteString(fmt.Sprintf("Binary check:  error - %v\n", statErr))
	}
	sb.WriteString("\n")

	// Include any output that was captured
	if output != "" {
		sb.WriteString("Captured output:\n")
		sb.WriteString(output)
		sb.WriteString("\n")
	}

	return sb.String()
}

// createCommand creates an exec.Cmd for running commands.
// Uses paths.CommandsBinaryPath() to locate the pre-built binary.
// The binary must exist - tests should have @depm:eac-commands dependency.
func (c *TestContext) createCommand(parts []string) *exec.Cmd {
	binaryPath := paths.CommandsBinaryPath(c.OriginalRepoRoot)

	// Check if binary exists and log comprehensive diagnostics if it doesn't
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		c.logBinaryNotFoundDiagnostics(binaryPath)
	}

	return exec.Command(binaryPath, parts...)
}

// logBinaryNotFoundDiagnostics outputs comprehensive diagnostic information
// when the commands binary cannot be found. This helps debug CI failures.
func (c *TestContext) logBinaryNotFoundDiagnostics(binaryPath string) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "╔══════════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(os.Stderr, "║  DIAGNOSTIC: Commands binary not found                           ║\n")
	fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════════════════════════════════╝\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Environment:\n")
	fmt.Fprintf(os.Stderr, "  GOOS:     %s\n", runtime.GOOS)
	fmt.Fprintf(os.Stderr, "  GOARCH:   %s\n", runtime.GOARCH)
	fmt.Fprintf(os.Stderr, "  PWD:      %s\n", mustGetwd())
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Paths:\n")
	fmt.Fprintf(os.Stderr, "  Expected binary:    %s\n", binaryPath)
	fmt.Fprintf(os.Stderr, "  OriginalRepoRoot:   %s\n", c.OriginalRepoRoot)
	fmt.Fprintf(os.Stderr, "  IsolatedDir:        %s\n", c.IsolatedDir)
	fmt.Fprintf(os.Stderr, "  R2R_CONTAINER_ROOT: %s\n", os.Getenv("R2R_CONTAINER_ROOT"))
	fmt.Fprintf(os.Stderr, "\n")

	// Check tools directory
	toolsDir := filepath.Join(c.OriginalRepoRoot, "out", "tools")
	fmt.Fprintf(os.Stderr, "Tools directory contents (%s):\n", toolsDir)
	if entries, err := os.ReadDir(toolsDir); err == nil {
		if len(entries) == 0 {
			fmt.Fprintf(os.Stderr, "  (empty directory)\n")
		}
		for _, entry := range entries {
			info, _ := entry.Info()
			if info != nil {
				fmt.Fprintf(os.Stderr, "  - %s (%d bytes, mode: %s)\n", entry.Name(), info.Size(), info.Mode())
			} else {
				fmt.Fprintf(os.Stderr, "  - %s\n", entry.Name())
			}
		}
	} else if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "  (directory does not exist)\n")
	} else {
		fmt.Fprintf(os.Stderr, "  (error reading: %v)\n", err)
	}

	// Check build directory for eac-commands
	buildDir := filepath.Join(c.OriginalRepoRoot, "out", "build", "eac-commands")
	fmt.Fprintf(os.Stderr, "\nBuild directory contents (%s):\n", buildDir)
	if entries, err := os.ReadDir(buildDir); err == nil {
		if len(entries) == 0 {
			fmt.Fprintf(os.Stderr, "  (empty directory)\n")
		}
		for _, entry := range entries {
			info, _ := entry.Info()
			if info != nil {
				fmt.Fprintf(os.Stderr, "  - %s (%d bytes)\n", entry.Name(), info.Size())
			} else {
				fmt.Fprintf(os.Stderr, "  - %s\n", entry.Name())
			}
		}
	} else if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "  (directory does not exist)\n")
	} else {
		fmt.Fprintf(os.Stderr, "  (error reading: %v)\n", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Troubleshooting:\n")
	fmt.Fprintf(os.Stderr, "  1. Ensure eac-commands is built before running tests\n")
	fmt.Fprintf(os.Stderr, "  2. Check that @depm:eac-commands tag is on the test\n")
	fmt.Fprintf(os.Stderr, "  3. In CI, verify the build artifact was downloaded\n")
	fmt.Fprintf(os.Stderr, "\n")
}

// mustGetwd returns the current working directory or "(unknown)" on error.
func mustGetwd() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "(unknown)"
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
