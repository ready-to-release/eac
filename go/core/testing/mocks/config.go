package mocks

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

// MockConfig implements core.ConfigPort for testing.
type MockConfig struct {
	repoRoot       string
	configRoot     string
	repository     core.RepositoryConfigPort
	environments   core.EnvironmentsConfigPort
	testingTags    core.TestingTagsConfigPort
	testSuites     core.TestSuitesConfigPort
	componentKinds core.ComponentKindsConfigPort
}

// NewMockConfig creates a new MockConfig with sensible defaults.
func NewMockConfig() *MockConfig {
	return &MockConfig{
		repoRoot:   "/mock/workspace",
		configRoot: "/mock/workspace/.eac",
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
func (m *MockConfig) WithRepository(repo core.RepositoryConfigPort) *MockConfig {
	m.repository = repo
	return m
}

// WithEnvironments sets the environments configuration.
func (m *MockConfig) WithEnvironments(envs core.EnvironmentsConfigPort) *MockConfig {
	m.environments = envs
	return m
}

// WithTestingTags sets the testing tags configuration.
func (m *MockConfig) WithTestingTags(tags core.TestingTagsConfigPort) *MockConfig {
	m.testingTags = tags
	return m
}

// WithTestSuites sets the test suites configuration.
func (m *MockConfig) WithTestSuites(suites core.TestSuitesConfigPort) *MockConfig {
	m.testSuites = suites
	return m
}

// WithComponentKinds sets the component kinds configuration.
func (m *MockConfig) WithComponentKinds(kinds core.ComponentKindsConfigPort) *MockConfig {
	m.componentKinds = kinds
	return m
}

// GetRepoRoot implements core.ConfigPort.
func (m *MockConfig) GetRepoRoot() string {
	return m.repoRoot
}

// GetConfigRoot implements core.ConfigPort.
func (m *MockConfig) GetConfigRoot() string {
	return m.configRoot
}

// GetRepository implements core.ConfigPort.
func (m *MockConfig) GetRepository() core.RepositoryConfigPort {
	if m.repository != nil {
		return m.repository
	}
	return NewMockRepositoryConfig()
}

// GetEnvironments implements core.ConfigPort.
func (m *MockConfig) GetEnvironments() core.EnvironmentsConfigPort {
	return m.environments
}

// GetTestingTags implements core.ConfigPort.
func (m *MockConfig) GetTestingTags() core.TestingTagsConfigPort {
	return m.testingTags
}

// GetTestSuites implements core.ConfigPort.
func (m *MockConfig) GetTestSuites() core.TestSuitesConfigPort {
	return m.testSuites
}

// GetComponentKinds implements core.ConfigPort.
func (m *MockConfig) GetComponentKinds() core.ComponentKindsConfigPort {
	return m.componentKinds
}

// Interface compliance check
var _ core.ConfigPort = (*MockConfig)(nil)

// MockRepositoryConfig implements core.RepositoryConfigPort for testing.
type MockRepositoryConfig struct {
	modules      map[string]core.ModuleContractPort
	monikers     []string
	testOutputDir string
}

// NewMockRepositoryConfig creates a new MockRepositoryConfig.
func NewMockRepositoryConfig() *MockRepositoryConfig {
	return &MockRepositoryConfig{
		modules:       make(map[string]core.ModuleContractPort),
		testOutputDir: "out/test",
	}
}

// WithModule adds a module to the repository config.
func (m *MockRepositoryConfig) WithModule(module core.ModuleContractPort) *MockRepositoryConfig {
	m.modules[module.GetMoniker()] = module
	m.monikers = append(m.monikers, module.GetMoniker())
	return m
}

// AllMonikers implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) AllMonikers() []string {
	return m.monikers
}

// GetModule implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetModule(moniker string) (core.ModuleContractPort, bool) {
	mod, ok := m.modules[moniker]
	return mod, ok
}

// GetBuildOutputPath implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetBuildOutputPath(workspaceRoot, moniker string) string {
	return workspaceRoot + "/out/build/" + moniker
}

// GetTestOutputPath implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetTestOutputPath(workspaceRoot, moniker string) string {
	return workspaceRoot + "/out/test/" + moniker
}

// GetScanOutputPath implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetScanOutputPath(workspaceRoot, moniker string) string {
	return workspaceRoot + "/out/scan/" + moniker
}

// GetLintOutputPath implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetLintOutputPath(workspaceRoot, moniker string) string {
	return workspaceRoot + "/out/lint/" + moniker
}

// TestOutputDir implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) TestOutputDir() string {
	return m.testOutputDir
}

// GetBuildArtifacts implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetBuildArtifacts(moniker string, includeAll bool) []core.ArtifactPort {
	return nil
}

// GetBuildArtifactIDs implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetBuildArtifactIDs(moniker string, buildAll bool) []string {
	return nil
}

// GetBooksByModule implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetBooksByModule(moniker string) []core.BookConfigPort {
	return nil
}

// GetDefaultBooksByModule implements core.RepositoryConfigPort.
func (m *MockRepositoryConfig) GetDefaultBooksByModule(moniker string) []core.BookConfigPort {
	return nil
}

// Interface compliance check
var _ core.RepositoryConfigPort = (*MockRepositoryConfig)(nil)
