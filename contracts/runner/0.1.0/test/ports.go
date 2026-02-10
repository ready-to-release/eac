// Package test defines the port interfaces and types for test runner adapters.
// Adapters implement TestRunnerPort; the execution engine (clibase/testrunners)
// consumes them via registry.
package test

import (
	"context"
	"io"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// TestRunnerPort is the interface that test runner adapters implement.
// Each adapter handles one or more test types (gotest, godog, tscucumber, mocha).
//
// The cfg parameter in GetTestInfo and FindTestRoot is typed as any to avoid
// importing implementation-specific config types. Adapters should type-assert
// to *config.EACConfig (or their specific config type).
type TestRunnerPort interface {
	// TestTypes returns the test types this runner handles.
	// Most runners handle a single type, but some (like GoRunner) handle multiple.
	TestTypes() []string

	// IsBDD returns true if this runner handles BDD-style test types.
	IsBDD() bool

	// GetTestInfo extracts structured test metadata from a test reference.
	// cfg is expected to be *config.EACConfig.
	GetTestInfo(test TestReference, workspaceRoot string, cfg any) *TestInfo

	// FindTestRoot finds the test runner location for a feature file.
	// cfg is expected to be *config.EACConfig.
	FindTestRoot(featurePath string, cfg any) string

	// BuildPackagePath constructs the package path for test grouping.
	BuildPackagePath(testRoot, featurePath string) string

	// Execute runs tests for a package and returns results.
	Execute(pkgPath string, tests []TestReference, tuiWriter io.Writer, cfg RunConfig) RunResult
}

// TestReference identifies a specific test with its tags.
type TestReference struct {
	FilePath string   // Path to test file
	Type     string   // "godog", "gotest", "tscucumber", "mocha"
	TestName string   // Name of test/scenario
	Tags     []string // All effective tags (source + inherited + inferred)

	// Tag provenance tracking
	SourceTags   []string // Tags explicitly in source file (physical tags before inference)
	InferredTags []string // Tags added by inference rules
	InferredDeps []string // System deps inferred from module types (@deps:go from go-* modules)
	InferredDepm []string // Module deps inferred from file path (@depm:X from specs/X/...)

	// Execution control
	IsIgnored    bool   // Has @skip:<reason> tag
	SkipReason   string // Reason code from @skip:<reason> (e.g., "wip", "broken")
	IsManual     bool   // Has @Manual tag
	IsSequential bool   // Has @sequential tag - must run sequentially, not in parallel

	// Dependency tracking
	SystemDependencies []string // System deps extracted from @deps:<name> (e.g., "docker", "git")
	ModuleDependencies []string // Module deps extracted from @depm:<module> (e.g., "clie", "eac")

	// Risk control linkage
	RiskControls []string // All @risk-control:<name>-<id> references

	// GxP regulatory classification
	IsGxP            bool // Has @gxp tag
	IsCriticalAspect bool // Has @gmp-critical-aspect tag
}

// TestInfo provides structured metadata for test execution and reporting.
type TestInfo struct {
	// ModuleMoniker is the module identifier for aggregation (e.g., "eac", "books")
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
	// Ctx is the worker context for cancellation/timeout propagation.
	// When the orchestrator's worker timeout fires, this context is cancelled,
	// allowing runners to kill subprocesses cleanly.
	// If nil, runners fall back to context.Background().
	Ctx context.Context

	WorkspaceRoot  string
	TestRunDir     string
	Coverage       bool
	SuiteTagFilter core.TagFilter
	Parallelism    int

	// ModuleMoniker is the module this package belongs to (for result aggregation).
	ModuleMoniker string

	// ModuleOutputPath is the module-based output path for this package's results.
	// Format: "<module-moniker>/<subpath>" e.g., "core/contracts"
	ModuleOutputPath string

	// OutputDir is the pre-created output directory for this test.
	// When set, runners use this directory instead of creating their own.
	// Structure: out/test/<module>/<component>
	OutputDir string
}

// TestTypeDescriptor provides metadata about a test type that the core
// needs for discovery, validation, and display -- without knowing the
// specific test framework technology.
type TestTypeDescriptor struct {
	// TestType is the canonical identifier (e.g., "godog", "tscucumber", "mocha").
	// This must match what TestRunnerPort.TestTypes() returns.
	TestType string

	// IsBDD indicates this is a BDD/Gherkin test type.
	// BDD types discover tests from .feature files.
	IsBDD bool

	// ComponentType is the component type this test type maps to
	// for UoW naming (e.g., "gherkin", "go", "typescript").
	ComponentType string

	// MonikerStyle controls how test monikers are generated.
	// "feature" = module_feature-name_scenario-name (for BDD)
	// "file"    = module_test-file_TestName (for unit tests)
	MonikerStyle string

	// RunnerFileConvention is the conventional filename for the test
	// runner bootstrap file (e.g., "godog_test.go").
	// Used by discovery to skip these files when scanning for unit tests.
	// Empty string means no runner file convention.
	RunnerFileConvention string

	// FeatureTestTypeResolver determines which BDD test type to assign
	// to .feature files discovered for a given module.
	// Only set for BDD types. The registry uses this to resolve
	// which BDD type owns a module's feature files.
	// Returns true if this BDD type should own features for the module.
	FeatureTestTypeResolver func(info FeatureModuleInfo) bool

	// DefaultInferences are tag inference rules contributed by this test type.
	// These are merged into the global inference set.
	DefaultInferences []Inference
}

// FeatureModuleInfo provides module metadata for BDD type resolution.
type FeatureModuleInfo struct {
	HasTypeScript bool
	HasGo         bool
	HasPython     bool
	HasDotnet     bool
}

// Inference represents a tag inference rule contributed by an adapter.
type Inference struct {
	TestTypes   []string
	IfTags      []string
	ThenAddTags []string
	Description string
}
