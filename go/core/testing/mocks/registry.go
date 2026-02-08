package mocks

import (
	"sort"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// MockModuleRegistry implements core.ModuleRegistryPort for testing.
type MockModuleRegistry struct {
	modules       map[string]core.ModuleContractPort
	workspaceRoot string
}

// NewMockModuleRegistry creates a new MockModuleRegistry.
func NewMockModuleRegistry() *MockModuleRegistry {
	return &MockModuleRegistry{
		modules:       make(map[string]core.ModuleContractPort),
		workspaceRoot: "/mock/workspace",
	}
}

// WithWorkspaceRoot sets the workspace root path.
func (m *MockModuleRegistry) WithWorkspaceRoot(root string) *MockModuleRegistry {
	m.workspaceRoot = root
	return m
}

// WithModule adds a module to the registry.
func (m *MockModuleRegistry) WithModule(module core.ModuleContractPort) *MockModuleRegistry {
	m.modules[module.GetMoniker()] = module
	return m
}

// AddModule adds a module to the registry (for compatibility with existing tests).
func (m *MockModuleRegistry) AddModule(module core.ModuleContractPort) {
	m.modules[module.GetMoniker()] = module
}

// Get implements core.ModuleRegistryPort.
func (m *MockModuleRegistry) Get(moniker string) (core.ModuleContractPort, bool) {
	mod, ok := m.modules[moniker]
	return mod, ok
}

// Has implements core.ModuleRegistryPort.
func (m *MockModuleRegistry) Has(moniker string) bool {
	_, ok := m.modules[moniker]
	return ok
}

// All implements core.ModuleRegistryPort.
func (m *MockModuleRegistry) All() []core.ModuleContractPort {
	result := make([]core.ModuleContractPort, 0, len(m.modules))
	for _, mod := range m.modules {
		result = append(result, mod)
	}
	return result
}

// AllMonikers implements core.ModuleRegistryPort.
func (m *MockModuleRegistry) AllMonikers() []string {
	result := make([]string, 0, len(m.modules))
	for moniker := range m.modules {
		result = append(result, moniker)
	}
	sort.Strings(result)
	return result
}

// Count implements core.ModuleRegistryPort.
func (m *MockModuleRegistry) Count() int {
	return len(m.modules)
}

// FilterByComponent implements core.ModuleRegistryPort.
func (m *MockModuleRegistry) FilterByComponent(componentType string) []core.ModuleContractPort {
	var result []core.ModuleContractPort
	for _, mod := range m.modules {
		if mod.HasComponent(componentType) {
			result = append(result, mod)
		}
	}
	return result
}

// FindModulesForFile implements core.ModuleRegistryPort.
func (m *MockModuleRegistry) FindModulesForFile(filePath string) []core.ModuleContractPort {
	// Simple implementation: check if file path starts with any component root
	var result []core.ModuleContractPort
	for _, mod := range m.modules {
		for _, root := range mod.GetComponentRoots() {
			if len(filePath) >= len(root) && filePath[:len(root)] == root {
				result = append(result, mod)
				break
			}
		}
	}
	return result
}

// WorkspaceRoot implements core.ModuleRegistryPort.
func (m *MockModuleRegistry) WorkspaceRoot() string {
	return m.workspaceRoot
}

// Interface compliance check
var _ core.ModuleRegistryPort = (*MockModuleRegistry)(nil)

// MockModuleReport implements core.ModuleReportPort for testing.
type MockModuleReport struct {
	registry core.ModuleRegistryPort
	errors   []error
}

// NewMockModuleReport creates a new MockModuleReport.
func NewMockModuleReport() *MockModuleReport {
	return &MockModuleReport{
		registry: NewMockModuleRegistry(),
	}
}

// WithRegistry sets the module registry.
func (m *MockModuleReport) WithRegistry(registry core.ModuleRegistryPort) *MockModuleReport {
	m.registry = registry
	return m
}

// WithErrors sets the loading errors.
func (m *MockModuleReport) WithErrors(errs ...error) *MockModuleReport {
	m.errors = errs
	return m
}

// Registry implements core.ModuleReportPort.
func (m *MockModuleReport) Registry() core.ModuleRegistryPort {
	return m.registry
}

// Errors implements core.ModuleReportPort.
func (m *MockModuleReport) Errors() []error {
	return m.errors
}

// Interface compliance check
var _ core.ModuleReportPort = (*MockModuleReport)(nil)
