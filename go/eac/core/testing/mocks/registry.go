package mocks

import (
	"sort"

	"github.com/ready-to-release/eac/contracts/eac-core-interfaces"
)

// MockModuleRegistry implements interfaces.ModuleRegistryPort for testing.
type MockModuleRegistry struct {
	modules       map[string]interfaces.ModuleContractPort
	workspaceRoot string
}

// NewMockModuleRegistry creates a new MockModuleRegistry.
func NewMockModuleRegistry() *MockModuleRegistry {
	return &MockModuleRegistry{
		modules:       make(map[string]interfaces.ModuleContractPort),
		workspaceRoot: "/mock/workspace",
	}
}

// WithWorkspaceRoot sets the workspace root path.
func (m *MockModuleRegistry) WithWorkspaceRoot(root string) *MockModuleRegistry {
	m.workspaceRoot = root
	return m
}

// WithModule adds a module to the registry.
func (m *MockModuleRegistry) WithModule(module interfaces.ModuleContractPort) *MockModuleRegistry {
	m.modules[module.GetMoniker()] = module
	return m
}

// AddModule adds a module to the registry (for compatibility with existing tests).
func (m *MockModuleRegistry) AddModule(module interfaces.ModuleContractPort) {
	m.modules[module.GetMoniker()] = module
}

// Get implements interfaces.ModuleRegistryPort.
func (m *MockModuleRegistry) Get(moniker string) (interfaces.ModuleContractPort, bool) {
	mod, ok := m.modules[moniker]
	return mod, ok
}

// Has implements interfaces.ModuleRegistryPort.
func (m *MockModuleRegistry) Has(moniker string) bool {
	_, ok := m.modules[moniker]
	return ok
}

// All implements interfaces.ModuleRegistryPort.
func (m *MockModuleRegistry) All() []interfaces.ModuleContractPort {
	result := make([]interfaces.ModuleContractPort, 0, len(m.modules))
	for _, mod := range m.modules {
		result = append(result, mod)
	}
	return result
}

// AllMonikers implements interfaces.ModuleRegistryPort.
func (m *MockModuleRegistry) AllMonikers() []string {
	result := make([]string, 0, len(m.modules))
	for moniker := range m.modules {
		result = append(result, moniker)
	}
	sort.Strings(result)
	return result
}

// Count implements interfaces.ModuleRegistryPort.
func (m *MockModuleRegistry) Count() int {
	return len(m.modules)
}

// FilterByComponent implements interfaces.ModuleRegistryPort.
func (m *MockModuleRegistry) FilterByComponent(componentType string) []interfaces.ModuleContractPort {
	var result []interfaces.ModuleContractPort
	for _, mod := range m.modules {
		if mod.HasComponent(componentType) {
			result = append(result, mod)
		}
	}
	return result
}

// FindModulesForFile implements interfaces.ModuleRegistryPort.
func (m *MockModuleRegistry) FindModulesForFile(filePath string) []interfaces.ModuleContractPort {
	// Simple implementation: check if file path starts with any component root
	var result []interfaces.ModuleContractPort
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

// WorkspaceRoot implements interfaces.ModuleRegistryPort.
func (m *MockModuleRegistry) WorkspaceRoot() string {
	return m.workspaceRoot
}

// Interface compliance check
var _ interfaces.ModuleRegistryPort = (*MockModuleRegistry)(nil)

// MockModuleReport implements interfaces.ModuleReportPort for testing.
type MockModuleReport struct {
	registry interfaces.ModuleRegistryPort
	errors   []error
}

// NewMockModuleReport creates a new MockModuleReport.
func NewMockModuleReport() *MockModuleReport {
	return &MockModuleReport{
		registry: NewMockModuleRegistry(),
	}
}

// WithRegistry sets the module registry.
func (m *MockModuleReport) WithRegistry(registry interfaces.ModuleRegistryPort) *MockModuleReport {
	m.registry = registry
	return m
}

// WithErrors sets the loading errors.
func (m *MockModuleReport) WithErrors(errs ...error) *MockModuleReport {
	m.errors = errs
	return m
}

// Registry implements interfaces.ModuleReportPort.
func (m *MockModuleReport) Registry() interfaces.ModuleRegistryPort {
	return m.registry
}

// Errors implements interfaces.ModuleReportPort.
func (m *MockModuleReport) Errors() []error {
	return m.errors
}

// Interface compliance check
var _ interfaces.ModuleReportPort = (*MockModuleReport)(nil)
