package core

// ConfigPort provides access to EAC repository configuration.
type ConfigPort interface {
	// Paths
	GetRepoRoot() string
	GetConfigRoot() string

	// Repository config
	GetRepository() RepositoryConfigPort

	// Optional configs
	GetEnvironments() EnvironmentsConfigPort
	GetTestingTags() TestingTagsConfigPort
	GetTestSuites() TestSuitesConfigPort
	GetComponentKinds() ComponentKindsConfigPort
}

// RepositoryConfigPort provides repository-level configuration.
type RepositoryConfigPort interface {
	// Module access
	AllMonikers() []string
	GetModule(moniker string) (ModuleContractPort, bool)

	// Output paths
	GetBuildOutputPath(workspaceRoot, moniker string) string
	GetTestOutputPath(workspaceRoot, moniker string) string
	GetScanOutputPath(workspaceRoot, moniker string) string
	GetLintOutputPath(workspaceRoot, moniker string) string
	TestOutputDir() string

	// Artifacts
	GetBuildArtifacts(moniker string, includeAll bool) []ArtifactPort
	GetBuildArtifactIDs(moniker string, buildAll bool) []string

	// Books
	GetBooksByModule(moniker string) []BookConfigPort
	GetDefaultBooksByModule(moniker string) []BookConfigPort
}

// ArtifactPort represents a build artifact.
type ArtifactPort interface {
	GetID() string
	GetType() string
	GetPattern() string
}

// BookConfigPort represents a book configuration.
type BookConfigPort interface {
	GetName() string
	GetType() string
	GetRoot() string
}

// EnvironmentsConfigPort provides environment configuration.
type EnvironmentsConfigPort interface {
	GetEnvironment(name string) (EnvironmentPort, bool)
	ListEnvironments() []string
}

// EnvironmentPort represents a single environment.
type EnvironmentPort interface {
	GetName() string
	GetVariables() map[string]string
}

// TestingTagsConfigPort provides testing tag configuration.
type TestingTagsConfigPort interface {
	GetSkipReasons() map[string]string
	GetTags() []string
}

// TestSuitesConfigPort provides test suite configuration.
type TestSuitesConfigPort interface {
	ListDefault() []string
	GetSuiteLTags(suiteName string) []string
	GetSuites() []TestSuitePort
}

// TestSuitePort represents a test suite.
type TestSuitePort interface {
	GetName() string
	GetTags() []string
	IsDefault() bool
}

// ComponentKindsConfigPort provides component kind configuration.
type ComponentKindsConfigPort interface {
	Get(name string) ComponentTypePort
	List() []string
}

// ConfigLoaderPort loads configuration.
type ConfigLoaderPort interface {
	// Load loads all EAC configuration from the repository.
	Load(opts ConfigLoadOptions) (ConfigPort, error)
}

// ConfigLoadOptions configures the loading behavior.
type ConfigLoadOptions struct {
	RepoRoot        string
	ValidateSchemas bool
	LazyLoad        bool
}
