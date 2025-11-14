// File: src/core/ai/executor.go
// Intent: Orchestrate AI provider invocation with configuration loading, fallback behavior, and logging
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Single Execute() method with clear signature
//   - Explicit fallback chain: configured provider → claude-cli
//   - Provider loading separated into loadProvider() function
//   - Functional options pattern for flexibility
//
// Easy to change:
//   - Provider factories registered in map (easy to add new providers)
//   - Config loading delegated to LoadConfig()
//   - Fallback logic isolated in loadProvider()
//   - Logging delegated to separate logger
//
// Hard to break:
//   - Always falls back to claude-cli (never fails due to missing config)
//   - Provider validation before use
//   - Context passed through for cancellation
//   - Comprehensive tests cover all fallback scenarios

package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Executor orchestrates AI provider execution
type Executor struct {
	workspaceRoot     string
	lastUsedProvider  Provider
	providerFactories map[string]ProviderFactory
	fallbackFactory   ProviderFactory // Factory for fallback provider
	logger            Logger
}

// ProviderFactory creates a provider from configuration
type ProviderFactory func(config *Config) (Provider, error)

// Logger interface for execution logging
type Logger interface {
	LogExecution(ctx context.Context, entry *LogEntry)
}

// NewExecutor creates a new executor for the given workspace
func NewExecutor(workspaceRoot string) *Executor {
	executor := &Executor{
		workspaceRoot:     workspaceRoot,
		providerFactories: make(map[string]ProviderFactory),
		logger:            newNoOpLogger(), // Default no-op logger
	}

	// Provider factories are registered externally to avoid import cycles
	// Call RegisterBuiltInProviders() after creating executor

	return executor
}

// RegisterProvider registers a provider factory
func (e *Executor) RegisterProvider(name string, factory ProviderFactory) {
	e.providerFactories[name] = factory
}

// SetFallbackProvider sets the fallback provider factory
func (e *Executor) SetFallbackProvider(factory ProviderFactory) {
	e.fallbackFactory = factory
}

// SetLogger sets the execution logger
func (e *Executor) SetLogger(logger Logger) {
	e.logger = logger
}

// Execute runs an AI prompt through the configured provider
func (e *Executor) Execute(ctx context.Context, input string, opts ...Option) (string, error) {
	// Load configuration
	config, err := e.loadConfig()
	var configErr error
	if err != nil {
		configErr = err
		// Config error will trigger fallback
	}

	// Load provider (with fallback to claude-cli)
	provider, didFallback := e.LoadProvider(config)
	e.lastUsedProvider = provider

	// Log warning if we fell back due to config error
	if configErr != nil && didFallback {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config (%v), using claude-cli fallback\n", configErr)
	}

	// Execute with provider
	response, err := provider.Execute(ctx, input, opts...)

	// Log execution
	logEntry := &LogEntry{
		Provider:  provider.Name(),
		Input:     input,
		Response:  response,
		Success:   err == nil,
		Error:     err,
		DidFallback: didFallback,
	}
	e.logger.LogExecution(ctx, logEntry)

	return response, err
}

// GetLastUsedProvider returns the last provider used (for testing)
func (e *Executor) GetLastUsedProvider() Provider {
	return e.lastUsedProvider
}

// loadConfig loads the agent configuration
func (e *Executor) loadConfig() (*Config, error) {
	configPath := filepath.Join(e.workspaceRoot, ".r2r", "agent-config.yml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found")
	}

	return LoadConfig(configPath)
}

// LoadProvider loads the configured provider or falls back
// Exported for testing
func (e *Executor) LoadProvider(config *Config) (Provider, bool) {
	// If no config, fall back
	if config == nil {
		return e.createFallback(), true
	}

	// Try to create configured provider
	factory, exists := e.providerFactories[config.ProviderName]
	if !exists {
		// Unknown provider, fall back
		fmt.Fprintf(os.Stderr, "Warning: Unknown provider '%s', using fallback\n", config.ProviderName)
		return e.createFallback(), true
	}

	// Try to create provider
	provider, err := factory(config)
	if err != nil {
		// Provider creation failed (e.g., missing API key), fall back
		fmt.Fprintf(os.Stderr, "Warning: Failed to create %s provider (%v), using fallback\n", config.ProviderName, err)
		return e.createFallback(), true
	}

	// Success - use configured provider
	return provider, false
}

// createFallback creates the fallback provider
func (e *Executor) createFallback() Provider {
	if e.fallbackFactory != nil {
		provider, err := e.fallbackFactory(nil)
		if err == nil {
			return provider
		}
	}
	// If no fallback factory is set, panic (programming error)
	panic("executor: no fallback provider configured")
}

// noOpLogger is a logger that does nothing
type noOpLogger struct{}

func newNoOpLogger() Logger {
	return &noOpLogger{}
}

func (l *noOpLogger) LogExecution(ctx context.Context, entry *LogEntry) {
	// No-op
}

// LogEntry represents a single AI execution log
type LogEntry struct {
	Provider    string
	Input       string
	Response    string
	Success     bool
	Error       error
	DidFallback bool
}
