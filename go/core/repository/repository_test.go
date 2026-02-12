//go:build L1 && ov
// +build L1,ov

package repository

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/git"
)

// createTestGitRepo creates a temporary git repository for testing using go-git.
// The repo is scoped to the test via t.TempDir() and needs no manual cleanup.
func createTestGitRepo(t *testing.T) (string, git.GitRepository) {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo using go-git
	repo, err := git.Init(tmpDir)
	if err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user for commits
	if err := repo.ConfigSet("user", "name", "Test User"); err != nil {
		t.Fatalf("Failed to set git user.name: %v", err)
	}

	if err := repo.ConfigSet("user", "email", "test@example.com"); err != nil {
		t.Fatalf("Failed to set git user.email: %v", err)
	}

	return tmpDir, repo
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
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", fullPath, err)
	}
}

// gitAdd adds a file to git staging using go-git
func gitAdd(t *testing.T, repo git.GitRepository, file string) {
	if err := repo.Add(file); err != nil {
		t.Fatalf("Failed to git add %s: %v", file, err)
	}
}

// gitCommit creates a commit using go-git
func gitCommit(t *testing.T, repo git.GitRepository, message string) {
	_, err := repo.Commit(message, "Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}
}

func TestGetRepositoryRoot_FromRoot(t *testing.T) {
	repoDir, _ := createTestGitRepo(t)

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
	repoDir, _ := createTestGitRepo(t)

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
	tmpDir := t.TempDir()

	// Don't initialize git
	_, err := GetRepositoryRoot(tmpDir)
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
	repoDir, repo := createTestGitRepo(t)

	// Create and track files
	createTestFile(t, repoDir, "file1.txt", "content1")
	createTestFile(t, repoDir, "file2.txt", "content2")
	gitAdd(t, repo, "file1.txt")
	gitAdd(t, repo, "file2.txt")
	gitCommit(t, repo, "Initial commit")

	// Create untracked file
	createTestFile(t, repoDir, "untracked.txt", "untracked")

	files, err := GetRepositoryFiles(repo, true, false, false, false)
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
	repoDir, repo := createTestGitRepo(t)

	// Create tracked files
	createTestFile(t, repoDir, "tracked.txt", "tracked")
	gitAdd(t, repo, "tracked.txt")
	gitCommit(t, repo, "Initial commit")

	// Create untracked file
	createTestFile(t, repoDir, "untracked.txt", "untracked")

	files, err := GetRepositoryFiles(repo, false, false, false, false)
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
	repoDir, repo := createTestGitRepo(t)

	// Create .gitignore
	createTestFile(t, repoDir, ".gitignore", "ignored.txt\n")
	gitAdd(t, repo, ".gitignore")
	gitCommit(t, repo, "Add gitignore")

	// Create ignored file
	createTestFile(t, repoDir, "ignored.txt", "ignored")

	// Create tracked file
	createTestFile(t, repoDir, "tracked.txt", "tracked")
	gitAdd(t, repo, "tracked.txt")
	gitCommit(t, repo, "Add tracked")

	// Get all files excluding ignored
	filesExcluded, err := GetRepositoryFiles(repo, false, false, false, false)
	if err != nil {
		t.Fatalf("GetRepositoryFiles failed: %v", err)
	}

	// Get all files including ignored
	filesIncluded, err := GetRepositoryFiles(repo, false, true, false, false)
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
	repoDir, _ := createTestGitRepo(t)

	repo, err := New(repoDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	expectedRoot := filepath.Clean(repoDir)
	gotRoot := filepath.Clean(repo.Root())

	if gotRoot != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, gotRoot)
	}

	// Verify GitRepo() returns a valid interface
	if repo.GitRepo() == nil {
		t.Error("GitRepo() returned nil")
	}
}

func TestRepository_Files(t *testing.T) {
	repoDir, gitRepo := createTestGitRepo(t)

	// Create tracked file
	createTestFile(t, repoDir, "file.txt", "content")
	gitAdd(t, gitRepo, "file.txt")
	gitCommit(t, gitRepo, "Initial commit")

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
	repoDir, _ := createTestGitRepo(t)
	nonRepoDir := t.TempDir()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"git repository", repoDir, true},
		{"non-git directory", nonRepoDir, false},
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

// ============================================================================
// Mock-based tests for error paths and edge cases
// ============================================================================

func TestGetRepositoryFiles_WithMock_TrackedFiles(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"main.go", "go.mod", "README.md"})

	files, err := GetRepositoryFiles(mock, true, false, false, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}

	// All should be tracked
	for _, f := range files {
		if !f.IsTracked {
			t.Errorf("File %s should be tracked", f.Path)
		}
	}
}

func TestGetRepositoryFiles_WithMock_StagedFiles(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithStagedFiles([]string{"new.go", "updated.go"})

	files, err := GetRepositoryFiles(mock, false, false, false, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 staged files, got %d", len(files))
	}
}

func TestGetRepositoryFiles_WithMock_TrackedFilesError(t *testing.T) {
	expectedErr := errors.New("git index corrupted")
	mock := git.NewMockRepository("/test/repo").
		WithError("TrackedFiles", expectedErr)

	_, err := GetRepositoryFiles(mock, true, false, false, false)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should be wrapped in RepositoryError
	var repoErr *RepositoryError
	if !errors.As(err, &repoErr) {
		t.Errorf("Expected RepositoryError, got %T", err)
	}
}

func TestGetRepositoryFiles_WithMock_StagedFilesError(t *testing.T) {
	expectedErr := errors.New("staging area error")
	mock := git.NewMockRepository("/test/repo").
		WithError("StagedFiles", expectedErr)

	_, err := GetRepositoryFiles(mock, false, false, false, true)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var repoErr *RepositoryError
	if !errors.As(err, &repoErr) {
		t.Errorf("Expected RepositoryError, got %T", err)
	}
}

func TestGetRepositoryFiles_WithMock_FiltersGitInternalFiles(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"main.go", ".gitignore", ".gitkeep", "src/.gitkeep"})

	// Without git internal files
	files, err := GetRepositoryFiles(mock, true, false, false, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file (excluding .gitignore/.gitkeep), got %d", len(files))
	}

	// With git internal files
	filesWithInternal, err := GetRepositoryFiles(mock, true, false, true, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(filesWithInternal) != 4 {
		t.Errorf("Expected 4 files (including .gitignore/.gitkeep), got %d", len(filesWithInternal))
	}
}

func TestGetRepositoryFiles_WithMock_EmptyRepository(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{})

	files, err := GetRepositoryFiles(mock, true, false, false, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(files))
	}
}

func TestGetRepositoryFiles_WithMock_WhitespaceInFilenames(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"  ", "main.go", "", "test.go"})

	files, err := GetRepositoryFiles(mock, true, false, false, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should filter out empty/whitespace-only entries
	if len(files) != 2 {
		t.Errorf("Expected 2 files (filtering whitespace), got %d", len(files))
	}
}
