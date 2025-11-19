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
// Long:
// Long: Example:
// Long:   work pull
// Long:   work pull --target=develop
// Long:   work pull --autostash
// Flag.target: type=string, default=main, usage=Target branch to rebase onto
// Flag.autostash: type=bool, default=false, usage=Automatically stash and unstash uncommitted changes
// Flag.no-fetch: type=bool, default=false, usage=Skip fetching from remote (use local target branch)
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

	// Phase 2: Validate environment
	if err := validatePullEnvironment(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 3: Handle uncommitted changes
	if config.autostash {
		if err := stashChanges(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		defer unstashChanges() // Unstash after rebase
	}

	// Phase 4: Fetch target branch
	if !config.noFetch {
		fmt.Printf("Fetching origin/%s...\n", config.targetBranch)
		if err := fetchTargetBranch(config.targetBranch); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	// Phase 5: Get rebase info
	info, err := getRebaseInfo(config.currentBranch, config.targetBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Phase 6: Check if already up to date
	if info.upToDate {
		fmt.Println("Already up to date")
		return 0
	}

	// Phase 7: Show rebase preview
	showRebasePreview(config.currentBranch, config.targetBranch, info)

	// Phase 8: Perform rebase
	fmt.Printf("\nRebasing %s onto origin/%s...\n", config.currentBranch, config.targetBranch)
	if err := performRebase(config.targetBranch); err != nil {
		handleRebaseError(err)
		return 1
	}

	// Phase 9: Success
	fmt.Printf("\n✓ Rebased %s onto origin/%s\n", config.currentBranch, config.targetBranch)
	fmt.Printf("  Your commits: %d\n", info.currentCommits)
	fmt.Printf("  New commits from %s: %d\n", config.targetBranch, info.newCommits)
	return 0
}

// pullConfig holds configuration for the pull command
type pullConfig struct {
	targetBranch  string
	autostash     bool
	noFetch       bool
	repoRoot      string
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

	config := &pullConfig{
		targetBranch: "main",
		autostash:    false,
		noFetch:      false,
	}

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--target=") {
			config.targetBranch = strings.TrimPrefix(arg, "--target=")
		} else if arg == "--autostash" {
			config.autostash = true
		} else if arg == "--no-fetch" {
			config.noFetch = true
		}
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}
	config.repoRoot = repoRoot

	// Get current branch
	currentBranch, err := internal.GetCurrentBranch(repoRoot)
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
		clean, err := internal.IsWorktreeClean(config.repoRoot)
		if err != nil {
			return fmt.Errorf("failed to check working tree status: %w", err)
		}
		if !clean {
			return fmt.Errorf("uncommitted changes detected\nCommit or stash your changes, or use --autostash")
		}
	}

	return nil
}

// stashChanges stashes uncommitted changes
func stashChanges() error {
	cmd := exec.Command("git", "stash", "push", "-m", "work-pull autostash")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stash changes: %w\nOutput: %s", err, string(output))
	}
	fmt.Println("Stashed uncommitted changes")
	return nil
}

// unstashChanges unstashes previously stashed changes
func unstashChanges() {
	cmd := exec.Command("git", "stash", "pop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to reapply stashed changes: %v\nOutput: %s\n", err, string(output))
		fmt.Fprintf(os.Stderr, "You can manually apply with: git stash pop\n")
		return
	}
	fmt.Println("Reapplied stashed changes")
}

// fetchTargetBranch fetches the target branch from origin
func fetchTargetBranch(targetBranch string) error {
	cmd := exec.Command("git", "fetch", "origin", targetBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fetch origin/%s: %w\nOutput: %s", targetBranch, err, string(output))
	}
	return nil
}

// getRebaseInfo gets information about the rebase operation
func getRebaseInfo(currentBranch, targetBranch string) (*rebaseInfo, error) {
	info := &rebaseInfo{}

	// Count commits on current branch ahead of target
	cmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("origin/%s..HEAD", targetBranch))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to count current commits: %w", err)
	}
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &info.currentCommits)

	// Count new commits on target
	cmd = exec.Command("git", "rev-list", "--count", fmt.Sprintf("HEAD..origin/%s", targetBranch))
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to count new commits: %w", err)
	}
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &info.newCommits)

	// Check if already up to date
	info.upToDate = info.newCommits == 0

	return info, nil
}

// showRebasePreview shows what will be rebased
func showRebasePreview(currentBranch, targetBranch string, info *rebaseInfo) {
	fmt.Printf("\nRebasing %s onto origin/%s\n", currentBranch, targetBranch)
	fmt.Printf("  Current branch: %d commits ahead of %s\n", info.currentCommits, targetBranch)
	fmt.Printf("  %s branch: %d new commits\n", targetBranch, info.newCommits)
	if info.currentCommits > 0 && info.newCommits > 0 {
		fmt.Printf("  Rebase will replay your %d commits on top of %d new commits\n", info.currentCommits, info.newCommits)
	}
}

// performRebase performs the git rebase operation
func performRebase(targetBranch string) error {
	cmd := exec.Command("git", "rebase", fmt.Sprintf("origin/%s", targetBranch))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// handleRebaseError handles rebase errors and provides guidance
func handleRebaseError(err error) {
	fmt.Fprintf(os.Stderr, "\n⚠️  Rebase conflict detected\n\n")

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
	fmt.Fprintf(os.Stderr, "  git rebase --continue\n")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Or abort:\n")
	fmt.Fprintf(os.Stderr, "  git rebase --abort\n")
}
