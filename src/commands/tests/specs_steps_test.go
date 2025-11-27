// Godog BDD step definitions for specs commands - Bridge to specs module tests
//
// This file bridges the specs step definitions from the specs module
// to the cross-module test runner in src/commands/tests.
//
// The actual step implementations live in:
//   src/commands/impl/specs/tests/
//
// State Synchronization Strategy:
// We use a shared context pointer (sharedCtx) that is passed to child packages.
// Changes made by child packages are immediately visible to the main test runner.
package tests

import (
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
	specstests "github.com/ready-to-release/eac/src/commands/impl/specs/tests"
)

// syncToSpecsCtx copies state from main ctx to specstests.Ctx for backward compatibility.
func syncToSpecsCtx() {
	if specstests.Ctx == nil {
		specstests.Ctx = &specstests.TestContext{}
	}
	if ctx != nil {
		specstests.Ctx.CommandOutput = ctx.commandOutput
		specstests.Ctx.ExitCode = ctx.exitCode
		specstests.Ctx.CommandError = ctx.commandError
		specstests.Ctx.ValidationErrors = ctx.validationErrors
	}
	// Pass shared context to specs tests
	specstests.SharedCtx = sharedCtx
}

// syncFromSpecsCtx copies state back from child packages to main ctx.
// We sync ctx TO sharedCtx first (to capture any changes from non-specs steps),
// then sync back if children made changes.
func syncFromSpecsCtx() {
	// First sync ctx TO sharedCtx to preserve any changes from main test runner
	syncCtxToShared()

	// Sync back from sharedCtx to ctx only if there are actual changes
	if sharedCtx != nil && ctx != nil && sharedCtx.HasOutput() {
		ctx.commandOutput = sharedCtx.CommandOutput
		ctx.exitCode = sharedCtx.ExitCode
		ctx.commandError = sharedCtx.CommandError
		ctx.validationErrors = sharedCtx.ValidationErrors
	}
}

// setupSpecsMocks bridges to the specs module's mock setup
// Called automatically when @env:isolated-test-project tag is present
func setupSpecsMocks() error {
	// Initialize shared state
	syncToSpecsCtx()
	specstests.OriginalRepoRoot = originalRepoRoot
	specstests.IsolatedTestProjectDir = isolatedTestProjectDir
	specstests.RunCommand = iRunTheCommand
	specstests.SharedCtx = sharedCtx

	// Set up test directory for file creation
	if specstests.Ctx != nil {
		specstests.Ctx.TestDir = isolatedTestProjectDir
		if specstests.Ctx.TestDir == "" {
			// Fallback to original repo root
			specstests.Ctx.TestDir = originalRepoRoot
		}
	}

	return specstests.SetupSpecsMocks()
}

// cleanupSpecsMocks bridges to the specs module's mock cleanup
func cleanupSpecsMocks() {
	syncFromSpecsCtx()
	specstests.CleanupSpecsMocks()

	// Clean up any test files created (using shared context)
	if sharedCtx != nil {
		for _, file := range sharedCtx.CreatedFiles {
			os.Remove(file)
			dir := filepath.Dir(file)
			os.Remove(dir) // Remove empty directories
		}
		sharedCtx.CreatedFiles = nil
	}
}

// InitializeSpecsScenario registers specs-specific step definitions
func InitializeSpecsScenario(sc *godog.ScenarioContext) {
	// Set up shared state - pass the shared context pointer
	specstests.OriginalRepoRoot = originalRepoRoot
	specstests.IsolatedTestProjectDir = isolatedTestProjectDir
	specstests.RunCommand = iRunTheCommand
	specstests.SharedCtx = sharedCtx

	// Sync context before and after every step for backward compatibility
	sc.BeforeStep(func(st *godog.Step) {
		syncToSpecsCtx()
	})
	sc.AfterStep(func(st *godog.Step, err error) {
		syncFromSpecsCtx()
	})

	// Register the specs module's steps
	specstests.InitializeSpecsScenario(sc)
}
