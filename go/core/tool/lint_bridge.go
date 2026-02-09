// Package tool provides a lint bridge that integrates the tool system.
package tool

import (
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
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
		// All mappings come from component-tools in tool-config.yml via the resolver.
		// No hardcoded fallbacks - tool names should not be embedded in code.
		providerHandlers: map[string]string{},
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

	handlerName := b.resolveHandlerName(providerName)

	// Check tool registry
	if b.registry != nil && b.executor != nil {
		if tool, ok := b.registry.Get(handlerName); ok {
			return NewLintHandlerAdapter(tool, b.executor)
		}
	}

	return nil
}

// GetProviderWeight returns the scheduling weight for a lint provider.
// Weight is derived from the tool's Resources configuration.
// Returns 1 if no tool is configured or resources not specified.
func (b *LintBridge) GetProviderWeight(providerName string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlerName := b.resolveHandlerName(providerName)

	// Check tool registry for resources
	if b.registry != nil {
		if tool, ok := b.registry.Get(handlerName); ok && tool.Resources != nil {
			return tool.Resources.Weight()
		}
	}

	return 1
}

// resolveHandlerName maps a provider name to its handler name.
// Must be called with mu held.
func (b *LintBridge) resolveHandlerName(providerName string) string {
	if handlerName, ok := b.providerHandlers[providerName]; ok {
		return handlerName
	}
	return providerName
}

// GetProvidersForModule returns all applicable lint providers for a module's components.
func (b *LintBridge) GetProvidersForModule(module interface{ GetEnabledComponents() []string }, lintProviders *config.LintProvidersConfig, componentTypes *config.ComponentKindsConfig) []string {
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

// ResolveTool returns the tool definition for a component type and operation.
// Returns nil if no tool is configured or resolver is not available.
func (b *LintBridge) ResolveTool(componentType string, operation core.ActionType) *ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.resolver == nil {
		return nil
	}

	t, err := b.resolver.Resolve(componentType, operation)
	if err != nil {
		return nil
	}

	return t
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
