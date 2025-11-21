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
// Long: Example:
// Long:   work remove                              # Remove current workspace
// Long:   work remove feature/old-feature          # Remove specific workspace
// Long:   work remove --keep-branch                # Keep local branch
// Long:   work remove --delete-remote              # Delete remote branch too
// Long:   work remove --force                      # Remove even with uncommitted changes
// Flag.keep-branch: type=bool, default=false, usage=Keep local branch after removing workspace
// Flag.delete-remote: type=bool, default=false, usage=Delete remote branch as well
// Flag.force: type=bool, shorthand=f, default=false, usage=Force remove even with uncommitted changes
// HasSideEffects: true
package work

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/repository"
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 2: Validate environment
	fmt.Println("Checking workspace status...")
	if err := validateRemoveEnvironment(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 3: Check for uncommitted changes
	if !config.force {
		clean, err := internal.IsWorktreeClean(config.worktreePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to check workspace status: %v\n", err)
			return 1
		}
		if !clean {
			fmt.Fprintf(os.Stderr, "Error: Uncommitted changes detected\n")
			fmt.Fprintf(os.Stderr, "Commit, stash, or use --force to discard changes\n")
			return 1
		}
	} else {
		// Warn about force flag
		clean, _ := internal.IsWorktreeClean(config.worktreePath)
		if !clean {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: Uncommitted changes will be lost\n")
		}
	}

	// Phase 4: Switch to main if we're in the workspace being removed
	inWorkspace := isInWorkspace(config.worktreePath, config.repoRoot)
	if inWorkspace {
		fmt.Println("Switching to main...")
		if err := switchToMain(config.repoRoot); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	// Phase 5: Remove workspace
	fmt.Println("Removing workspace...")
	if err := removeWorktree(config.worktreePath, config.force); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Check if folder still exists and inform user
	if _, err := os.Stat(config.worktreePath); err == nil {
		fmt.Printf("ℹ️  Workspace folder still exists: %s\n", config.worktreePath)
		fmt.Println("   You can manually delete this folder if needed")
	}

	// Phase 6: Delete local branch (unless --keep-branch)
	if !config.keepBranch {
		fmt.Printf("Deleting local branch %s...\n", config.branchName)
		if err := deleteBranch(config.branchName, config.force); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	} else {
		fmt.Printf("Branch %s preserved\n", config.branchName)
	}

	// Phase 7: Delete remote branch (if --delete-remote)
	if config.deleteRemote {
		fmt.Printf("Deleting remote branch %s...\n", config.branchName)
		if err := deleteRemoteBranch(config.branchName); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		} else {
			fmt.Println("Deleted remote branch")
		}
	} else {
		// Check if remote branch exists
		if remoteExists(config.branchName) {
			fmt.Printf("Remote branch origin/%s preserved\n", config.branchName)
		}
	}

	// Phase 8: Success
	fmt.Printf("\n✓ Removed workspace for %s\n", config.branchName)
	return 0
}

// removeConfig holds configuration for the remove command
type removeConfig struct {
	branchName   string // Branch to remove (from args or current branch)
	keepBranch   bool
	deleteRemote bool
	force        bool
	repoRoot     string
	worktreePath string
}

// parseRemoveConfig parses command line arguments
func parseRemoveConfig() (*removeConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "remove"

	config := &removeConfig{
		keepBranch:   false,
		deleteRemote: false,
		force:        false,
	}

	// Parse flags and branch name
	var branchArg string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--keep-branch" {
			config.keepBranch = true
		} else if arg == "--delete-remote" {
			config.deleteRemote = true
		} else if arg == "--force" || arg == "-f" {
			config.force = true
		} else if !strings.HasPrefix(arg, "--") && !strings.HasPrefix(arg, "-") {
			// This is the branch name argument
			branchArg = arg
		}
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}
	config.repoRoot = repoRoot

	// Determine which branch to remove
	if branchArg != "" {
		// Remove specific workspace by branch name
		config.branchName = branchArg
		worktree, err := internal.FindWorktreeByBranch(branchArg, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("workspace not found for branch %s: %w", branchArg, err)
		}
		config.worktreePath = worktree.Path
	} else {
		// Remove current workspace
		currentBranch, err := internal.GetCurrentBranch(repoRoot)
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

// removeWorktree removes the worktree from git tracking
func removeWorktree(path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// deleteBranch deletes the local branch
func deleteBranch(branch string, force bool) error {
	var cmd *exec.Cmd
	if force {
		cmd = exec.Command("git", "branch", "-D", branch)
	} else {
		cmd = exec.Command("git", "branch", "-d", branch)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// deleteRemoteBranch deletes the remote branch
func deleteRemoteBranch(branch string) error {
	cmd := exec.Command("git", "push", "origin", "--delete", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete remote branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// remoteExists checks if a branch exists on the remote
func remoteExists(branch string) bool {
	cmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}
