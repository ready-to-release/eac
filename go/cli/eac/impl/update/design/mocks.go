// Package design provides the update design subcommand.
//
// This file contains mock infrastructure for testing the update design command.
// It follows the same pattern as the create design command's mocks.go.
package design

import (
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/logging"
)

// mockAIResponse holds the mock response for testing. When set, AI calls return this.
var mockAIResponse string

// SetMockAIResponse sets a mock AI response for testing.
// When set, the generateUpdatedWorkspace function will return this response instead of calling the AI.
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
var (
	gitRepo git.GitRepository
	gitMgr  *git.RepositoryManager
)

// initGitManager initializes the git repository manager if needed.
func initGitManager() {
	if gitMgr == nil {
		gitMgr = git.NewManager(logging.C().Zap())
	}
}

// getGitRepo returns the git repository, initializing it if needed.
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	repo, err := gitMgr.Open(workspaceRoot)
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
	gitMgr = nil
}
