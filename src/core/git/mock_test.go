//go:build L1 && ov
// +build L1,ov

package git

import (
	"errors"
	"testing"
)

func TestMockRepository_RootPath(t *testing.T) {
	mock := NewMockRepository("/test/repo")
	if mock.RootPath() != "/test/repo" {
		t.Errorf("Expected /test/repo, got %s", mock.RootPath())
	}
}

func TestMockRepository_RemoteURL(t *testing.T) {
	mock := NewMockRepository("/test").
		WithRemote("origin", "https://github.com/test/repo.git")

	url, err := mock.RemoteURL("origin")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if url != "https://github.com/test/repo.git" {
		t.Errorf("Expected https://github.com/test/repo.git, got %s", url)
	}

	// Test missing remote
	_, err = mock.RemoteURL("upstream")
	if err == nil {
		t.Error("Expected error for missing remote")
	}
}

func TestMockRepository_RemoteURL_Error(t *testing.T) {
	expectedErr := errors.New("network error")
	mock := NewMockRepository("/test").
		WithRemote("origin", "https://github.com/test/repo.git").
		WithError("RemoteURL", expectedErr)

	_, err := mock.RemoteURL("origin")
	if err != expectedErr {
		t.Errorf("Expected %v, got %v", expectedErr, err)
	}
}

func TestMockRepository_CurrentBranch(t *testing.T) {
	mock := NewMockRepository("/test").
		WithCurrentBranch("main")

	branch, err := mock.CurrentBranch()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if branch != "main" {
		t.Errorf("Expected main, got %s", branch)
	}
}

func TestMockRepository_CurrentBranch_NotSet(t *testing.T) {
	mock := NewMockRepository("/test")

	_, err := mock.CurrentBranch()
	if err == nil {
		t.Error("Expected error when branch not set")
	}
}

func TestMockRepository_HeadShortSHA(t *testing.T) {
	mock := NewMockRepository("/test").
		WithHeadSHA("abc1234567890")

	sha, err := mock.HeadShortSHA()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if sha != "abc1234" {
		t.Errorf("Expected abc1234, got %s", sha)
	}
}

func TestMockRepository_TrackedFiles(t *testing.T) {
	mock := NewMockRepository("/test").
		WithTrackedFiles([]string{"main.go", "go.mod", "README.md"})

	files, err := mock.TrackedFiles()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}
}

func TestMockRepository_StagedFiles(t *testing.T) {
	mock := NewMockRepository("/test").
		WithStagedFiles([]string{"new.go"})

	files, err := mock.StagedFiles()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
}

func TestMockRepository_IsFileTracked(t *testing.T) {
	mock := NewMockRepository("/test").
		WithTrackedFiles([]string{"main.go", "go.mod"})

	tests := []struct {
		path     string
		expected bool
	}{
		{"main.go", true},
		{"go.mod", true},
		{"new.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := mock.IsFileTracked(tt.path)
			if result != tt.expected {
				t.Errorf("IsFileTracked(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMockRepository_IsFileIgnored(t *testing.T) {
	mock := NewMockRepository("/test").
		WithIgnoredFiles([]string{"*.log", "build/"})

	tests := []struct {
		path     string
		expected bool
	}{
		{"*.log", true},
		{"build/", true},
		{"main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := mock.IsFileIgnored(tt.path)
			if result != tt.expected {
				t.Errorf("IsFileIgnored(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMockRepository_Add(t *testing.T) {
	mock := NewMockRepository("/test")

	err := mock.Add("new.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = mock.Add("another.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	added := mock.AddedFiles()
	if len(added) != 2 {
		t.Errorf("Expected 2 added files, got %d", len(added))
	}
}

func TestMockRepository_Add_Error(t *testing.T) {
	expectedErr := errors.New("permission denied")
	mock := NewMockRepository("/test").
		WithError("Add", expectedErr)

	err := mock.Add("file.go")
	if err != expectedErr {
		t.Errorf("Expected %v, got %v", expectedErr, err)
	}
}

func TestMockRepository_Commit(t *testing.T) {
	mock := NewMockRepository("/test")

	hash, err := mock.Commit("Initial commit", "Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	commits := mock.Commits()
	if len(commits) != 1 {
		t.Fatalf("Expected 1 commit, got %d", len(commits))
	}
	if commits[0].Message != "Initial commit" {
		t.Errorf("Expected 'Initial commit', got '%s'", commits[0].Message)
	}
	if commits[0].AuthorName != "Test User" {
		t.Errorf("Expected 'Test User', got '%s'", commits[0].AuthorName)
	}
}

func TestMockRepository_ConfigSet(t *testing.T) {
	mock := NewMockRepository("/test")

	err := mock.ConfigSet("user", "name", "Test User")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	val, ok := mock.Config("user", "name")
	if !ok {
		t.Error("Config not found")
	}
	if val != "Test User" {
		t.Errorf("Expected 'Test User', got '%s'", val)
	}
}

func TestMockRepository_AddRemote(t *testing.T) {
	mock := NewMockRepository("/test")

	err := mock.AddRemote("origin", "https://github.com/test/repo.git")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should fail for duplicate
	err = mock.AddRemote("origin", "https://other.com/repo.git")
	if err == nil {
		t.Error("Expected error for duplicate remote")
	}
}

func TestMockRepository_GoGitRepo(t *testing.T) {
	mock := NewMockRepository("/test")

	repo := mock.GoGitRepo()
	if repo != nil {
		t.Error("Expected nil for mock GoGitRepo")
	}
}

func TestMockRepository_BuilderChaining(t *testing.T) {
	mock := NewMockRepository("/test").
		WithRemote("origin", "https://github.com/test/repo.git").
		WithCurrentBranch("main").
		WithHeadSHA("abc1234567890").
		WithTrackedFiles([]string{"main.go"}).
		WithStagedFiles([]string{"new.go"}).
		WithIgnoredFile("*.log")

	// Verify all settings work
	if mock.RootPath() != "/test" {
		t.Error("RootPath failed")
	}

	url, _ := mock.RemoteURL("origin")
	if url != "https://github.com/test/repo.git" {
		t.Error("RemoteURL failed")
	}

	branch, _ := mock.CurrentBranch()
	if branch != "main" {
		t.Error("CurrentBranch failed")
	}

	sha, _ := mock.HeadShortSHA()
	if sha != "abc1234" {
		t.Error("HeadShortSHA failed")
	}

	files, _ := mock.TrackedFiles()
	if len(files) != 1 {
		t.Error("TrackedFiles failed")
	}

	staged, _ := mock.StagedFiles()
	if len(staged) != 1 {
		t.Error("StagedFiles failed")
	}

	if !mock.IsFileIgnored("*.log") {
		t.Error("IsFileIgnored failed")
	}
}

func TestMockRepository_StagedDiff(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
new file mode 100644
--- /dev/null
+++ b/main.go
@@ -0,0 +1 @@
+package main`

	mock := NewMockRepository("/test").
		WithStagedDiff(diff)

	result, err := mock.StagedDiff()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != diff {
		t.Errorf("Expected diff content, got %s", result)
	}
}

func TestMockRepository_StagedDiff_Error(t *testing.T) {
	expectedErr := errors.New("diff error")
	mock := NewMockRepository("/test").
		WithStagedDiff("some diff").
		WithError("StagedDiff", expectedErr)

	_, err := mock.StagedDiff()
	if err != expectedErr {
		t.Errorf("Expected %v, got %v", expectedErr, err)
	}
}

func TestMockRepository_StagedDiffStats(t *testing.T) {
	stats := " main.go | 1 +\n 1 file changed, 1 insertion(+)"

	mock := NewMockRepository("/test").
		WithStagedDiffStats(stats)

	result, err := mock.StagedDiffStats()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != stats {
		t.Errorf("Expected stats content, got %s", result)
	}
}

func TestMockRepository_StagedDiffStats_Error(t *testing.T) {
	expectedErr := errors.New("stats error")
	mock := NewMockRepository("/test").
		WithStagedDiffStats("some stats").
		WithError("StagedDiffStats", expectedErr)

	_, err := mock.StagedDiffStats()
	if err != expectedErr {
		t.Errorf("Expected %v, got %v", expectedErr, err)
	}
}
