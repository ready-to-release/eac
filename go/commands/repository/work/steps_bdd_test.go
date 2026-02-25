// Package work contains godog step implementations for work commands.
//
// This file contains work command step definitions.
package work

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// registerSteps registers step definitions for work command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Output verification steps are registered in common steps (go/adapters/godog/steps.go)
	// Do not register `I see "..."` or `I see error "..."` here to avoid ambiguity

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
		if err := ensureMainWorkspace(ctx); err != nil {
			return err
		}
		return verifyCurrentBranch(ctx, "main")
	})
	sc.Step(`^the local branch "([^"]*)" is deleted$`, func(branch string) error {
		// Ensure we're checking from the main worktree
		if err := ensureMainWorkspace(ctx); err != nil {
			return err
		}
		return verifyBranchDeleted(ctx, branch)
	})
	sc.Step(`^the local branch "([^"]*)" still exists$`, func(branch string) error {
		// Ensure we're checking from the main worktree
		if err := ensureMainWorkspace(ctx); err != nil {
			return err
		}
		return verifyBranchExists(ctx, branch)
	})
}

// commitAllUncommittedFiles commits any uncommitted files in the current directory.
func commitAllUncommittedFiles(ctx *eacgodog.TestContext) error {
	if ctx.IsolatedDir == "" {
		return nil // Not in isolated mode
	}

	workDir := ctx.CurrentWorkDir
	if workDir == "" {
		workDir = ctx.IsolatedDir
	}

	return gitCommitIfDirty(workDir, "Auto-commit before test")
}

// setupWorkspace creates a git worktree for the specified branch.
func setupWorkspace(ctx *eacgodog.TestContext, branch string) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("workspace setup requires isolation")
	}

	// Verify git repository exists (should be set up by test isolation)
	gitDir := filepath.Join(ctx.IsolatedDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("git repository not initialized - test isolation should have set this up")
	}

	// Ensure main worktree is clean before creating branch worktree
	if err := gitCommitIfDirty(ctx.IsolatedDir, "Pre-worktree setup"); err != nil {
		return fmt.Errorf("failed to clean main worktree: %w", err)
	}

	// Create branch (ignore error if it already exists) and add worktree
	worktreePath := filepath.Join(ctx.IsolatedDir, "worktrees", branch)
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o750); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}

	_ = runGitCommand(ctx.IsolatedDir, "branch", branch) // OK if exists
	if err := runGitCommand(ctx.IsolatedDir, "worktree", "add", worktreePath, branch); err != nil {
		return fmt.Errorf("failed to add worktree: %w", err)
	}

	// Switch current working directory to worktree (but keep IsolatedDir as main repo root)
	ctx.CurrentWorkDir = worktreePath

	// Ensure the worktree has a clean state
	if err := gitCommitIfDirty(ctx.CurrentWorkDir, "Worktree setup"); err != nil {
		return fmt.Errorf("failed to clean worktree: %w", err)
	}

	return nil
}

// ensureMainWorkspace ensures we're in the main workspace.
func ensureMainWorkspace(ctx *eacgodog.TestContext) error {
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
func verifyCurrentBranch(ctx *eacgodog.TestContext, expectedBranch string) error {
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
func ensureCleanWorkspace(ctx *eacgodog.TestContext) error {
	if err := commitAllUncommittedFiles(ctx); err != nil {
		return err
	}

	// Verify the workspace is actually clean after commit
	workDir := ctx.CurrentWorkDir
	if workDir == "" {
		workDir = ctx.IsolatedDir
	}
	status, err := getGitCommandOutput(workDir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to verify clean state: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		// Diagnostic: dump git config and diff for first dirty file
		configOut, _ := getGitCommandOutput(workDir, "config", "--list", "--show-origin")
		envDump := fmt.Sprintf("GIT_CONFIG_GLOBAL=%q GIT_CONFIG_SYSTEM=%q", os.Getenv("GIT_CONFIG_GLOBAL"), os.Getenv("GIT_CONFIG_SYSTEM"))
		diffOut, _ := getGitCommandOutput(workDir, "diff", "--stat", "HEAD")
		return fmt.Errorf("workspace still dirty after cleanup:\n%s\nworkDir: %s\nisolatedDir: %s\nenv: %s\ngit config:\n%s\ngit diff --stat HEAD:\n%s",
			status, workDir, ctx.IsolatedDir, envDump, configOut, diffOut)
	}
	return nil
}

// createUncommittedChanges creates uncommitted changes in the workspace.
func createUncommittedChanges(ctx *eacgodog.TestContext) error {
	// Create file in current working directory
	workDir := ctx.CurrentWorkDir
	if workDir == "" {
		workDir = ctx.IsolatedDir
	}
	testFile := filepath.Join(workDir, "test-change.txt")
	return os.WriteFile(testFile, []byte("uncommitted change\n"), 0o644)
}

// verifyBranchDeleted verifies that a branch has been deleted.
func verifyBranchDeleted(ctx *eacgodog.TestContext, branch string) error {
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
func verifyBranchExists(ctx *eacgodog.TestContext, branch string) error {
	output, err := getGitCommandOutput(ctx.IsolatedDir, "branch", "--list", branch)
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("branch %q does not exist", branch)
	}
	return nil
}

// gitCommitIfDirty stages and commits all changes in one subprocess call.
// Equivalent to: git status --porcelain → git add . → git commit, but batched
// into a single bash invocation to reduce subprocess overhead.
func gitCommitIfDirty(dir, message string) error {
	script := fmt.Sprintf(
		`if [ -n "$(git status --porcelain)" ]; then git add . && git commit -m %q; fi`,
		message,
	)
	cmd := exec.Command("bash", "-c", script)
	cmd.Dir = dir
	cmd.Env = gitTestEnv()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit-if-dirty failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// runGitCommand runs a git command in the specified directory.
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitTestEnv()
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
	cmd.Env = gitTestEnv()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %v failed: %w\nOutput: %s", args, err, string(output))
	}
	return string(output), nil
}

// gitTestEnv returns environment variables for subprocess git commands,
// isolated from global/system git config. This prevents Windows
// core.autocrlf=true from causing CRLF conversion in test worktrees.
func gitTestEnv() []string {
	env := os.Environ()
	env = append(env, "GIT_CONFIG_GLOBAL=", "GIT_CONFIG_SYSTEM=")
	return env
}
