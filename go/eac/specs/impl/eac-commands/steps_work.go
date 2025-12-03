// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains work command step definitions.
package srccommands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// registerWorkSteps registers step definitions for work command features.
func registerWorkSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Work-specific output verification
	sc.Step(`^I see "([^"]*)"$`, func(text string) error {
		return workOutputContains(ctx, text)
	})
	sc.Step(`^I see error "([^"]*)"$`, func(text string) error {
		return workOutputContains(ctx, text)
	})

	// Workspace setup steps
	sc.Step(`^I am in a workspace for "([^"]*)"$`, func(branch string) error {
		return setupWorkspace(ctx, branch)
	})
	sc.Step(`^I am in the main workspace on branch "([^"]*)"$`, func(branch string) error {
		if err := ensureMainWorkspace(ctx); err != nil {
			return err
		}
		return verifyCurrentBranch(ctx, branch)
	})

	// Uncommitted changes steps
	sc.Step(`^I have no uncommitted changes$`, func() error {
		return ensureCleanWorkspace(ctx)
	})
	sc.Step(`^I have uncommitted changes$`, func() error {
		return createUncommittedChanges(ctx)
	})

	// Verification steps
	sc.Step(`^I am switched to main branch$`, func() error {
		// After work remove, we need to verify from the main worktree
		ensureMainWorkspace(ctx)
		return verifyCurrentBranch(ctx, "main")
	})
	sc.Step(`^the local branch "([^"]*)" is deleted$`, func(branch string) error {
		// Ensure we're checking from the main worktree
		ensureMainWorkspace(ctx)
		return verifyBranchDeleted(ctx, branch)
	})
	sc.Step(`^the local branch "([^"]*)" still exists$`, func(branch string) error {
		// Ensure we're checking from the main worktree
		ensureMainWorkspace(ctx)
		return verifyBranchExists(ctx, branch)
	})
}

// workOutputContains checks if command output contains text.
func workOutputContains(ctx *internal.TestContext, text string) error {
	if !strings.Contains(ctx.CommandOutput, text) {
		return fmt.Errorf("expected output to contain %q, got:\n%s", text, ctx.CommandOutput)
	}
	return nil
}

// commitAllUncommittedFiles commits any uncommitted files in the current directory.
func commitAllUncommittedFiles(ctx *internal.TestContext) error {
	if ctx.IsolatedDir == "" {
		return nil // Not in isolated mode
	}

	// Use current working directory (may be worktree or main)
	workDir := ctx.CurrentWorkDir
	if workDir == "" {
		workDir = ctx.IsolatedDir
	}

	// Check for uncommitted changes
	status, err := getGitCommandOutput(workDir, "status", "--porcelain")
	if err != nil {
		// If git status fails, maybe we're not in a git repo, just continue
		return nil
	}

	// If there are uncommitted changes, commit them
	if strings.TrimSpace(status) != "" {
		if err := runGitCommand(workDir, "add", "."); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}
		if err := runGitCommand(workDir, "commit", "-m", "Auto-commit before test"); err != nil {
			return fmt.Errorf("failed to commit files: %w", err)
		}
	}

	return nil
}

// setupWorkspace creates a git worktree for the specified branch.
func setupWorkspace(ctx *internal.TestContext, branch string) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("workspace setup requires isolation")
	}

	// Store the main worktree path
	mainWorktreePath := ctx.IsolatedDir

	// Verify git repository exists (should be set up by test isolation)
	gitDir := filepath.Join(ctx.IsolatedDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("git repository not initialized - test isolation should have set this up")
	}

	// Ensure main worktree is clean before creating branch worktree
	status, err := getGitCommandOutput(mainWorktreePath, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		// Commit any uncommitted files in main worktree
		if err := runGitCommand(mainWorktreePath, "add", "."); err != nil {
			return fmt.Errorf("failed to stage files in main: %w", err)
		}
		if err := runGitCommand(mainWorktreePath, "commit", "-m", "Pre-worktree setup"); err != nil {
			return fmt.Errorf("failed to commit in main: %w", err)
		}
	}

	// Create branch if it doesn't exist
	if err := runGitCommand(ctx.IsolatedDir, "branch", branch); err != nil {
		// Branch might already exist, that's okay
	}

	// Create worktree directory
	worktreePath := filepath.Join(ctx.IsolatedDir, "worktrees", branch)
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}

	// Add worktree
	if err := runGitCommand(ctx.IsolatedDir, "worktree", "add", worktreePath, branch); err != nil {
		return fmt.Errorf("failed to add worktree: %w", err)
	}

	// Switch current working directory to worktree (but keep IsolatedDir as main repo root)
	ctx.CurrentWorkDir = worktreePath

	// Ensure the worktree has a clean state by committing any files
	status, err = getGitCommandOutput(ctx.CurrentWorkDir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check worktree status: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		if err := runGitCommand(ctx.CurrentWorkDir, "add", "."); err != nil {
			return fmt.Errorf("failed to stage files in worktree: %w", err)
		}
		if err := runGitCommand(ctx.CurrentWorkDir, "commit", "-m", "Worktree setup"); err != nil {
			return fmt.Errorf("failed to commit worktree files: %w", err)
		}
	}

	return nil
}

// ensureMainWorkspace ensures we're in the main workspace.
func ensureMainWorkspace(ctx *internal.TestContext) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("workspace operations require isolation")
	}

	// Reset current working directory to main isolated directory
	ctx.CurrentWorkDir = ctx.IsolatedDir

	// Verify git repository exists (should be set up by test isolation)
	gitDir := filepath.Join(ctx.IsolatedDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("git repository not initialized - test isolation should have set this up")
	}

	return nil
}

// verifyCurrentBranch verifies we're on the specified branch.
func verifyCurrentBranch(ctx *internal.TestContext, expectedBranch string) error {
	// Use CurrentWorkDir to check the branch where we currently are
	workDir := ctx.CurrentWorkDir
	if workDir == "" {
		workDir = ctx.IsolatedDir
	}
	output, err := getGitCommandOutput(workDir, "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(output)

	if currentBranch != expectedBranch {
		return fmt.Errorf("expected branch %q, got %q", expectedBranch, currentBranch)
	}
	return nil
}

// ensureCleanWorkspace ensures there are no uncommitted changes.
func ensureCleanWorkspace(ctx *internal.TestContext) error {
	// Use the helper to commit any uncommitted files
	return commitAllUncommittedFiles(ctx)
}

// createUncommittedChanges creates uncommitted changes in the workspace.
func createUncommittedChanges(ctx *internal.TestContext) error {
	// Create file in current working directory
	workDir := ctx.CurrentWorkDir
	if workDir == "" {
		workDir = ctx.IsolatedDir
	}
	testFile := filepath.Join(workDir, "test-change.txt")
	return os.WriteFile(testFile, []byte("uncommitted change\n"), 0644)
}

// verifyWorkspaceRemoved verifies that the worktree was removed.
func verifyWorkspaceRemoved(ctx *internal.TestContext) error {
	// Check git worktree list to ensure the worktree is not present
	output, err := getGitCommandOutput(ctx.IsolatedDir, "worktree", "list")
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}
	// If we're checking for removal, we should be back in main
	// The output should only show the main worktree
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "worktrees") {
			return fmt.Errorf("worktree still exists: %s", line)
		}
	}
	return nil
}

// verifyBranchDeleted verifies that a branch has been deleted.
func verifyBranchDeleted(ctx *internal.TestContext, branch string) error {
	output, err := getGitCommandOutput(ctx.IsolatedDir, "branch", "--list", branch)
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}
	if strings.TrimSpace(output) != "" {
		return fmt.Errorf("branch %q still exists", branch)
	}
	return nil
}

// verifyBranchExists verifies that a branch still exists.
func verifyBranchExists(ctx *internal.TestContext, branch string) error {
	output, err := getGitCommandOutput(ctx.IsolatedDir, "branch", "--list", branch)
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("branch %q does not exist", branch)
	}
	return nil
}

// runGitCommand runs a git command in the specified directory.
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v failed: %w\nOutput: %s", args, err, string(output))
	}
	return nil
}

// getGitCommandOutput runs a git command and returns its output.
func getGitCommandOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %v failed: %w\nOutput: %s", args, err, string(output))
	}
	return string(output), nil
}
