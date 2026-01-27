// Package tool provides a lint bridge that integrates the tool system.
package tool

import (
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// LintBridge provides a unified interface for resolving lint handlers.
// All handlers are resolved from tool-config.yml definitions.
type LintBridge struct {
	mu sync.RWMutex

	// Provider-to-handler mapping (from lint-providers.yml style)
	providerHandlers map[string]string

	// Tool system integration
	registry Registry
	resolver *DefaultResolver
	executor Executor
}

// NewLintBridge creates a new lint bridge.
func NewLintBridge() *LintBridge {
	return &LintBridge{
		providerHandlers: map[string]string{
			// Default provider-to-handler mapping
			// Provider names (from lint-providers.yml) map to handler names
			"golangci-lint":     "golangci-lint-system",
			"markdownlint-cli2": "markdownlint-system",
			"eslint":            "eslint-system",
		},
	}
}

// SetToolSystem configures the tool system for tool-config.yml defined tools.
func (b *LintBridge) SetToolSystem(registry Registry, resolver *DefaultResolver, executor Executor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry = registry
	b.resolver = resolver
	b.executor = executor
}

// SetProviderMapping sets a provider-to-handler mapping.
func (b *LintBridge) SetProviderMapping(provider, handler string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.providerHandlers[provider] = handler
}

// GetHandler returns a lint handler by name from the tool registry.
func (b *LintBridge) GetHandler(name string) LintHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.registry != nil && b.executor != nil {
		if tool, ok := b.registry.Get(name); ok {
			return NewLintHandlerAdapter(tool, b.executor)
		}
	}

	return nil
}

// GetAllHandlers returns all available handlers from the tool registry.
func (b *LintBridge) GetAllHandlers() map[string]LintHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]LintHandler)

	if b.registry != nil && b.executor != nil {
		for name, tool := range b.registry.GetAll() {
			result[name] = NewLintHandlerAdapter(tool, b.executor)
		}
	}

	return result
}

// GetHandlerForProvider returns the handler for a lint provider.
// Provider names map to handler names via the providerHandlers mapping.
func (b *LintBridge) GetHandlerForProvider(providerName string) LintHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Look up handler name from provider mapping
	handlerName, ok := b.providerHandlers[providerName]
	if !ok {
		// Direct mapping: provider name == handler name
		handlerName = providerName
	}

	// Check tool registry
	if b.registry != nil && b.executor != nil {
		if tool, ok := b.registry.Get(handlerName); ok {
			return NewLintHandlerAdapter(tool, b.executor)
		}
	}

	return nil
}

// GetProvidersForModule returns all applicable lint providers for a module's components.
func (b *LintBridge) GetProvidersForModule(module interface{ GetEnabledComponents() []string }, lintProviders *config.LintProvidersConfig, componentTypes *config.ComponentTypesConfig) []string {
	if module == nil || lintProviders == nil {
		return nil
	}

	providerSet := make(map[string]bool)
	for _, compName := range module.GetEnabledComponents() {
		compType := compName

		// Find providers that apply to this component type
		for name, provider := range lintProviders.LintProviders {
			for _, applies := range provider.AppliesTo {
				if applies == compType {
					providerSet[name] = true
					break
				}
			}
		}
	}

	result := make([]string, 0, len(providerSet))
	for name := range providerSet {
		result = append(result, name)
	}
	return result
}

// HasHandler checks if a handler exists by name.
func (b *LintBridge) HasHandler(name string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.registry != nil {
		if _, ok := b.registry.Get(name); ok {
			return true
		}
	}

	return false
}

// Global lint bridge instance.
var (
	globalLintBridge     *LintBridge
	globalLintBridgeOnce sync.Once
)

// GlobalLintBridge returns the global lint bridge instance.
func GlobalLintBridge() *LintBridge {
	globalLintBridgeOnce.Do(func() {
		globalLintBridge = NewLintBridge()
	})
	return globalLintBridge
}
