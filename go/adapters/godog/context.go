package godog

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
	coretesting "github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
)

// TestContext wraps the core SharedTestContext with additional spec-specific state.
type TestContext struct {
	*coretesting.SharedTestContext

	// OriginalRepoRoot is the actual repository root (for running go commands)
	OriginalRepoRoot string

	// AssetsPath is the path to the assets directory relative to the repository root.
	// Example: "go/cli/eac/specs/assets" for eac.
	AssetsPath string

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
	// Keys are env var names (e.g., "CLIE_MOCK_AI_SPECS"), values are mock file names.
	MockOverrides map[string]string

	// OriginalRepoCache provides cached data from the ORIGINAL repository root.
	// This is NOT for isolated/mocked test repositories - only the real repo.
	// This dramatically improves performance by avoiding repeated git/file operations.
	OriginalRepoCache *TestCache

	// CommandDispatcher, when set, routes command execution in-process instead of
	// spawning a subprocess. The function receives parsed command args (e.g., ["show", "test-results"])
	// and returns (combined output, exit code). This dramatically speeds up BDD tests.
	CommandDispatcher func(args []string) (string, int)
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

	// Reset mock Docker client state to prevent state leakage between scenarios
	// Delete the state file directly to avoid dependency on commands package
	deleteMockDockerStateFile()
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
// The key should be the env var name (e.g., "CLIE_MOCK_AI_SPECS"),
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
	if err := os.MkdirAll(specsDir, 0o750); err != nil {
		return fmt.Errorf("failed to create specs directory in isolation: %w", err)
	}

	// Ensure .git directory exists and is valid in isolated directory
	// This is critical for repository.GetRepositoryRoot() to find the correct root
	gitDir := filepath.Join(c.IsolatedDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf(".git directory not found in isolated directory: %s", c.IsolatedDir)
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
// For isolated tests, filters out CLIE_TEST_LOGGING_ACTIVE so unified log works.
func (c *TestContext) buildBaseEnvironment() []string {
	env := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		if c.IsolatedDir != "" && strings.HasPrefix(e, environments.EnvCLIETestLoggingActive+"=") {
			continue // Don't inherit - isolated tests need unified log to work
		}
		env = append(env, e)
	}
	return env
}

// buildIsolationEnvironment adds isolation-specific environment variables.
// Sets CLIE_PWD, CLIE_REPO_ROOT, and CLIE_TEST_SCOPE for isolated test execution.
func (c *TestContext) buildIsolationEnvironment(env []string) []string {
	if c.IsolatedDir == "" {
		return env
	}

	// CLIE_PWD: current working directory (may be worktree)
	// CLIE_REPO_ROOT: main repository root (never changes)
	// CLIE_TEST_SCOPE: marker indicating we're inside a spec test
	workDir := c.getEffectiveWorkDir()
	env = append(env, fmt.Sprintf("%s=%s", environments.EnvCLIEPWD, workDir))
	env = append(env, fmt.Sprintf("%s=%s", environments.EnvCLIERepoRoot, c.IsolatedDir))
	env = append(env, fmt.Sprintf("%s=1", environments.EnvCLIETestScope))
	return env
}

// loadMockConfig loads the mock configuration from .eac/testing-mocks.yml.
// Returns an empty config (not error) if the file is missing, for backward compatibility.
func (c *TestContext) loadMockConfig() config.TestingMocksConfig {
	mockConfigPath := filepath.Join(c.OriginalRepoRoot, ".eac", "testing-mocks.yml")
	mockConfig, err := config.LoadTestingMocks(mockConfigPath)
	if err != nil {
		log.Warnf("Failed to load mock config from %s: %v - falling back to environment variables", mockConfigPath, err)
		return config.TestingMocksConfig{}
	}
	return mockConfig
}

// resolveMockEnvVars converts mock configuration to environment variables,
// resolving relative paths to absolute paths based on the repo root.
func (c *TestContext) resolveMockEnvVars(mockConfig config.TestingMocksConfig) []string {
	mockEnvVars := mockConfig.ToEnvironmentVariables()

	for i, e := range mockEnvVars {
		if strings.HasPrefix(e, environments.EnvCLIEMockAIDir+"=") {
			mockDir := strings.TrimPrefix(e, environments.EnvCLIEMockAIDir+"=")
			if mockDir != "" && !filepath.IsAbs(mockDir) {
				mockDir = filepath.Join(c.OriginalRepoRoot, mockDir)
			}
			mockEnvVars[i] = fmt.Sprintf("%s=%s", environments.EnvCLIEMockAIDir, mockDir)
		}
	}

	return mockEnvVars
}

// resolveAIDir returns the mock AI directory env var, or empty string if not needed.
// This provides backward compatibility when the config file does not set the AI mock directory.
func (c *TestContext) resolveAIDir(mockConfig config.TestingMocksConfig, existingVars []string) string {
	for _, e := range existingVars {
		if strings.HasPrefix(e, environments.EnvCLIEMockAIDir+"=") {
			return "" // Already set by config
		}
	}

	if !mockConfig.Mocks.AI.Enabled || c.AssetsPath == "" {
		return ""
	}

	assetsRoot := repository.GetDistRoot(c.OriginalRepoRoot)
	if assetsRoot == "" {
		return ""
	}
	assetsDir := filepath.Join(assetsRoot, c.AssetsPath)
	if _, err := os.Stat(assetsDir); err != nil {
		return ""
	}
	return fmt.Sprintf("%s=%s", environments.EnvCLIEMockAIDir, assetsDir)
}

// buildMockingEnvironment adds mocking environment variables for tests.
// Sets CLIE_CONTAINER_ROOT, CLIE_MOCK_AI_DIR, CLIE_MOCK_SECURITY, and CLIE_MOCK_STRUCTURIZR.
func (c *TestContext) buildMockingEnvironment(env []string) []string {
	env = append(env, fmt.Sprintf("%s=%s", environments.EnvCLIEContainerRoot, c.OriginalRepoRoot))

	mockConfig := c.loadMockConfig()
	mockEnvVars := c.resolveMockEnvVars(mockConfig)
	env = append(env, mockEnvVars...)

	if aiDir := c.resolveAIDir(mockConfig, mockEnvVars); aiDir != "" {
		env = append(env, aiDir)
	}

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

	// Use in-process dispatch if available (avoids subprocess overhead)
	if c.CommandDispatcher != nil {
		output, exitCode := c.CommandDispatcher(parts)
		c.CommandOutput = output
		c.ExitCode = exitCode
		c.CommandError = nil
		if exitCode != 0 {
			c.CommandError = fmt.Errorf("exit code %d", exitCode)
		}
		return nil
	}

	// Try to use pre-built binary for better performance
	cmd := c.createCommand(parts)

	// Set working directory for isolated tests
	// CRITICAL: Always set cmd.Dir for isolated tests to ensure subprocess runs in correct directory
	if c.IsolatedDir != "" {
		cmd.Dir = c.getEffectiveWorkDir()
		// Verify that the directory exists before running command
		if _, err := os.Stat(cmd.Dir); os.IsNotExist(err) {
			return fmt.Errorf("isolated directory does not exist: %s", cmd.Dir)
		}
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
//	err := ctx.RunCommand("build core")
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
// The binary must exist - tests should have @depm:eac dependency.
// The command is placed in its own process group so it can be killed cleanly on timeout.
func (c *TestContext) createCommand(parts []string) *exec.Cmd {
	binaryPath := paths.CommandsBinaryPath(c.OriginalRepoRoot)

	// Check if binary exists and log comprehensive diagnostics if it doesn't
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		c.logBinaryNotFoundDiagnostics(binaryPath)
	}

	cmd := exec.Command(binaryPath, parts...)
	tool.SetProcessGroup(cmd)
	return cmd
}

// logBinaryNotFoundDiagnostics outputs comprehensive diagnostic information
// when the commands binary cannot be found. This helps debug CI failures.
// Uses the structured logger so output is captured by CI log aggregation.
func (c *TestContext) logBinaryNotFoundDiagnostics(binaryPath string) {
	log.Warnf("Commands binary not found: %s", binaryPath)
	log.Warnf("Environment: GOOS=%s GOARCH=%s PWD=%s", runtime.GOOS, runtime.GOARCH, mustGetwd())
	log.Warnf("Paths: expected=%s repoRoot=%s isolatedDir=%s containerRoot=%s",
		binaryPath, c.OriginalRepoRoot, c.IsolatedDir, os.Getenv(environments.EnvCLIEContainerRoot))

	// Check tools directory
	toolsDir := filepath.Join(c.OriginalRepoRoot, "out", "tools")
	if entries, err := os.ReadDir(toolsDir); err == nil {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		log.Warnf("Tools directory (%s): %v", toolsDir, names)
	} else if os.IsNotExist(err) {
		log.Warnf("Tools directory does not exist: %s", toolsDir)
	} else {
		log.Warnf("Tools directory error: %v", err)
	}

	// Check build directory for eac
	buildDir := filepath.Join(c.OriginalRepoRoot, "out", "build", "eac")
	if entries, err := os.ReadDir(buildDir); err == nil {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		log.Warnf("Build directory (%s): %v", buildDir, names)
	} else if os.IsNotExist(err) {
		log.Warnf("Build directory does not exist: %s", buildDir)
	} else {
		log.Warnf("Build directory error: %v", err)
	}

	log.Warnf("Troubleshooting: 1) Ensure eac is built before running tests 2) Check @depm:eac tag 3) In CI, verify build artifact was downloaded")
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

// deleteMockDockerStateFile removes the mock Docker state file.
// This is called between scenarios to prevent state leakage.
// The file path matches what the mock Docker client uses (in commands/internal/serve).
func deleteMockDockerStateFile() {
	stateFile := filepath.Join(os.TempDir(), "eac-mock-docker-state.json")
	_ = os.Remove(stateFile) // Ignore errors - file may not exist
}
