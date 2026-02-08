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
// Long: Expected Output:
// Long:   - Branch rebased onto latest target
// Long:   - Conflict instructions if conflicts occur
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
	"time"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/cli/eac/impl/work/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/environments"
)

// Pull syncs the current branch with target branch via rebase.
func Pull() int {
	startTime := time.Now()

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Phase 1: Parse configuration
	config, err := parsePullConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer func() { _ = config.base.Logger.Sync() }() //nolint:errcheck // best-effort logger sync

	logger := config.base.Logger
	logger.Debug("Phase 1: Completed",
		zap.String("phase", "phase1"),
		zap.String("currentBranch", config.currentBranch),
		zap.String("targetBranch", config.targetBranch),
		zap.Bool("autostash", config.autostash),
		zap.Bool("noFetch", config.noFetch))

	config.base.Logger.Debug("Starting work pull command",
		zap.String("currentBranch", config.currentBranch),
		zap.String("targetBranch", config.targetBranch),
		zap.Bool("autostash", config.autostash),
		zap.Bool("noFetch", config.noFetch))

	// Phase 2: Validate environment
	phase2Start := time.Now()
	config.base.Logger.Debug("Phase 2: Starting validate environment", zap.String("phase", "phase2"))

	if err := validatePullEnvironment(config); err != nil {
		config.base.Logger.Debug("Phase 2: Failed", zap.String("phase", "phase2"), zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}

	phase2Duration := time.Since(phase2Start)
	config.base.Logger.Debug("Phase 2: Completed", zap.String("phase", "phase2"), zap.Duration("duration", phase2Duration))

	// Phase 3: Handle uncommitted changes
	phase3Start := time.Now()
	config.base.Logger.Debug("Phase 3: Starting handle uncommitted changes", zap.String("phase", "phase3"))

	if config.autostash {
		if err := config.base.GitOps.Stash("work-pull autostash"); err != nil {
			config.base.Logger.Debug("Phase 3: Failed", zap.String("phase", "phase3"), zap.Error(err))
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

	phase3Duration := time.Since(phase3Start)
	config.base.Logger.Debug("Phase 3: Completed", zap.String("phase", "phase3"), zap.Duration("duration", phase3Duration))

	// Phase 4: Fetch target branch
	phase4Start := time.Now()
	config.base.Logger.Debug("Phase 4: Starting fetch target branch", zap.String("phase", "phase4"))

	if !config.noFetch {
		config.base.Logger.Info(fmt.Sprintf("Fetching origin/%s...", config.targetBranch))
		if err := config.base.GitOps.FetchBranch(config.targetBranch); err != nil {
			config.base.Logger.Debug("Phase 4: Failed", zap.String("phase", "phase4"), zap.Error(err))
			config.base.Logger.Error(fmt.Sprintf("Failed to fetch: %v", err))
			return 1
		}
	}

	phase4Duration := time.Since(phase4Start)
	config.base.Logger.Debug("Phase 4: Completed", zap.String("phase", "phase4"), zap.Duration("duration", phase4Duration))

	// Phase 5: Get rebase info
	phase5Start := time.Now()
	config.base.Logger.Debug("Phase 5: Starting get rebase info", zap.String("phase", "phase5"))

	info, err := getRebaseInfo(config)
	if err != nil {
		config.base.Logger.Debug("Phase 5: Failed", zap.String("phase", "phase5"), zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Failed to get rebase info: %v", err))
		return 1
	}

	phase5Duration := time.Since(phase5Start)
	config.base.Logger.Debug("Phase 5: Completed", zap.String("phase", "phase5"), zap.Duration("duration", phase5Duration))

	// Phase 6: Check if already up to date
	phase6Start := time.Now()
	config.base.Logger.Debug("Phase 6: Starting check if up to date", zap.String("phase", "phase6"))

	if info.upToDate {
		phase6Duration := time.Since(phase6Start)
		config.base.Logger.Debug("Phase 6: Completed", zap.String("phase", "phase6"), zap.Duration("duration", phase6Duration))
		config.base.Logger.Info("Already up to date")
		totalDuration := time.Since(startTime)
		config.base.Logger.Debug("Work pull command completed successfully", zap.Duration("totalDuration", totalDuration))
		return 0
	}

	phase6Duration := time.Since(phase6Start)
	config.base.Logger.Debug("Phase 6: Completed", zap.String("phase", "phase6"), zap.Duration("duration", phase6Duration))

	// Phase 7: Show rebase preview
	phase7Start := time.Now()
	config.base.Logger.Debug("Phase 7: Starting show rebase preview", zap.String("phase", "phase7"))

	showRebasePreview(config.currentBranch, config.targetBranch, info)

	phase7Duration := time.Since(phase7Start)
	config.base.Logger.Debug("Phase 7: Completed", zap.String("phase", "phase7"), zap.Duration("duration", phase7Duration))

	// Phase 8: Perform rebase
	phase8Start := time.Now()
	config.base.Logger.Debug("Phase 8: Starting perform rebase", zap.String("phase", "phase8"))
	config.base.Logger.Info(fmt.Sprintf("\nRebasing %s onto origin/%s...", config.currentBranch, config.targetBranch))

	if err := config.base.GitOps.Rebase(config.targetBranch); err != nil {
		config.base.Logger.Debug("Phase 8: Failed", zap.String("phase", "phase8"), zap.Error(err))
		handleRebaseError(config, err)
		return 1
	}

	phase8Duration := time.Since(phase8Start)
	config.base.Logger.Debug("Phase 8: Completed", zap.String("phase", "phase8"), zap.Duration("duration", phase8Duration))

	// Phase 9: Success
	phase9Start := time.Now()
	config.base.Logger.Debug("Phase 9: Starting report success", zap.String("phase", "phase9"))

	config.base.Logger.Info("")
	config.base.Logger.Info(fmt.Sprintf("✓ Rebased %s onto origin/%s", config.currentBranch, config.targetBranch))
	config.base.Logger.Info(fmt.Sprintf("  Your commits: %d", info.currentCommits))
	config.base.Logger.Info(fmt.Sprintf("  New commits from %s: %d", config.targetBranch, info.newCommits))

	phase9Duration := time.Since(phase9Start)
	config.base.Logger.Debug("Phase 9: Completed", zap.String("phase", "phase9"), zap.Duration("duration", phase9Duration))

	totalDuration := time.Since(startTime)
	config.base.Logger.Debug("Work pull command completed successfully", zap.Duration("totalDuration", totalDuration))
	return 0
}

// pullConfig holds configuration for the pull command.
type pullConfig struct {
	base          *internal.BaseConfig
	targetBranch  string
	autostash     bool
	noFetch       bool
	currentBranch string
}

// rebaseInfo holds information about the rebase operation.
type rebaseInfo struct {
	currentCommits int
	newCommits     int
	upToDate       bool
}

// parsePullConfig parses command line arguments.
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
	if targetValue := flags.GetFlagValue(args, "--target"); targetValue != "" {
		config.targetBranch = targetValue
	}
	config.autostash = flags.HasFlag(args, "--autostash", "")
	config.noFetch = flags.HasFlag(args, "--no-fetch", "")

	// Get current branch from current working directory (not repoRoot)
	// This ensures we get the correct branch in worktree environments
	// Check CLIE_PWD first (for test isolation)
	cwd := os.Getenv(environments.EnvCLIEPWD)
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

	return config, nil
}

// validatePullEnvironment validates the environment before pulling.
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
		// Check CLIE_PWD first (for test isolation)
		cwd := os.Getenv(environments.EnvCLIEPWD)
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
			return fmt.Errorf("uncommitted changes detected\nCommit or stash your changes, or use --autostash")
		}
	}

	return nil
}

// getRebaseInfo gets information about the rebase operation.
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

// showRebasePreview shows what will be rebased.
func showRebasePreview(currentBranch, targetBranch string, info *rebaseInfo) {
	log.Infof("\nRebasing %s onto origin/%s", currentBranch, targetBranch)
	log.Infof("  Current branch: %d commits ahead of %s", info.currentCommits, targetBranch)
	log.Infof("  %s branch: %d new commits", targetBranch, info.newCommits)
	if info.currentCommits > 0 && info.newCommits > 0 {
		log.Infof("  Rebase will replay your %d commits on top of %d new commits", info.currentCommits, info.newCommits)
	}
}

// handleRebaseError handles rebase errors and provides guidance.
func handleRebaseError(config *pullConfig, err error) {
	config.base.Logger.Error("\n⚠️  Rebase conflict detected\n")

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
	config.base.Logger.Error("  git rebase --continue")
	config.base.Logger.Error("")
	config.base.Logger.Error("Or abort:")
	config.base.Logger.Error("  git rebase --abort")
}
