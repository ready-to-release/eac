package tests

import (
	"fmt"
	"os"

	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// applyTestContext holds state for apply command tests
type applyTestContext struct {
	workDir       string                       // Isolated test working directory
	testDir       string                       // Directory where tests are running from
	commandOutput string                       // Combined stdout/stderr from command
	errorOutput   string                       // Error output from command
	exitCode      int                          // Exit code from command
	isolation     *coretesting.TestIsolation   // Test isolation instance
}

var applyCtx *applyTestContext

// GetContext returns the current apply test context
func GetContext() *applyTestContext {
	return applyCtx
}

// InitializeContext sets up the test context for a scenario
func InitializeContext() error {
	// Get the test directory before creating temp directory
	testDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Use TestIsolation for consistent test isolation
	isolation := coretesting.NewTestIsolation()
	if err := isolation.Setup(); err != nil {
		return fmt.Errorf("failed to setup apply test isolation: %w", err)
	}

	applyCtx = &applyTestContext{
		workDir:   isolation.IsolatedDir(),
		testDir:   testDir,
		isolation: isolation,
	}

	return nil
}

// CleanupContext tears down the test context after a scenario
func CleanupContext() error {
	if applyCtx != nil {
		if applyCtx.isolation != nil {
			applyCtx.isolation.Cleanup()
		}
		applyCtx = nil
	}
	return nil
}

// Interface methods for shared context access
func (c *applyTestContext) WorkDir() string {
	return c.workDir
}

func (c *applyTestContext) TestDir() string {
	return c.testDir
}

func (c *applyTestContext) CommandOutput() string {
	return c.commandOutput
}

func (c *applyTestContext) ErrorOutput() string {
	return c.errorOutput
}

func (c *applyTestContext) ExitCode() int {
	return c.exitCode
}

func (c *applyTestContext) SetCommandOutput(output string) {
	c.commandOutput = output
}

func (c *applyTestContext) SetErrorOutput(output string) {
	c.errorOutput = output
}

func (c *applyTestContext) SetExitCode(code int) {
	c.exitCode = code
}

func (c *applyTestContext) Isolation() interface{} {
	return c.isolation
}
