// Command: work merge
// Short: Merge workspace changes back to main (squash by default)
// Long: Merges the current workspace branch back into the target branch (default: main)
// Long: using squash merge to create a single, well-documented commit.
// Long:
// Long: By default, this command:
// Long:   1. Validates workspace is clean and up to date
// Long:   2. Switches to target branch and updates it
// Long:   3. Squash merges all workspace commits into a single commit
// Long:   4. Uses commit-ai to generate a comprehensive commit message
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
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode (pass through to commit-ai)
// HasSideEffects: true
package work

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/commit"
	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/repository"
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
//   - Uses commit-ai for high-quality squash commit messages
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 2: Validate environment
	if err := validateMergeEnvironment(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 3: Check branch is up to date
	if err := checkBranchUpToDate(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 4: Switch to target branch
	fmt.Printf("Switching to %s...\n", config.targetBranch)
	if err := switchToTargetBranch(config.targetBranch, config.repoRoot); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 5: Update target branch
	fmt.Printf("Updating %s from remote...\n", config.targetBranch)
	if err := updateTargetBranch(config.targetBranch); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 6: Perform merge
	var mergeType string
	if config.noSquash {
		fmt.Printf("\nMerging %s into %s (regular merge)...\n", config.currentBranch, config.targetBranch)
		if err := performRegularMerge(config.currentBranch); err != nil {
			handleMergeError(err, config)
			return 1
		}
		mergeType = "fast-forward"
	} else {
		fmt.Printf("\nMerging %s into %s (squash)...\n", config.currentBranch, config.targetBranch)
		if err := performSquashMerge(config); err != nil {
			handleMergeError(err, config)
			return 1
		}
		mergeType = "squash"
	}

	// Phase 7: Cleanup workspace
	if !config.keepWorktree {
		fmt.Printf("\nRemoving workspace...\n")
		if err := removeWorkspace(config.worktreePath, config.currentBranch); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		} else {
			fmt.Printf("✓ Removed workspace: %s\n", config.worktreePath)
		}
	} else {
		fmt.Printf("\n✓ Workspace preserved at: %s\n", config.worktreePath)
	}

	// Phase 8: Success
	fmt.Printf("\n✓ Merged %s into %s (%s)\n", config.currentBranch, config.targetBranch, mergeType)
	return 0
}

// mergeConfig holds configuration for the merge command
type mergeConfig struct {
	targetBranch  string
	noSquash      bool
	keepWorktree  bool
	debug         bool
	repoRoot      string
	currentBranch string
	worktreePath  string
}

// parseMergeConfig parses command line arguments
func parseMergeConfig() (*mergeConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "merge"

	config := &mergeConfig{
		targetBranch: "main",
		noSquash:     false,
		keepWorktree: false,
		debug:        false,
	}

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--target=") {
			config.targetBranch = strings.TrimPrefix(arg, "--target=")
		} else if arg == "--no-squash" {
			config.noSquash = true
		} else if arg == "--keep-worktree" {
			config.keepWorktree = true
		} else if arg == "--debug" || arg == "-d" {
			config.debug = true
		}
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}
	config.repoRoot = repoRoot

	// Get current branch
	currentBranch, err := internal.GetCurrentBranch(config.repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	config.currentBranch = currentBranch

	// Get worktree path for current directory
	worktreePath, err := getWorktreePath(config.repoRoot)
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
	clean, err := internal.IsWorktreeClean(config.repoRoot)
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
	cmd := exec.Command("git", "fetch", "origin", config.targetBranch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch %s: %w", config.targetBranch, err)
	}

	// Check if current branch has all commits from target
	cmd = exec.Command("git", "rev-list", "--count", fmt.Sprintf("HEAD..origin/%s", config.targetBranch))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check if branch is up to date: %w", err)
	}

	var behindCount int
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &behindCount)

	if behindCount > 0 {
		return fmt.Errorf("branch not up to date with %s (behind by %d commits)\nRun 'work pull' first to sync with %s", config.targetBranch, behindCount, config.targetBranch)
	}

	return nil
}

// switchToTargetBranch switches to the target branch
func switchToTargetBranch(targetBranch, repoRoot string) error {
	cmd := exec.Command("git", "checkout", targetBranch)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to switch to %s: %w\nOutput: %s", targetBranch, err, string(output))
	}
	return nil
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

// performSquashMerge performs a squash merge and uses commit-ai for the message
func performSquashMerge(config *mergeConfig) error {
	// Perform squash merge (stages changes but doesn't commit)
	cmd := exec.Command("git", "merge", "--squash", config.currentBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to squash merge: %w\nOutput: %s", err, string(output))
	}

	// Use commit-ai to generate commit message and create commit
	fmt.Println("\nGenerating commit message...")
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	if config.debug {
		os.Args = []string{"r2r", "commit-ai", "--debug"}
	} else {
		os.Args = []string{"r2r", "commit-ai"}
	}

	exitCode := commit.CommitAI()
	if exitCode != 0 {
		return fmt.Errorf("commit-ai failed with exit code %d", exitCode)
	}

	return nil
}

// performRegularMerge performs a regular merge
func performRegularMerge(branch string) error {
	cmd := exec.Command("git", "merge", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// removeWorkspace removes the workspace and deletes the branch
func removeWorkspace(worktreePath, branch string) error {
	// Remove worktree
	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w\nOutput: %s", err, string(output))
	}

	// Delete branch
	cmd = exec.Command("git", "branch", "-d", branch)
	output, err = cmd.CombinedOutput()
	if err != nil {
		// If branch deletion fails, warn but don't error (branch might be needed)
		fmt.Fprintf(os.Stderr, "Warning: failed to delete branch %s: %v\nOutput: %s\n", branch, err, string(output))
		fmt.Fprintf(os.Stderr, "You can manually delete with: git branch -D %s\n", branch)
	}

	return nil
}

// getWorktreePath returns the path of the current worktree
func getWorktreePath(repoRoot string) (string, error) {
	// Get list of all worktrees
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Parse worktree list to find current worktree path
	cwd, _ := os.Getwd()
	lines := strings.Split(string(output), "\n")
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "worktree ") {
			path := strings.TrimPrefix(lines[i], "worktree ")
			// Normalize paths for comparison
			if strings.EqualFold(path, cwd) || strings.EqualFold(path, repoRoot) {
				return path, nil
			}
		}
	}

	return repoRoot, nil
}

// handleMergeError handles merge errors and provides guidance
func handleMergeError(err error, config *mergeConfig) {
	fmt.Fprintf(os.Stderr, "\n⚠️  Merge conflict detected\n\n")

	// Get conflicting files
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	output, _ := cmd.Output()
	conflicts := strings.Split(strings.TrimSpace(string(output)), "\n")

	if len(conflicts) > 0 && conflicts[0] != "" {
		fmt.Fprintf(os.Stderr, "Conflicting files:\n")
		for _, file := range conflicts {
			fmt.Fprintf(os.Stderr, "  - %s\n", file)
		}
		fmt.Fprintln(os.Stderr)
	}

	fmt.Fprintf(os.Stderr, "Resolve conflicts then:\n")
	fmt.Fprintf(os.Stderr, "  git add <files>\n")
	fmt.Fprintf(os.Stderr, "  git commit\n")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Or abort:\n")
	fmt.Fprintf(os.Stderr, "  git merge --abort\n")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Note: Workspace preserved at %s\n", config.worktreePath)
}
