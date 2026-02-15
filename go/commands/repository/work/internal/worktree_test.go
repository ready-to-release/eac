//go:build L1

package internal

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// parseWorktreeList tests
// ---------------------------------------------------------------------------

func TestParseWorktreeList(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCount  int
		expectedFirst  *Worktree
		expectedSecond *Worktree
	}{
		{
			name:          "empty input",
			input:         "",
			expectedCount: 0,
		},
		{
			name: "single worktree (main)",
			input: "worktree /home/user/repo\n" +
				"HEAD abc123def456789\n" +
				"branch refs/heads/main\n",
			expectedCount: 1,
			expectedFirst: &Worktree{
				Path:   "/home/user/repo",
				SHA:    "abc123def456789",
				Branch: "main",
			},
		},
		{
			name: "two worktrees separated by blank line",
			input: "worktree /home/user/repo\n" +
				"HEAD aaa111\n" +
				"branch refs/heads/main\n" +
				"\n" +
				"worktree /home/user/repo-feature\n" +
				"HEAD bbb222\n" +
				"branch refs/heads/feature/auth\n",
			expectedCount: 2,
			expectedFirst: &Worktree{
				Path:   "/home/user/repo",
				SHA:    "aaa111",
				Branch: "main",
			},
			expectedSecond: &Worktree{
				Path:   "/home/user/repo-feature",
				SHA:    "bbb222",
				Branch: "feature/auth",
			},
		},
		{
			name: "worktree without branch (detached HEAD)",
			input: "worktree /home/user/repo\n" +
				"HEAD abc123\n" +
				"detached\n",
			expectedCount: 1,
			expectedFirst: &Worktree{
				Path:   "/home/user/repo",
				SHA:    "abc123",
				Branch: "",
			},
		},
		{
			name: "branch with nested refs/heads prefix stripped",
			input: "worktree /repo\n" +
				"HEAD fff000\n" +
				"branch refs/heads/feature/deep/nested\n",
			expectedCount: 1,
			expectedFirst: &Worktree{
				Path:   "/repo",
				SHA:    "fff000",
				Branch: "feature/deep/nested",
			},
		},
		{
			name: "trailing newlines handled",
			input: "worktree /repo\n" +
				"HEAD abc\n" +
				"branch refs/heads/main\n" +
				"\n\n",
			expectedCount: 1,
			expectedFirst: &Worktree{
				Path:   "/repo",
				SHA:    "abc",
				Branch: "main",
			},
		},
		{
			name: "three worktrees",
			input: "worktree /r1\n" +
				"HEAD a1\n" +
				"branch refs/heads/main\n" +
				"\n" +
				"worktree /r2\n" +
				"HEAD b2\n" +
				"branch refs/heads/feature/a\n" +
				"\n" +
				"worktree /r3\n" +
				"HEAD c3\n" +
				"branch refs/heads/feature/b\n",
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: parseWorktreeList also calls IsWorktreeClean for each worktree,
			// which will fail for non-existent paths. The Clean field will default
			// to false in that case, which is acceptable for this test.
			worktrees, err := parseWorktreeList(tt.input)
			require.NoError(t, err)
			assert.Len(t, worktrees, tt.expectedCount)

			if tt.expectedFirst != nil && len(worktrees) > 0 {
				assert.Equal(t, tt.expectedFirst.Path, worktrees[0].Path)
				assert.Equal(t, tt.expectedFirst.SHA, worktrees[0].SHA)
				assert.Equal(t, tt.expectedFirst.Branch, worktrees[0].Branch)
			}

			if tt.expectedSecond != nil && len(worktrees) > 1 {
				assert.Equal(t, tt.expectedSecond.Path, worktrees[1].Path)
				assert.Equal(t, tt.expectedSecond.SHA, worktrees[1].SHA)
				assert.Equal(t, tt.expectedSecond.Branch, worktrees[1].Branch)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GenerateWorktreePath tests
// ---------------------------------------------------------------------------

func TestGenerateWorktreePath(t *testing.T) {
	tests := []struct {
		name       string
		repoName   string
		branchName string
		expected   string
	}{
		{
			name:       "feature branch with slash",
			repoName:   "my-repo",
			branchName: "feature/auth",
			expected:   filepath.Join("..", "my-repo-feature-auth"),
		},
		{
			name:       "bugfix with number",
			repoName:   "project",
			branchName: "bugfix/issue-123",
			expected:   filepath.Join("..", "project-bugfix-issue-123"),
		},
		{
			name:       "simple branch without slash",
			repoName:   "repo",
			branchName: "develop",
			expected:   filepath.Join("..", "repo-develop"),
		},
		{
			name:       "branch with multiple slashes",
			repoName:   "repo",
			branchName: "feat/scope/detail",
			expected:   filepath.Join("..", "repo-feat-scope-detail"),
		},
		{
			name:       "empty branch name",
			repoName:   "repo",
			branchName: "",
			expected:   filepath.Join("..", "repo-"),
		},
		{
			name:       "empty repo name",
			repoName:   "",
			branchName: "main",
			expected:   filepath.Join("..", "-main"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateWorktreePath(tt.repoName, tt.branchName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// GetRepoName tests
// ---------------------------------------------------------------------------

func TestGetRepoName(t *testing.T) {
	tests := []struct {
		name        string
		repoRoot    string
		expected    string
		windowsOnly bool
	}{
		{
			name:     "unix style path",
			repoRoot: "/home/user/projects/my-repo",
			expected: "my-repo",
		},
		{
			name:        "windows style path",
			repoRoot:    `C:\source\projects\my-repo`,
			expected:    "my-repo",
			windowsOnly: true,
		},
		{
			name:     "single directory",
			repoRoot: "my-repo",
			expected: "my-repo",
		},
		{
			name:     "dot (current directory)",
			repoRoot: ".",
			expected: ".",
		},
		{
			name:     "trailing slash is handled by filepath.Base",
			repoRoot: "/home/user/repo/",
			expected: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.windowsOnly && runtime.GOOS != "windows" {
				t.Skip("test only runs on Windows")
			}
			result := GetRepoName(tt.repoRoot)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// Worktree struct tests
// ---------------------------------------------------------------------------

func TestWorktreeStruct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		wt := Worktree{}
		assert.Empty(t, wt.Path)
		assert.Empty(t, wt.Branch)
		assert.Empty(t, wt.SHA)
		assert.False(t, wt.Clean)
	})

	t.Run("populated struct", func(t *testing.T) {
		wt := Worktree{
			Path:   "/home/user/repo",
			Branch: "feature/auth",
			SHA:    "abc123",
			Clean:  true,
		}
		assert.Equal(t, "/home/user/repo", wt.Path)
		assert.Equal(t, "feature/auth", wt.Branch)
		assert.Equal(t, "abc123", wt.SHA)
		assert.True(t, wt.Clean)
	})
}

// ---------------------------------------------------------------------------
// BaseConfig struct tests
// ---------------------------------------------------------------------------

func TestBaseConfigStruct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		config := BaseConfig{}
		assert.False(t, config.Debug)
		assert.Empty(t, config.RepoRoot)
		assert.Nil(t, config.Logger)
		assert.Nil(t, config.GitOps)
	})

	t.Run("with mock git ops", func(t *testing.T) {
		mock := &MockGitOps{}
		config := BaseConfig{
			Debug:    true,
			RepoRoot: "/tmp/repo",
			GitOps:   mock,
		}
		assert.True(t, config.Debug)
		assert.Equal(t, "/tmp/repo", config.RepoRoot)
		assert.NotNil(t, config.GitOps)
	})
}

// ---------------------------------------------------------------------------
// Deps struct tests
// ---------------------------------------------------------------------------

func TestDepsStruct(t *testing.T) {
	t.Run("DefaultDeps returns valid deps", func(t *testing.T) {
		deps := DefaultDeps("/tmp/repo")
		require.NotNil(t, deps)
		require.NotNil(t, deps.GitOps)

		_, ok := deps.GitOps.(*defaultGitOps)
		assert.True(t, ok, "GitOps should be a *defaultGitOps")
	})

	t.Run("Deps accepts mock", func(t *testing.T) {
		mock := &MockGitOps{}
		deps := &Deps{GitOps: mock}
		assert.Equal(t, mock, deps.GitOps)
	})
}

// ---------------------------------------------------------------------------
// GetAbsolutePath tests
// ---------------------------------------------------------------------------

func TestGetAbsolutePath(t *testing.T) {
	t.Run("absolute path returned as-is", func(t *testing.T) {
		var input string
		if runtime.GOOS == "windows" {
			input = `C:\projects\repo`
		} else {
			input = "/home/user/repo"
		}
		result, err := GetAbsolutePath(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("relative path is converted to absolute", func(t *testing.T) {
		result, err := GetAbsolutePath("relative/path")
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(result),
			"expected absolute path, got: %s", result)
		assert.Contains(t, result, "relative")
	})

	t.Run("dot path is converted to absolute", func(t *testing.T) {
		result, err := GetAbsolutePath(".")
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(result),
			"expected absolute path, got: %s", result)
	})
}

// ---------------------------------------------------------------------------
// MockGitOps default behavior tests
// ---------------------------------------------------------------------------

func TestMockGitOps_DefaultBehavior(t *testing.T) {
	mock := &MockGitOps{}

	t.Run("CreateWorktree returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.CreateWorktree("/path", "branch", "base"))
	})

	t.Run("RemoveWorktree returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.RemoveWorktree("/path"))
	})

	t.Run("ListWorktrees returns empty slice by default", func(t *testing.T) {
		wts, err := mock.ListWorktrees()
		assert.NoError(t, err)
		assert.Empty(t, wts)
	})

	t.Run("WorktreeExists returns false by default", func(t *testing.T) {
		exists, err := mock.WorktreeExists("branch")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("BranchExists returns false by default", func(t *testing.T) {
		exists, err := mock.BranchExists("branch")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("GetCurrentBranch returns main by default", func(t *testing.T) {
		branch, err := mock.GetCurrentBranch("/path")
		assert.NoError(t, err)
		assert.Equal(t, "main", branch)
	})

	t.Run("DeleteBranch returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.DeleteBranch("branch", false))
	})

	t.Run("IsWorktreeClean returns true by default", func(t *testing.T) {
		clean, err := mock.IsWorktreeClean("/path")
		assert.NoError(t, err)
		assert.True(t, clean)
	})

	t.Run("FetchBranch returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.FetchBranch("main"))
	})

	t.Run("PushBranch returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.PushBranch("branch", false))
	})

	t.Run("Rebase returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.Rebase("main"))
	})

	t.Run("Merge returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.Merge("branch", true))
	})

	t.Run("MergeAbort returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.MergeAbort())
	})

	t.Run("RebaseAbort returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.RebaseAbort())
	})

	t.Run("Stash returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.Stash("message"))
	})

	t.Run("StashPop returns nil by default", func(t *testing.T) {
		assert.NoError(t, mock.StashPop())
	})

	t.Run("GetCommitCount returns 0 by default", func(t *testing.T) {
		count, err := mock.GetCommitCount("base", "head")
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("GetConflictingFiles returns empty by default", func(t *testing.T) {
		files, err := mock.GetConflictingFiles()
		assert.NoError(t, err)
		assert.Empty(t, files)
	})
}

func TestMockGitOps_CustomFunctions(t *testing.T) {
	t.Run("custom GetCurrentBranch", func(t *testing.T) {
		mock := &MockGitOps{
			GetCurrentBranchFunc: func(path string) (string, error) {
				return "feature/test", nil
			},
		}
		branch, err := mock.GetCurrentBranch("/path")
		assert.NoError(t, err)
		assert.Equal(t, "feature/test", branch)
	})

	t.Run("custom IsWorktreeClean returns false", func(t *testing.T) {
		mock := &MockGitOps{
			IsWorktreeCleanFunc: func(path string) (bool, error) {
				return false, nil
			},
		}
		clean, err := mock.IsWorktreeClean("/path")
		assert.NoError(t, err)
		assert.False(t, clean)
	})

	t.Run("custom GetCommitCount", func(t *testing.T) {
		mock := &MockGitOps{
			GetCommitCountFunc: func(base, head string) (int, error) {
				return 42, nil
			},
		}
		count, err := mock.GetCommitCount("origin/main", "HEAD")
		assert.NoError(t, err)
		assert.Equal(t, 42, count)
	})

	t.Run("custom ListWorktrees", func(t *testing.T) {
		mock := &MockGitOps{
			ListWorktreesFunc: func() ([]Worktree, error) {
				return []Worktree{
					{Path: "/r1", Branch: "main", SHA: "a1"},
					{Path: "/r2", Branch: "feature/x", SHA: "b2"},
				}, nil
			},
		}
		wts, err := mock.ListWorktrees()
		assert.NoError(t, err)
		assert.Len(t, wts, 2)
		assert.Equal(t, "main", wts[0].Branch)
		assert.Equal(t, "feature/x", wts[1].Branch)
	})
}
