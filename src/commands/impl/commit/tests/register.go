// Package tests provides shared test context and step registration for commit subcommands.
//
// This file delegates step registration to subcommand-specific packages.
// Context is shared via direct pointer references to avoid complex bidirectional synchronization.
package tests

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/ready-to-release/eac/src/commands/impl/commit/tests/message"
	"github.com/ready-to-release/eac/src/commands/impl/commit/tests/reset"
)

// InitializeCommitScenario registers all commit-related step definitions.
// It delegates to subcommand-specific packages for their respective steps.
func InitializeCommitScenario(sc *godog.ScenarioContext) {
	// Initialize child package contexts before scenario runs
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		initializeChildContexts()
		return ctx, nil
	})

	// Register message subcommand steps
	message.InitializeScenario(sc)

	// Register reset subcommand steps
	reset.InitializeScenario(sc)
}

// initializeChildContexts sets up child package contexts.
// Child packages reference these contexts directly, so changes are automatically visible.
func initializeChildContexts() {
	// Initialize message context - shares same underlying data
	if Ctx != nil {
		message.Ctx = &message.Context{
			CommandOutput:     Ctx.CommandOutput,
			ExitCode:          Ctx.ExitCode,
			CommandError:      Ctx.CommandError,
			TestCommitMessage: Ctx.TestCommitMessage,
			AffectedModules:   Ctx.AffectedModules,
			ValidationErrors:  Ctx.ValidationErrors,
		}
	}
	message.IsolatedTestProjectDir = IsolatedTestProjectDir
	message.OriginalRepoRoot = OriginalRepoRoot
	message.RunCommand = RunCommand
	message.TestAIOutput = TestAIOutput

	// Initialize reset context - shares same underlying data
	if Ctx != nil {
		reset.Ctx = &reset.Context{
			CommandOutput: Ctx.CommandOutput,
			ExitCode:      Ctx.ExitCode,
			CommandError:  Ctx.CommandError,
			ResetTestFile: Ctx.ResetTestFile,
		}
	}
	reset.IsolatedTestProjectDir = IsolatedTestProjectDir
	reset.OriginalRepoRoot = OriginalRepoRoot
	reset.RunCommand = RunCommand
}

// SyncContextFromChildren synchronizes changes back from child packages.
// Called after scenarios to capture state changes made by child step definitions.
func SyncContextFromChildren() {
	if Ctx == nil {
		return
	}

	// Sync from message package - message is the primary context for most scenarios
	if message.Ctx != nil {
		Ctx.CommandOutput = message.Ctx.CommandOutput
		Ctx.ExitCode = message.Ctx.ExitCode
		Ctx.CommandError = message.Ctx.CommandError
		Ctx.TestCommitMessage = message.Ctx.TestCommitMessage
		Ctx.AffectedModules = message.Ctx.AffectedModules
		Ctx.ValidationErrors = message.Ctx.ValidationErrors
	}
	TestAIOutput = message.TestAIOutput

	// Sync from reset package only if it was actually used
	if reset.Ctx != nil && reset.Ctx.WasUsed {
		Ctx.CommandOutput = reset.Ctx.CommandOutput
		Ctx.ExitCode = reset.Ctx.ExitCode
		Ctx.CommandError = reset.Ctx.CommandError
		Ctx.ResetTestFile = reset.Ctx.ResetTestFile
	}
}

// ============================================================================
// Exported Mock Setup Functions
// ============================================================================
// These functions delegate to subcommand packages and handle context sync.

// SetupCommitMocks initializes mocks for commit message tests.
func SetupCommitMocks() error {
	initializeChildContexts()
	return message.SetupMocks()
}

// CleanupCommitMocks cleans up mocks after commit message tests.
func CleanupCommitMocks() {
	message.CleanupMocks()
	SyncContextFromChildren()
}

// SetupResetMocks initializes mocks for reset tests.
func SetupResetMocks() error {
	initializeChildContexts()
	return reset.SetupMocks()
}

// CleanupResetMocks cleans up mocks after reset tests.
func CleanupResetMocks() {
	reset.CleanupMocks()
	SyncContextFromChildren()
}

// SyncContextToChildren is an exported alias for initializeChildContexts.
// This maintains backward compatibility with external test bridges.
func SyncContextToChildren() {
	initializeChildContexts()
}
