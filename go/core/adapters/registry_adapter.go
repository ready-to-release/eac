package adapters

import (
	"github.com/ready-to-release/eac/go/core/domain/modules"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Compile-time interface check.
var _ core.ModuleRegistryPort = (*ModuleRegistryAdapter)(nil)

// ModuleRegistryAdapter wraps a *modules.Registry to implement core.ModuleRegistryPort.
type ModuleRegistryAdapter struct {
	registry *modules.Registry
}

// NewModuleRegistryAdapter creates a new adapter wrapping a module registry.
func NewModuleRegistryAdapter(r *modules.Registry) *ModuleRegistryAdapter {
	return &ModuleRegistryAdapter{registry: r}
}

// Unwrap returns the underlying concrete registry.
// Use this when you need access to methods not exposed through the port interface.
func (a *ModuleRegistryAdapter) Unwrap() *modules.Registry {
	return a.registry
}

// Get retrieves a module contract by moniker.
func (a *ModuleRegistryAdapter) Get(moniker string) (core.ModuleContractPort, bool) {
	m, ok := a.registry.Get(moniker)
	if !ok {
		return nil, false
	}
	return AdaptModule(m), true
}

// Has checks if a module exists in the registry.
func (a *ModuleRegistryAdapter) Has(moniker string) bool {
	return a.registry.Has(moniker)
}

// All returns all module domain.
func (a *ModuleRegistryAdapter) All() []core.ModuleContractPort {
	return AdaptModules(a.registry.All())
}

// AllMonikers returns all module monikers sorted alphabetically.
func (a *ModuleRegistryAdapter) AllMonikers() []string {
	return a.registry.AllMonikers()
}

// Count returns the number of modules.
func (a *ModuleRegistryAdapter) Count() int {
	return a.registry.Count()
}

// FilterByComponent returns modules with the specified component type.
func (a *ModuleRegistryAdapter) FilterByComponent(componentType string) []core.ModuleContractPort {
	return AdaptModules(a.registry.FilterByComponent(componentType))
}

// FindModulesForFile returns modules that own the given file path.
func (a *ModuleRegistryAdapter) FindModulesForFile(filePath string) []core.ModuleContractPort {
	return AdaptModules(a.registry.FindModulesForFile(filePath))
}

// WorkspaceRoot returns the workspace root path.
func (a *ModuleRegistryAdapter) WorkspaceRoot() string {
	return a.registry.WorkspaceRoot()
}

// AdaptRegistry is a convenience function to wrap a module registry.
func AdaptRegistry(r *modules.Registry) core.ModuleRegistryPort {
	if r == nil {
		return nil
	}
	return NewModuleRegistryAdapter(r)
}
