package repository

import (
	"errors"
	"testing"

	"github.com/ready-to-release/eac/src/core/git"
)

func TestGetGitContext_WithMock(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithRemote("origin", "https://github.com/test/repo.git").
		WithCurrentBranch("main")

	ctx, err := GetGitContext(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if ctx.RepositoryURL != "https://github.com/test/repo" {
		t.Errorf("Expected URL 'https://github.com/test/repo', got '%s'", ctx.RepositoryURL)
	}

	if ctx.CurrentBranch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", ctx.CurrentBranch)
	}

	if ctx.BaseCommit != "main" {
		t.Errorf("Expected base commit 'main', got '%s'", ctx.BaseCommit)
	}
}

func TestGetGitContext_SSHRemoteURL(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithRemote("origin", "git@github.com:owner/repo.git").
		WithCurrentBranch("feature-branch")

	ctx, err := GetGitContext(mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should convert SSH to HTTPS
	if ctx.RepositoryURL != "https://github.com/owner/repo" {
		t.Errorf("Expected URL 'https://github.com/owner/repo', got '%s'", ctx.RepositoryURL)
	}
}

func TestGetGitContext_RemoteURLError(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithCurrentBranch("main").
		WithError("RemoteURL", errors.New("remote not found"))

	_, err := GetGitContext(mock)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var repoErr *RepositoryError
	if !errors.As(err, &repoErr) {
		t.Errorf("Expected RepositoryError, got %T", err)
	}
}

func TestGetGitContext_BranchError(t *testing.T) {
	mock := git.NewMockRepository("/test/repo").
		WithRemote("origin", "https://github.com/test/repo.git").
		WithError("CurrentBranch", errors.New("detached HEAD"))

	_, err := GetGitContext(mock)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var repoErr *RepositoryError
	if !errors.As(err, &repoErr) {
		t.Errorf("Expected RepositoryError, got %T", err)
	}
}

func TestNormalizeGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTTPS with .git suffix",
			input:    "https://github.com/owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "HTTPS without .git suffix",
			input:    "https://github.com/owner/repo",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "SSH format with .git suffix",
			input:    "git@github.com:owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "SSH format without .git suffix",
			input:    "git@github.com:owner/repo",
			expected: "https://github.com/owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeGitHubURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeGitHubURL(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGitContext_BuildGitHubFileURL(t *testing.T) {
	ctx := &GitContext{
		RepositoryURL: "https://github.com/owner/repo",
		BaseCommit:    "main",
		CurrentBranch: "feature",
	}

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "simple path",
			filePath: "src/main.go",
			expected: "https://github.com/owner/repo/blob/main/src/main.go",
		},
		{
			name:     "nested path",
			filePath: "src/pkg/internal/handler.go",
			expected: "https://github.com/owner/repo/blob/main/src/pkg/internal/handler.go",
		},
		{
			name:     "windows path separators",
			filePath: "src\\pkg\\file.go",
			expected: "https://github.com/owner/repo/blob/main/src/pkg/file.go",
		},
		{
			name:     "root file",
			filePath: "README.md",
			expected: "https://github.com/owner/repo/blob/main/README.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.BuildGitHubFileURL(tt.filePath)
			if result != tt.expected {
				t.Errorf("BuildGitHubFileURL(%s) = %s, expected %s", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestGitContext_BuildGitHubBlobURL(t *testing.T) {
	ctx := &GitContext{
		RepositoryURL: "https://github.com/owner/repo",
		BaseCommit:    "abc123",
		CurrentBranch: "main",
	}

	// BuildGitHubBlobURL is an alias for BuildGitHubFileURL
	result := ctx.BuildGitHubBlobURL("file.txt")
	expected := "https://github.com/owner/repo/blob/abc123/file.txt"

	if result != expected {
		t.Errorf("BuildGitHubBlobURL() = %s, expected %s", result, expected)
	}
}
