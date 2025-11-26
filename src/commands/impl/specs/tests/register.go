// Package tests provides shared test context and step registration for specs subcommands.
//
// This file delegates step registration to subcommand-specific packages.
package tests

import (
	"context"

	"github.com/cucumber/godog"
)

// InitializeSpecsScenario registers all specs-related step definitions.
func InitializeSpecsScenario(sc *godog.ScenarioContext) {
	// Sync context before scenario runs
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		SyncContextFromParent()
		return ctx, nil
	})

	// Register shared steps
	registerSharedSteps(sc)

	// Register create subcommand steps
	registerCreateSetupSteps(sc)
	registerCreateVerifySteps(sc)

	// Register validate subcommand steps
	registerValidateSetupSteps(sc)
	registerValidateVerifySteps(sc)
}

// SyncContextFromParent synchronizes context from the parent test runner.
// This must be called before each scenario to ensure we have the latest context.
func SyncContextFromParent() {
	// Context is synced by the bridge file in src/commands/tests/specs_steps_test.go
}

// SyncContextToParent synchronizes changes back to the parent test runner.
// This should be called after each scenario to capture any state changes.
func SyncContextToParent() {
	// Context is synced by the bridge file in src/commands/tests/specs_steps_test.go
}
