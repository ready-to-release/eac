// File: go/eac/commands/internal/ai/executor.go
package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// Executor orchestrates AI provider execution.
type Executor struct {
	workspaceRoot     string
	lastUsedProvider  Provider
	providerFactories map[string]ProviderFactory
}

// ProviderFactory creates a provider from configuration.
type ProviderFactory func(config *Config) (Provider, error)

// NewExecutor creates a new executor for the given workspace.
func NewExecutor(workspaceRoot string) *Executor {
	executor := &Executor{
		workspaceRoot:     workspaceRoot,
		providerFactories: make(map[string]ProviderFactory),
	}

	// Provider factories are registered externally to avoid import cycles
	// Call RegisterBuiltInProviders() after creating executor

	return executor
}

// RegisterProvider registers a provider factory.
func (e *Executor) RegisterProvider(name string, factory ProviderFactory) {
	e.providerFactories[name] = factory
}

// Execute runs an AI prompt through the configured provider.
func (e *Executor) Execute(ctx context.Context, input string, opts ...Option) (string, error) {
	startTime := time.Now()
	options := ApplyOptions(opts...)

	// Load configuration
	config, err := e.loadConfig()
	if err != nil {
		return "", err
	}

	// Load provider (no fallback - error if invalid)
	provider, err := e.LoadProvider(config)
	if err != nil {
		return "", err
	}
	e.lastUsedProvider = provider

	// Execute with provider
	response, err := provider.Execute(ctx, input, opts...)

	// Include debug info in output if requested
	if options.Debug {
		debugInfo := e.formatDebugInfo(provider.Name(), input, response, err, time.Since(startTime))
		if err != nil {
			// Include debug info even on error
			return debugInfo, err
		}
		// Prepend debug info to response
		return debugInfo + "\n\n" + response, nil
	}

	return response, err
}

// GetLastUsedProvider returns the last provider used (for testing).
func (e *Executor) GetLastUsedProvider() Provider {
	return e.lastUsedProvider
}

// loadConfig loads the EAC configuration from .r2r/eac directory.
// Loads team config and merges with personal overrides if present.
// Personal config can override: api_key, model, provider name, endpoint, git token.
// Both configs are validated against the ai-provider schema.
func (e *Executor) loadConfig() (*Config, error) {
	teamConfigPath := paths.EACConfigFilePath(e.workspaceRoot)
	personalConfigPath := paths.EACConfigPersonalFilePath(e.workspaceRoot)

	// Use merge-based loading with schema validation
	return LoadConfigWithOverrides(e.workspaceRoot, teamConfigPath, personalConfigPath)
}

// LoadProvider loads the configured provider without fallback
// Returns error if provider is unknown or cannot be created.
func (e *Executor) LoadProvider(config *Config) (Provider, error) {
	// Config is required
	if config == nil {
		return nil, ErrAIProviderNotConfigured
	}

	// Check if provider factory exists
	factory, exists := e.providerFactories[config.AI.Provider]
	if !exists {
		return nil, fmt.Errorf("unknown provider: %s\n\nPlease run: eac init --ai-provider <provider>\nSupported providers: claude-api, openai, gemini", config.AI.Provider)
	}

	// Try to create provider
	provider, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w\n\nPlease run: eac init --ai-provider <provider>\nSupported providers: claude-api, openai, gemini", config.AI.Provider, err)
	}

	return provider, nil
}

// formatDebugInfo formats debug information for inclusion in output.
func (e *Executor) formatDebugInfo(provider, input, response string, err error, duration time.Duration) string {
	var status string
	if err != nil {
		status = fmt.Sprintf("failure: %v", err)
	} else {
		status = "success"
	}

	debugInfo := fmt.Sprintf(`DEBUG INFO:
timestamp: %s
provider: %s
duration: %v
status: %s
input_length: %d
response_length: %d`,
		time.Now().Format(time.RFC3339),
		provider,
		duration,
		status,
		len(input),
		len(response),
	)

	return debugInfo
}
