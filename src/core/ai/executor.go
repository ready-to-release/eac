// File: src/core/ai/executor.go
// Intent: Orchestrate AI provider invocation with configuration loading and error handling
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Single Execute() method with clear signature
//   - No fallback behavior - fails fast with clear error messages
//   - Provider loading separated into LoadProvider() function
//   - Functional options pattern for flexibility
//   - Debug logging included in output when requested, never written to files
//
// Easy to change:
//   - Provider factories registered in map (easy to add new providers)
//   - Config loading delegated to LoadConfig()
//   - Error handling provides clear recovery instructions
//   - Debug output controlled by option, not hardcoded
//
// Hard to break:
//   - Always validates configuration before use
//   - Provider validation before execution
//   - Context passed through for cancellation
//   - Comprehensive tests cover all error scenarios
//   - No fallback means clear failure modes

package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Executor orchestrates AI provider execution
type Executor struct {
	workspaceRoot     string
	lastUsedProvider  Provider
	providerFactories map[string]ProviderFactory
}

// ProviderFactory creates a provider from configuration
type ProviderFactory func(config *Config) (Provider, error)

// NewExecutor creates a new executor for the given workspace
func NewExecutor(workspaceRoot string) *Executor {
	executor := &Executor{
		workspaceRoot:     workspaceRoot,
		providerFactories: make(map[string]ProviderFactory),
	}

	// Provider factories are registered externally to avoid import cycles
	// Call RegisterBuiltInProviders() after creating executor

	return executor
}

// RegisterProvider registers a provider factory
func (e *Executor) RegisterProvider(name string, factory ProviderFactory) {
	e.providerFactories[name] = factory
}

// Execute runs an AI prompt through the configured provider
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

// GetLastUsedProvider returns the last provider used (for testing)
func (e *Executor) GetLastUsedProvider() Provider {
	return e.lastUsedProvider
}

// loadConfig loads the agent configuration from .r2r directory
func (e *Executor) loadConfig() (*Config, error) {
	configPath := filepath.Join(e.workspaceRoot, ".r2r", "agent-config.yml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("agent-config.yml not found in .r2r directory\n\nPlease run: r2r agent init --ai <provider>\nSupported providers: claude-cli, claude-api, openai, gemini")
	}

	return LoadConfig(configPath)
}

// LoadProvider loads the configured provider without fallback
// Returns error if provider is unknown or cannot be created
func (e *Executor) LoadProvider(config *Config) (Provider, error) {
	// Config is required
	if config == nil {
		return nil, fmt.Errorf("configuration is required\n\nPlease run: r2r agent init --ai <provider>\nSupported providers: claude-cli, claude-api, openai, gemini")
	}

	// Check if provider factory exists
	factory, exists := e.providerFactories[config.ProviderName]
	if !exists {
		return nil, fmt.Errorf("unknown provider: %s\n\nPlease run: r2r agent init --ai <provider>\nSupported providers: claude-cli, claude-api, openai, gemini", config.ProviderName)
	}

	// Try to create provider
	provider, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w\n\nPlease run: r2r agent init --ai <provider>\nSupported providers: claude-cli, claude-api, openai, gemini", config.ProviderName, err)
	}

	return provider, nil
}

// formatDebugInfo formats debug information for inclusion in output
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
