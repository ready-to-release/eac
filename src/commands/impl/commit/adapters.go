package commit

import (
	"context"

	"github.com/ready-to-release/eac/src/core/ai"
)

// aiExecutorAdapter adapts ai.Executor to contract.AIExecutor interface
type aiExecutorAdapter struct {
	executor *ai.Executor
	model    string
}

// Execute adapts the ai.Executor.Execute signature to contract.AIExecutor.Execute
func (a *aiExecutorAdapter) Execute(ctx context.Context, prompt string, opts ...interface{}) (string, error) {
	// Convert interface{} options to ai.Option
	var aiOpts []ai.Option

	// Add model option if specified
	if a.model != "" {
		aiOpts = append(aiOpts, ai.WithModel(a.model))
	}

	// Add any additional options
	for _, opt := range opts {
		if aiOpt, ok := opt.(ai.Option); ok {
			aiOpts = append(aiOpts, aiOpt)
		}
	}

	return a.executor.Execute(ctx, prompt, aiOpts...)
}
