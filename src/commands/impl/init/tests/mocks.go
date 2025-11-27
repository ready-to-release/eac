// Package tests provides BDD step definitions for the init command.
//
// This file contains mock setup and cleanup functions for isolated testing.
package tests

import (
	initcmd "github.com/ready-to-release/eac/src/commands/impl/init"
	"github.com/ready-to-release/eac/src/core/git"
)

// SetupInitMocks sets up git mocks for isolated testing.
// Called automatically when @env:isolated-test-project tag is present.
func SetupInitMocks() error {
	if IsolatedTestProjectDir == "" {
		return nil // Not in isolated mode
	}

	// Set up mock git repository
	TestMockRepo = git.NewMockRepository(IsolatedTestProjectDir).
		WithCurrentBranch("main").
		WithHeadSHA("abc1234567890")

	initcmd.SetGitRepo(TestMockRepo)
	return nil
}

// CleanupInitMocks resets all mocks.
func CleanupInitMocks() {
	initcmd.ResetGitRepo()
	TestMockRepo = nil
}
