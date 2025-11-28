package tests

import (
	"os"

	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// listTestContext holds state for templates list tests
type listTestContext struct {
	workDir       string
	testDir       string
	commandOutput string
	errorOutput   string
	exitCode      int
	isolation     *coretesting.TestIsolation
}

var listCtx *listTestContext

// initializeListContext sets up test context for list scenarios
func initializeListContext() error {
	// Get the test directory before creating temp directory
	testDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Use TestIsolation for consistent test isolation
	isolation := coretesting.NewTestIsolation()
	if err := isolation.Setup(); err != nil {
		return err
	}

	listCtx = &listTestContext{
		workDir:   isolation.IsolatedDir(),
		testDir:   testDir,
		isolation: isolation,
	}
	return nil
}

// cleanupListContext cleans up test context
func cleanupListContext() error {
	if listCtx != nil {
		if listCtx.isolation != nil {
			listCtx.isolation.Cleanup()
		}
	}
	listCtx = nil
	return nil
}
