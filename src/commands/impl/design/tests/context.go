// Package tests provides BDD test context for the design command.
//
// This file contains the shared test context that tracks state across
// test scenarios, including command execution results and test artifacts.
package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/commands/impl/design/create"
	"github.com/ready-to-release/eac/src/commands/impl/design/update"
	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// DesignTestContext holds state for design command tests.
type DesignTestContext struct {
	// Test isolation
	Isolation *coretesting.TestIsolation

	// Command execution state
	LastOutput   string
	LastExitCode int

	// Design-specific fields
	generatedWorkspace string            // Last generated workspace content
	validationResult   *ValidationResult // Last validation result
	servePort          int               // Port allocated for serve command
	containerID        string            // Docker container ID for serve
}

// ValidationResult captures workspace validation output
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// sharedContext is the singleton instance shared across all scenarios
var (
	sharedContext *DesignTestContext
	contextOnce   sync.Once
)

// GetContext returns the shared design test context.
// Creates it if it doesn't exist (singleton pattern).
func GetContext() *DesignTestContext {
	contextOnce.Do(func() {
		sharedContext = &DesignTestContext{}
	})
	return sharedContext
}

// registerContext registers hooks to manage test context lifecycle.
func registerContext(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		beforeScenario()
		return ctx, nil
	})
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		afterScenario()
		return ctx, nil
	})
}

// beforeScenario sets up test context before each scenario.
func beforeScenario() {
	ctx := GetContext()

	// Initialize test isolation with temporary directory
	// This prevents tests from creating files in the real specs/ directory
	ctx.Isolation = coretesting.NewTestIsolation()
	if err := ctx.Isolation.Setup(); err != nil {
		// Log error but don't fail - some tests may not need isolation
		fmt.Printf("Warning: Failed to setup test isolation: %v\n", err)
	}

	// Reset design-specific state
	ctx.generatedWorkspace = ""
	ctx.validationResult = nil
	ctx.servePort = 0
	ctx.containerID = ""

	// Reset mocks
	create.ResetMockAIResponse()
	create.ResetGitRepo()
	update.ResetMockAIResponse()
	update.ResetGitRepo()
}

// afterScenario cleans up test context after each scenario.
func afterScenario() {
	ctx := GetContext()

	// Clean up Docker container if running
	if ctx.containerID != "" {
		// TODO: Stop and remove container
		ctx.containerID = ""
	}

	// Clean up test isolation (removes temporary directory)
	if ctx.Isolation != nil {
		ctx.Isolation.Cleanup()
		ctx.Isolation = nil
	}

	// Reset mocks
	create.ResetMockAIResponse()
	create.ResetGitRepo()
	update.ResetMockAIResponse()
	update.ResetGitRepo()
}

// CaptureOutput runs a function while capturing its stdout/stderr.
// Returns captured output and any error from the function.
func (ctx *DesignTestContext) CaptureOutput(fn func() error) (string, error) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	err := fn()

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	buf.ReadFrom(r)
	return buf.String(), err
}

// GetWorkspacePath returns the path to a module's workspace file.
func (ctx *DesignTestContext) GetWorkspacePath(module string) string {
	// Get repository root from shared context or use current working directory
	repoRoot := ctx.GetRepositoryRoot()
	if repoRoot == "" {
		cwd, _ := os.Getwd()
		repoRoot = cwd
	}
	return filepath.Join(repoRoot, "specs", module, "design", "workspace.dsl")
}

// GetRepositoryRoot returns the repository root directory.
func (ctx *DesignTestContext) GetRepositoryRoot() string {
	if ctx.Isolation != nil && ctx.Isolation.IsolatedDir() != "" {
		return ctx.Isolation.IsolatedDir()
	}
	// Fallback: try to find it
	cwd, _ := os.Getwd()
	return cwd
}

// RunCommand executes a command with arguments and captures output.
func (ctx *DesignTestContext) RunCommand(args []string) error {
	// Build command with "go run ." prefix
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)

	// Set working directory and environment
	if ctx.Isolation != nil {
		cmd.Dir = ctx.GetRepositoryRoot()
		cmd.Env = ctx.Isolation.AppendToEnvironment(os.Environ())
	}

	// Capture output
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Run command
	err := cmd.Run()

	// Store results
	ctx.LastOutput = outBuf.String() + errBuf.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ctx.LastExitCode = exitErr.ExitCode()
		} else {
			ctx.LastExitCode = 1
		}
	} else {
		ctx.LastExitCode = 0
	}

	return nil // Don't return error, tests check LastExitCode
}

// WorkspaceExists checks if a workspace file exists for a module.
func (ctx *DesignTestContext) WorkspaceExists(module string) bool {
	path := ctx.GetWorkspacePath(module)
	_, err := os.Stat(path)
	return err == nil
}

// ReadWorkspace reads the workspace content for a module.
func (ctx *DesignTestContext) ReadWorkspace(module string) (string, error) {
	path := ctx.GetWorkspacePath(module)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteWorkspace writes workspace content for a module.
func (ctx *DesignTestContext) WriteWorkspace(module string, content string) error {
	path := ctx.GetWorkspacePath(module)
	dir := filepath.Dir(path)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write workspace: %w", err)
	}

	return nil
}

// SetMockAIResponseForCreate sets a mock AI response for create command tests.
func (ctx *DesignTestContext) SetMockAIResponseForCreate(response string) {
	create.SetMockAIResponse(response)
}

// SetMockAIResponseForUpdate sets a mock AI response for update command tests.
func (ctx *DesignTestContext) SetMockAIResponseForUpdate(response string) {
	update.SetMockAIResponse(response)
}
