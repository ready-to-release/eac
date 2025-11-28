// Godog BDD step definitions for risks command - Bridge to risks module tests
//
// This file bridges the risks step definitions from the risks module
// to the cross-module test runner in src/commands/tests.
//
// The actual step implementations live in:
//   src/commands/impl/risks/tests/
package tests

import (
	"github.com/cucumber/godog"
	riskstests "github.com/ready-to-release/eac/src/commands/impl/risks/tests"
)

// syncToRisksCtx copies state from main ctx to riskstests
func syncToRisksCtx() {
	if riskstests.Ctx == nil {
		riskstests.Ctx = riskstests.NewTestContext()
	}
	// Pass shared context to risks tests
	riskstests.SharedCtx = sharedCtx
	riskstests.OriginalRepoRoot = originalRepoRoot
	riskstests.IsolatedTestProjectDir = isolatedTestProjectDir
}

// syncFromRisksCtx copies state back from child packages to main ctx
func syncFromRisksCtx() {
	// First sync ctx TO sharedCtx to preserve any changes from main test runner
	syncCtxToShared()

	// Finally sync back from sharedCtx to ctx if children made changes
	if sharedCtx != nil && ctx != nil && sharedCtx.HasOutput() {
		ctx.commandOutput = sharedCtx.CommandOutput
		ctx.exitCode = sharedCtx.ExitCode
		ctx.commandError = sharedCtx.CommandError
	}
}

// setupRisksMocks bridges to the risks module's mock setup
func setupRisksMocks() error {
	// Initialize shared state
	syncToRisksCtx()
	riskstests.OriginalRepoRoot = originalRepoRoot
	riskstests.IsolatedTestProjectDir = isolatedTestProjectDir
	riskstests.SharedCtx = sharedCtx
	return riskstests.SetupRisksMocks()
}

// cleanupRisksMocks bridges to the risks module's mock cleanup
func cleanupRisksMocks() {
	syncFromRisksCtx()
	riskstests.CleanupRisksMocks()
}

// InitializeRisksScenario registers risks-specific step definitions
func InitializeRisksScenario(sc *godog.ScenarioContext) {
	// Set up shared state - pass the shared context pointer
	riskstests.OriginalRepoRoot = originalRepoRoot
	riskstests.IsolatedTestProjectDir = isolatedTestProjectDir
	riskstests.SharedCtx = sharedCtx

	// Sync context before and after every step
	sc.BeforeStep(func(st *godog.Step) {
		syncToRisksCtx()
	})
	sc.AfterStep(func(st *godog.Step, err error) {
		syncFromRisksCtx()
	})

	// Register the risks module's steps
	riskstests.InitializeRisksScenario(sc)
}
