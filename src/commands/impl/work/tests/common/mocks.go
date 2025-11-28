package common

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
)

// MockGitOps provides a mock implementation of WorkGitOperations
type MockGitOps struct {
	Worktrees        []internal.Worktree
	Branches         map[string]bool
	RemoteBranches   map[string]bool
	CurrentBranch    string
	IsClean          bool
	ConflictingFiles []string
	CommitCount      int

	// Call tracking
	CreateWorktreeCalled   bool
	RemoveWorktreeCalled   bool
	DeleteBranchCalled     bool
	DeleteRemoteCalled     bool
	FetchCalled            bool
	PushCalled             bool
	RebaseCalled           bool
	MergeCalled            bool
	StashCalled            bool

	// Error simulation
	ShouldFailCreate       bool
	ShouldFailRemove       bool
	ShouldFailFetch        bool
	ShouldFailPush         bool
	ShouldFailRebase       bool
	ShouldFailMerge        bool
}

// NewMockGitOps creates a new mock GitOps
func NewMockGitOps() *MockGitOps {
	return &MockGitOps{
		Worktrees:      []internal.Worktree{},
		Branches:       make(map[string]bool),
		RemoteBranches: make(map[string]bool),
		CurrentBranch:  "main",
		IsClean:        true,
		CommitCount:    0,
	}
}

func (m *MockGitOps) CreateWorktree(path, branch, base string) error {
	m.CreateWorktreeCalled = true
	if m.ShouldFailCreate {
		return fmt.Errorf("failed to create worktree")
	}
	m.Worktrees = append(m.Worktrees, internal.Worktree{
		Path:   path,
		Branch: branch,
	})
	m.Branches[branch] = true
	return nil
}

func (m *MockGitOps) RemoveWorktree(path string) error {
	m.RemoveWorktreeCalled = true
	if m.ShouldFailRemove {
		return fmt.Errorf("failed to remove worktree")
	}
	for i, wt := range m.Worktrees {
		if wt.Path == path {
			m.Worktrees = append(m.Worktrees[:i], m.Worktrees[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockGitOps) ListWorktrees() ([]internal.Worktree, error) {
	return m.Worktrees, nil
}

func (m *MockGitOps) WorktreeExists(branch string) (bool, error) {
	for _, wt := range m.Worktrees {
		if wt.Branch == branch {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockGitOps) BranchExists(branch string) (bool, error) {
	return m.Branches[branch], nil
}

func (m *MockGitOps) GetCurrentBranch(path string) (string, error) {
	return m.CurrentBranch, nil
}

func (m *MockGitOps) DeleteBranch(branch string, force bool) error {
	m.DeleteBranchCalled = true
	delete(m.Branches, branch)
	return nil
}

func (m *MockGitOps) IsWorktreeClean(path string) (bool, error) {
	return m.IsClean, nil
}

func (m *MockGitOps) FetchBranch(branch string) error {
	m.FetchCalled = true
	if m.ShouldFailFetch {
		return fmt.Errorf("failed to fetch")
	}
	return nil
}

func (m *MockGitOps) PushBranch(branch string, force bool) error {
	m.PushCalled = true
	if m.ShouldFailPush {
		return fmt.Errorf("failed to push")
	}
	m.RemoteBranches[branch] = true
	return nil
}

func (m *MockGitOps) Rebase(target string) error {
	m.RebaseCalled = true
	if m.ShouldFailRebase {
		return fmt.Errorf("rebase conflict")
	}
	return nil
}

func (m *MockGitOps) Merge(branch string, squash bool) error {
	m.MergeCalled = true
	if m.ShouldFailMerge {
		return fmt.Errorf("merge conflict")
	}
	return nil
}

func (m *MockGitOps) MergeAbort() error {
	return nil
}

func (m *MockGitOps) RebaseAbort() error {
	return nil
}

func (m *MockGitOps) Stash(message string) error {
	m.StashCalled = true
	return nil
}

func (m *MockGitOps) StashPop() error {
	return nil
}

func (m *MockGitOps) GetCommitCount(base, head string) (int, error) {
	return m.CommitCount, nil
}

func (m *MockGitOps) GetConflictingFiles() ([]string, error) {
	return m.ConflictingFiles, nil
}

// Helper methods for test setup

func (m *MockGitOps) AddWorktree(path, branch string) {
	m.Worktrees = append(m.Worktrees, internal.Worktree{
		Path:   path,
		Branch: branch,
	})
	m.Branches[branch] = true
}

func (m *MockGitOps) AddBranch(branch string) {
	m.Branches[branch] = true
}

func (m *MockGitOps) AddRemoteBranch(branch string) {
	m.RemoteBranches[branch] = true
}

func (m *MockGitOps) SetCurrentBranch(branch string) {
	m.CurrentBranch = branch
}

func (m *MockGitOps) SetClean(clean bool) {
	m.IsClean = clean
}

func (m *MockGitOps) SetConflicts(files []string) {
	m.ConflictingFiles = files
}

func (m *MockGitOps) SetCommitCount(count int) {
	m.CommitCount = count
}

// RemoteBranchExists checks if a remote branch exists
func (m *MockGitOps) RemoteBranchExists(branch string) bool {
	// Check if it's a remote ref format
	if strings.HasPrefix(branch, "origin/") {
		branch = strings.TrimPrefix(branch, "origin/")
	}
	return m.RemoteBranches[branch]
}
