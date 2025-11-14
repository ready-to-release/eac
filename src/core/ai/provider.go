// File: src/core/ai/provider.go
// Package ai provides an abstraction layer for AI provider integrations
package ai

import "context"

// Provider defines the simple interface that all AI providers must implement.
//
// Intent: Provide a minimal, consistent interface for AI provider integrations.
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Single Execute() method with clear signature
//   - Name() method for provider identification
//   - Functional options pattern for optional parameters
//   - No hidden state or complex lifecycle
//
// Easy to change:
//   - Providers are independently replaceable
//   - Adding new providers doesn't modify existing code
//   - Options can be extended without breaking existing calls
//   - Context.Context allows for cancellation and timeouts
//
// Hard to break:
//   - Interface is minimal - hard to misimplement
//   - Compile-time check ensures all providers implement interface
//   - Context parameter prevents hanging operations
//   - Options are optional - no required configuration in method signature
type Provider interface {
	// Name returns the provider name for identification and logging
	Name() string

	// Execute sends input to the AI provider and returns the response.
	// Options can modify behavior (model, temperature, max tokens, etc.)
	Execute(ctx context.Context, input string, opts ...Option) (string, error)
}

// Option is a functional option for configuring AI provider execution
type Option func(*ExecuteOptions)

// ExecuteOptions holds optional parameters for AI execution
type ExecuteOptions struct {
	Model       string  // Model to use (e.g., "haiku", "sonnet", "gpt-4")
	Temperature float64 // Randomness (0.0 - 1.0), default 0.3
	MaxTokens   int     // Max response length, default 4000
}

// WithModel sets the AI model to use for execution
func WithModel(model string) Option {
	return func(opts *ExecuteOptions) {
		opts.Model = model
	}
}

// WithTemperature sets the randomness/creativity level (0.0 = deterministic, 1.0 = creative)
func WithTemperature(temp float64) Option {
	return func(opts *ExecuteOptions) {
		opts.Temperature = temp
	}
}

// WithMaxTokens sets the maximum number of tokens in the response
func WithMaxTokens(max int) Option {
	return func(opts *ExecuteOptions) {
		opts.MaxTokens = max
	}
}

// ApplyOptions applies functional options to ExecuteOptions with defaults
func ApplyOptions(opts ...Option) *ExecuteOptions {
	// Set defaults
	options := &ExecuteOptions{
		Temperature: 0.3,
		MaxTokens:   4000,
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	return options
}
