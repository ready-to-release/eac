// Package tests provides BDD test implementations for the design command.
//
// This package contains step definitions organized into subfolders by subcommand,
// plus shared test context, mocks, and helper utilities.
package tests

import (
	"github.com/cucumber/godog"
)

// RegisterDesignSteps registers all design-related step definitions with godog.
// This is called by the main test suite in src/commands/tests.
func RegisterDesignSteps(sc *godog.ScenarioContext) {
	// Register shared context and mocks
	registerContext(sc)

	// Register subcommand-specific steps inline to avoid import cycles
	registerCreateSteps(sc)
	registerUpdateSteps(sc)
	registerValidateSteps(sc)
	registerServeSteps(sc)
}
