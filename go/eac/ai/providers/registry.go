// File: go/eac/ai/providers/registry.go
package providers

import (
	"fmt"

	"github.com/ready-to-release/eac/go/eac/ai"
)

// ExecutorRegistry defines the interface for registering providers
type ExecutorRegistry interface {
	RegisterProvider(name string, factory ai.ProviderFactory)
}

// RegisterBuiltIn registers all built-in providers with an executor
func RegisterBuiltIn(executor ExecutorRegistry) {
	// Register claude-cli provider
	executor.RegisterProvider("claude-cli", func(config *ai.Config) (ai.Provider, error) {
		return NewClaudeCLI(), nil
	})

	// Register claude-api provider
	executor.RegisterProvider("claude-api", func(config *ai.Config) (ai.Provider, error) {
		if config.APIKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is required for claude-api provider")
		}
		return NewClaudeAPI(config.APIKey, config.Model)
	})

	// Register openai provider
	executor.RegisterProvider("openai", func(config *ai.Config) (ai.Provider, error) {
		if config.APIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required for openai provider")
		}
		return NewOpenAI(config.APIKey, config.Model)
	})

	// Register gemini provider
	executor.RegisterProvider("gemini", func(config *ai.Config) (ai.Provider, error) {
		if config.APIKey == "" {
			return nil, fmt.Errorf("GOOGLE_API_KEY is required for gemini provider")
		}
		return NewGemini(config.APIKey, config.Model)
	})

	// Register test provider for acceptance tests
	// Reads mock response from .r2r/test/ai-mock.txt or R2R_TEST_AI_RESPONSE env var
	executor.RegisterProvider("test", func(config *ai.Config) (ai.Provider, error) {
		return NewTestProvider(), nil
	})
}
