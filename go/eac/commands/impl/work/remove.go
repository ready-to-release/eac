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
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/work/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(Remove)
}

// Intent: Remove workspace from git tracking and clean up associated branches
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear flow: validate → switch branch → remove worktree → delete branch
//   - Explicit warnings for destructive operations
//   - Sensible defaults (delete local, keep remote)
//   - Informs user if manual folder deletion is needed
//
// Easy to change:
//   - Branch deletion is optional
//   - Remote deletion configurable
//   - Force flag for edge cases
//
// Hard to break:
//   - Validates uncommitted changes
//   - Prevents removing main workspace
//   - Confirms workspace exists before removing
//   - Switches away from workspace before removing it
//   - Does not force folder deletion to avoid data loss

// Remove removes a workspace and optionally deletes branches
func Remove() int {
	// Phase 1: Parse configuration
	config, err := parseRemoveConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Starting work remove command")
	internal.WriteDebugFile(config.base.Logger, config.base.RepoRoot, "remove-config.txt",
		fmt.Sprintf("Branch: %s\nKeepBranch: %v\nDeleteRemote: %v\nForce: %v\nPath: %s\n",
			config.branchName, config.keepBranch, config.deleteRemote, config.force, config.worktreePath))

	// Phase 2: Validate environment
	config.base.Logger.Info("Checking workspace status...")
	if err := validateRemoveEnvironment(config); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return 1
	}

	// Phase 3: Check for uncommitted changes
	if !config.force {
		clean, err := config.base.GitOps.IsWorktreeClean(config.worktreePath)
		if err != nil {
			config.base.Logger.Error(fmt.Sprintf("Failed to check workspace status: %v", err))
			return 1
		}
		if !clean {
			config.base.Logger.Error("Uncommitted changes detected")
			config.base.Logger.Error("Commit, stash, or use --force to discard changes")
			return 1
		}
	} else {
		// Warn about force flag
		clean, _ := config.base.GitOps.IsWorktreeClean(config.worktreePath)
		if !clean {
			config.base.Logger.Warn("⚠️  Warning: Uncommitted changes will be lost")
		}
	}

	// Phase 4: Switch to main if we're in the workspace being removed
	inWorkspace := isInWorkspace(config.worktreePath, config.base.RepoRoot)
	if inWorkspace {
		config.base.Logger.Info("Switching to main...")
		if err := switchToMain(config.base.RepoRoot); err != nil {
			config.base.Logger.Error(fmt.Sprintf("Failed to switch to main: %v", err))
			return 1
		}
	}

	// Phase 5: Remove workspace
	config.base.Logger.Info("Removing workspace...")
	if err := config.base.GitOps.RemoveWorktree(config.worktreePath); err != nil {
		config.base.Logger.Error(fmt.Sprintf("Failed to remove worktree: %v", err))
		return 1
	}

	// Check if folder still exists and inform user
	if _, err := os.Stat(config.worktreePath); err == nil {
		config.base.Logger.Info(fmt.Sprintf("ℹ️  Workspace folder still exists: %s", config.worktreePath))
		config.base.Logger.Info("   You can manually delete this folder if needed")
	}

	// Phase 6: Delete local branch (unless --keep-branch)
	if !config.keepBranch {
		config.base.Logger.Info(fmt.Sprintf("Deleting local branch %s...", config.branchName))
		if err := config.base.GitOps.DeleteBranch(config.branchName, config.force); err != nil {
			config.base.Logger.Warn(fmt.Sprintf("Failed to delete branch: %v", err))
		}
	} else {
		config.base.Logger.Info(fmt.Sprintf("Branch %s preserved", config.branchName))
	}

	// Phase 7: Delete remote branch (if --delete-remote)
	if config.deleteRemote {
		config.base.Logger.Info(fmt.Sprintf("Deleting remote branch %s...", config.branchName))
		if err := deleteRemoteBranch(config.base.RepoRoot, config.branchName); err != nil {
			config.base.Logger.Warn(fmt.Sprintf("Failed to delete remote branch: %v", err))
		} else {
			config.base.Logger.Info("Deleted remote branch")
		}
	} else {
		// Check if remote branch exists
		if remoteExists(config.base.RepoRoot, config.branchName) {
			config.base.Logger.Info(fmt.Sprintf("Remote branch origin/%s preserved", config.branchName))
		}
	}

	// Phase 8: Success
	config.base.Logger.Info("")
	config.base.Logger.Info(fmt.Sprintf("✓ Removed workspace for %s", config.branchName))
	config.base.Logger.Debug("Work remove command completed successfully")
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
	config.keepBranch = internal.HasFlag(args, "--keep-branch", "")
	config.deleteRemote = internal.HasFlag(args, "--delete-remote", "")
	config.force = internal.HasFlag(args, "--force", "-f")

	// Get positional arguments
	positionalArgs := internal.GetPositionalArgs(args)
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
		currentBranch, err := config.base.GitOps.GetCurrentBranch(config.base.RepoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
		config.branchName = currentBranch

		// Get current worktree path
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		config.worktreePath = cwd
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
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to switch to main: %w\nOutput: %s", err, string(output))
	}
	return nil
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
