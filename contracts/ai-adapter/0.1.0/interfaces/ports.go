// Package ai defines the interface contract for AI provider integration.
package ai

import "context"

// Provider defines the interface that all AI providers must implement.
type Provider interface {
	// Name returns the provider name for identification and logging
	Name() string

	// Execute sends input to the AI provider and returns the response.
	// Options can modify behavior (model, temperature, max tokens, etc.)
	Execute(ctx context.Context, input string, opts ...Option) (string, error)
}

// Option is a functional option for configuring AI provider execution.
type Option func(*ExecuteOptions)

// ExecuteOptions holds optional parameters for AI execution.
type ExecuteOptions struct {
	Model       string  // Model to use (e.g., "haiku", "sonnet", "gpt-4")
	Temperature float64 // Randomness (0.0 - 1.0), default 0.3
	MaxTokens   int     // Max response length, default 4000
	Debug       bool    // Include debug logs in output, default false
}

// Executor orchestrates AI provider selection and execution.
type Executor interface {
	Execute(ctx context.Context, input string, opts ...Option) (string, error)
	RegisterProvider(name string, factory ProviderFactory)
	GetLastUsedProvider() Provider
}

// ProviderFactory creates a provider from configuration.
type ProviderFactory func(config *Config) (Provider, error)

// Config represents AI and Git configuration loaded from .eac/ai-provider.yml.
type Config struct {
	AI  AIConfig  `yaml:"ai"`  // AI provider configuration
	Git GitConfig `yaml:"git"` // Git provider configuration
}

// AIConfig represents AI provider configuration.
type AIConfig struct {
	Provider string `yaml:"provider"` // Provider identifier (claude-api, claude-cli, openai, gemini)
	Model    string `yaml:"model"`    // AI model to use
	Endpoint string `yaml:"endpoint"` // API endpoint URL (empty for claude-cli)
	APIKey   string `yaml:"api_key"`  // API key with ${VAR} substitution (empty for claude-cli)
}

// GitConfig represents Git provider configuration.
type GitConfig struct {
	Token string `yaml:"token"` // Generic git provider token (GitHub, GitLab, Bitbucket, etc.) with ${VAR} substitution
}

// ConfigLoader loads AI configuration.
type ConfigLoader interface {
	Load(workspaceRoot string) (*Config, error)
	LoadWithOverrides(workspaceRoot, teamPath, personalPath string) (*Config, error)
}
