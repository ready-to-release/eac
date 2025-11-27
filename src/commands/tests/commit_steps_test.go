// Godog BDD step definitions for commit command - Bridge to commit module tests
//
// This file bridges the commit step definitions from the commit module
// to the cross-module test runner in src/commands/tests.
//
// The actual step implementations live in:
//   src/commands/impl/commit/tests/
//
// State Synchronization Strategy:
// We use a shared context pointer (sharedCtx) that is passed to child packages.
// Changes made by child packages are immediately visible to the main test runner
// because they all reference the same context object. This eliminates the need
// for BeforeStep/AfterStep synchronization.
package tests

import (
	"github.com/cucumber/godog"
	committests "github.com/ready-to-release/eac/src/commands/impl/commit/tests"
)

// syncToCommitCtx copies state from main ctx to committests for backward compatibility.
// With the new shared context, this is mainly used to initialize committests.Ctx
// for legacy code paths that still use it.
func syncToCommitCtx() {
	if committests.Ctx == nil {
		committests.Ctx = &committests.TestContext{}
	}
	if ctx != nil {
		committests.Ctx.CommandOutput = ctx.commandOutput
		committests.Ctx.ExitCode = ctx.exitCode
		committests.Ctx.CommandError = ctx.commandError
		committests.Ctx.TestCommitMessage = ctx.testCommitMessage
		committests.Ctx.AffectedModules = ctx.affectedModules
		committests.Ctx.ValidationErrors = ctx.validationErrors
	}
	// Pass shared context to commit tests
	committests.SharedCtx = sharedCtx
	committests.SyncContextToChildren()
}

// syncFromCommitCtx copies state back from child packages to main ctx.
// We sync ctx TO sharedCtx first (to capture any changes from non-commit steps),
// then sync from children to pick up commit-specific changes.
func syncFromCommitCtx() {
	// First sync ctx TO sharedCtx to preserve any changes from main test runner
	syncCtxToShared()

	// Then sync from child packages
	committests.SyncContextFromChildren()

	// Finally sync back from sharedCtx to ctx if children made changes
	if sharedCtx != nil && ctx != nil && sharedCtx.HasOutput() {
		ctx.commandOutput = sharedCtx.CommandOutput
		ctx.exitCode = sharedCtx.ExitCode
		ctx.commandError = sharedCtx.CommandError
		ctx.testCommitMessage = sharedCtx.TestCommitMessage
		ctx.affectedModules = sharedCtx.AffectedModules
		ctx.validationErrors = sharedCtx.ValidationErrors
	}
}

// setupCommitMocks bridges to the commit module's mock setup
func setupCommitMocks() error {
	// Initialize shared state
	syncToCommitCtx()
	committests.OriginalRepoRoot = originalRepoRoot
	committests.IsolatedTestProjectDir = isolatedTestProjectDir
	committests.RunCommand = iRunTheCommand
	committests.SharedCtx = sharedCtx
	if err := committests.SetupCommitMocks(); err != nil {
		return err
	}
	// Also setup reset mocks
	return committests.SetupResetMocks()
}

// cleanupCommitMocks bridges to the commit module's mock cleanup
func cleanupCommitMocks() {
	syncFromCommitCtx()
	committests.CleanupCommitMocks()
	committests.CleanupResetMocks()
}

// InitializeCommitScenario registers commit-specific step definitions
func InitializeCommitScenario(sc *godog.ScenarioContext) {
	// Set up shared state - pass the shared context pointer
	committests.OriginalRepoRoot = originalRepoRoot
	committests.IsolatedTestProjectDir = isolatedTestProjectDir
	committests.RunCommand = iRunTheCommand
	committests.SharedCtx = sharedCtx

	// Sync context before and after every step for backward compatibility
	// With shared context, changes are automatically visible, but we sync
	// to keep the old ctx fields up to date for legacy step definitions
	sc.BeforeStep(func(st *godog.Step) {
		syncToCommitCtx()
	})
	sc.AfterStep(func(st *godog.Step, err error) {
		syncFromCommitCtx()
	})

	// Register the commit module's steps
	committests.InitializeCommitScenario(sc)
}
