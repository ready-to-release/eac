// Command: work remove
// Short: Remove workspace and optionally delete associated branches
// Long: Removes a workspace and optionally deletes the associated local and remote branches.
// Long:
// Long: By default, this command:
// Long:   1. Validates the workspace can be safely removed
// Long:   2. Switches to main branch (if in the workspace being removed)
// Long:   3. Removes the workspace from git tracking
// Long:   4. Deletes the local branch
// Long:   5. Preserves the remote branch
// Long:   6. Informs you if the workspace folder still exists for manual deletion
// Long:
// Long: Use --debug to enable detailed logging to out/logs/work/.
// Long:
// Long: Expected Output:
// Long:   - Workspace removed from git tracking
// Long:   - Local branch deleted (unless --keep-branch)
// Long:   - Information about manual folder deletion if needed
// Long:
// Long: Example:
// Long:   work remove                              # Remove current workspace
// Long:   work remove feature/old-feature          # Remove specific workspace
// Long:   work remove --keep-branch                # Keep local branch
// Long:   work remove --delete-remote              # Delete remote branch too
// Long:   work remove --force                      # Remove even with uncommitted changes
// Long:   work remove --debug                      # Enable debug logging
// Flag.keep-branch: type=bool, default=false, usage=Keep local branch after removing workspace
// Flag.delete-remote: type=bool, default=false, usage=Delete remote branch as well
// Flag.force: type=bool, shorthand=f, default=false, usage=Force remove even with uncommitted changes
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
package work

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/impl/work/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(Remove)
}

// Remove removes a workspace and optionally deletes branches
func Remove() int {
	startTime := time.Now()

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Phase 1: Parse configuration
	phaseStart := time.Now()
	config, err := parseRemoveConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Phase 1: Starting parse configuration", zap.String("phase", "phase1"))

	config.base.Logger.Debug("Phase 1: Completed",
		zap.String("phase", "phase1"),
		zap.Duration("duration", time.Since(phaseStart)),
		zap.String("branchName", config.branchName),
		zap.String("worktreePath", config.worktreePath),
		zap.Bool("keepBranch", config.keepBranch),
		zap.Bool("deleteRemote", config.deleteRemote),
		zap.Bool("force", config.force))

	// Phase 2: Validate environment
	phaseStart = time.Now()
	config.base.Logger.Debug("Phase 2: Starting validate environment", zap.String("phase", "phase2"))
	config.base.Logger.Info("Checking workspace status...")
	if err := validateRemoveEnvironment(config); err != nil {
		config.base.Logger.Debug("Phase 2: Failed",
			zap.String("phase", "phase2"),
			zap.Duration("duration", time.Since(phaseStart)),
			zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}
	config.base.Logger.Debug("Phase 2: Completed",
		zap.String("phase", "phase2"),
		zap.Duration("duration", time.Since(phaseStart)))

	// Phase 3: Check for uncommitted changes
	phaseStart = time.Now()
	config.base.Logger.Debug("Phase 3: Starting check uncommitted changes", zap.String("phase", "phase3"))
	if !config.force {
		clean, err := config.base.GitOps.IsWorktreeClean(config.worktreePath)
		if err != nil {
			config.base.Logger.Debug("Phase 3: Failed",
				zap.String("phase", "phase3"),
				zap.Duration("duration", time.Since(phaseStart)),
				zap.Error(err))
			config.base.Logger.Error(fmt.Sprintf("Failed to check workspace status: %v", err))
			return 1
		}
		if !clean {
			config.base.Logger.Debug("Phase 3: Failed - uncommitted changes",
				zap.String("phase", "phase3"),
				zap.Duration("duration", time.Since(phaseStart)))
			config.base.Logger.Error("Uncommitted changes detected")
			config.base.Logger.Error("Commit, stash, or use --force to discard changes")
			return 1
		}
		config.base.Logger.Debug("Phase 3: Completed - worktree clean",
			zap.String("phase", "phase3"),
			zap.Duration("duration", time.Since(phaseStart)))
	} else {
		// Warn about force flag
		clean, _ := config.base.GitOps.IsWorktreeClean(config.worktreePath)
		if !clean {
			config.base.Logger.Warn("⚠️  Warning: Uncommitted changes will be lost")
			config.base.Logger.Debug("Phase 3: Completed - force mode with uncommitted changes",
				zap.String("phase", "phase3"),
				zap.Duration("duration", time.Since(phaseStart)))
		} else {
			config.base.Logger.Debug("Phase 3: Completed - force mode but worktree clean",
				zap.String("phase", "phase3"),
				zap.Duration("duration", time.Since(phaseStart)))
		}
	}

	// Phase 4: Note about switching (actual directory change happens outside this command)
	// When running from within the workspace being removed, the caller is responsible
	// for changing to a safe directory after this command completes.
	phaseStart = time.Now()
	config.base.Logger.Debug("Phase 4: Starting check workspace location", zap.String("phase", "phase4"))
	inWorkspace := isInWorkspace(config.worktreePath, config.base.RepoRoot)
	if inWorkspace {
		config.base.Logger.Info("Switching to main...")
		config.base.Logger.Debug("Note: You are currently in the workspace being removed")
		config.base.Logger.Debug("Your shell will need to change directories after removal")
		config.base.Logger.Debug("Phase 4: Completed - in workspace being removed",
			zap.String("phase", "phase4"),
			zap.Duration("duration", time.Since(phaseStart)),
			zap.Bool("inWorkspace", true))
	} else {
		config.base.Logger.Debug("Phase 4: Completed - not in workspace",
			zap.String("phase", "phase4"),
			zap.Duration("duration", time.Since(phaseStart)),
			zap.Bool("inWorkspace", false))
	}

	// Phase 5: Remove workspace
	phaseStart = time.Now()
	config.base.Logger.Debug("Phase 5: Starting remove workspace", zap.String("phase", "phase5"))
	config.base.Logger.Info("Removing workspace...")

	// Use git to find the actual common git directory (works correctly in worktrees)
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = config.worktreePath
	output, err := cmd.Output()
	if err != nil {
		config.base.Logger.Debug("Phase 5: Failed - git common dir detection",
			zap.String("phase", "phase5"),
			zap.Duration("duration", time.Since(phaseStart)),
			zap.Error(err))
		config.base.Logger.Error(fmt.Sprintf("Failed to detect git directory: %v", err))
		return 1
	}
	gitCommonDir := strings.TrimSpace(string(output))

	// The common git dir's parent is the main repository root
	actualRepoRoot := filepath.Dir(gitCommonDir)
	config.base.Logger.Debug(fmt.Sprintf("Detected actual repository root: %s", actualRepoRoot),
		zap.String("gitCommonDir", gitCommonDir),
		zap.String("actualRepoRoot", actualRepoRoot))

	// Run git worktree remove from the detected repository root
	// On Windows, use --force by default to handle locked files
	cmd = exec.Command("git", "worktree", "remove", "--force", config.worktreePath)
	cmd.Dir = actualRepoRoot
	output, err = cmd.CombinedOutput()
	if err != nil {
		// On Windows, even with --force, git may fail to delete locked directories
		// Check if it's a permission error and if the worktree was successfully unregistered
		outputStr := string(output)
		if strings.Contains(outputStr, "Permission denied") || strings.Contains(outputStr, "failed to delete") {
			// Check if worktree is still registered
			cmd = exec.Command("git", "worktree", "list")
			cmd.Dir = actualRepoRoot
			listOutput, listErr := cmd.Output()
			if listErr == nil && !strings.Contains(string(listOutput), config.worktreePath) {
				// Worktree was successfully unregistered, just couldn't delete directory
				config.base.Logger.Debug("Worktree unregistered but directory couldn't be deleted",
					zap.String("output", outputStr))
				// Continue execution - folder will be reported as still existing below
			} else {
				// Worktree is still registered, this is a real failure
				config.base.Logger.Debug("Phase 5: Failed - worktree still registered",
					zap.String("phase", "phase5"),
					zap.Duration("duration", time.Since(phaseStart)),
					zap.Error(err))
				config.base.Logger.Error(fmt.Sprintf("Failed to remove worktree: %v\nOutput: %s", err, outputStr))
				return 1
			}
		} else {
			// Some other error
			config.base.Logger.Debug("Phase 5: Failed - worktree remove",
				zap.String("phase", "phase5"),
				zap.Duration("duration", time.Since(phaseStart)),
				zap.Error(err))
			config.base.Logger.Error(fmt.Sprintf("Failed to remove worktree: %v\nOutput: %s", err, string(output)))
			return 1
		}
	}

	// Check if folder still exists and inform user
	folderExists := false
	if _, err := os.Stat(config.worktreePath); err == nil {
		folderExists = true
		config.base.Logger.Info(fmt.Sprintf("ℹ️  Workspace folder still exists: %s", config.worktreePath))
		config.base.Logger.Info("   You can manually delete this folder if needed")
	}

	config.base.Logger.Debug("Phase 5: Completed",
		zap.String("phase", "phase5"),
		zap.Duration("duration", time.Since(phaseStart)),
		zap.Bool("folderStillExists", folderExists))

	// Phase 6: Delete local branch (unless --keep-branch)
	phaseStart = time.Now()
	config.base.Logger.Debug("Phase 6: Starting delete local branch", zap.String("phase", "phase6"))
	if !config.keepBranch {
		config.base.Logger.Info(fmt.Sprintf("Deleting local branch %s...", config.branchName))
		if err := config.base.GitOps.DeleteBranch(config.branchName, config.force); err != nil {
			config.base.Logger.Warn(fmt.Sprintf("Failed to delete branch: %v", err))
			config.base.Logger.Debug("Phase 6: Completed - branch deletion failed",
				zap.String("phase", "phase6"),
				zap.Duration("duration", time.Since(phaseStart)),
				zap.Error(err))
		} else {
			config.base.Logger.Debug("Phase 6: Completed - branch deleted",
				zap.String("phase", "phase6"),
				zap.Duration("duration", time.Since(phaseStart)))
		}
	} else {
		config.base.Logger.Info(fmt.Sprintf("Branch %s preserved", config.branchName))
		config.base.Logger.Debug("Phase 6: Completed - branch preserved",
			zap.String("phase", "phase6"),
			zap.Duration("duration", time.Since(phaseStart)))
	}

	// Phase 7: Delete remote branch (if --delete-remote)
	phaseStart = time.Now()
	config.base.Logger.Debug("Phase 7: Starting delete remote branch", zap.String("phase", "phase7"))
	if config.deleteRemote {
		config.base.Logger.Info(fmt.Sprintf("Deleting remote branch %s...", config.branchName))
		if err := deleteRemoteBranch(config.base.RepoRoot, config.branchName); err != nil {
			config.base.Logger.Warn(fmt.Sprintf("Failed to delete remote branch: %v", err))
			config.base.Logger.Debug("Phase 7: Completed - remote deletion failed",
				zap.String("phase", "phase7"),
				zap.Duration("duration", time.Since(phaseStart)),
				zap.Error(err))
		} else {
			config.base.Logger.Info("Deleted remote branch")
			config.base.Logger.Debug("Phase 7: Completed - remote deleted",
				zap.String("phase", "phase7"),
				zap.Duration("duration", time.Since(phaseStart)))
		}
	} else {
		// Check if remote branch exists
		remoteExistsFlag := remoteExists(config.base.RepoRoot, config.branchName)
		if remoteExistsFlag {
			config.base.Logger.Info(fmt.Sprintf("Remote branch origin/%s preserved", config.branchName))
		}
		config.base.Logger.Debug("Phase 7: Completed - remote preserved",
			zap.String("phase", "phase7"),
			zap.Duration("duration", time.Since(phaseStart)),
			zap.Bool("remoteExists", remoteExistsFlag))
	}

	// Phase 8: Success
	config.base.Logger.Info("")
	config.base.Logger.Info(fmt.Sprintf("✓ Removed workspace for %s", config.branchName))
	config.base.Logger.Debug("Work remove command completed successfully",
		zap.Duration("totalDuration", time.Since(startTime)))
	return 0
}

// removeConfig holds configuration for the remove command
type removeConfig struct {
	base         *internal.BaseConfig
	branchName   string // Branch to remove (from args or current branch)
	keepBranch   bool
	deleteRemote bool
	force        bool
	worktreePath string
}

// parseRemoveConfig parses command line arguments
func parseRemoveConfig() (*removeConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "remove"

	// Parse base config (debug flag, repo root, logger, git ops)
	baseConfig, err := internal.ParseBaseConfig(args)
	if err != nil {
		return nil, err
	}

	config := &removeConfig{
		base:         baseConfig,
		keepBranch:   false,
		deleteRemote: false,
		force:        false,
	}

	// Parse flags
	config.keepBranch = flags.HasFlag(args, "--keep-branch", "")
	config.deleteRemote = flags.HasFlag(args, "--delete-remote", "")
	config.force = flags.HasFlag(args, "--force", "")

	// Get positional arguments
	positionalArgs := flags.GetPositionalArgs(args)
	var branchArg string
	if len(positionalArgs) > 0 {
		branchArg = positionalArgs[0]
	}

	// Determine which branch to remove
	if branchArg != "" {
		// Remove specific workspace by branch name
		config.branchName = branchArg
		worktree, err := internal.FindWorktreeByBranch(branchArg, config.base.RepoRoot)
		if err != nil {
			return nil, fmt.Errorf("workspace not found for branch %s: %w", branchArg, err)
		}
		config.worktreePath = worktree.Path
	} else {
		// Remove current workspace
		// Get current working directory - check R2R_PWD first (for test isolation)
		cwd := os.Getenv("R2R_PWD")
		if cwd == "" {
			// Fall back to actual working directory
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current directory: %w", err)
			}
		}
		config.worktreePath = cwd

		// Get branch from current working directory (not repo root)
		currentBranch, err := config.base.GitOps.GetCurrentBranch(cwd)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
		config.branchName = currentBranch
	}

	return config, nil
}

// validateRemoveEnvironment validates the environment before removing
func validateRemoveEnvironment(config *removeConfig) error {
	// Check we're in a git repository
	if err := internal.EnsureInGitRepo(); err != nil {
		return err
	}

	// Prevent removing main workspace
	if config.branchName == "main" || config.branchName == "master" {
		return fmt.Errorf("cannot remove main workspace\nYou are trying to remove the main branch workspace.")
	}

	return nil
}

// isInWorkspace checks if the current directory is in the workspace being removed
func isInWorkspace(worktreePath, repoRoot string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	// Normalize paths for comparison
	return strings.EqualFold(cwd, worktreePath) || strings.HasPrefix(cwd, worktreePath)
}

// switchToMain switches to the main branch in the main workspace
func switchToMain(repoRoot string) error {
	// First verify which branch exists
	// Try "main" first (preferred for modern git repositories)
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/main")
	cmd.Dir = repoRoot
	mainExists := cmd.Run() == nil

	cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/master")
	cmd.Dir = repoRoot
	masterExists := cmd.Run() == nil

	// Try to checkout main if it exists
	if mainExists {
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = repoRoot
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		// If main exists but checkout failed, return detailed error
		return fmt.Errorf("branch 'main' exists but checkout failed: %w\nOutput: %s", err, string(output))
	}

	// Try to checkout master if it exists
	if masterExists {
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = repoRoot
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		return fmt.Errorf("branch 'master' exists but checkout failed: %w\nOutput: %s", err, string(output))
	}

	// Neither branch exists - this shouldn't happen in a valid repository
	return fmt.Errorf("neither 'main' nor 'master' branch found in repository at %s", repoRoot)
}

// getDefaultBranch detects the default/trunk branch name (main or master)
func getDefaultBranch(repoRoot string) string {
	// Try to get the default branch from git config
	cmd := exec.Command("git", "config", "--get", "init.defaultBranch")
	cmd.Dir = repoRoot
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		branch := strings.TrimSpace(string(output))
		if branch != "" {
			return branch
		}
	}

	// Check which branches actually exist locally using git branch --list
	// This is more reliable than rev-parse in worktree contexts
	cmd = exec.Command("git", "branch", "--list")
	cmd.Dir = repoRoot
	if output, err := cmd.Output(); err == nil {
		branches := string(output)
		// Check for main first (preferred)
		if strings.Contains(branches, " main\n") || strings.Contains(branches, "* main\n") {
			return "main"
		}
		// Fall back to master if main doesn't exist
		if strings.Contains(branches, " master\n") || strings.Contains(branches, "* master\n") {
			return "master"
		}
	}

	// Default to main (safest default for modern git)
	return "main"
}

// deleteRemoteBranch deletes the remote branch
func deleteRemoteBranch(repoRoot, branch string) error {
	cmd := exec.Command("git", "push", "origin", "--delete", branch)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete remote branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// remoteExists checks if a branch exists on the remote
func remoteExists(repoRoot, branch string) bool {
	cmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}
