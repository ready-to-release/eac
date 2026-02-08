package init

import (
	"github.com/ready-to-release/eac/go/adapters/ai"
	"github.com/ready-to-release/eac/go/adapters/ai/providers"
	"github.com/ready-to-release/eac/go/core/git"
)

// Deps holds injectable dependencies for testing.
type Deps struct {
	GetGitRepo    func(workspaceRoot string) (git.GitRepository, error)
	GetAIExecutor func(repoRoot string) *ai.Executor
}

func defaultDeps() *Deps {
	lazyRepo := &git.LazyRepo{}
	return &Deps{
		GetGitRepo: lazyRepo.Get,
		GetAIExecutor: func(repoRoot string) *ai.Executor {
			executor := ai.NewExecutor(repoRoot)
			providers.RegisterBuiltIn(executor)
			return executor
		},
	}
}
