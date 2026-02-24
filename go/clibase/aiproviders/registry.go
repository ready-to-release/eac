// Package aiproviders provides a registry-based AI provider dispatch system.
// Each AI provider (claude, openai, gemini) has a dedicated adapter that self-registers
// via init() functions, following the same pattern as testrunners/.
//
// Port types (Provider, ProviderConfig, Option, etc.) are defined in
// contracts/ai-provider/0.1.0 and re-exported here for convenience.
package aiproviders

import (
	"sync"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
)

// Type aliases -- new code should import directly from contracts/ai-provider/0.1.0.
type (
	Provider         = ai.Provider
	ProviderFactory  = ai.ProviderFactory
	ProviderConfig   = ai.ProviderConfig
	ConnectionConfig = ai.ConnectionConfig
	GitConfig        = ai.GitConfig
	Option           = ai.Option
	ExecuteOptions   = ai.ExecuteOptions
)

// Registry holds registered AI provider factories. Use NewRegistry() to create an
// isolated instance for testing, or use the package-level functions which
// delegate to the default registry.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]ai.ProviderFactory
}

// NewRegistry creates a new, empty Registry with initialized maps.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]ai.ProviderFactory)}
}

// defaultRegistry is the package-level registry instance.
var (
	defaultRegistry     *Registry
	defaultRegistryOnce sync.Once
)

// DefaultRegistry returns the package-level default registry instance.
func DefaultRegistry() *Registry {
	defaultRegistryOnce.Do(func() {
		defaultRegistry = NewRegistry()
	})
	return defaultRegistry
}

// --- Registry methods ---

// Register registers a named provider factory.
// Called from adapter init() functions.
func (r *Registry) Register(name string, factory ai.ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Get returns the factory for a named provider.
func (r *Registry) Get(name string) (ai.ProviderFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[name]
	return f, ok
}

// SupportedProviders returns the names of all registered providers.
func (r *Registry) SupportedProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// ResetForTesting clears all registrations. Use only in tests.
func (r *Registry) ResetForTesting() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories = make(map[string]ai.ProviderFactory)
}

// --- Package-level functions (delegate to defaultRegistry) ---

// Register registers a named provider factory with the default registry.
// Call this from init() in your adapter implementation file.
func Register(name string, factory ai.ProviderFactory) {
	DefaultRegistry().Register(name, factory)
}

// Get returns the factory for a named provider from the default registry.
func Get(name string) (ai.ProviderFactory, bool) {
	return DefaultRegistry().Get(name)
}

// SupportedProviders returns the names of all registered providers.
func SupportedProviders() []string {
	return DefaultRegistry().SupportedProviders()
}

// ResetForTesting clears all registrations in the default registry. Use only in tests.
func ResetForTesting() {
	DefaultRegistry().ResetForTesting()
}
