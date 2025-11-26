// Godog BDD step definitions for commit command - Bridge to commit module tests
//
// This file bridges the commit step definitions from the commit module
// to the cross-module test runner in src/commands/tests.
//
// The actual step implementations live in:
//   src/commands/impl/commit/tests/commit_steps.go
//
// State Synchronization Strategy:
// The commit tests package uses its own TestContext (committests.Ctx) which must
// be kept in sync with the main test context (ctx). We achieve this by:
// 1. Setting up a shared pointer: committests.Ctx points to an adapter that
//    wraps the main ctx
// 2. This avoids the need to sync before/after each step
package tests

import (
	"strings"

	"github.com/cucumber/godog"
	committests "github.com/ready-to-release/eac/src/commands/impl/commit/tests"
)

// commitTestContextAdapter wraps the main ctx to provide the committests.TestContext interface
// by directly reading/writing to the main ctx fields
type commitTestContextAdapter struct{}

// getCommitCtx returns an adapter that syncs with the main ctx
func getCommitCtx() *committests.TestContext {
	if committests.Ctx == nil {
		committests.Ctx = &committests.TestContext{}
	}
	// Sync from main ctx
	if ctx != nil {
		committests.Ctx.CommandOutput = ctx.commandOutput
		committests.Ctx.ExitCode = ctx.exitCode
		committests.Ctx.CommandError = ctx.commandError
		committests.Ctx.TestCommitMessage = ctx.testCommitMessage
		committests.Ctx.AffectedModules = ctx.affectedModules
		committests.Ctx.ValidationErrors = ctx.validationErrors
	}
	return committests.Ctx
}

// syncBackFromCommitCtx copies state back from committests.Ctx to main ctx
func syncBackFromCommitCtx() {
	// First sync from child packages (message, reset) to committests.Ctx
	committests.SyncContextFromChildren()

	// Then sync from committests.Ctx to main ctx
	if committests.Ctx != nil && ctx != nil {
		ctx.commandOutput = committests.Ctx.CommandOutput
		ctx.exitCode = committests.Ctx.ExitCode
		ctx.commandError = committests.Ctx.CommandError
		ctx.testCommitMessage = committests.Ctx.TestCommitMessage
		ctx.affectedModules = committests.Ctx.AffectedModules
		ctx.validationErrors = committests.Ctx.ValidationErrors
	}
}

// setupCommitMocks bridges to the commit module's mock setup
func setupCommitMocks() error {
	// Initialize shared state
	getCommitCtx()
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
	syncBackFromCommitCtx()
	committests.CleanupCommitMocks()
	committests.CleanupResetMocks()
}

// commitStepPatterns contains regex patterns that identify commit-specific steps
// We only sync state for these steps to avoid interfering with other commands
// IMPORTANT: Patterns must be specific enough to NOT match generic verification steps
// like "I should see X" - otherwise syncing will overwrite the real command output
var commitStepPatterns = []string{
	"commit message", "Auditor-Summary", "HEADER_TOO_LONG", "HEADER_TRAILING_PERIOD",
	"MISSING_AUDITOR_SUMMARY", "module sections", "noise filtering", "auto-cleanup",
	"AI output", "code fence", "blank lines", "git diff", "staged files",
	"contract", "validation", "module with", "affected module", "error should occur",
	"message is validated", "closing fence", "period should be", "wrapped at word",
	"output should start with", "body section", "semantic types",
	// Reset command patterns - use specific phrases to avoid matching generic "I should see" steps
	"latest commit should be", "commit reset directly", "soft reset", "in the staging area",
	"in the working directory", "uncommitted changes should be", "detached HEAD state",
	"commits to reset", "initial commit", "reset should succeed",
}

// isCommitStep returns true if the step text matches a commit-specific pattern
func isCommitStep(stepText string) bool {
	for _, pattern := range commitStepPatterns {
		if strings.Contains(strings.ToLower(stepText), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// InitializeCommitScenario registers commit-specific step definitions
func InitializeCommitScenario(sc *godog.ScenarioContext) {
	// Set up shared state
	committests.OriginalRepoRoot = originalRepoRoot
	committests.IsolatedTestProjectDir = isolatedTestProjectDir
	committests.RunCommand = iRunTheCommand

	// Use BeforeStep/AfterStep hooks to sync state (using old-style API)
	// Only sync for commit-specific steps to avoid interfering with other tests
	sc.BeforeStep(func(st *godog.Step) {
		if isCommitStep(st.Text) {
			getCommitCtx()
		}
	})
	sc.AfterStep(func(st *godog.Step, err error) {
		if isCommitStep(st.Text) {
			syncBackFromCommitCtx()
		}
	})

	// Register the commit module's steps
	committests.InitializeCommitScenario(sc)
}
