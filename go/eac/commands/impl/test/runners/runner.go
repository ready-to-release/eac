// Package runners provides a registry-based test type dispatch system.
// Each test type (gotest, godog, tscucumber, mocha) has a dedicated runner
// that knows how to:
// - Find test runner locations for feature files
// - Build package paths for test grouping
// - Execute tests with appropriate commands
package runners

import (
	"io"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

// TestInfo provides structured metadata for test execution and reporting.
// This normalizes information across all test types for consistent aggregation.
type TestInfo struct {
	// ModuleMoniker is the module identifier for aggregation (e.g., "eac-commands", "books")
	ModuleMoniker string

	// Language is the programming language (e.g., "go", "ts")
	Language string

	// PackageKey is the unique key for grouping tests (used internally)
	PackageKey string

	// DisplayName is the human-readable name for TUI display
	DisplayName string

	// TestRoot is the directory where tests are executed from
	TestRoot string
}

// RunResult holds the results from running a package's tests.
type RunResult struct {
	// ModuleMoniker is the module this package belongs to (for aggregation)
	ModuleMoniker string
	PackageName   string
	LogFilePath   string
	TestsPassed   int
	TestsFailed   int
	TestsSkipped  int
	TestsTotal    int
	PackageFailed bool
	Duration      time.Duration
}

// RunConfig holds configuration for test execution.
type RunConfig struct {
	WorkspaceRoot  string
	TestRunDir     string
	Coverage       bool
	SuiteTagFilter string
	Parallelism    int

	// ModuleMoniker is the module this package belongs to (for result aggregation).
	ModuleMoniker string

	// ModuleOutputPath is the module-based output path for this package's results.
	// Format: "<module-moniker>/<subpath>" e.g., "eac-core/contracts"
	// This is used instead of the raw package path for cleaner output organization.
	ModuleOutputPath string

	// OutputDir is the pre-created output directory for this test.
	// When set, runners use this directory instead of creating their own.
	// Structure: out/test/<module>/<component>
	OutputDir string
}

// TestTypeRunner defines the interface for test type-specific runners.
// Each runner handles a specific test type (gotest, godog, tscucumber, mocha).
type TestTypeRunner interface {
	// TestTypes returns the test types this runner handles.
	// Most runners handle a single type, but some (like GoRunner) handle multiple.
	TestTypes() []string

	// GetTestInfo extracts structured test metadata from a test reference.
	// This provides consistent information for grouping, display, and aggregation.
	// Parameters:
	//   - test: the test reference to analyze
	//   - workspaceRoot: the workspace root directory
	//   - cfg: loaded configuration
	// Returns structured test info, or nil if the test cannot be processed.
	GetTestInfo(test testing.TestReference, workspaceRoot string, cfg *config.EACConfig) *TestInfo

	// FindTestRoot finds the test runner location for a feature file.
	// For BDD tests, this returns the package/module where test implementation lives.
	// For unit tests, this returns empty string (tests are in the same directory).
	// Parameters:
	//   - featurePath: relative path to the feature file from workspace root
	//   - cfg: loaded configuration
	// Returns the relative path to the test runner directory, or empty if not found.
	FindTestRoot(featurePath string, cfg *config.EACConfig) string

	// BuildPackagePath constructs the package path for test grouping.
	// For BDD tests, this typically returns "testRoot:featurePath".
	// For unit tests, this returns the directory path.
	// Parameters:
	//   - testRoot: the result from FindTestRoot (or directory for unit tests)
	//   - featurePath: relative path to the feature file (empty for unit tests)
	// Returns the package path used as a key for test grouping.
	BuildPackagePath(testRoot, featurePath string) string

	// Execute runs tests for a package and returns results.
	// Parameters:
	//   - pkgPath: the package path (from BuildPackagePath)
	//   - tests: the tests to run in this package
	//   - tuiWriter: writer for TUI output (status messages)
	//   - cfg: run configuration
	// Returns the test execution results.
	Execute(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, cfg RunConfig) RunResult
}
