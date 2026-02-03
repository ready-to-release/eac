//go:build L1 && ov
// +build L1,ov

package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if repo.RootPath() != tmpDir {
		t.Errorf("Expected root path %s, got %s", tmpDir, repo.RootPath())
	}

	// Verify .git directory exists
	gitDir := filepath.Join(tmpDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory was not created")
	}
}

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a repo first
	_, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Open from root
	repo, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if repo.RootPath() != tmpDir {
		t.Errorf("Expected root path %s, got %s", tmpDir, repo.RootPath())
	}
}

func TestOpen_FromSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a repo
	_, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "src", "test")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Open from subdirectory
	repo, err := Open(subDir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if repo.RootPath() != tmpDir {
		t.Errorf("Expected root path %s, got %s", tmpDir, repo.RootPath())
	}
}

func TestOpen_NotARepository(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Open(tmpDir)
	if err == nil {
		t.Error("Expected error for non-repository directory")
	}
}

func TestConfigSet(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err = repo.ConfigSet("user", "name", "Test User")
	if err != nil {
		t.Fatalf("ConfigSet user.name failed: %v", err)
	}

	err = repo.ConfigSet("user", "email", "test@example.com")
	if err != nil {
		t.Fatalf("ConfigSet user.email failed: %v", err)
	}
}

func TestAddAndCommit(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Configure user for commit
	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add file
	err = repo.Add("test.txt")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Set deterministic time for test
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	oldTimeNow := timeNow
	timeNow = func() time.Time { return fixedTime }
	defer func() { timeNow = oldTimeNow }()

	// Commit
	hash, err := repo.Commit("Initial commit", "Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	if len(hash) != 40 {
		t.Errorf("Expected 40-character hash, got %d characters", len(hash))
	}
}

func TestTrackedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create and commit files
	for _, name := range []string{"file1.txt", "file2.txt"} {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
		if err := repo.Add(name); err != nil {
			t.Fatalf("Failed to add %s: %v", name, err)
		}
	}

	_, err = repo.Commit("Add files", "Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Create untracked file
	untracked := filepath.Join(tmpDir, "untracked.txt")
	if err := os.WriteFile(untracked, []byte("untracked"), 0644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	files, err := repo.TrackedFiles()
	if err != nil {
		t.Fatalf("TrackedFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 tracked files, got %d", len(files))
	}
}

func TestStagedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create and stage a file
	testFile := filepath.Join(tmpDir, "staged.txt")
	if err := os.WriteFile(testFile, []byte("staged"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	if err := repo.Add("staged.txt"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	files, err := repo.StagedFiles()
	if err != nil {
		t.Fatalf("StagedFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 staged file, got %d", len(files))
	}

	if len(files) > 0 && files[0] != "staged.txt" {
		t.Errorf("Expected 'staged.txt', got %s", files[0])
	}
}

func TestIsFileTracked(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create and commit a file
	trackedPath := filepath.Join(tmpDir, "tracked.txt")
	if err := os.WriteFile(trackedPath, []byte("tracked"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	repo.Add("tracked.txt")
	repo.Commit("Add tracked", "Test User", "test@example.com")

	// Create untracked file
	untrackedPath := filepath.Join(tmpDir, "untracked.txt")
	if err := os.WriteFile(untrackedPath, []byte("untracked"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"tracked.txt", true},
		{"untracked.txt", false},
		{"nonexistent.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := repo.IsFileTracked(tt.path)
			if result != tt.expected {
				t.Errorf("IsFileTracked(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsFileIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create .gitignore
	gitignore := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignore, []byte("*.log\nbuild/\n"), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"test.log", true},
		{"build/output.txt", true},
		{"src/main.go", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := repo.IsFileIgnored(tt.path)
			if result != tt.expected {
				t.Errorf("IsFileIgnored(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCurrentBranch(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Need at least one commit to have a branch
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	repo.Add("test.txt")
	repo.Commit("Initial", "Test User", "test@example.com")

	branch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch failed: %v", err)
	}

	// Default branch is typically "master" or "main" depending on git version
	if branch != "master" && branch != "main" {
		t.Errorf("Expected 'master' or 'main', got %s", branch)
	}
}

func TestRemoteURL(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Add a remote
	err = repo.AddRemote("origin", "https://github.com/test/repo.git")
	if err != nil {
		t.Fatalf("AddRemote failed: %v", err)
	}

	url, err := repo.RemoteURL("origin")
	if err != nil {
		t.Fatalf("RemoteURL failed: %v", err)
	}

	if url != "https://github.com/test/repo.git" {
		t.Errorf("Expected 'https://github.com/test/repo.git', got %s", url)
	}
}

func TestRemoteURL_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	_, err = repo.RemoteURL("origin")
	if err == nil {
		t.Error("Expected error for non-existent remote")
	}
}

func TestHeadShortSHA(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create a commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	repo.Add("test.txt")
	repo.Commit("Initial", "Test User", "test@example.com")

	sha, err := repo.HeadShortSHA()
	if err != nil {
		t.Fatalf("HeadShortSHA failed: %v", err)
	}

	if len(sha) != 7 {
		t.Errorf("Expected 7-character SHA, got %d characters: %s", len(sha), sha)
	}
}

// TestUncommittedFiles verifies that UncommittedFiles returns all files with
// uncommitted changes (staged, unstaged, and untracked changes).
func TestUncommittedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create and commit initial file
	initialFile := filepath.Join(tmpDir, "initial.txt")
	if err := os.WriteFile(initialFile, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}
	if err := repo.Add("initial.txt"); err != nil {
		t.Fatalf("Failed to add initial file: %v", err)
	}
	if _, err := repo.Commit("Initial commit", "Test User", "test@example.com"); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Now create various uncommitted changes:
	// 1. Modify a tracked file (unstaged)
	if err := os.WriteFile(initialFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify initial file: %v", err)
	}

	// 2. Stage a new file
	stagedFile := filepath.Join(tmpDir, "staged.txt")
	if err := os.WriteFile(stagedFile, []byte("staged content"), 0644); err != nil {
		t.Fatalf("Failed to create staged file: %v", err)
	}
	if err := repo.Add("staged.txt"); err != nil {
		t.Fatalf("Failed to add staged file: %v", err)
	}

	// 3. Create an untracked file
	untrackedFile := filepath.Join(tmpDir, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("untracked content"), 0644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	// Get uncommitted files
	files, err := repo.UncommittedFiles()
	if err != nil {
		t.Fatalf("UncommittedFiles failed: %v", err)
	}

	// Should include all three: modified, staged, and untracked
	if len(files) != 3 {
		t.Errorf("Expected 3 uncommitted files, got %d: %v", len(files), files)
	}

	// Verify specific files are present
	fileMap := make(map[string]bool)
	for _, f := range files {
		fileMap[f] = true
	}

	expected := []string{"initial.txt", "staged.txt", "untracked.txt"}
	for _, exp := range expected {
		if !fileMap[exp] {
			t.Errorf("Expected %q in uncommitted files, but not found. Got: %v", exp, files)
		}
	}
}

// TestUncommittedFiles_Empty verifies that a clean repo returns no uncommitted files.
func TestUncommittedFiles_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := repo.Add("test.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	if _, err := repo.Commit("Initial commit", "Test User", "test@example.com"); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Get uncommitted files from clean repo
	files, err := repo.UncommittedFiles()
	if err != nil {
		t.Fatalf("UncommittedFiles failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 uncommitted files in clean repo, got %d: %v", len(files), files)
	}
}

// TestStagedFiles_ModifiedTrackedFile verifies that modifying and staging
// a tracked file returns it in StagedFiles.
func TestStagedFiles_ModifiedTrackedFile(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create and commit initial file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := repo.Add("test.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	if _, err := repo.Commit("Initial commit", "Test User", "test@example.com"); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Modify the file and stage it
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}
	if err := repo.Add("test.txt"); err != nil {
		t.Fatalf("Failed to stage modified file: %v", err)
	}

	files, err := repo.StagedFiles()
	if err != nil {
		t.Fatalf("StagedFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 staged file, got %d: %v", len(files), files)
	}
	if len(files) > 0 && files[0] != "test.txt" {
		t.Errorf("Expected 'test.txt', got %q", files[0])
	}
}

// TestStagedFiles_ExcludesUnstagedModifications verifies that modifications
// that are NOT staged do NOT appear in StagedFiles.
func TestStagedFiles_ExcludesUnstagedModifications(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create and commit initial file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := repo.Add("test.txt"); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	if _, err := repo.Commit("Initial commit", "Test User", "test@example.com"); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Modify the file but do NOT stage it
	if err := os.WriteFile(testFile, []byte("modified but not staged"), 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	files, err := repo.StagedFiles()
	if err != nil {
		t.Fatalf("StagedFiles failed: %v", err)
	}

	// Should return empty since no changes are staged
	if len(files) != 0 {
		t.Errorf("Expected 0 staged files (modification not staged), got %d: %v", len(files), files)
	}
}

// TestTrackedFiles_IncludesStagedNewFiles verifies that newly staged files
// (not yet committed) appear in TrackedFiles.
func TestTrackedFiles_IncludesStagedNewFiles(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create and commit initial file
	initialFile := filepath.Join(tmpDir, "initial.txt")
	if err := os.WriteFile(initialFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}
	if err := repo.Add("initial.txt"); err != nil {
		t.Fatalf("Failed to add initial file: %v", err)
	}
	if _, err := repo.Commit("Initial commit", "Test User", "test@example.com"); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Stage a new file (not committed yet)
	newFile := filepath.Join(tmpDir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}
	if err := repo.Add("new.txt"); err != nil {
		t.Fatalf("Failed to add new file: %v", err)
	}

	files, err := repo.TrackedFiles()
	if err != nil {
		t.Fatalf("TrackedFiles failed: %v", err)
	}

	// Should include both the committed file and the staged file
	if len(files) != 2 {
		t.Errorf("Expected 2 tracked files, got %d: %v", len(files), files)
	}

	fileMap := make(map[string]bool)
	for _, f := range files {
		fileMap[f] = true
	}

	if !fileMap["initial.txt"] {
		t.Errorf("Expected 'initial.txt' in tracked files, got: %v", files)
	}
	if !fileMap["new.txt"] {
		t.Errorf("Expected 'new.txt' in tracked files (staged), got: %v", files)
	}
}

// TestTrackedFiles_Subdirectories verifies that files in subdirectories
// are returned with their relative paths.
func TestTrackedFiles_Subdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo.ConfigSet("user", "name", "Test User")
	repo.ConfigSet("user", "email", "test@example.com")

	// Create subdirectory structure
	subDir := filepath.Join(tmpDir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create files in subdirectory
	file1 := filepath.Join(subDir, "file1.go")
	if err := os.WriteFile(file1, []byte("package pkg"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := repo.Add("src/pkg/file1.go"); err != nil {
		t.Fatalf("Failed to add file1: %v", err)
	}

	// Create file at root
	rootFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(rootFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}
	if err := repo.Add("main.go"); err != nil {
		t.Fatalf("Failed to add main.go: %v", err)
	}

	if _, err := repo.Commit("Add files", "Test User", "test@example.com"); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	files, err := repo.TrackedFiles()
	if err != nil {
		t.Fatalf("TrackedFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 tracked files, got %d: %v", len(files), files)
	}

	// Check paths use forward slashes (git convention)
	fileMap := make(map[string]bool)
	for _, f := range files {
		fileMap[f] = true
	}

	if !fileMap["main.go"] {
		t.Errorf("Expected 'main.go' in tracked files, got: %v", files)
	}
	if !fileMap["src/pkg/file1.go"] {
		t.Errorf("Expected 'src/pkg/file1.go' in tracked files, got: %v", files)
	}
}
