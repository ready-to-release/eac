// Package reset provides BDD step definitions for the commit reset subcommand.
//
// This file contains mock setup and cleanup functions for git operations testing.
package reset

import (
	"github.com/ready-to-release/eac/src/commands/impl/commit"
)

// MockResetGitOps implements commit.ResetGitOperations for testing.
type MockResetGitOps struct {
	HasParentCommitResult bool
	HasParentCommitError  error
	SoftResetHeadError    error
	SoftResetHeadCalled   bool
}

// HasParentCommit returns the configured result for parent commit check.
func (m *MockResetGitOps) HasParentCommit() (bool, error) {
	return m.HasParentCommitResult, m.HasParentCommitError
}

// SoftResetHead records that reset was called and returns configured error.
func (m *MockResetGitOps) SoftResetHead() error {
	m.SoftResetHeadCalled = true
	return m.SoftResetHeadError
}

// MockOps holds the mock for reset tests
var MockOps *MockResetGitOps

// SetupMocks sets up git mocks for reset command testing.
// Called automatically when @env:isolated-test-project tag is present.
func SetupMocks() error {
	if IsolatedTestProjectDir == "" {
		return nil // Not in isolated mode
	}

	MockOps = &MockResetGitOps{
		HasParentCommitResult: true, // Default: has commits to reset
	}
	commit.SetResetGitOps(MockOps)
	return nil
}

// CleanupMocks resets all reset mocks.
func CleanupMocks() {
	commit.ResetResetGitOps()
	MockOps = nil
}
