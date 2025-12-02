package commit

import (
	"context"

	"github.com/ready-to-release/eac/go/eac/ai"
)

// aiExecutorAdapter adapts ai.Executor to contracts.AIExecutor interface.
// It also implements contracts.AIExecutorWithProviderInfo to expose provider metadata.
type aiExecutorAdapter struct {
	executor *ai.Executor
	model    string
}

// Execute adapts the ai.Executor.Execute signature to contracts.AIExecutor.Execute
func (a *aiExecutorAdapter) Execute(ctx interface{}, prompt string, opts ...interface{}) (string, error) {
	// Convert interface{} context to context.Context
	var actualCtx context.Context
	if c, ok := ctx.(context.Context); ok {
		actualCtx = c
	} else {
		actualCtx = context.Background()
	}

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

	return a.executor.Execute(actualCtx, prompt, aiOpts...)
}

// GetProviderName returns the name of the AI provider used for the last execution.
// Implements contracts.AIExecutorWithProviderInfo interface.
func (a *aiExecutorAdapter) GetProviderName() string {
	if provider := a.executor.GetLastUsedProvider(); provider != nil {
		return provider.Name()
	}
	return ""
}
