package adapters

import (
	"github.com/ready-to-release/eac/go/core/domain/modules"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Compile-time interface check.
var _ core.ModuleContractPort = (*ModuleContractAdapter)(nil)

// ModuleContractAdapter wraps a *modules.ModuleContract to implement core.ModuleContractPort.
type ModuleContractAdapter struct {
	module *modules.ModuleContract
}

// NewModuleContractAdapter creates a new adapter wrapping a module contract.
func NewModuleContractAdapter(m *modules.ModuleContract) *ModuleContractAdapter {
	return &ModuleContractAdapter{module: m}
}

// Unwrap returns the underlying concrete module contract.
// Use this when you need access to methods not exposed through the port interface.
func (a *ModuleContractAdapter) Unwrap() *modules.ModuleContract {
	return a.module
}

// GetMoniker returns the module's unique identifier.
func (a *ModuleContractAdapter) GetMoniker() string {
	return a.module.GetMoniker()
}

// GetName returns the module's display name.
func (a *ModuleContractAdapter) GetName() string {
	return a.module.GetName()
}

// GetDescription returns the module's description.
func (a *ModuleContractAdapter) GetDescription() string {
	return a.module.GetDescription()
}

// GetGroup returns the module's group name.
func (a *ModuleContractAdapter) GetGroup() string {
	return a.module.GetGroup()
}

// HasComponent returns true if the module has the specified component type.
func (a *ModuleContractAdapter) HasComponent(componentType string) bool {
	return a.module.HasComponent(componentType)
}

// GetComponentRoot returns the root directory for a component type.
func (a *ModuleContractAdapter) GetComponentRoot(componentType string) string {
	return a.module.GetComponentRoot(componentType)
}

// GetComponentRoots returns all component roots as a map.
func (a *ModuleContractAdapter) GetComponentRoots() map[string]string {
	return a.module.GetComponentRoots()
}

// GetComponentTypesDisplay returns a comma-separated list of component types.
func (a *ModuleContractAdapter) GetComponentTypesDisplay() string {
	return a.module.GetComponentTypesDisplay()
}

// GetComponentGroup returns the component group for a named component.
func (a *ModuleContractAdapter) GetComponentGroup(componentName string) string {
	return a.module.GetComponentGroup(componentName)
}

// GetComponentAmp returns the resource amplifier for a component's operation.
func (a *ModuleContractAdapter) GetComponentAmp(componentName, operation string) float64 {
	if a.module.Components == nil {
		return 1.0
	}
	comp, ok := a.module.Components[componentName]
	if !ok || comp == nil {
		return 1.0
	}
	return comp.GetAmpForOperation(operation)
}

// GetDependsOn returns the list of module dependencies.
func (a *ModuleContractAdapter) GetDependsOn() []string {
	return a.module.DependsOn
}

// GetVersioningScheme returns the versioning scheme (SemVer, CalVer, Implicit).
func (a *ModuleContractAdapter) GetVersioningScheme() string {
	if a.module.Versioning == nil {
		return ""
	}
	return a.module.Versioning.Scheme
}

// GetReleaseType returns the release type (published, internal, bundle, none).
func (a *ModuleContractAdapter) GetReleaseType() string {
	if a.module.Versioning == nil {
		return ""
	}
	return a.module.Versioning.ReleaseType
}

// GetChangelog returns the path to the changelog file.
func (a *ModuleContractAdapter) GetChangelog() string {
	return a.module.GetChangelog()
}

// HasVersioning returns true if versioning is configured.
func (a *ModuleContractAdapter) HasVersioning() bool {
	return a.module.Versioning != nil
}

// GetMetadata returns the module's metadata as a map.
func (a *ModuleContractAdapter) GetMetadata() map[string]interface{} {
	if a.module.Metadata == nil {
		return nil
	}
	// Convert map[string]string to map[string]interface{}
	result := make(map[string]interface{}, len(a.module.Metadata))
	for k, v := range a.module.Metadata {
		result[k] = v
	}
	return result
}

// GetContentHash returns a SHA256 hash of the module's owned files.
// Delegates to the underlying ModuleContract.
func (a *ModuleContractAdapter) GetContentHash() (string, error) {
	return a.module.GetContentHash()
}

// AdaptModule is a convenience function to wrap a module contract.
func AdaptModule(m *modules.ModuleContract) core.ModuleContractPort {
	if m == nil {
		return nil
	}
	return NewModuleContractAdapter(m)
}

// AdaptModules wraps a slice of module contracts.
func AdaptModules(mods []*modules.ModuleContract) []core.ModuleContractPort {
	return AdaptSlice(mods, AdaptModule)
}

// UnwrapModule extracts the concrete ModuleContract from a port interface.
// Returns nil if the port is nil or not backed by a ModuleContractAdapter.
// Use this in native handlers that need access to methods not in the port interface.
func UnwrapModule(port core.ModuleContractPort) *modules.ModuleContract {
	if port == nil {
		return nil
	}
	if adapter, ok := port.(*ModuleContractAdapter); ok {
		return adapter.Unwrap()
	}
	return nil
}
