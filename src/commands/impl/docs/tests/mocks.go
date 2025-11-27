// Package tests provides BDD step definitions for the docs command.
//
// This file contains mock setup and cleanup functions for isolated testing.
// It follows the same pattern as the commit and specs commands' mocks.go.
package tests

// SetupDocsMocks sets up Docker mocks for isolated testing.
// Called automatically when @env:isolated-test-project tag is present.
func SetupDocsMocks() error {
	// Docker mocking not yet implemented for docs command
	// Tests will use real Docker client
	return nil
}

// CleanupDocsMocks resets all mocks.
func CleanupDocsMocks() {
	// No mocks to clean up yet
}
