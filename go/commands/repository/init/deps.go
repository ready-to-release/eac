package init

import (
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
	"github.com/ready-to-release/eac/go/core/git"
)

// Deps holds injectable dependencies for testing.
type Deps struct {
	GetGitRepo    func(workspaceRoot string) (git.GitRepository, error)
	GetAIExecutor func(repoRoot string) *aiproviders.Executor
}

func defaultDeps() *Deps {
	lazyRepo := &git.LazyRepo{}
	return &Deps{
		GetGitRepo: lazyRepo.Get,
		GetAIExecutor: func(repoRoot string) *aiproviders.Executor {
			return aiproviders.NewExecutor(repoRoot, nil)
		},
	}
}
