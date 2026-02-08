// Package create provides the design create subcommand.
//
// This file contains mock infrastructure for testing the design create command.
// It follows the same pattern as the commit command's mocks.go.
package design

import (
	"github.com/ready-to-release/eac/go/core/git"
)

// mockAIResponse holds the mock response for testing. When set, AI calls return this.
var mockAIResponse string

// SetMockAIResponse sets a mock AI response for testing.
// When set, the generateWithAI function will return this response instead of calling the AI.
func SetMockAIResponse(response string) {
	mockAIResponse = response
}

// ResetMockAIResponse clears the mock AI response.
// This should be called in test cleanup to ensure tests don't affect each other.
func ResetMockAIResponse() {
	mockAIResponse = ""
}

// GetMockAIResponse returns the current mock AI response.
// Returns empty string if no mock is set.
func GetMockAIResponse() string {
	return mockAIResponse
}

// gitRepoProvider provides lazy-initialized git repository with test injection support.
var gitRepoProvider = &git.LazyRepo{}

// getGitRepo returns the git repository, initializing it if needed.
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	return gitRepoProvider.Get(workspaceRoot)
}

// SetGitRepo allows tests to inject a mock repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepoProvider.Set(repo)
}

// ResetGitRepo clears the repository for test cleanup.
func ResetGitRepo() {
	gitRepoProvider.Reset()
}
