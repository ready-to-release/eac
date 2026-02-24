package aiproviders

import (
	"context"
	"fmt"
	"sync"
	"time"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/core/paths"
)

// Executor orchestrates AI provider selection and prompt execution.
// Uses the registry to find provider factories.
type Executor struct {
	workspaceRoot    string
	lastUsedProvider ai.Provider
	registry         *Registry

	cachedConfig     *ai.ProviderConfig
	cachedConfigOnce sync.Once
	cachedConfigErr  error
}

// NewExecutor creates a new executor for the given workspace.
// If registry is nil, the default registry is used.
func NewExecutor(workspaceRoot string, registry *Registry) *Executor {
	if registry == nil {
		registry = DefaultRegistry()
	}
	return &Executor{
		workspaceRoot: workspaceRoot,
		registry:      registry,
	}
}

// Execute runs an AI prompt through the configured provider.
func (e *Executor) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	startTime := time.Now()
	options := ai.ApplyOptions(opts...)

	// Load configuration
	config, err := e.loadConfig()
	if err != nil {
		return "", err
	}

	// Load provider from registry
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
			return debugInfo, err
		}
		return debugInfo + "\n\n" + response, nil
	}

	return response, err
}

// GetLastUsedProvider returns the last provider used (for testing/debugging).
func (e *Executor) GetLastUsedProvider() ai.Provider {
	return e.lastUsedProvider
}

// GetProviderName returns the name of the last used provider.
func (e *Executor) GetProviderName() string {
	if e.lastUsedProvider != nil {
		return e.lastUsedProvider.Name()
	}
	return ""
}

// loadConfig loads the EAC configuration from .eac directory.
func (e *Executor) loadConfig() (*ai.ProviderConfig, error) {
	e.cachedConfigOnce.Do(func() {
		teamConfigPath := paths.EACConfigFilePath(e.workspaceRoot)
		personalConfigPath := paths.EACConfigPersonalFilePath(e.workspaceRoot)
		e.cachedConfig, e.cachedConfigErr = LoadConfigWithOverrides(e.workspaceRoot, teamConfigPath, personalConfigPath)
	})
	return e.cachedConfig, e.cachedConfigErr
}

// LoadProvider loads the configured provider from the registry.
// Returns error if provider is unknown or cannot be created.
func (e *Executor) LoadProvider(config *ai.ProviderConfig) (ai.Provider, error) {
	if config == nil {
		return nil, ai.ErrProviderNotConfigured
	}

	factory, exists := e.registry.Get(config.AI.Provider)
	if !exists {
		return nil, fmt.Errorf("unknown provider: %s\n\nPlease run: eac init --ai-provider <provider>\nSupported providers: %v", config.AI.Provider, e.registry.SupportedProviders())
	}

	provider, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w\n\nPlease run: eac init --ai-provider <provider>", config.AI.Provider, err)
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

	return fmt.Sprintf(`DEBUG INFO:
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
}

// WithModelExecutor wraps a GenerationExecutor to add a model override to every call.
type WithModelExecutor struct {
	inner ai.GenerationExecutor
	model string
}

// NewWithModelExecutor creates an executor wrapper that overrides the model for every call.
func NewWithModelExecutor(inner ai.GenerationExecutor, model string) *WithModelExecutor {
	return &WithModelExecutor{inner: inner, model: model}
}

// Execute delegates to the inner executor, appending a model override option.
func (e *WithModelExecutor) Execute(ctx context.Context, prompt string, opts ...ai.Option) (string, error) {
	if e.model != "" {
		opts = append(opts, ai.WithModel(e.model))
	}
	return e.inner.Execute(ctx, prompt, opts...)
}

// GetProviderName returns the provider name from the inner executor if available.
func (e *WithModelExecutor) GetProviderName() string {
	if p, ok := e.inner.(ai.GenerationExecutorWithProviderInfo); ok {
		return p.GetProviderName()
	}
	return ""
}

// Compile-time interface checks.
var _ ai.GenerationExecutor = (*Executor)(nil)
var _ ai.GenerationExecutorWithProviderInfo = (*Executor)(nil)
var _ ai.GenerationExecutor = (*WithModelExecutor)(nil)
var _ ai.GenerationExecutorWithProviderInfo = (*WithModelExecutor)(nil)
