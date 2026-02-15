package design

import "github.com/ready-to-release/eac/go/core/git"

// Deps holds injectable dependencies for testing.
type Deps struct {
	AIResponse string // Non-empty bypasses AI, returns this response directly
	GetGitRepo func(workspaceRoot string) (git.GitRepository, error)
}

func defaultDeps() *Deps {
	lazyRepo := &git.LazyRepo{}
	return &Deps{
		GetGitRepo: lazyRepo.Get,
	}
}
