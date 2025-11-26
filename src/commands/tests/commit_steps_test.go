// Godog BDD step definitions for commit command - Bridge to commit module tests
//
// This file bridges the commit step definitions from the commit module
// to the cross-module test runner in src/commands/tests.
//
// The actual step implementations live in:
//   src/commands/impl/commit/tests/
//
// State Synchronization Strategy:
// We sync context at the start of every step to ensure commit subpackages
// have access to the latest state, and sync back after every step to
// capture any changes they made. This is simpler and more reliable than
// trying to detect "commit-specific" steps via pattern matching.
package tests

import (
	"github.com/cucumber/godog"
	committests "github.com/ready-to-release/eac/src/commands/impl/commit/tests"
)

// syncToCommitCtx copies state from main ctx to committests.Ctx
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
	// Also sync to child packages
	committests.SyncContextToChildren()
}

// syncFromCommitCtx copies state back from committests.Ctx to main ctx
// Only syncs if commit tests actually modified something (non-empty output or non-zero exit)
func syncFromCommitCtx() {
	// First sync from child packages (message, reset) to committests.Ctx
	committests.SyncContextFromChildren()

	// Only sync back if commit context has meaningful data
	// This prevents overwriting main context with empty commit context values
	// when running non-commit tests
	if committests.Ctx != nil && ctx != nil {
		// Only sync if commit tests actually produced output or set state
		hasCommitOutput := committests.Ctx.CommandOutput != "" ||
			committests.Ctx.ExitCode != 0 ||
			committests.Ctx.TestCommitMessage != "" ||
			len(committests.Ctx.AffectedModules) > 0 ||
			len(committests.Ctx.ValidationErrors) > 0

		if hasCommitOutput {
			ctx.commandOutput = committests.Ctx.CommandOutput
			ctx.exitCode = committests.Ctx.ExitCode
			ctx.commandError = committests.Ctx.CommandError
			ctx.testCommitMessage = committests.Ctx.TestCommitMessage
			ctx.affectedModules = committests.Ctx.AffectedModules
			ctx.validationErrors = committests.Ctx.ValidationErrors
		}
	}
}

// setupCommitMocks bridges to the commit module's mock setup
func setupCommitMocks() error {
	// Initialize shared state
	syncToCommitCtx()
	committests.OriginalRepoRoot = originalRepoRoot
	committests.IsolatedTestProjectDir = isolatedTestProjectDir
	committests.RunCommand = iRunTheCommand
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
	// Set up shared state
	committests.OriginalRepoRoot = originalRepoRoot
	committests.IsolatedTestProjectDir = isolatedTestProjectDir
	committests.RunCommand = iRunTheCommand

	// Sync context before and after every step
	// This ensures commit subpackages always have access to the latest state
	// and any changes they make are propagated back to the main context
	sc.BeforeStep(func(st *godog.Step) {
		syncToCommitCtx()
	})
	sc.AfterStep(func(st *godog.Step, err error) {
		syncFromCommitCtx()
	})

	// Register the commit module's steps
	committests.InitializeCommitScenario(sc)
}
