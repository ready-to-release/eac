// Godog BDD step definitions for docs command - Bridge to docs module tests
//
// This file bridges the docs step definitions from the docs module
// to the cross-module test runner in src/commands/tests.
//
// The actual step implementations live in:
//   src/commands/impl/docs/tests/
//
// State Synchronization Strategy:
// We use a shared context pointer (sharedCtx) that is passed to child packages.
// Changes made by child packages are immediately visible to the main test runner
// because they all reference the same context object.
package tests

import (
	"github.com/cucumber/godog"
	docstests "github.com/ready-to-release/eac/src/commands/impl/docs/tests"
)

// syncToDocsCtx copies state from main ctx to docstests for backward compatibility.
func syncToDocsCtx() {
	if docstests.Ctx == nil {
		docstests.Ctx = &docstests.TestContext{}
	}
	if ctx != nil {
		docstests.Ctx.CommandOutput = ctx.commandOutput
		docstests.Ctx.ExitCode = ctx.exitCode
		docstests.Ctx.CommandError = ctx.commandError
	}
	// Pass shared context to docs tests
	docstests.SharedCtx = sharedCtx
	docstests.SyncContextFromParent()
}

// syncFromDocsCtx copies state back from child packages to main ctx.
func syncFromDocsCtx() {
	// First sync ctx TO sharedCtx to preserve any changes from main test runner
	syncCtxToShared()

	// Then sync from child packages
	docstests.SyncContextToParent()

	// Finally sync back from sharedCtx to ctx if children made changes
	if sharedCtx != nil && ctx != nil && sharedCtx.HasOutput() {
		ctx.commandOutput = sharedCtx.CommandOutput
		ctx.exitCode = sharedCtx.ExitCode
		ctx.commandError = sharedCtx.CommandError
	}
}

// setupDocsMocks bridges to the docs module's mock setup
func setupDocsMocks() error {
	// Initialize shared state
	syncToDocsCtx()
	docstests.OriginalRepoRoot = originalRepoRoot
	docstests.IsolatedTestProjectDir = isolatedTestProjectDir
	docstests.RunCommand = iRunTheCommand
	docstests.SharedCtx = sharedCtx
	return docstests.SetupDocsMocks()
}

// cleanupDocsMocks bridges to the docs module's mock cleanup
func cleanupDocsMocks() {
	syncFromDocsCtx()
	docstests.CleanupDocsMocks()
}

// InitializeDocsScenario registers docs-specific step definitions
func InitializeDocsScenario(sc *godog.ScenarioContext) {
	// Set up shared state - pass the shared context pointer
	docstests.OriginalRepoRoot = originalRepoRoot
	docstests.IsolatedTestProjectDir = isolatedTestProjectDir
	docstests.RunCommand = iRunTheCommand
	docstests.SharedCtx = sharedCtx

	// Sync context before and after every step for backward compatibility
	sc.BeforeStep(func(st *godog.Step) {
		syncToDocsCtx()
	})
	sc.AfterStep(func(st *godog.Step, err error) {
		syncFromDocsCtx()
	})

	// Register the docs module's steps
	docstests.InitializeDocsScenario(sc)
}
