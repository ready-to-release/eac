package repository

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// createTestGitRepo creates a temporary git repository for testing
func createTestGitRepo(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "repo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user for commits
	exec.Command("git", "config", "user.name", "Test User").Dir = tmpDir
	exec.Command("git", "config", "user.email", "test@example.com").Run()

	return tmpDir
}

// createTestFile creates a test file in the given directory
func createTestFile(t *testing.T, dir, relativePath, content string) {
	fullPath := filepath.Join(dir, relativePath)
	dirPath := filepath.Dir(fullPath)

	// Create parent directories
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dirPath, err)
	}

	// Create file
	if err := ioutil.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", fullPath, err)
	}
}

// gitAdd adds a file to git staging
func gitAdd(t *testing.T, repoDir, file string) {
	cmd := exec.Command("git", "add", file)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add %s: %v", file, err)
	}
}

// gitCommit creates a commit
func gitCommit(t *testing.T, repoDir, message string) {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}
}

func TestGetRepositoryRoot_FromRoot(t *testing.T) {
	repoDir := createTestGitRepo(t)
	defer os.RemoveAll(repoDir)

	root, err := GetRepositoryRoot(repoDir)
	if err != nil {
		t.Fatalf("GetRepositoryRoot failed: %v", err)
	}

	// Normalize paths for comparison
	expectedRoot := filepath.Clean(repoDir)
	gotRoot := filepath.Clean(root)

	if gotRoot != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, gotRoot)
	}
}

func TestGetRepositoryRoot_FromSubdirectory(t *testing.T) {
	repoDir := createTestGitRepo(t)
	defer os.RemoveAll(repoDir)

	// Create subdirectories
	subDir := filepath.Join(repoDir, "src", "test")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	root, err := GetRepositoryRoot(subDir)
	if err != nil {
		t.Fatalf("GetRepositoryRoot failed: %v", err)
	}

	expectedRoot := filepath.Clean(repoDir)
	gotRoot := filepath.Clean(root)

	if gotRoot != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, gotRoot)
	}
}

func TestGetRepositoryRoot_NotARepository(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "repo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Don't initialize git
	_, err = GetRepositoryRoot(tmpDir)
	if err == nil {
		t.Error("Expected error for non-repository directory")
	}
}

func TestGetRepositoryRoot_EmptyPath(t *testing.T) {
	// Should use current directory
	// This test will pass if the test is run from within a git repository
	root, err := GetRepositoryRoot("")
	if err != nil {
		// It's okay if we're not in a git repo during test
		t.Skip("Not in a git repository")
	}

	if root == "" {
		t.Error("Expected non-empty root path")
	}
}

func TestGetRepositoryFiles_TrackedOnly(t *testing.T) {
	repoDir := createTestGitRepo(t)
	defer os.RemoveAll(repoDir)

	// Create and track files
	createTestFile(t, repoDir, "file1.txt", "content1")
	createTestFile(t, repoDir, "file2.txt", "content2")
	gitAdd(t, repoDir, "file1.txt")
	gitAdd(t, repoDir, "file2.txt")
	gitCommit(t, repoDir, "Initial commit")

	// Create untracked file
	createTestFile(t, repoDir, "untracked.txt", "untracked")

	files, err := GetRepositoryFiles(true, false, false, false, repoDir)
	if err != nil {
		t.Fatalf("GetRepositoryFiles failed: %v", err)
	}

	// Should only have 2 tracked files
	if len(files) != 2 {
		t.Errorf("Expected 2 tracked files, got %d", len(files))
	}

	// All should be tracked
	for _, file := range files {
		if !file.IsTracked {
			t.Errorf("File %s should be tracked", file.Path)
		}
	}
}

func TestGetRepositoryFiles_AllFiles(t *testing.T) {
	repoDir := createTestGitRepo(t)
	defer os.RemoveAll(repoDir)

	// Create tracked files
	createTestFile(t, repoDir, "tracked.txt", "tracked")
	gitAdd(t, repoDir, "tracked.txt")
	gitCommit(t, repoDir, "Initial commit")

	// Create untracked file
	createTestFile(t, repoDir, "untracked.txt", "untracked")

	files, err := GetRepositoryFiles(false, false, false, false, repoDir)
	if err != nil {
		t.Fatalf("GetRepositoryFiles failed: %v", err)
	}

	// Should have at least 2 files (tracked + untracked)
	if len(files) < 2 {
		t.Errorf("Expected at least 2 files, got %d", len(files))
	}

	trackedCount := 0
	untrackedCount := 0
	for _, file := range files {
		if file.IsTracked {
			trackedCount++
		} else {
			untrackedCount++
		}
	}

	if trackedCount < 1 {
		t.Error("Expected at least 1 tracked file")
	}
	if untrackedCount < 1 {
		t.Error("Expected at least 1 untracked file")
	}
}

func TestGetRepositoryFiles_WithIgnored(t *testing.T) {
	repoDir := createTestGitRepo(t)
	defer os.RemoveAll(repoDir)

	// Create .gitignore
	createTestFile(t, repoDir, ".gitignore", "ignored.txt\n")
	gitAdd(t, repoDir, ".gitignore")
	gitCommit(t, repoDir, "Add gitignore")

	// Create ignored file
	createTestFile(t, repoDir, "ignored.txt", "ignored")

	// Create tracked file
	createTestFile(t, repoDir, "tracked.txt", "tracked")
	gitAdd(t, repoDir, "tracked.txt")
	gitCommit(t, repoDir, "Add tracked")

	// Get all files excluding ignored
	filesExcluded, err := GetRepositoryFiles(false, false, false, false, repoDir)
	if err != nil {
		t.Fatalf("GetRepositoryFiles failed: %v", err)
	}

	// Get all files including ignored
	filesIncluded, err := GetRepositoryFiles(false, true, false, false, repoDir)
	if err != nil {
		t.Fatalf("GetRepositoryFiles failed: %v", err)
	}

	// Should have more files when including ignored
	if len(filesIncluded) <= len(filesExcluded) {
		t.Error("Expected more files when including ignored")
	}

	// Check that ignored file is marked as ignored
	foundIgnored := false
	for _, file := range filesIncluded {
		if file.Path == "ignored.txt" {
			foundIgnored = true
			if !file.IsIgnored {
				t.Error("ignored.txt should be marked as ignored")
			}
		}
	}

	if !foundIgnored {
		t.Error("ignored.txt should be in files when includeIgnored=true")
	}
}

func TestNew(t *testing.T) {
	repoDir := createTestGitRepo(t)
	defer os.RemoveAll(repoDir)

	repo, err := New(repoDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	expectedRoot := filepath.Clean(repoDir)
	gotRoot := filepath.Clean(repo.Root())

	if gotRoot != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, gotRoot)
	}
}

func TestRepository_Files(t *testing.T) {
	repoDir := createTestGitRepo(t)
	defer os.RemoveAll(repoDir)

	// Create tracked file
	createTestFile(t, repoDir, "file.txt", "content")
	gitAdd(t, repoDir, "file.txt")
	gitCommit(t, repoDir, "Initial commit")

	repo, err := New(repoDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	files, err := repo.Files(true, false)
	if err != nil {
		t.Fatalf("Files failed: %v", err)
	}

	if len(files) < 1 {
		t.Error("Expected at least 1 tracked file")
	}
}

func TestIsGitRepository(t *testing.T) {
	repoDir := createTestGitRepo(t)
	defer os.RemoveAll(repoDir)

	tmpDir, err := ioutil.TempDir("", "repo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"git repository", repoDir, true},
		{"non-git directory", tmpDir, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGitRepository(tt.path)
			if got != tt.expected {
				t.Errorf("IsGitRepository(%s) = %v, expected %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestFileInfo_Fields(t *testing.T) {
	info := FileInfo{
		Path:         "src/test.go",
		AbsolutePath: "/workspace/src/test.go",
		IsTracked:    true,
		IsIgnored:    false,
	}

	if info.Path != "src/test.go" {
		t.Errorf("Expected Path 'src/test.go', got '%s'", info.Path)
	}

	if info.AbsolutePath != "/workspace/src/test.go" {
		t.Errorf("Expected AbsolutePath '/workspace/src/test.go', got '%s'", info.AbsolutePath)
	}

	if !info.IsTracked {
		t.Error("Expected IsTracked to be true")
	}

	if info.IsIgnored {
		t.Error("Expected IsIgnored to be false")
	}
}

func TestRepositoryError_Error(t *testing.T) {
	err := &RepositoryError{
		Op:      "find",
		Path:    "/test/path",
		Message: "test error",
	}

	errorMsg := err.Error()
	if errorMsg == "" {
		t.Error("Error message should not be empty")
	}

	// Should contain key information
	if len(errorMsg) < 10 {
		t.Error("Error message seems too short")
	}
}

func TestRepositoryError_Unwrap(t *testing.T) {
	underlying := os.ErrNotExist
	err := &RepositoryError{
		Op:      "test",
		Err:     underlying,
		Message: "test",
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlying {
		t.Error("Unwrap returned wrong error")
	}
}
