// Package tests provides BDD step definitions for the specs commands.
//
// This file contains mock setup and cleanup functions for isolated testing.
// It follows the same pattern as the commit command's mocks.go.
package tests

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/commands/impl/specs/create"
	"github.com/ready-to-release/eac/src/commands/impl/specs/validate"
	"github.com/ready-to-release/eac/src/core/git"
)

// SetupSpecsMocks sets up git and AI mocks for isolated testing.
// Called automatically when @env:isolated-test-project tag is present.
//
// Note: Contracts are now copied by TestIsolation in the main test runner,
// so this function only handles git mock and AI response setup.
func SetupSpecsMocks() error {
	if IsolatedTestProjectDir == "" {
		return nil // Not in isolated mode
	}

	// Set up mock git repository for both commands
	TestMockRepo = git.NewMockRepository(IsolatedTestProjectDir).
		WithCurrentBranch("main").
		WithHeadSHA("abc1234567890")

	// Inject into both create and validate commands
	create.SetGitRepo(TestMockRepo)
	validate.SetGitRepo(TestMockRepo)

	// Load and set mock AI response for create command
	if OriginalRepoRoot != "" {
		mockResponsePath := filepath.Join(OriginalRepoRoot,
			"src/commands/impl/specs/tests/assets/mock-response.txt")
		mockResponse, err := os.ReadFile(mockResponsePath)
		if err == nil {
			create.SetMockAIResponse(string(mockResponse))
		}
	}

	return nil
}

// CleanupSpecsMocks resets all mocks.
func CleanupSpecsMocks() {
	create.ResetGitRepo()
	create.ResetMockAIResponse()
	validate.ResetGitRepo()
	TestMockRepo = nil
}
