package testing

import (
	"os/exec"
	"testing"
)

// RequireGitRepository skips the test if not running inside a git repository.
func RequireGitRepository(t *testing.T) {
	t.Helper()
	if !isGitRepository() {
		t.Skip("Test requires git repository (deps:git)")
	}
}

// RequireCleanWorktree skips if git worktree has uncommitted changes.
func RequireCleanWorktree(t *testing.T) {
	t.Helper()
	RequireGitRepository(t)

	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil || len(out) > 0 {
		t.Skip("Test requires clean git worktree")
	}
}

// isGitRepository checks if the current directory is inside a git repository.
func isGitRepository() bool {
	_, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	return err == nil
}
