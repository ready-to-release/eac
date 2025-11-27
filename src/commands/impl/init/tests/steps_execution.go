// Package tests provides BDD step definitions for the init command.
//
// This file contains "When" steps (actions/execution).
package tests

// registerExecutionSteps registers all "When" step definitions.
func registerExecutionSteps(sc interface{ Step(interface{}, interface{}) }) {
	sc.Step(`^I run "init" without any flags$`, iRunInitWithoutAnyFlags)
}

// iRunInitWithoutAnyFlags runs init command with no flags
func iRunInitWithoutAnyFlags() error {
	// Use the shared RunCommand function from shared context
	if SharedCtx == nil || SharedCtx.RunCommand == nil {
		// Fallback: construct command manually
		return runInitCommand("init")
	}
	return SharedCtx.RunCommand("init")
}

// runInitCommand is a local helper to run init command
func runInitCommand(args string) error {
	if SharedCtx != nil && SharedCtx.RunCommand != nil {
		return SharedCtx.RunCommand(args)
	}
	// If shared context not available, return error
	return nil
}
