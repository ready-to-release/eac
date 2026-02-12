package commitmessage

import (
	"fmt"
	"os"
	"strings"

	aimock "github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/environments"
)

// verifyContractImplementation checks if the AI config is valid.
func verifyContractImplementation(workspaceRoot string) error {
	log.Debug("verifyContractImplementation: start")
	// Verify that ai-config.yml can be loaded and has commit-message type
	aiConfig, err := aimock.LoadAIConfig(workspaceRoot)
	if err != nil {
		log.Error("AI config verification failed")
		log.Errorf("config load error: %v", err)
		return fmt.Errorf("config verification failed: %w", err)
	}

	// Check that commit-message type exists
	if _, ok := aiConfig.Types["commit-message"]; !ok {
		log.Error("AI config missing commit-message type")
		return fmt.Errorf("ai-config.yml must define 'commit-message' type")
	}

	log.Debug("verifyContractImplementation: config verified")
	return nil
}

// performAutoCommit creates a git commit with the generated message.
// This is called when --commit flag is provided and message generation succeeds.
func performAutoCommit(deps *Deps, workspaceRoot, message string) int {
	repo, err := deps.GetGitRepo(workspaceRoot)
	if err != nil {
		log.Errorf("auto-commit failed: %v", err)
		return 1
	}

	// Get author info from git config (use git command for reliability)
	authorName, authorEmail := getGitAuthorInfo(deps, workspaceRoot)
	if authorName == "" || authorEmail == "" {
		log.Error("Auto-commit failed: git user.name and user.email must be configured")
		log.Info("Run: git config user.name \"Your Name\"")
		log.Info("Run: git config user.email \"your@email.com\"")
		return 1
	}

	// Perform the commit
	hash, err := repo.Commit(message, authorName, authorEmail)
	if err != nil {
		log.Errorf("git commit failed: %v", err)
		return 1
	}

	// Success message to stderr (so it doesn't interfere with stdout output)
	log.Infof("\n✓ Committed: %s", hash[:7])
	return 0
}

// getGitAuthorInfo retrieves git user.name and user.email from git config.
// Returns empty strings if not configured.
func getGitAuthorInfo(deps *Deps, workspaceRoot string) (name, email string) {
	// Try to get from environment first (GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL)
	name = os.Getenv(environments.EnvGitAuthorName)
	email = os.Getenv(environments.EnvGitAuthorEmail)
	if name != "" && email != "" {
		return name, email
	}

	// Fall back to git config command
	// Using exec.Command because go-git config reading can be complex
	nameCmd := deps.ExecCmd("git", "config", "user.name")
	nameCmd.Dir = workspaceRoot
	if nameOut, err := nameCmd.Output(); err == nil {
		name = strings.TrimSpace(string(nameOut))
	}

	emailCmd := deps.ExecCmd("git", "config", "user.email")
	emailCmd.Dir = workspaceRoot
	if emailOut, err := emailCmd.Output(); err == nil {
		email = strings.TrimSpace(string(emailOut))
	}

	return name, email
}
