// Package tests provides shared test context and step registration for commit subcommands.
//
// This file delegates step registration to subcommand-specific packages.
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
	// Sync context to child packages before scenario runs
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		SyncContextToChildren()
		return ctx, nil
	})

	// Register message subcommand steps
	message.InitializeScenario(sc)

	// Register reset subcommand steps
	reset.InitializeScenario(sc)
}

// SyncContextToChildren synchronizes the shared context to child packages.
// This must be called before each scenario to ensure child packages have
// access to the test context.
func SyncContextToChildren() {
	// Sync to message package
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

	// Sync to reset package
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
// This should be called after each scenario to capture any state changes.
func SyncContextFromChildren() {
	if Ctx == nil {
		return
	}

	// Sync from message package
	if message.Ctx != nil {
		Ctx.CommandOutput = message.Ctx.CommandOutput
		Ctx.ExitCode = message.Ctx.ExitCode
		Ctx.CommandError = message.Ctx.CommandError
		Ctx.TestCommitMessage = message.Ctx.TestCommitMessage
		Ctx.AffectedModules = message.Ctx.AffectedModules
		Ctx.ValidationErrors = message.Ctx.ValidationErrors
	}
	TestAIOutput = message.TestAIOutput

	// Sync from reset package - only if reset context was actually used
	// This prevents reset's default zero values from overwriting message context
	if reset.Ctx != nil && reset.Ctx.WasUsed {
		Ctx.CommandOutput = reset.Ctx.CommandOutput
		Ctx.ExitCode = reset.Ctx.ExitCode
		Ctx.CommandError = reset.Ctx.CommandError
		Ctx.ResetTestFile = reset.Ctx.ResetTestFile
	}
}

// ============================================================================
// Exported Step Functions (for cross-package state synchronization)
// ============================================================================
// These wrappers allow the main test runner to call into message package steps
// when needed for backward compatibility.

// SetupCommitMocks delegates to message package for backward compatibility.
func SetupCommitMocks() error {
	SyncContextToChildren()
	return message.SetupMocks()
}

// CleanupCommitMocks delegates to message package for backward compatibility.
func CleanupCommitMocks() {
	message.CleanupMocks()
	SyncContextFromChildren()
}

// SetupResetMocks delegates to reset package.
func SetupResetMocks() error {
	SyncContextToChildren()
	return reset.SetupMocks()
}

// CleanupResetMocks delegates to reset package.
func CleanupResetMocks() {
	reset.CleanupMocks()
	SyncContextFromChildren()
}
