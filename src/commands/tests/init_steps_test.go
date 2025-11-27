// Godog BDD step definitions for init command - BRIDGE FILE
//
// This file delegates to the init/tests package for actual step implementations.
// It syncs test context between the parent tests package and the child init/tests package.
//
// Features:
// - specs/src-commands/init/
package tests

import (
	"github.com/cucumber/godog"
	initTests "github.com/ready-to-release/eac/src/commands/impl/init/tests"
)

func InitializeInitScenario(sc *godog.ScenarioContext) {
	// Sync context to child package before delegating
	initTests.IsolatedTestProjectDir = isolatedTestProjectDir
	initTests.OriginalRepoRoot = originalRepoRoot
	initTests.SharedCtx = sharedCtx // Pass shared context to child package

	// Delegate to child package for step registration
	initTests.InitializeInitScenario(sc)
}

// setupInitMocks bridges to the init module's mock setup
func setupInitMocks() error {
	// Initialize shared state
	initTests.OriginalRepoRoot = originalRepoRoot
	initTests.IsolatedTestProjectDir = isolatedTestProjectDir
	initTests.SharedCtx = sharedCtx
	return initTests.SetupInitMocks()
}

// cleanupInitMocks bridges to the init module's mock cleanup
func cleanupInitMocks() {
	initTests.CleanupInitMocks()
}
