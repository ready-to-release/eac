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
// Long: Expected Output:
// Long:   - Squash merge commit on target branch
// Long:   - Workspace removed (unless --keep-worktree)
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
	"strings"
	"time"

	"go.uber.org/zap"

	commitmessage "github.com/ready-to-release/eac/go/cli/eac/impl/create/commit-message"
	"github.com/ready-to-release/eac/go/cli/eac/impl/work/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/gitexec"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func init() {
	registry.Register(Merge)
}

// Merge merges the current workspace into the target branch.
func Merge() int {
	startTime := time.Now()

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Phase 1: Parse configuration
	config, err := parseMergeConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer func() { _ = config.base.Logger.Sync() }() //nolint:errcheck // best-effort logger sync

	config.base.Logger.Debug("Phase 1: Starting parse configuration", zap.String("phase", "phase1"))

	config.base.Logger.Debug("Phase 1: Completed",
		zap.String("phase", "phase1"),
		zap.String("currentBranch", config.currentBranch),
		zap.String("targetBranch", config.targetBranch),
		zap.Bool("noSquash", config.noSquash),
		zap.Bool("keepWorktree", config.keepWorktree))

	// Phase 2: Validate environment
	config.base.Logger.Debug("Phase 2: Starting validate environment", zap.String("phase", "phase2"))
	if err := validateMergeEnvironment(config); err != nil {
		config.base.Logger.Debug("Phase 2: Failed", zap.String("phase", "phase2"), zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}
	config.base.Logger.Debug("Phase 2: Completed", zap.String("phase", "phase2"))

	// Phase 3: Check branch is up to date
	config.base.Logger.Debug("Phase 3: Starting check branch is up to date",
		zap.String("phase", "phase3"),
		zap.String("currentBranch", config.currentBranch),
		zap.String("targetBranch", config.targetBranch))
	if err := checkBranchUpToDate(config); err != nil {
		config.base.Logger.Debug("Phase 3: Failed", zap.String("phase", "phase3"), zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Branch check failed: %v", err))
		return 1
	}
	config.base.Logger.Debug("Phase 3: Completed", zap.String("phase", "phase3"))

	// Phase 4: Switch to target branch
	config.base.Logger.Debug("Phase 4: Starting switch to target branch",
		zap.String("phase", "phase4"),
		zap.String("targetBranch", config.targetBranch))
	config.base.Logger.Info(fmt.Sprintf("Switching to %s...", config.targetBranch))
	if err := switchToTargetBranch(config.targetBranch, config.base.RepoRoot); err != nil {
		config.base.Logger.Debug("Phase 4: Failed", zap.String("phase", "phase4"), zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Failed to switch: %v", err))
		return 1
	}
	config.base.Logger.Debug("Phase 4: Completed", zap.String("phase", "phase4"))

	// Phase 5: Update target branch
	config.base.Logger.Debug("Phase 5: Starting update target branch",
		zap.String("phase", "phase5"),
		zap.String("targetBranch", config.targetBranch))
	config.base.Logger.Info(fmt.Sprintf("Updating %s from remote...", config.targetBranch))
	if err := updateTargetBranch(config.targetBranch); err != nil {
		config.base.Logger.Debug("Phase 5: Failed", zap.String("phase", "phase5"), zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Failed to update: %v", err))
		return 1
	}
	config.base.Logger.Debug("Phase 5: Completed", zap.String("phase", "phase5"))

	// Phase 6: Perform merge
	config.base.Logger.Debug("Phase 6: Starting perform merge",
		zap.String("phase", "phase6"),
		zap.String("currentBranch", config.currentBranch),
		zap.String("targetBranch", config.targetBranch),
		zap.Bool("squash", !config.noSquash))
	var mergeType string
	if config.noSquash {
		config.base.Logger.Info(fmt.Sprintf("\nMerging %s into %s (regular merge)...", config.currentBranch, config.targetBranch))
		if err := config.base.GitOps.Merge(config.currentBranch, false); err != nil {
			config.base.Logger.Debug("Phase 6: Failed", zap.String("phase", "phase6"), zap.Error(err))
			handleMergeError(config, err)
			return 1
		}
		mergeType = "fast-forward"
	} else {
		config.base.Logger.Info(fmt.Sprintf("\nMerging %s into %s (squash)...", config.currentBranch, config.targetBranch))
		if err := performSquashMerge(config); err != nil {
			config.base.Logger.Debug("Phase 6: Failed", zap.String("phase", "phase6"), zap.Error(err))
			handleMergeError(config, err)
			return 1
		}
		mergeType = "squash"
	}
	config.base.Logger.Debug("Phase 6: Completed",
		zap.String("phase", "phase6"),
		zap.String("mergeType", mergeType))

	// Phase 7: Cleanup workspace
	config.base.Logger.Debug("Phase 7: Starting cleanup workspace",
		zap.String("phase", "phase7"),
		zap.Bool("keepWorktree", config.keepWorktree),
		zap.String("worktreePath", config.worktreePath))
	if !config.keepWorktree {
		config.base.Logger.Info("\nRemoving workspace...")
		if err := removeWorkspace(config); err != nil {
			config.base.Logger.Debug("Phase 7: Failed to remove workspace",
				zap.String("phase", "phase7"),
				zap.Error(err))
			config.base.Logger.Warn(fmt.Sprintf("Failed to remove workspace: %v", err))
		} else {
			config.base.Logger.Info(fmt.Sprintf("✓ Removed workspace: %s", config.worktreePath))
			config.base.Logger.Debug("Phase 7: Completed",
				zap.String("phase", "phase7"),
				zap.String("worktreePath", config.worktreePath))
		}
	} else {
		config.base.Logger.Info(fmt.Sprintf("\n✓ Workspace preserved at: %s", config.worktreePath))
		config.base.Logger.Debug("Phase 7: Completed (workspace preserved)",
			zap.String("phase", "phase7"),
			zap.String("worktreePath", config.worktreePath))
	}

	// Phase 8: Success
	duration := time.Since(startTime)
	config.base.Logger.Info("")
	config.base.Logger.Info(fmt.Sprintf("✓ Merged %s into %s (%s)", config.currentBranch, config.targetBranch, mergeType))
	config.base.Logger.Debug("Phase 8: Completed",
		zap.String("phase", "phase8"),
		zap.String("currentBranch", config.currentBranch),
		zap.String("targetBranch", config.targetBranch),
		zap.String("mergeType", mergeType),
		zap.Duration("totalDuration", duration))
	return 0
}

// mergeConfig holds configuration for the merge command.
type mergeConfig struct {
	base          *internal.BaseConfig
	targetBranch  string
	noSquash      bool
	keepWorktree  bool
	currentBranch string
	worktreePath  string
}

// parseMergeConfig parses command line arguments.
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
	if targetValue := flags.GetFlagValue(args, "--target"); targetValue != "" {
		config.targetBranch = targetValue
	}
	config.noSquash = flags.HasFlag(args, "--no-squash", "")
	config.keepWorktree = flags.HasFlag(args, "--keep-worktree", "")

	// Get current branch from current working directory (not repoRoot)
	// This ensures we get the correct branch in worktree environments
	// Check R2R_PWD first (for test isolation)
	cwd := os.Getenv("R2R_PWD")
	if cwd == "" {
		// Fall back to actual working directory
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	currentBranch, err := config.base.GitOps.GetCurrentBranch(cwd)
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

// validateMergeEnvironment validates the environment before merging.
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
	// Check R2R_PWD first (for test isolation)
	cwd := os.Getenv("R2R_PWD")
	if cwd == "" {
		// Fall back to actual working directory
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	clean, err := config.base.GitOps.IsWorktreeClean(cwd)
	if err != nil {
		return fmt.Errorf("failed to check working tree status: %w", err)
	}
	if !clean {
		return fmt.Errorf("uncommitted changes detected\nCommit or stash your changes before merging")
	}

	return nil
}

// checkBranchUpToDate checks if the current branch is up to date with target.
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
// In multi-worktree setups, it finds and switches to the worktree where the target branch is checked out.
func switchToTargetBranch(targetBranch, repoRoot string) error {
	// First, find where the target branch is checked out
	targetWorktree, err := findWorktreeForBranch(targetBranch)
	if err != nil {
		// Branch not checked out anywhere, try normal checkout
		_, checkoutErr := gitexec.Run(repoRoot, "checkout", targetBranch)
		if checkoutErr != nil {
			return fmt.Errorf("failed to switch to %s: %w", targetBranch, checkoutErr)
		}
		return nil
	}

	// Branch is checked out in a worktree - switch to that directory
	if err := os.Chdir(targetWorktree); err != nil {
		return fmt.Errorf("failed to switch to worktree at %s: %w", targetWorktree, err)
	}

	log.Infof("📂 Switched to worktree: %s", targetWorktree)
	return nil
}

// findWorktreeForBranch finds the worktree path where a branch is checked out
// Returns empty string and error if branch is not checked out in any worktree.
func findWorktreeForBranch(branch string) (string, error) {
	output, err := gitexec.Run(".", "worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w", err)
	}

	lines := string(output)
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

// updateTargetBranch updates the target branch from remote.
func updateTargetBranch(targetBranch string) error {
	_, err := gitexec.Run(".", "pull", "origin", targetBranch)
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", targetBranch, err)
	}
	return nil
}

// performSquashMerge performs a squash merge and uses commit message for the message.
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

	exitCode := commitmessage.CreateCommitMessage()
	if exitCode != 0 {
		return fmt.Errorf("commit message failed with exit code %d", exitCode)
	}

	return nil
}

// removeWorkspace removes the workspace and deletes the branch.
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

// getWorktreePath returns the path of the current worktree.
func getWorktreePath(base *internal.BaseConfig) (string, error) {
	// Get list of all worktrees
	worktrees, err := base.GitOps.ListWorktrees()
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Find current worktree path
	// Check R2R_PWD first (for test isolation)
	cwd := os.Getenv("R2R_PWD")
	if cwd == "" {
		var wdErr error
		cwd, wdErr = os.Getwd()
		if wdErr != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", wdErr)
		}
	}
	for _, wt := range worktrees {
		// Normalize paths for comparison
		if wt.Path == cwd || wt.Path == base.RepoRoot {
			return wt.Path, nil
		}
	}

	return base.RepoRoot, nil
}

// handleMergeError handles merge errors and provides guidance.
func handleMergeError(config *mergeConfig, err error) {
	config.base.Logger.Error("\n⚠️  Merge conflict detected\n")

	// Get conflicting files
	conflicts, conflictErr := config.base.GitOps.GetConflictingFiles()
	if conflictErr == nil && len(conflicts) > 0 {
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
