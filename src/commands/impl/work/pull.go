// Command: work pull
// Short: Sync workspace with latest main via rebase
// Long: Fetches the latest changes from the target branch (default: main) and rebases
// Long: the current branch onto it, keeping your commit history linear.
// Long:
// Long: This command:
// Long:   1. Fetches latest changes from origin/main
// Long:   2. Rebases your commits on top of the fetched changes
// Long:   3. Handles conflicts with clear instructions
// Long:
// Long: Use --autostash to automatically stash uncommitted changes before rebasing.
// Long: Use --debug to enable detailed logging to out/logs/work/.
// Long:
// Long: Example:
// Long:   work pull
// Long:   work pull --target=develop
// Long:   work pull --autostash
// Long:   work pull --debug
// Flag.target: type=string, default=main, usage=Target branch to rebase onto
// Flag.autostash: type=bool, default=false, usage=Automatically stash and unstash uncommitted changes
// Flag.no-fetch: type=bool, default=false, usage=Skip fetching from remote (use local target branch)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
package work

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/logging"
)

func init() {
	registry.Register(Pull)
}

// Intent: Sync workspace branch with latest main via rebase
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear flow: validate → fetch → check status → rebase → report
//   - Explicit error messages for each failure case
//   - Progress feedback at each step
//
// Easy to change:
//   - Autostash logic isolated
//   - Rebase execution separated from status checking
//   - Target branch configurable
//
// Hard to break:
//   - Validates uncommitted changes before rebasing
//   - Detects and reports conflicts clearly
//   - Provides resolution instructions
//   - Prevents rebasing main onto itself

// Pull syncs the current branch with target branch via rebase
func Pull() int {
	// Phase 1: Parse configuration
	config, err := parsePullConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Starting work pull command")
	internal.WriteDebugFile(config.base.Logger, config.base.RepoRoot, "pull-config.txt",
		fmt.Sprintf("CurrentBranch: %s\nTarget: %s\nAutostash: %v\nNoFetch: %v\n",
			config.currentBranch, config.targetBranch, config.autostash, config.noFetch))

	// Phase 2: Validate environment
	if err := validatePullEnvironment(config); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}

	// Phase 3: Handle uncommitted changes
	if config.autostash {
		if err := config.base.GitOps.Stash("work-pull autostash"); err != nil {
			config.base.Logger.Error(fmt.Sprintf("Failed to stash changes: %v", err))
			return 1
		}
		config.base.Logger.Info("Stashed uncommitted changes")
		defer func() {
			if err := config.base.GitOps.StashPop(); err != nil {
				config.base.Logger.Warn(fmt.Sprintf("Failed to reapply stashed changes: %v", err))
				config.base.Logger.Warn("You can manually apply with: git stash pop")
			} else {
				config.base.Logger.Info("Reapplied stashed changes")
			}
		}()
	}

	// Phase 4: Fetch target branch
	if !config.noFetch {
		config.base.Logger.Info(fmt.Sprintf("Fetching origin/%s...", config.targetBranch))
		if err := config.base.GitOps.FetchBranch(config.targetBranch); err != nil {
			config.base.Logger.Error(fmt.Sprintf("Failed to fetch: %v", err))
			return 1
		}
	}

	// Phase 5: Get rebase info
	info, err := getRebaseInfo(config)
	if err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to get rebase info: %v", err))
		return 1
	}

	// Phase 6: Check if already up to date
	if info.upToDate {
		config.base.Logger.Info("Already up to date")
		return 0
	}

	// Phase 7: Show rebase preview
	showRebasePreview(config.base.Logger, config.currentBranch, config.targetBranch, info)

	// Phase 8: Perform rebase
	config.base.Logger.Info(fmt.Sprintf("\nRebasing %s onto origin/%s...", config.currentBranch, config.targetBranch))
	if err := config.base.GitOps.Rebase(config.targetBranch); err != nil {
		handleRebaseError(config, err)
		return 1
	}

	// Phase 9: Success
	config.base.Logger.Info("")
	config.base.Logger.Info(fmt.Sprintf("✓ Rebased %s onto origin/%s", config.currentBranch, config.targetBranch))
	config.base.Logger.Info(fmt.Sprintf("  Your commits: %d", info.currentCommits))
	config.base.Logger.Info(fmt.Sprintf("  New commits from %s: %d", config.targetBranch, info.newCommits))
	config.base.Logger.Debug("Work pull command completed successfully")
	return 0
}

// pullConfig holds configuration for the pull command
type pullConfig struct {
	base          *internal.BaseConfig
	targetBranch  string
	autostash     bool
	noFetch       bool
	currentBranch string
}

// rebaseInfo holds information about the rebase operation
type rebaseInfo struct {
	currentCommits int
	newCommits     int
	upToDate       bool
}

// parsePullConfig parses command line arguments
func parsePullConfig() (*pullConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "pull"

	// Parse base config (debug flag, repo root, logger, git ops)
	baseConfig, err := internal.ParseBaseConfig(args)
	if err != nil {
		return nil, err
	}

	config := &pullConfig{
		base:         baseConfig,
		targetBranch: "main",
		autostash:    false,
		noFetch:      false,
	}

	// Parse flags
	if targetValue := internal.GetFlagValue(args, "--target"); targetValue != "" {
		config.targetBranch = targetValue
	}
	config.autostash = internal.HasFlag(args, "--autostash", "")
	config.noFetch = internal.HasFlag(args, "--no-fetch", "")

	// Get current branch from current working directory (not repoRoot)
	// This ensures we get the correct branch in worktree environments
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	currentBranch, err := config.base.GitOps.GetCurrentBranch(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	config.currentBranch = currentBranch

	return config, nil
}

// validatePullEnvironment validates the environment before pulling
func validatePullEnvironment(config *pullConfig) error {
	// Check we're in a git repository
	if err := internal.EnsureInGitRepo(); err != nil {
		return err
	}

	// Prevent rebasing main onto itself
	if config.currentBranch == "main" || config.currentBranch == "master" {
		return fmt.Errorf("cannot rebase main onto itself\nYou are on the main branch. Switch to a feature branch first.")
	}

	// Check for uncommitted changes if not using autostash
	if !config.autostash {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		clean, err := config.base.GitOps.IsWorktreeClean(cwd)
		if err != nil {
			return fmt.Errorf("failed to check working tree status: %w", err)
		}
		if !clean {
			return fmt.Errorf("uncommitted changes detected\nCommit or stash your changes, or use --autostash")
		}
	}

	return nil
}

// getRebaseInfo gets information about the rebase operation
func getRebaseInfo(config *pullConfig) (*rebaseInfo, error) {
	info := &rebaseInfo{}

	// Count commits on current branch ahead of target
	currentCommits, err := config.base.GitOps.GetCommitCount(fmt.Sprintf("origin/%s", config.targetBranch), "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to count current commits: %w", err)
	}
	info.currentCommits = currentCommits

	// Count new commits on target
	newCommits, err := config.base.GitOps.GetCommitCount("HEAD", fmt.Sprintf("origin/%s", config.targetBranch))
	if err != nil {
		return nil, fmt.Errorf("failed to count new commits: %w", err)
	}
	info.newCommits = newCommits

	// Check if already up to date
	info.upToDate = info.newCommits == 0

	return info, nil
}

// showRebasePreview shows what will be rebased
func showRebasePreview(logger *logging.Logger, currentBranch, targetBranch string, info *rebaseInfo) {
	logger.Info(fmt.Sprintf("\nRebasing %s onto origin/%s", currentBranch, targetBranch))
	logger.Info(fmt.Sprintf("  Current branch: %d commits ahead of %s", info.currentCommits, targetBranch))
	logger.Info(fmt.Sprintf("  %s branch: %d new commits", targetBranch, info.newCommits))
	if info.currentCommits > 0 && info.newCommits > 0 {
		logger.Info(fmt.Sprintf("  Rebase will replay your %d commits on top of %d new commits", info.currentCommits, info.newCommits))
	}
}

// handleRebaseError handles rebase errors and provides guidance
func handleRebaseError(config *pullConfig, err error) {
	config.base.Logger.Error("\n⚠️  Rebase conflict detected\n")

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
	config.base.Logger.Error("  git rebase --continue")
	config.base.Logger.Error("")
	config.base.Logger.Error("Or abort:")
	config.base.Logger.Error("  git rebase --abort")
}
