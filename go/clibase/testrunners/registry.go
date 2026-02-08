// Package testrunners provides a registry-based test type dispatch system.
// Each test type (gotest, godog, tscucumber, mocha) has a dedicated runner.
package testrunners

import (
	"context"
	"io"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
)

// TestInfo provides structured metadata for test execution and reporting.
type TestInfo struct {
	// ModuleMoniker is the module identifier for aggregation (e.g., "eac-cli", "books")
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

	// IsBDD returns true if this runner handles BDD-style test types.
	IsBDD() bool

	// GetTestInfo extracts structured test metadata from a test reference.
	GetTestInfo(test testing.TestReference, workspaceRoot string, cfg *config.EACConfig) *TestInfo

	// FindTestRoot finds the test runner location for a feature file.
	FindTestRoot(featurePath string, cfg *config.EACConfig) string

	// BuildPackagePath constructs the package path for test grouping.
	BuildPackagePath(testRoot, featurePath string) string

	// Execute runs tests for a package and returns results.
	Execute(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, cfg RunConfig) RunResult
}

// TestTypeDescriptor provides metadata about a test type that the core
// needs for discovery, validation, and display -- without knowing the
// specific test framework technology.
type TestTypeDescriptor struct {
	// TestType is the canonical identifier (e.g., "godog", "tscucumber", "mocha").
	// This must match what TestTypeRunner.TestTypes() returns.
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
// This is a minimal struct to avoid importing module types into testrunners.
type FeatureModuleInfo struct {
	HasTypeScript bool
	HasGo         bool
}

// Inference represents a tag inference rule contributed by an adapter.
type Inference struct {
	TestTypes   []string
	IfTags      []string
	ThenAddTags []string
	Description string
}

var (
	mu          sync.RWMutex
	runners     = make(map[string]TestTypeRunner)
	descriptors = make(map[string]*TestTypeDescriptor)
	fallback    TestTypeRunner
)

// Register registers a test type runner for one or more test types.
// Call this from init() in your runner implementation file.
func Register(runner TestTypeRunner) {
	mu.Lock()
	defer mu.Unlock()
	for _, testType := range runner.TestTypes() {
		runners[testType] = runner
	}
}

// RegisterFallback registers a fallback runner for unknown test types.
func RegisterFallback(runner TestTypeRunner) {
	mu.Lock()
	defer mu.Unlock()
	fallback = runner
}

// Get returns the runner for a specific test type.
// Returns nil if no runner is registered for that type.
func Get(testType string) TestTypeRunner {
	mu.RLock()
	defer mu.RUnlock()
	if runner, ok := runners[testType]; ok {
		return runner
	}
	return fallback
}

// GetAll returns all registered runners.
func GetAll() map[string]TestTypeRunner {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]TestTypeRunner, len(runners))
	for k, v := range runners {
		result[k] = v
	}
	return result
}

// SupportedTypes returns a list of all registered test types.
func SupportedTypes() []string {
	mu.RLock()
	defer mu.RUnlock()
	types := make([]string, 0, len(runners))
	for t := range runners {
		types = append(types, t)
	}
	return types
}

// RegisterDescriptor registers a test type descriptor.
// Call this from init() alongside Register().
func RegisterDescriptor(desc *TestTypeDescriptor) {
	if desc == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	descriptors[desc.TestType] = desc
}

// GetDescriptor returns the descriptor for a test type.
func GetDescriptor(testType string) *TestTypeDescriptor {
	mu.RLock()
	defer mu.RUnlock()
	return descriptors[testType]
}

// AllDescriptors returns all registered descriptors (deduplicated).
func AllDescriptors() []*TestTypeDescriptor {
	mu.RLock()
	defer mu.RUnlock()
	seen := make(map[*TestTypeDescriptor]bool)
	var result []*TestTypeDescriptor
	for _, d := range descriptors {
		if !seen[d] {
			result = append(result, d)
			seen[d] = true
		}
	}
	return result
}

// GetComponentType returns the component type for a test type from the registry.
// Returns empty string if not registered.
func GetComponentType(testType string) string {
	mu.RLock()
	defer mu.RUnlock()
	if d, ok := descriptors[testType]; ok {
		return d.ComponentType
	}
	return ""
}

// GetRunnerFileConventions returns all runner file conventions as a set.
// Used by discovery to know which files to skip.
func GetRunnerFileConventions() map[string]bool {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]bool)
	for _, d := range descriptors {
		if d.RunnerFileConvention != "" {
			result[d.RunnerFileConvention] = true
		}
	}
	return result
}

// ResolveFeatureTestType determines which BDD test type should own
// .feature files for a module with the given characteristics.
// Returns "godog" as ultimate fallback if no resolver matches.
func ResolveFeatureTestType(info FeatureModuleInfo) string {
	mu.RLock()
	defer mu.RUnlock()
	for _, d := range descriptors {
		if d.IsBDD && d.FeatureTestTypeResolver != nil {
			if d.FeatureTestTypeResolver(info) {
				return d.TestType
			}
		}
	}
	// Default to first registered BDD type, or "godog" as ultimate fallback
	for _, d := range descriptors {
		if d.IsBDD {
			return d.TestType
		}
	}
	return "godog"
}

// GetMonikerStyle returns the moniker generation style for a test type.
// Returns "file" if not registered (default for unit tests).
func GetMonikerStyle(testType string) string {
	mu.RLock()
	defer mu.RUnlock()
	if d, ok := descriptors[testType]; ok {
		return d.MonikerStyle
	}
	return "file"
}

// CollectInferences returns all default inferences from all registered types.
func CollectInferences() []Inference {
	mu.RLock()
	defer mu.RUnlock()
	var all []Inference
	seen := make(map[*TestTypeDescriptor]bool)
	for _, d := range descriptors {
		if !seen[d] {
			all = append(all, d.DefaultInferences...)
			seen[d] = true
		}
	}
	return all
}

// BDDComponentNames returns the component names used by BDD test types.
// Used to find test implementation paths without hardcoding "godog".
func BDDComponentNames() []string {
	mu.RLock()
	defer mu.RUnlock()
	seen := make(map[string]bool)
	var names []string
	for _, d := range descriptors {
		if d.IsBDD && d.TestType != "" && !seen[d.TestType] {
			seen[d.TestType] = true
			names = append(names, d.TestType)
		}
	}
	if len(names) == 0 {
		return []string{"godog"}
	}
	return names
}

// ResetForTesting clears all registrations. Use only in tests.
func ResetForTesting() {
	mu.Lock()
	defer mu.Unlock()
	runners = make(map[string]TestTypeRunner)
	descriptors = make(map[string]*TestTypeDescriptor)
	fallback = nil
}

func init() {
	// Wire testrunners registry as provider for go/core/testing.
	// This bridges the dependency gap: core defines the provider interface,
	// clibase/testrunners implements it with the actual registry.
	testing.SetSupportedTypesProvider(SupportedTypes)
	testing.SetComponentTypeProvider(GetComponentType)
	testing.SetMonikerStyleProvider(GetMonikerStyle)
	testing.SetRunnerFileConventionsProvider(GetRunnerFileConventions)
	testing.SetFeatureTestTypeProvider(func(hasTypeScript, hasGo bool) string {
		return ResolveFeatureTestType(FeatureModuleInfo{
			HasTypeScript: hasTypeScript,
			HasGo:         hasGo,
		})
	})
	testing.SetInferenceProvider(func() []testing.Inference {
		adapterInferences := CollectInferences()
		result := make([]testing.Inference, len(adapterInferences))
		for i, inf := range adapterInferences {
			result[i] = testing.Inference{
				TestTypes:   inf.TestTypes,
				IfTags:      inf.IfTags,
				ThenAddTags: inf.ThenAddTags,
				Description: inf.Description,
			}
		}
		return result
	})
	testing.SetBDDComponentNamesProvider(BDDComponentNames)
}
