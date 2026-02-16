package godog

import (
	"sort"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	contractsreports "github.com/ready-to-release/eac/go/core/domain/reports"
)

// ============================================================================
// Adapters for concrete types to port interfaces
//
// These adapt concrete types (ModuleContractReport, modules.Registry,
// modules.ModuleContract) to the port interfaces defined in contracts/core.
// ============================================================================

// moduleReportAdapter wraps *contractsreports.ModuleContractReport to implement core.ModuleReportPort.
type moduleReportAdapter struct {
	report *contractsreports.ModuleContractReport
}

func newModuleReportAdapter(report *contractsreports.ModuleContractReport) *moduleReportAdapter {
	return &moduleReportAdapter{report: report}
}

// Registry implements core.ModuleReportPort.
func (a *moduleReportAdapter) Registry() core.ModuleRegistryPort {
	if a.report.Registry == nil {
		return nil
	}
	return newModuleRegistryAdapter(a.report.Registry)
}

// Errors implements core.ModuleReportPort.
func (a *moduleReportAdapter) Errors() []error {
	return nil // ModuleContractReport doesn't track errors
}

var _ core.ModuleReportPort = (*moduleReportAdapter)(nil)

// moduleRegistryAdapter wraps *modules.Registry to implement core.ModuleRegistryPort.
type moduleRegistryAdapter struct {
	registry *modules.Registry
}

func newModuleRegistryAdapter(registry *modules.Registry) *moduleRegistryAdapter {
	return &moduleRegistryAdapter{registry: registry}
}

// Get implements core.ModuleRegistryPort.
func (a *moduleRegistryAdapter) Get(moniker string) (core.ModuleContractPort, bool) {
	mod, ok := a.registry.Get(moniker)
	if !ok || mod == nil {
		return nil, false
	}
	return newModuleContractAdapter(mod), true
}

// Has implements core.ModuleRegistryPort.
func (a *moduleRegistryAdapter) Has(moniker string) bool {
	_, ok := a.registry.Get(moniker)
	return ok
}

// All implements core.ModuleRegistryPort.
func (a *moduleRegistryAdapter) All() []core.ModuleContractPort {
	mods := a.registry.All()
	result := make([]core.ModuleContractPort, len(mods))
	for i, mod := range mods {
		result[i] = newModuleContractAdapter(mod)
	}
	return result
}

// AllMonikers implements core.ModuleRegistryPort.
func (a *moduleRegistryAdapter) AllMonikers() []string {
	mods := a.registry.All()
	result := make([]string, len(mods))
	for i, mod := range mods {
		result[i] = mod.Moniker
	}
	sort.Strings(result)
	return result
}

// Count implements core.ModuleRegistryPort.
func (a *moduleRegistryAdapter) Count() int {
	return len(a.registry.All())
}

// FilterByComponent implements core.ModuleRegistryPort.
func (a *moduleRegistryAdapter) FilterByComponent(componentType string) []core.ModuleContractPort {
	mods := a.registry.All()
	var result []core.ModuleContractPort
	for _, mod := range mods {
		if mod.HasComponent(componentType) {
			result = append(result, newModuleContractAdapter(mod))
		}
	}
	return result
}

// FindModulesForFile implements core.ModuleRegistryPort.
func (a *moduleRegistryAdapter) FindModulesForFile(filePath string) []core.ModuleContractPort {
	mods := a.registry.FindModulesForFile(filePath)
	result := make([]core.ModuleContractPort, len(mods))
	for i, mod := range mods {
		result[i] = newModuleContractAdapter(mod)
	}
	return result
}

// WorkspaceRoot implements core.ModuleRegistryPort.
func (a *moduleRegistryAdapter) WorkspaceRoot() string {
	return a.registry.WorkspaceRoot()
}

var _ core.ModuleRegistryPort = (*moduleRegistryAdapter)(nil)

// moduleContractAdapter wraps *modules.ModuleContract to implement core.ModuleContractPort.
type moduleContractAdapter struct {
	module *modules.ModuleContract
}

func newModuleContractAdapter(module *modules.ModuleContract) *moduleContractAdapter {
	return &moduleContractAdapter{module: module}
}

// GetMoniker implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetMoniker() string {
	return a.module.Moniker
}

// GetName implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetName() string {
	return a.module.GetName()
}

// GetDescription implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetDescription() string {
	return a.module.GetDescription()
}

// GetGroup implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetGroup() string {
	return a.module.GetGroup()
}

// HasComponent implements core.ModuleContractPort.
func (a *moduleContractAdapter) HasComponent(componentType string) bool {
	return a.module.HasComponent(componentType)
}

// GetComponentRoot implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentRoot(componentType string) string {
	return a.module.GetComponentRoot(componentType)
}

// GetComponentRoots implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentRoots() map[string]string {
	return a.module.GetComponentRoots()
}

// GetComponentTypesDisplay implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentTypesDisplay() string {
	return a.module.GetComponentTypesDisplay()
}

// GetComponentGroup implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentGroup(componentName string) string {
	return a.module.GetComponentGroup(componentName)
}

// GetComponentAmp implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentAmp(componentName, operation string) float64 {
	if comp, ok := a.module.Components[componentName]; ok && comp != nil {
		return comp.Amp.GetAmp(operation)
	}
	return 1.0
}

// GetDependsOn implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetDependsOn() []string {
	return a.module.DependsOn
}

// GetVersioningScheme implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetVersioningScheme() string {
	if a.module.Versioning != nil {
		return a.module.Versioning.Scheme
	}
	return ""
}

// GetReleaseType implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetReleaseType() string {
	if a.module.Versioning != nil {
		return a.module.Versioning.ReleaseType
	}
	return ""
}

// GetChangelog implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetChangelog() string {
	return a.module.GetChangelog()
}

// HasVersioning implements core.ModuleContractPort.
func (a *moduleContractAdapter) HasVersioning() bool {
	return a.module.Versioning != nil && a.module.Versioning.Scheme != ""
}

// GetMetadata implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetMetadata() map[string]interface{} {
	if a.module.Metadata == nil {
		return nil
	}
	result := make(map[string]interface{}, len(a.module.Metadata))
	for k, v := range a.module.Metadata {
		result[k] = v
	}
	return result
}

// GetContentHash implements core.ModuleContractPort.
func (a *moduleContractAdapter) GetContentHash() (string, error) {
	return a.module.GetContentHash()
}

var _ core.ModuleContractPort = (*moduleContractAdapter)(nil)
