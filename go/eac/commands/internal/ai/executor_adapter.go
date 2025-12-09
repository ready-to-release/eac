// Package ai provides AI execution and adaptation utilities
package ai

import "context"

// ExecutorAdapter adapts ai.Executor to contracts.AIExecutor interface
//
// This adapter bridges the gap between the concrete ai.Executor type
// and the interface expected by the contracts package, allowing for
// flexible dependency injection and easier testing.
//
// Intent: Provide a clean adapter for AI executor in contract-based generation
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Simple adapter pattern with single responsibility
//   - Clear interface conversion logic
//   - Minimal abstraction
//
// Easy to change:
//   - Decoupled from contract implementation details
//   - Optional model override support
//   - Pure function for type conversion
//
// Hard to break:
//   - Nil-safe context handling
//   - Type-safe option conversion
//   - Graceful fallback for missing context
type ExecutorAdapter struct {
	executor *Executor
	model    string // Optional model override
}

// NewExecutorAdapter creates a new adapter for an AI executor
func NewExecutorAdapter(executor *Executor) *ExecutorAdapter {
	return &ExecutorAdapter{
		executor: executor,
	}
}

// NewExecutorAdapterWithModel creates an adapter with a specific model override
func NewExecutorAdapterWithModel(executor *Executor, model string) *ExecutorAdapter {
	return &ExecutorAdapter{
		executor: executor,
		model:    model,
	}
}

// Execute adapts the ai.Executor.Execute signature to contracts.AIExecutor.Execute
//
// This method performs the necessary type conversions to bridge between
// the interface{} types used by contracts and the concrete types used by ai.Executor.
//
// Context Handling:
//   - Attempts to convert interface{} context to context.Context
//   - Falls back to context.Background() if conversion fails
//   - Never panics on type assertion failure
//
// Options Handling:
//   - Filters interface{} options to ai.Option types
//   - Ignores incompatible option types
//   - Adds model override if specified
func (a *ExecutorAdapter) Execute(ctx interface{}, prompt string, opts ...interface{}) (string, error) {
	// Convert interface{} context to context.Context
	var actualCtx context.Context
	if c, ok := ctx.(context.Context); ok {
		actualCtx = c
	} else {
		actualCtx = context.Background()
	}

	// Convert interface{} options to ai.Option
	var aiOpts []Option
	for _, opt := range opts {
		if aiOpt, ok := opt.(Option); ok {
			aiOpts = append(aiOpts, aiOpt)
		}
	}

	// Add model override if specified
	if a.model != "" {
		aiOpts = append(aiOpts, WithModel(a.model))
	}

	return a.executor.Execute(actualCtx, prompt, aiOpts...)
}
