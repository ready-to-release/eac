// Command: work merge
// Short: Merge workspace changes back to main (squash by default)
// Long: Merges the current workspace branch back into the target branch (default: main)
// Long: using squash merge to create a single, well-documented commit.
// Long:
// Long: By default, this command:
// Long:   1. Validates workspace is clean and up to date
// Long:   2. Switches to target branch and updates it
// Long:   3. Squash merges all workspace commits into a single commit
// Long:   4. Uses commit to generate a comprehensive commit message
// Long:   5. Removes the workspace after successful merge
// Long:
// Long: Example:
// Long:   work merge
// Long:   work merge --target=develop
// Long:   work merge --no-squash
// Long:   work merge --keep-worktree
// Flag.target: type=string, default=main, usage=Target branch to merge into
// Flag.no-squash: type=bool, default=false, usage=Use regular merge instead of squash merge
// Flag.keep-worktree: type=bool, default=false, usage=Keep workspace after merge (don't remove)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode (pass through to commit)
package work

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/commit"
	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/logging"
)

func init() {
	registry.Register(Merge)
}

// Intent: Merge workspace changes back to main with squash merge as default
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear flow: validate → update target → merge → commit → cleanup
//   - Squash merge is the sensible default for feature branches
//   - Uses commit for high-quality squash commit messages
//
// Easy to change:
//   - Squash and regular merge paths are separate
//   - Worktree cleanup is optional
//   - Target branch is configurable
//
// Hard to break:
//   - Validates uncommitted changes before proceeding
//   - Checks branch is up to date with target
//   - Prevents merging main into itself
//   - Preserves workspace on merge conflicts

// Merge merges the current workspace into the target branch
func Merge() int {
	// Phase 1: Parse configuration
	config, err := parseMergeConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Starting work merge command")
	internal.WriteDebugFile(config.base.Logger, config.base.RepoRoot, "merge-config.txt",
		fmt.Sprintf("CurrentBranch: %s\nTarget: %s\nNoSquash: %v\nKeepWorktree: %v\n",
			config.currentBranch, config.targetBranch, config.noSquash, config.keepWorktree))

	// Phase 2: Validate environment
	if err := validateMergeEnvironment(config); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}

	// Phase 3: Check branch is up to date
	if err := checkBranchUpToDate(config); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Branch check failed: %v", err))
		return 1
	}

	// Phase 4: Switch to target branch
	config.base.Logger.Info(fmt.Sprintf("Switching to %s...", config.targetBranch))
	if err := switchToTargetBranch(config.base.Logger, config.targetBranch, config.base.RepoRoot); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to switch: %v", err))
		return 1
	}

	// Phase 5: Update target branch
	config.base.Logger.Info(fmt.Sprintf("Updating %s from remote...", config.targetBranch))
	if err := updateTargetBranch(config.targetBranch); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to update: %v", err))
		return 1
	}

	// Phase 6: Perform merge
	var mergeType string
	if config.noSquash {
		config.base.Logger.Info(fmt.Sprintf("\nMerging %s into %s (regular merge)...", config.currentBranch, config.targetBranch))
		if err := config.base.GitOps.Merge(config.currentBranch, false); err != nil {
			handleMergeError(config, err)
			return 1
		}
		mergeType = "fast-forward"
	} else {
		config.base.Logger.Info(fmt.Sprintf("\nMerging %s into %s (squash)...", config.currentBranch, config.targetBranch))
		if err := performSquashMerge(config); err != nil {
			handleMergeError(config, err)
			return 1
		}
		mergeType = "squash"
	}

	// Phase 7: Cleanup workspace
	if !config.keepWorktree {
		config.base.Logger.Info("\nRemoving workspace...")
		if err := removeWorkspace(config); err != nil {
			config.base.Logger.Warn(fmt.Sprintf("Failed to remove workspace: %v", err))
		} else {
			config.base.Logger.Info(fmt.Sprintf("✓ Removed workspace: %s", config.worktreePath))
		}
	} else {
		config.base.Logger.Info(fmt.Sprintf("\n✓ Workspace preserved at: %s", config.worktreePath))
	}

	// Phase 8: Success
	config.base.Logger.Info("")
	config.base.Logger.Info(fmt.Sprintf("✓ Merged %s into %s (%s)", config.currentBranch, config.targetBranch, mergeType))
	config.base.Logger.Debug("Work merge command completed successfully")
	return 0
}

// mergeConfig holds configuration for the merge command
type mergeConfig struct {
	base          *internal.BaseConfig
	targetBranch  string
	noSquash      bool
	keepWorktree  bool
	currentBranch string
	worktreePath  string
}

// parseMergeConfig parses command line arguments
func parseMergeConfig() (*mergeConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "merge"

	// Parse base config (debug flag, repo root, logger, git ops)
	baseConfig, err := internal.ParseBaseConfig(args)
	if err != nil {
		return nil, err
	}

	config := &mergeConfig{
		base:         baseConfig,
		targetBranch: "main",
		noSquash:     false,
		keepWorktree: false,
	}

	// Parse flags
	if targetValue := internal.GetFlagValue(args, "--target"); targetValue != "" {
		config.targetBranch = targetValue
	}
	config.noSquash = internal.HasFlag(args, "--no-squash", "")
	config.keepWorktree = internal.HasFlag(args, "--keep-worktree", "")

	// Get current branch
	currentBranch, err := config.base.GitOps.GetCurrentBranch(config.base.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	config.currentBranch = currentBranch

	// Get worktree path for current directory
	worktreePath, err := getWorktreePath(config.base)
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree path: %w", err)
	}
	config.worktreePath = worktreePath

	return config, nil
}

// validateMergeEnvironment validates the environment before merging
func validateMergeEnvironment(config *mergeConfig) error {
	// Check we're in a git repository
	if err := internal.EnsureInGitRepo(); err != nil {
		return err
	}

	// Prevent merging main into itself
	if config.currentBranch == "main" || config.currentBranch == "master" {
		return fmt.Errorf("cannot merge main into itself\nYou are on the main branch. Switch to a workspace first.")
	}

	// Check for uncommitted changes
	cwd, _ := os.Getwd()
	clean, err := config.base.GitOps.IsWorktreeClean(cwd)
	if err != nil {
		return fmt.Errorf("failed to check working tree status: %w", err)
	}
	if !clean {
		return fmt.Errorf("uncommitted changes detected\nCommit or stash your changes before merging")
	}

	return nil
}

// checkBranchUpToDate checks if the current branch is up to date with target
func checkBranchUpToDate(config *mergeConfig) error {
	// Fetch target branch to ensure we have latest
	if err := config.base.GitOps.FetchBranch(config.targetBranch); err != nil {
		return fmt.Errorf("failed to fetch %s: %w", config.targetBranch, err)
	}

	// Check if current branch has all commits from target
	behindCount, err := config.base.GitOps.GetCommitCount("HEAD", fmt.Sprintf("origin/%s", config.targetBranch))
	if err != nil {
		return fmt.Errorf("failed to check if branch is up to date: %w", err)
	}

	if behindCount > 0 {
		return fmt.Errorf("branch not up to date with %s (behind by %d commits)\nRun 'work pull' first to sync with %s", config.targetBranch, behindCount, config.targetBranch)
	}

	return nil
}

// switchToTargetBranch switches to the target branch
// In multi-worktree setups, it finds and switches to the worktree where the target branch is checked out
func switchToTargetBranch(logger *logging.Logger, targetBranch, repoRoot string) error {
	// First, find where the target branch is checked out
	targetWorktree, err := findWorktreeForBranch(logger, targetBranch)
	if err != nil {
		// Branch not checked out anywhere, try normal checkout
		cmd := exec.Command("git", "checkout", targetBranch)
		cmd.Dir = repoRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to switch to %s: %w\nOutput: %s", targetBranch, err, string(output))
		}
		return nil
	}

	// Branch is checked out in a worktree - switch to that directory
	if err := os.Chdir(targetWorktree); err != nil {
		return fmt.Errorf("failed to switch to worktree at %s: %w", targetWorktree, err)
	}

	logger.Info(fmt.Sprintf("📂 Switched to worktree: %s", targetWorktree))
	return nil
}

// findWorktreeForBranch finds the worktree path where a branch is checked out
// Returns empty string and error if branch is not checked out in any worktree
func findWorktreeForBranch(logger *logging.Logger, branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w", err)
	}

	lines := fmt.Sprintf("%s", output)
	var currentPath string
	for _, line := range strings.Split(lines, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			branchName := strings.TrimPrefix(line, "branch refs/heads/")
			if branchName == branch {
				return currentPath, nil
			}
		}
	}

	return "", fmt.Errorf("branch %s not checked out in any worktree", branch)
}

// updateTargetBranch updates the target branch from remote
func updateTargetBranch(targetBranch string) error {
	cmd := exec.Command("git", "pull", "origin", targetBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update %s: %w\nOutput: %s", targetBranch, err, string(output))
	}
	return nil
}

// performSquashMerge performs a squash merge and uses commit message for the message
func performSquashMerge(config *mergeConfig) error {
	// Perform squash merge (stages changes but doesn't commit)
	if err := config.base.GitOps.Merge(config.currentBranch, true); err != nil {
		return err
	}

	// Use commit message to generate commit message
	config.base.Logger.Info("\nGenerating commit message...")
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	if config.base.Debug {
		os.Args = []string{"r2r", "commit", "message", "--debug"}
	} else {
		os.Args = []string{"r2r", "commit", "message"}
	}

	exitCode := commit.CreateCommitMessage()
	if exitCode != 0 {
		return fmt.Errorf("commit message failed with exit code %d", exitCode)
	}

	return nil
}

// removeWorkspace removes the workspace and deletes the branch
func removeWorkspace(config *mergeConfig) error {
	// Remove worktree
	if err := config.base.GitOps.RemoveWorktree(config.worktreePath); err != nil {
		return err
	}

	// Delete branch
	if err := config.base.GitOps.DeleteBranch(config.currentBranch, false); err != nil {
		// If branch deletion fails, warn but don't error (branch might be needed)
		config.base.Logger.Warn(fmt.Sprintf("Failed to delete branch %s: %v", config.currentBranch, err))
		config.base.Logger.Warn(fmt.Sprintf("You can manually delete with: git branch -D %s", config.currentBranch))
	}

	return nil
}

// getWorktreePath returns the path of the current worktree
func getWorktreePath(base *internal.BaseConfig) (string, error) {
	// Get list of all worktrees
	worktrees, err := base.GitOps.ListWorktrees()
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Find current worktree path
	cwd, _ := os.Getwd()
	for _, wt := range worktrees {
		// Normalize paths for comparison
		if wt.Path == cwd || wt.Path == base.RepoRoot {
			return wt.Path, nil
		}
	}

	return base.RepoRoot, nil
}

// handleMergeError handles merge errors and provides guidance
func handleMergeError(config *mergeConfig, err error) {
	config.base.Logger.Error("\n⚠️  Merge conflict detected\n")

	// Get conflicting files
	conflicts, _ := config.base.GitOps.GetConflictingFiles()
	if len(conflicts) > 0 {
		config.base.Logger.Error("Conflicting files:")
		for _, file := range conflicts {
			config.base.Logger.Error(fmt.Sprintf("  - %s", file))
		}
		config.base.Logger.Error("")
	}

	config.base.Logger.Error("Resolve conflicts then:")
	config.base.Logger.Error("  git add <files>")
	config.base.Logger.Error("  git commit")
	config.base.Logger.Error("")
	config.base.Logger.Error("Or abort:")
	config.base.Logger.Error("  git merge --abort")
	config.base.Logger.Error("")
	config.base.Logger.Error(fmt.Sprintf("Note: Workspace preserved at %s", config.worktreePath))
}
