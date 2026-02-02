package mocks

import "github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"

// MockConfig implements interfaces.ConfigPort for testing.
type MockConfig struct {
	repoRoot       string
	configRoot     string
	repository     interfaces.RepositoryConfigPort
	environments   interfaces.EnvironmentsConfigPort
	testingTags    interfaces.TestingTagsConfigPort
	testSuites     interfaces.TestSuitesConfigPort
	componentTypes interfaces.ComponentTypesConfigPort
}

// NewMockConfig creates a new MockConfig with sensible defaults.
func NewMockConfig() *MockConfig {
	return &MockConfig{
		repoRoot:   "/mock/workspace",
		configRoot: "/mock/workspace/.r2r/eac",
	}
}

// WithRepoRoot sets the repository root path.
func (m *MockConfig) WithRepoRoot(root string) *MockConfig {
	m.repoRoot = root
	return m
}

// WithConfigRoot sets the config root path.
func (m *MockConfig) WithConfigRoot(root string) *MockConfig {
	m.configRoot = root
	return m
}

// WithRepository sets the repository configuration.
func (m *MockConfig) WithRepository(repo interfaces.RepositoryConfigPort) *MockConfig {
	m.repository = repo
	return m
}

// WithEnvironments sets the environments configuration.
func (m *MockConfig) WithEnvironments(envs interfaces.EnvironmentsConfigPort) *MockConfig {
	m.environments = envs
	return m
}

// WithTestingTags sets the testing tags configuration.
func (m *MockConfig) WithTestingTags(tags interfaces.TestingTagsConfigPort) *MockConfig {
	m.testingTags = tags
	return m
}

// WithTestSuites sets the test suites configuration.
func (m *MockConfig) WithTestSuites(suites interfaces.TestSuitesConfigPort) *MockConfig {
	m.testSuites = suites
	return m
}

// WithComponentTypes sets the component types configuration.
func (m *MockConfig) WithComponentTypes(types interfaces.ComponentTypesConfigPort) *MockConfig {
	m.componentTypes = types
	return m
}

// GetRepoRoot implements interfaces.ConfigPort.
func (m *MockConfig) GetRepoRoot() string {
	return m.repoRoot
}

// GetConfigRoot implements interfaces.ConfigPort.
func (m *MockConfig) GetConfigRoot() string {
	return m.configRoot
}

// GetRepository implements interfaces.ConfigPort.
func (m *MockConfig) GetRepository() interfaces.RepositoryConfigPort {
	if m.repository != nil {
		return m.repository
	}
	return NewMockRepositoryConfig()
}

// GetEnvironments implements interfaces.ConfigPort.
func (m *MockConfig) GetEnvironments() interfaces.EnvironmentsConfigPort {
	return m.environments
}

// GetTestingTags implements interfaces.ConfigPort.
func (m *MockConfig) GetTestingTags() interfaces.TestingTagsConfigPort {
	return m.testingTags
}

// GetTestSuites implements interfaces.ConfigPort.
func (m *MockConfig) GetTestSuites() interfaces.TestSuitesConfigPort {
	return m.testSuites
}

// GetComponentTypes implements interfaces.ConfigPort.
func (m *MockConfig) GetComponentTypes() interfaces.ComponentTypesConfigPort {
	return m.componentTypes
}

// Interface compliance check
var _ interfaces.ConfigPort = (*MockConfig)(nil)

// MockRepositoryConfig implements interfaces.RepositoryConfigPort for testing.
type MockRepositoryConfig struct {
	modules      map[string]interfaces.ModuleContractPort
	monikers     []string
	testOutputDir string
}

// NewMockRepositoryConfig creates a new MockRepositoryConfig.
func NewMockRepositoryConfig() *MockRepositoryConfig {
	return &MockRepositoryConfig{
		modules:       make(map[string]interfaces.ModuleContractPort),
		testOutputDir: "out/test",
	}
}

// WithModule adds a module to the repository config.
func (m *MockRepositoryConfig) WithModule(module interfaces.ModuleContractPort) *MockRepositoryConfig {
	m.modules[module.GetMoniker()] = module
	m.monikers = append(m.monikers, module.GetMoniker())
	return m
}

// AllMonikers implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) AllMonikers() []string {
	return m.monikers
}

// GetModule implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetModule(moniker string) (interfaces.ModuleContractPort, bool) {
	mod, ok := m.modules[moniker]
	return mod, ok
}

// GetBuildOutputPath implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetBuildOutputPath(workspaceRoot, moniker string) string {
	return workspaceRoot + "/out/build/" + moniker
}

// GetTestOutputPath implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetTestOutputPath(workspaceRoot, moniker string) string {
	return workspaceRoot + "/out/test/" + moniker
}

// GetScanOutputPath implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetScanOutputPath(workspaceRoot, moniker string) string {
	return workspaceRoot + "/out/scan/" + moniker
}

// GetLintOutputPath implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetLintOutputPath(workspaceRoot, moniker string) string {
	return workspaceRoot + "/out/lint/" + moniker
}

// TestOutputDir implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) TestOutputDir() string {
	return m.testOutputDir
}

// GetBuildArtifacts implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetBuildArtifacts(moniker string, includeAll bool) []interfaces.ArtifactPort {
	return nil
}

// GetBuildArtifactIDs implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetBuildArtifactIDs(moniker string, buildAll bool) []string {
	return nil
}

// GetBooksByModule implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetBooksByModule(moniker string) []interfaces.BookConfigPort {
	return nil
}

// GetDefaultBooksByModule implements interfaces.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetDefaultBooksByModule(moniker string) []interfaces.BookConfigPort {
	return nil
}

// Interface compliance check
var _ interfaces.RepositoryConfigPort = (*MockRepositoryConfig)(nil)
