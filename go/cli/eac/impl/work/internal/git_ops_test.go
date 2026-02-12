//go:build L1

package internal

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/tool"
)

func TestNewDefaultGitOps(t *testing.T) {
	t.Run("returns non-nil implementation", func(t *testing.T) {
		ops := NewDefaultGitOps("/tmp/test-repo", tool.NewToolSystemForTesting())
		if ops == nil {
			t.Fatal("NewDefaultGitOps() returned nil")
		}

		// Should be default implementation
		if _, ok := ops.(*defaultGitOps); !ok {
			t.Error("Expected defaultGitOps implementation")
		}
	})
}

func TestDefaultDeps(t *testing.T) {
	t.Run("returns deps with production git ops", func(t *testing.T) {
		deps := DefaultDeps("/tmp/test-repo")
		if deps == nil {
			t.Fatal("DefaultDeps() returned nil")
		}
		if deps.GitOps == nil {
			t.Fatal("DefaultDeps().GitOps is nil")
		}

		// Should be default implementation
		if _, ok := deps.GitOps.(*defaultGitOps); !ok {
			t.Error("Expected defaultGitOps implementation")
		}
	})
}

func TestDepsWithMock(t *testing.T) {
	t.Run("deps accepts mock implementation", func(t *testing.T) {
		mock := &MockGitOps{}
		deps := &Deps{GitOps: mock}

		if deps.GitOps != mock {
			t.Error("Deps did not hold the mock implementation")
		}
	})
}

// MockGitOps is a mock implementation for testing
type MockGitOps struct {
	CreateWorktreeFunc      func(path, branch, base string) error
	RemoveWorktreeFunc      func(path string) error
	ListWorktreesFunc       func() ([]Worktree, error)
	WorktreeExistsFunc      func(branch string) (bool, error)
	BranchExistsFunc        func(branch string) (bool, error)
	GetCurrentBranchFunc    func(path string) (string, error)
	DeleteBranchFunc        func(branch string, force bool) error
	IsWorktreeCleanFunc     func(path string) (bool, error)
	FetchBranchFunc         func(branch string) error
	PushBranchFunc          func(branch string, force bool) error
	RebaseFunc              func(target string) error
	MergeFunc               func(branch string, squash bool) error
	MergeAbortFunc          func() error
	RebaseAbortFunc         func() error
	StashFunc               func(message string) error
	StashPopFunc            func() error
	GetCommitCountFunc      func(base, head string) (int, error)
	GetConflictingFilesFunc func() ([]string, error)
}

func (m *MockGitOps) CreateWorktree(path, branch, base string) error {
	if m.CreateWorktreeFunc != nil {
		return m.CreateWorktreeFunc(path, branch, base)
	}
	return nil
}

func (m *MockGitOps) RemoveWorktree(path string) error {
	if m.RemoveWorktreeFunc != nil {
		return m.RemoveWorktreeFunc(path)
	}
	return nil
}

func (m *MockGitOps) ListWorktrees() ([]Worktree, error) {
	if m.ListWorktreesFunc != nil {
		return m.ListWorktreesFunc()
	}
	return []Worktree{}, nil
}

func (m *MockGitOps) WorktreeExists(branch string) (bool, error) {
	if m.WorktreeExistsFunc != nil {
		return m.WorktreeExistsFunc(branch)
	}
	return false, nil
}

func (m *MockGitOps) BranchExists(branch string) (bool, error) {
	if m.BranchExistsFunc != nil {
		return m.BranchExistsFunc(branch)
	}
	return false, nil
}

func (m *MockGitOps) GetCurrentBranch(path string) (string, error) {
	if m.GetCurrentBranchFunc != nil {
		return m.GetCurrentBranchFunc(path)
	}
	return "main", nil
}

func (m *MockGitOps) DeleteBranch(branch string, force bool) error {
	if m.DeleteBranchFunc != nil {
		return m.DeleteBranchFunc(branch, force)
	}
	return nil
}

func (m *MockGitOps) IsWorktreeClean(path string) (bool, error) {
	if m.IsWorktreeCleanFunc != nil {
		return m.IsWorktreeCleanFunc(path)
	}
	return true, nil
}

func (m *MockGitOps) FetchBranch(branch string) error {
	if m.FetchBranchFunc != nil {
		return m.FetchBranchFunc(branch)
	}
	return nil
}

func (m *MockGitOps) PushBranch(branch string, force bool) error {
	if m.PushBranchFunc != nil {
		return m.PushBranchFunc(branch, force)
	}
	return nil
}

func (m *MockGitOps) Rebase(target string) error {
	if m.RebaseFunc != nil {
		return m.RebaseFunc(target)
	}
	return nil
}

func (m *MockGitOps) Merge(branch string, squash bool) error {
	if m.MergeFunc != nil {
		return m.MergeFunc(branch, squash)
	}
	return nil
}

func (m *MockGitOps) MergeAbort() error {
	if m.MergeAbortFunc != nil {
		return m.MergeAbortFunc()
	}
	return nil
}

func (m *MockGitOps) RebaseAbort() error {
	if m.RebaseAbortFunc != nil {
		return m.RebaseAbortFunc()
	}
	return nil
}

func (m *MockGitOps) Stash(message string) error {
	if m.StashFunc != nil {
		return m.StashFunc(message)
	}
	return nil
}

func (m *MockGitOps) StashPop() error {
	if m.StashPopFunc != nil {
		return m.StashPopFunc()
	}
	return nil
}

func (m *MockGitOps) GetCommitCount(base, head string) (int, error) {
	if m.GetCommitCountFunc != nil {
		return m.GetCommitCountFunc(base, head)
	}
	return 0, nil
}

func (m *MockGitOps) GetConflictingFiles() ([]string, error) {
	if m.GetConflictingFilesFunc != nil {
		return m.GetConflictingFilesFunc()
	}
	return []string{}, nil
}
