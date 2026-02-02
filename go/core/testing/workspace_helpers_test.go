package testing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupWorkspaceIsolation(t *testing.T) {
	// Get real repo root before isolation (for comparison)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	// Find real git root by walking up
	realRoot := findTestGitRoot(t, wd)

	// Run the function under test
	root := SetupWorkspaceIsolation(t)

	// Verify isolation was set up
	if envRoot := os.Getenv("R2R_REPO_ROOT"); envRoot == "" {
		t.Error("R2R_REPO_ROOT not set after SetupWorkspaceIsolation")
	}

	// Verify return value matches real repo root
	if root != realRoot {
		t.Errorf("SetupWorkspaceIsolation() returned %q, want %q", root, realRoot)
	}

	// Verify we can access real repo files
	repoYml := filepath.Join(root, ".eac", "repository.yml")
	if _, err := os.Stat(repoYml); os.IsNotExist(err) {
		t.Errorf("expected repository.yml at %s, but file not found", repoYml)
	}
}

func TestSetupTempWorkspaceIsolation(t *testing.T) {
	// Run the function under test
	tempRoot := SetupTempWorkspaceIsolation(t)

	// Verify isolation was set up
	if envRoot := os.Getenv("R2R_REPO_ROOT"); envRoot == "" {
		t.Error("R2R_REPO_ROOT not set after SetupTempWorkspaceIsolation")
	}

	// Verify the returned path is a temp directory
	if tempRoot == "" {
		t.Error("SetupTempWorkspaceIsolation() returned empty string")
	}

	// Verify .git marker was created
	gitDir := filepath.Join(tempRoot, ".git")
	info, err := os.Stat(gitDir)
	if os.IsNotExist(err) {
		t.Errorf(".git directory not created at %s", gitDir)
	} else if err != nil {
		t.Errorf("failed to stat .git: %v", err)
	} else if !info.IsDir() {
		t.Errorf(".git exists but is not a directory")
	}

	// Verify we can create files in the temp directory
	testFile := filepath.Join(tempRoot, "test-file.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Errorf("failed to create test file in temp workspace: %v", err)
	}
}

func TestRealRepoRoot(t *testing.T) {
	root := RealRepoRoot(t)

	// Verify it returns a valid path
	if root == "" {
		t.Error("RealRepoRoot() returned empty string")
	}

	// Verify it contains .git
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("RealRepoRoot() returned %q but no .git found there", root)
	}

	// Verify it contains .eac/repository.yml
	repoYml := filepath.Join(root, ".eac", "repository.yml")
	if _, err := os.Stat(repoYml); os.IsNotExist(err) {
		t.Errorf("RealRepoRoot() returned %q but no repository.yml found", root)
	}
}

func TestCopyToTempWorkspace(t *testing.T) {
	// Copy a specific file to temp workspace
	filesToCopy := []string{
		".eac/repository.yml",
	}

	tempRoot := CopyToTempWorkspace(t, filesToCopy)

	// Verify isolation was set up
	if envRoot := os.Getenv("R2R_REPO_ROOT"); envRoot == "" {
		t.Error("R2R_REPO_ROOT not set after CopyToTempWorkspace")
	}

	// Verify the file was copied
	copiedFile := filepath.Join(tempRoot, ".eac", "repository.yml")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("file not copied to %s", copiedFile)
	}

	// Verify .git marker was created
	gitDir := filepath.Join(tempRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf(".git directory not created at %s", gitDir)
	}

	// Verify we can modify the copied file without affecting the original
	if err := os.WriteFile(copiedFile, []byte("modified content"), 0o644); err != nil {
		t.Errorf("failed to modify copied file: %v", err)
	}
}

func TestCopyToTempWorkspaceMultipleFiles(t *testing.T) {
	// Copy multiple files that exist in the repo
	filesToCopy := []string{
		".eac/repository.yml",
		"SECURITY.md",
	}

	tempRoot := CopyToTempWorkspace(t, filesToCopy)

	// Verify both files were copied
	for _, relPath := range filesToCopy {
		copiedFile := filepath.Join(tempRoot, relPath)
		if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
			t.Errorf("file not copied: %s", relPath)
		}
	}
}

// findTestGitRoot walks up from startPath looking for .git directory.
// This is a test helper to find the real repo root for comparison.
func findTestGitRoot(t *testing.T, startPath string) string {
	t.Helper()
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	current := absPath
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatalf("not in a git repository: %s", startPath)
		}
		current = parent
	}
}
