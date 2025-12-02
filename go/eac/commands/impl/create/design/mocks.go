// Package create provides the design create subcommand.
//
// This file contains mock infrastructure for testing the design create command.
// It follows the same pattern as the commit command's mocks.go.
package design

import (
	"github.com/ready-to-release/eac/go/eac/core/git"
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

// gitRepo holds the git repository instance for git operations.
// In production, this is initialized lazily. For tests, it can be injected via SetGitRepo.
var gitRepo git.GitRepository

// getGitRepo returns the git repository, initializing it if needed.
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	repo, err := git.Open(workspaceRoot)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// SetGitRepo allows tests to inject a mock repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepo = repo
}

// ResetGitRepo clears the repository for test cleanup.
func ResetGitRepo() {
	gitRepo = nil
}
