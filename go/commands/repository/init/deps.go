package init

import (
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
	"github.com/ready-to-release/eac/go/core/git"
)

// Deps holds injectable dependencies for testing.
type Deps struct {
	GetGitRepo    func(workspaceRoot string) (git.GitRepository, error)
	GetAIExecutor func(repoRoot string) *aiproviders.Executor

	// PrimaryGenerator is the generator used when AI is requested.
	// Nil means the default AIGenerator is constructed from aiProvider at runtime.
	PrimaryGenerator ConfigGenerator

	// FallbackGenerator is the generator used when the primary fails or no AI is requested.
	// Nil means NewRuleBasedGenerator() is used.
	FallbackGenerator ConfigGenerator
}

func defaultDeps() *Deps {
	lazyRepo := &git.LazyRepo{}
	return &Deps{
		GetGitRepo: lazyRepo.Get,
		GetAIExecutor: func(repoRoot string) *aiproviders.Executor {
			return aiproviders.NewExecutor(repoRoot, nil)
		},
		// PrimaryGenerator and FallbackGenerator are nil — constructed at runtime
	}
}
