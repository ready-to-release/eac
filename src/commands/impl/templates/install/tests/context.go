package tests

import (
	"fmt"
	"os"

	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// installTestContext holds state for install command tests
type installTestContext struct {
	workDir       string                       // Isolated test working directory
	testDir       string                       // Directory where tests are running from
	commandOutput string                       // Combined stdout/stderr from command
	errorOutput   string                       // Error output from command
	exitCode      int                          // Exit code from command
	isolation     *coretesting.TestIsolation   // Test isolation instance
}

var installCtx *installTestContext

// GetContext returns the current install test context
func GetContext() *installTestContext {
	return installCtx
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
		return fmt.Errorf("failed to setup install test isolation: %w", err)
	}

	installCtx = &installTestContext{
		workDir:   isolation.IsolatedDir(),
		testDir:   testDir,
		isolation: isolation,
	}

	return nil
}

// CleanupContext tears down the test context after a scenario
func CleanupContext() error {
	if installCtx != nil {
		if installCtx.isolation != nil {
			installCtx.isolation.Cleanup()
		}
		installCtx = nil
	}
	return nil
}

// Interface methods for shared context access
func (c *installTestContext) WorkDir() string {
	return c.workDir
}

func (c *installTestContext) TestDir() string {
	return c.testDir
}

func (c *installTestContext) CommandOutput() string {
	return c.commandOutput
}

func (c *installTestContext) ErrorOutput() string {
	return c.errorOutput
}

func (c *installTestContext) ExitCode() int {
	return c.exitCode
}

func (c *installTestContext) SetCommandOutput(output string) {
	c.commandOutput = output
}

func (c *installTestContext) SetErrorOutput(output string) {
	c.errorOutput = output
}

func (c *installTestContext) SetExitCode(code int) {
	c.exitCode = code
}

func (c *installTestContext) Isolation() interface{} {
	return c.isolation
}
