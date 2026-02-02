package mocks

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
)

// MockModule implements interfaces.ModuleContractPort for testing.
type MockModule struct {
	moniker          string
	name             string
	description      string
	components       map[string]string // componentType -> root
	componentAmps    map[string]map[string]float64 // componentType -> operation -> amp
	dependsOn        []string
	versioningScheme string
	releaseType      string
	changelog        string
	metadata         map[string]interface{}
}

// NewMockModule creates a new MockModule with the given moniker.
func NewMockModule(moniker string) *MockModule {
	return &MockModule{
		moniker:    moniker,
		name:       moniker,
		components: make(map[string]string),
		metadata:   make(map[string]interface{}),
	}
}

// WithName sets the module name.
func (m *MockModule) WithName(name string) *MockModule {
	m.name = name
	return m
}

// WithDescription sets the module description.
func (m *MockModule) WithDescription(desc string) *MockModule {
	m.description = desc
	return m
}

// WithComponent adds a component to the module.
func (m *MockModule) WithComponent(componentType, root string) *MockModule {
	m.components[componentType] = root
	return m
}

// WithGoComponent adds a Go component to the module.
func (m *MockModule) WithGoComponent(root string) *MockModule {
	return m.WithComponent("go", root)
}

// WithDocsComponent adds a docs component to the module.
func (m *MockModule) WithDocsComponent(root string) *MockModule {
	return m.WithComponent("docs", root)
}

// WithSpecsComponent adds a specs component to the module.
func (m *MockModule) WithSpecsComponent(root string) *MockModule {
	return m.WithComponent("specs", root)
}

// WithDependsOn adds dependencies to the module.
func (m *MockModule) WithDependsOn(deps ...string) *MockModule {
	m.dependsOn = append(m.dependsOn, deps...)
	return m
}

// WithVersioning adds versioning configuration.
func (m *MockModule) WithVersioning(scheme string) *MockModule {
	m.versioningScheme = scheme
	return m
}

// WithCalVer adds CalVer versioning.
func (m *MockModule) WithCalVer() *MockModule {
	return m.WithVersioning("CalVer")
}

// WithSemVer adds SemVer versioning.
func (m *MockModule) WithSemVer() *MockModule {
	return m.WithVersioning("SemVer")
}

// WithReleaseType sets the release type.
func (m *MockModule) WithReleaseType(releaseType string) *MockModule {
	m.releaseType = releaseType
	return m
}

// WithChangelog sets the changelog path.
func (m *MockModule) WithChangelog(path string) *MockModule {
	m.changelog = path
	return m
}

// WithMetadata sets metadata key-value pair.
func (m *MockModule) WithMetadata(key string, value interface{}) *MockModule {
	m.metadata[key] = value
	return m
}

// WithComponentAmp sets the amplifier for a component operation.
func (m *MockModule) WithComponentAmp(componentType, operation string, amp float64) *MockModule {
	if m.componentAmps == nil {
		m.componentAmps = make(map[string]map[string]float64)
	}
	if m.componentAmps[componentType] == nil {
		m.componentAmps[componentType] = make(map[string]float64)
	}
	m.componentAmps[componentType][operation] = amp
	return m
}

// GetMoniker implements interfaces.ModuleContractPort.
func (m *MockModule) GetMoniker() string {
	return m.moniker
}

// GetName implements interfaces.ModuleContractPort.
func (m *MockModule) GetName() string {
	return m.name
}

// GetDescription implements interfaces.ModuleContractPort.
func (m *MockModule) GetDescription() string {
	return m.description
}

// HasComponent implements interfaces.ModuleContractPort.
func (m *MockModule) HasComponent(componentType string) bool {
	_, ok := m.components[componentType]
	return ok
}

// GetComponentRoot implements interfaces.ModuleContractPort.
func (m *MockModule) GetComponentRoot(componentType string) string {
	return m.components[componentType]
}

// GetComponentRoots implements interfaces.ModuleContractPort.
func (m *MockModule) GetComponentRoots() map[string]string {
	result := make(map[string]string)
	for k, v := range m.components {
		result[k] = v
	}
	return result
}

// GetComponentTypesDisplay implements interfaces.ModuleContractPort.
func (m *MockModule) GetComponentTypesDisplay() string {
	types := make([]string, 0, len(m.components))
	for t := range m.components {
		types = append(types, t)
	}
	return strings.Join(types, ", ")
}

// GetComponentAmp implements interfaces.ModuleContractPort.
func (m *MockModule) GetComponentAmp(componentName, operation string) float64 {
	if m.componentAmps == nil {
		return 1.0
	}
	if ops, ok := m.componentAmps[componentName]; ok {
		if amp, ok := ops[operation]; ok {
			return amp
		}
	}
	return 1.0
}

// GetDependsOn implements interfaces.ModuleContractPort.
func (m *MockModule) GetDependsOn() []string {
	return m.dependsOn
}

// GetVersioningScheme implements interfaces.ModuleContractPort.
func (m *MockModule) GetVersioningScheme() string {
	return m.versioningScheme
}

// GetReleaseType implements interfaces.ModuleContractPort.
func (m *MockModule) GetReleaseType() string {
	return m.releaseType
}

// GetChangelog implements interfaces.ModuleContractPort.
func (m *MockModule) GetChangelog() string {
	return m.changelog
}

// HasVersioning implements interfaces.ModuleContractPort.
func (m *MockModule) HasVersioning() bool {
	return m.versioningScheme != ""
}

// GetMetadata implements interfaces.ModuleContractPort.
func (m *MockModule) GetMetadata() map[string]interface{} {
	return m.metadata
}

// GetContentHash implements interfaces.ModuleContractPort.
// Returns a mock hash based on the module moniker for testing.
func (m *MockModule) GetContentHash() (string, error) {
	// Simple mock: hash the moniker for deterministic test values
	return fmt.Sprintf("%08x", len(m.moniker)*12345), nil
}

// Interface compliance check
var _ interfaces.ModuleContractPort = (*MockModule)(nil)
