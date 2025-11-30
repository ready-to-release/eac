// Command: commit reset
// Short: Soft reset the latest commit, preserving changes
// Long: Performs a soft reset of the most recent commit (git reset --soft HEAD~1).
// Long:
// Long: This undoes the last commit but keeps all changes staged in the index,
// Long: allowing you to modify the commit message, add more changes, or restructure
// Long: the commit before recommitting.
// Long:
// Long: Example:
// Long:   r2r commit reset
// HasSideEffects: true
package commit

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(CommitReset)
}

// resetGitOps holds the git operations interface for reset command.
// In production, this uses exec.Command. For tests, it can be injected via SetResetGitOps.
var resetGitOps ResetGitOperations

// ResetGitOperations defines the git operations needed for reset command.
type ResetGitOperations interface {
	HasParentCommit() (bool, error)
	SoftResetHead() error
}

// defaultResetGitOps implements ResetGitOperations using real git commands.
type defaultResetGitOps struct{}

func (d *defaultResetGitOps) HasParentCommit() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD~1")
	if err := cmd.Run(); err != nil {
		return false, nil // No parent commit
	}
	return true, nil
}

func (d *defaultResetGitOps) SoftResetHead() error {
	cmd := exec.Command("git", "reset", "--soft", "HEAD~1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reset commit: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// getResetGitOps returns the git operations interface, initializing if needed.
func getResetGitOps() ResetGitOperations {
	if resetGitOps != nil {
		return resetGitOps
	}
	return &defaultResetGitOps{}
}

// SetResetGitOps allows tests to inject a mock implementation.
func SetResetGitOps(ops ResetGitOperations) {
	resetGitOps = ops
}

// ResetResetGitOps clears the operations for test cleanup.
func ResetResetGitOps() {
	resetGitOps = nil
}

// CommitReset performs a soft reset of the latest commit
func CommitReset() int {
	args := os.Args[3:] // Skip program name, "commit", "reset"

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printResetUsage()
			return 0
		}
	}

	// Phase 1: Validate environment - must be in a git repository
	if err := ensureInGitRepo(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Get workspace root for context
	_, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	ops := getResetGitOps()

	// Phase 2: Check if there are commits to reset (need at least 2 commits)
	hasParent, err := ops.HasParentCommit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if !hasParent {
		fmt.Fprintf(os.Stderr, "Error: no commits to reset (may be initial commit)\n")
		return 1
	}

	// Phase 3: Execute soft reset
	if err := ops.SoftResetHead(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println("Successfully reset latest commit. Changes are now staged.")
	return 0
}

func printResetUsage() {
	fmt.Println("Soft reset the latest commit, preserving changes")
	fmt.Println()
	fmt.Println("Usage: r2r commit reset")
	fmt.Println()
	fmt.Println("This command performs 'git reset --soft HEAD~1', which:")
	fmt.Println("  - Undoes the most recent commit")
	fmt.Println("  - Keeps all changes staged in the index")
	fmt.Println("  - Preserves the working directory")
	fmt.Println()
	fmt.Println("Use this when you want to:")
	fmt.Println("  - Modify the commit message")
	fmt.Println("  - Add more changes to the commit")
	fmt.Println("  - Split the commit into multiple commits")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  r2r commit reset")
}

// ensureInGitRepo checks if we're in a git repository
func ensureInGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not in a git repository")
	}
	return nil
}
