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
