package testing

import (
	"os"
	"testing"
)

func TestRequireGitRepository_InGitRepo(t *testing.T) {
	RequireGitRepository(t)
}

func TestRequireGitRepository_NotInGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}()

	if err = os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if isGitRepository() {
		t.Skip("Test environment has git repo in temp dir")
	}
}

func TestRequireCleanWorktree_Clean(t *testing.T) {
	RequireGitRepository(t)
}
