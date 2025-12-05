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

// RunResult holds the results from running a package's tests
type RunResult struct {
	PackageName   string
	LogFilePath   string
	TestsPassed   int
	TestsFailed   int
	TestsSkipped  int
	TestsTotal    int
	PackageFailed bool
	Duration      time.Duration
}

// RunConfig holds configuration for test execution
type RunConfig struct {
	WorkspaceRoot  string
	TestRunDir     string
	ReportFormat   string
	Coverage       bool
	SuiteTagFilter string
	Parallelism    int

	// ModuleOutputPath is the module-based output path for this package's results.
	// Format: "<module-moniker>/<subpath>" e.g., "eac-core/contracts"
	// This is used instead of the raw package path for cleaner output organization.
	ModuleOutputPath string
}

// TestTypeRunner defines the interface for test type-specific runners.
// Each runner handles a specific test type (gotest, godog, tscucumber, mocha).
type TestTypeRunner interface {
	// TestTypes returns the test types this runner handles.
	// Most runners handle a single type, but some (like GoRunner) handle multiple.
	TestTypes() []string

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
	BuildPackagePath(testRoot string, featurePath string) string

	// Execute runs tests for a package and returns results.
	// Parameters:
	//   - pkgPath: the package path (from BuildPackagePath)
	//   - tests: the tests to run in this package
	//   - tuiWriter: writer for TUI output (status messages)
	//   - cfg: run configuration
	// Returns the test execution results.
	Execute(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, cfg RunConfig) RunResult
}
