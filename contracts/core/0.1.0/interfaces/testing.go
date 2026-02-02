package interfaces

// ModuleReportPort provides access to module loading results.
type ModuleReportPort interface {
	// Registry returns the module registry from the report.
	Registry() ModuleRegistryPort

	// Errors returns any errors encountered during module loading.
	Errors() []error
}

// TestCachePort provides cached repository state for tests.
// This cache is for the ORIGINAL repository root only, not isolated test repos.
type TestCachePort interface {
	// EnsurePopulated ensures the cache is populated for the given repo root.
	EnsurePopulated(repoRoot string) error

	// ModuleReport returns the module contract report.
	ModuleReport() (ModuleReportPort, error)

	// TrackedFiles returns all git-tracked files (normalized paths).
	TrackedFiles() []string

	// FilesByExtension returns files matching the given extension (e.g., ".md").
	FilesByExtension(ext string) []string

	// FilesInDir returns files under the given directory prefix.
	FilesInDir(dir string) []string

	// FilesInDirWithExtension returns files under dir matching extension.
	FilesInDirWithExtension(dir, ext string) []string

	// AbsolutePath returns the absolute path for a relative tracked file.
	AbsolutePath(relPath string) string

	// RepoRoot returns the cached repository root.
	RepoRoot() string
}

// TestIsolationPort provides isolated test environments.
type TestIsolationPort interface {
	// Setup creates the isolated environment.
	Setup() error

	// Cleanup removes the isolated environment.
	Cleanup()

	// IsolatedDir returns the path to the isolated directory.
	IsolatedDir() string

	// Environment returns environment variables for the isolated environment.
	// These should be passed to subprocesses.
	Environment() []string
}

// TestIsolationBuilderPort builds test isolation configurations.
type TestIsolationBuilderPort interface {
	// WithOriginalRepoRoot sets the original repository root.
	WithOriginalRepoRoot(root string) TestIsolationBuilderPort

	// WithCopyContracts enables copying contracts to isolated dir.
	WithCopyContracts(copy bool) TestIsolationBuilderPort

	// WithCopySpecs enables copying specs to isolated dir.
	WithCopySpecs(copy bool) TestIsolationBuilderPort

	// WithCopyAIContracts enables copying AI configs to isolated dir.
	WithCopyAIContracts(copy bool) TestIsolationBuilderPort

	// WithMockAIConfig enables mock AI configuration.
	WithMockAIConfig(mock bool) TestIsolationBuilderPort

	// Build creates the TestIsolationPort.
	Build() TestIsolationPort
}

// FixturePoolPort manages reusable test environment templates.
type FixturePoolPort interface {
	// CreateTemplate creates a fixture template for fast copying.
	CreateTemplate(originalRepoRoot string) (FixtureTemplatePort, error)

	// NewIsolationFromTemplate creates isolation by copying a template.
	NewIsolationFromTemplate(template FixtureTemplatePort) (TestIsolationPort, error)

	// Cleanup removes all templates and temp directories.
	Cleanup()
}

// FixtureTemplatePort represents a reusable test environment template.
type FixtureTemplatePort interface {
	// TemplateDir returns the template directory path.
	TemplateDir() string

	// OriginalRepoRoot returns the original repo root.
	OriginalRepoRoot() string
}

// TestingHelper provides test utility functions.
type TestingHelper interface {
	// RequireGitRepository skips the test if not in a git repository.
	RequireGitRepository(t TestingT)

	// RequireIsolation fails the test if not running in isolation.
	RequireIsolation(t TestingT)

	// ForTesting sets up workspace for testing, returns cleanup func.
	ForTesting(t TestingT, root string) func()
}

// TestingT is the minimal testing interface (subset of *testing.T).
type TestingT interface {
	Helper()
	Skip(args ...interface{})
	Skipf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Cleanup(func())
}
