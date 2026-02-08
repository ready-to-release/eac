package git

import "fmt"

// LazyRepo provides a lazy-initialized git repository with test injection support.
// Each consumer gets its own LazyRepo instance as a package-level var.
//
// Usage in consumer packages:
//
//	var gitRepoProvider = &git.LazyRepo{}
//
//	func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
//	    return gitRepoProvider.Get(workspaceRoot)
//	}
//
//	func SetGitRepo(repo git.GitRepository) {
//	    gitRepoProvider.Set(repo)
//	}
type LazyRepo struct {
	repo GitRepository
}

// Get returns the git repository, opening one at workspaceRoot if needed.
func (lr *LazyRepo) Get(workspaceRoot string) (GitRepository, error) {
	if lr.repo != nil {
		return lr.repo, nil
	}
	repo, err := NewManager(nil).Open(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	lr.repo = repo
	return lr.repo, nil
}

// Set injects a mock repository for testing.
func (lr *LazyRepo) Set(repo GitRepository) {
	lr.repo = repo
}

// Reset clears the injected repository.
func (lr *LazyRepo) Reset() {
	lr.repo = nil
}
