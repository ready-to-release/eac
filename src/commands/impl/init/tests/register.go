// Package tests provides BDD step definitions for the init command.
package tests

import (
	"github.com/cucumber/godog"
)

// Package-level variables set by the main test runner (from tests package)
var (
	IsolatedTestProjectDir string
	OriginalRepoRoot       string
)

// InitializeInitScenario registers all init-related step definitions.
func InitializeInitScenario(sc *godog.ScenarioContext) {
	// Register step definitions - pass sc directly (it implements Step method)
	// Note: Context syncing and mock setup is handled by the parent test package's Before hook
	registerSetupSteps(sc)
	registerExecutionSteps(sc)
	registerVerificationSteps(sc)
}

// SyncContextFromParent synchronizes context from the parent test runner.
// This must be called before each scenario to ensure we have the latest context.
func SyncContextFromParent() {
	// Context is synced by the bridge file in src/commands/tests/init_steps_test.go
}
