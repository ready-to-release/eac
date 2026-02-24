// Package ai defines the interface contract for AI provider integration.
package ai

import "context"

// Provider defines the port that all AI provider adapters must implement.
type Provider interface {
	// Name returns the provider name for identification and logging.
	Name() string

	// Execute sends input to the AI provider and returns the response.
	// Options can modify behavior (model, temperature, max tokens, etc.)
	Execute(ctx context.Context, input string, opts ...Option) (string, error)
}

// ProviderFactory creates a Provider from configuration.
type ProviderFactory func(config *ProviderConfig) (Provider, error)

// ConfigLoader loads AI provider configuration from the environment.
type ConfigLoader interface {
	Load(workspaceRoot string) (*ProviderConfig, error)
	LoadWithOverrides(workspaceRoot, teamPath, personalPath string) (*ProviderConfig, error)
}

// GenerationExecutor is the typed interface consumed by the generation/retry layer.
// This replaces validation.AIExecutor which uses interface{} parameters.
type GenerationExecutor interface {
	Execute(ctx context.Context, prompt string, opts ...Option) (string, error)
}

// GenerationExecutorWithProviderInfo extends GenerationExecutor with provider metadata.
type GenerationExecutorWithProviderInfo interface {
	GenerationExecutor
	GetProviderName() string
}
