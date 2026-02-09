package repository

import (
	"errors"
	"os"
	"testing"

	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/github"
)

func TestFileCache_DevBox_UsesGitLsFiles(t *testing.T) {
	// Setup: mock git repo with tracked files
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"file1.go", "dir/file2.go", "README.md"})

	// Create cache with useGitHubInCI=false (DevBox mode)
	cache := NewFileCache("/test/repo", false)
	cache.testGitRepo = mockGit

	// Act: get tracked files
	files, err := cache.TrackedFiles()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
}

func TestFileCache_DevBox_NormalizesPathSeparators(t *testing.T) {
	// Setup: mock with Windows-style paths
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"dir\\subdir\\file.go", "other\\file.md"})

	cache := NewFileCache("/test/repo", false)
	cache.testGitRepo = mockGit

	// Act
	files, err := cache.TrackedFiles()

	// Assert: paths should be normalized to forward slashes
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range files {
		if f == "dir\\subdir\\file.go" || f == "other\\file.md" {
			t.Errorf("path not normalized: %s", f)
		}
	}
	// Check normalized paths exist
	found := false
	for _, f := range files {
		if f == "dir/subdir/file.go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected normalized path 'dir/subdir/file.go' not found")
	}
}

func TestFileCache_CI_UsesGitHubAPI(t *testing.T) {
	// Save and restore env vars
	origCI := os.Getenv("CI")
	origSHA := os.Getenv("GITHUB_SHA")
	defer func() {
		os.Setenv("CI", origCI)
		os.Setenv("GITHUB_SHA", origSHA)
	}()

	// Setup CI environment
	os.Setenv("CI", "true")
	os.Setenv("GITHUB_SHA", "abc123def456")

	// Setup GitHub mock
	mockAPI := github.NewMockAPI().
		AddTreeFiles("abc123def456", []string{"src/main.go", "src/util.go", "README.md"})

	// Create cache with useGitHubInCI=true and inject mock API
	cache := NewFileCache("/test/repo", true)
	cache.githubAPI = mockAPI

	// Act
	files, err := cache.TrackedFiles()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
	if mockAPI.CallCount("GetTreeFiles") != 1 {
		t.Errorf("expected GetTreeFiles to be called once, got %d", mockAPI.CallCount("GetTreeFiles"))
	}
}

func TestFileCache_CI_FallsBackToGit_WhenGitHubAPIFails(t *testing.T) {
	// Save and restore env vars
	origCI := os.Getenv("CI")
	origSHA := os.Getenv("GITHUB_SHA")
	defer func() {
		os.Setenv("CI", origCI)
		os.Setenv("GITHUB_SHA", origSHA)
	}()

	// Setup CI environment
	os.Setenv("CI", "true")
	os.Setenv("GITHUB_SHA", "abc123def456")

	// Setup GitHub mock to fail
	mockAPI := github.NewMockAPI()
	mockAPI.TreeError = errors.New("API rate limit exceeded")

	// Setup git mock as fallback
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"fallback1.go", "fallback2.go"})

	// Create cache with injected mocks
	cache := NewFileCache("/test/repo", true)
	cache.githubAPI = mockAPI
	cache.testGitRepo = mockGit

	// Act
	files, err := cache.TrackedFiles()

	// Assert: should fall back to git and succeed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files from fallback, got %d", len(files))
	}
	if mockAPI.CallCount("GetTreeFiles") != 1 {
		t.Errorf("expected GetTreeFiles to be attempted, got %d calls", mockAPI.CallCount("GetTreeFiles"))
	}
}

func TestFileCache_CI_FallsBackToGit_WhenGitHubSHAMissing(t *testing.T) {
	// Save and restore env vars
	origCI := os.Getenv("CI")
	origSHA := os.Getenv("GITHUB_SHA")
	defer func() {
		os.Setenv("CI", origCI)
		os.Setenv("GITHUB_SHA", origSHA)
	}()

	// Setup CI environment without GITHUB_SHA
	os.Setenv("CI", "true")
	os.Unsetenv("GITHUB_SHA")

	// Setup GitHub mock (should not be called since SHA is missing)
	mockAPI := github.NewMockAPI()

	// Setup git mock as fallback
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"fallback.go"})

	// Create cache with injected mocks
	cache := NewFileCache("/test/repo", true)
	cache.githubAPI = mockAPI
	cache.testGitRepo = mockGit

	// Act
	files, err := cache.TrackedFiles()

	// Assert: should fall back to git
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file from fallback, got %d", len(files))
	}
}

func TestFileCache_CI_FallsBackToGit_WhenGitHubAPINotInitialized(t *testing.T) {
	// Save and restore env vars
	origCI := os.Getenv("CI")
	origSHA := os.Getenv("GITHUB_SHA")
	defer func() {
		os.Setenv("CI", origCI)
		os.Setenv("GITHUB_SHA", origSHA)
	}()

	// Setup CI environment
	os.Setenv("CI", "true")
	os.Setenv("GITHUB_SHA", "abc123")

	// Setup git mock as fallback (no githubAPI set — nil)
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"fallback.go"})

	// Create cache without githubAPI (simulates uninitialized)
	cache := NewFileCache("/test/repo", true)
	cache.testGitRepo = mockGit

	// Act
	files, err := cache.TrackedFiles()

	// Assert: should fall back to git
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file from fallback, got %d", len(files))
	}
}

func TestFileCache_FilesByExtension(t *testing.T) {
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"main.go", "util.go", "README.md", "config.yaml"})

	cache := NewFileCache("/test/repo", false)
	cache.testGitRepo = mockGit

	// Act
	goFiles, err := cache.FilesByExtension(".go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert
	if len(goFiles) != 2 {
		t.Errorf("expected 2 .go files, got %d", len(goFiles))
	}
}

func TestFileCache_FilesInDir(t *testing.T) {
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"src/main.go", "src/util.go", "test/main_test.go", "README.md"})

	cache := NewFileCache("/test/repo", false)
	cache.testGitRepo = mockGit

	// Act
	srcFiles, err := cache.FilesInDir("src")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert
	if len(srcFiles) != 2 {
		t.Errorf("expected 2 files in src/, got %d", len(srcFiles))
	}
}

func TestFileCache_CachesResults(t *testing.T) {
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"file1.go", "file2.go"})

	cache := NewFileCache("/test/repo", false)
	cache.testGitRepo = mockGit

	// First call
	files1, _ := cache.TrackedFiles()
	// Second call
	files2, _ := cache.TrackedFiles()

	// Assert: same slice returned (cached)
	if &files1[0] != &files2[0] {
		t.Error("expected cached result to be returned")
	}
}

func TestFileCache_Invalidate(t *testing.T) {
	mockGit := git.NewMockRepository("/test/repo").
		WithTrackedFiles([]string{"file1.go"})

	cache := NewFileCache("/test/repo", false)
	cache.testGitRepo = mockGit

	// Populate cache
	_, _ = cache.TrackedFiles()

	// Invalidate
	cache.Invalidate()

	// Update mock
	mockGit.WithTrackedFiles([]string{"file1.go", "file2.go"})

	// Get files again
	files, _ := cache.TrackedFiles()

	// Assert: should have new files
	if len(files) != 2 {
		t.Errorf("expected 2 files after invalidate, got %d", len(files))
	}
}
