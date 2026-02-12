// Package testrunners provides a registry-based test type dispatch system.
// Each test type (gotest, godog, tscucumber, mocha) has a dedicated runner.
//
// Port types (TestRunnerPort, TestInfo, RunConfig, etc.) are defined in
// contracts/runner/0.1.0/test and re-exported here for backward compatibility.
package testrunners

import (
	"sync"

	test "github.com/ready-to-release/eac/contracts/runner/0.1.0/test"
	"github.com/ready-to-release/eac/go/core/testing"
)

// Type aliases for backward compatibility during migration.
// New code should import directly from contracts/runner/0.1.0/test.
type (
	TestInfo           = test.TestInfo
	RunResult          = test.RunResult
	RunConfig          = test.RunConfig
	TestTypeRunner     = test.TestRunnerPort
	TestTypeDescriptor = test.TestTypeDescriptor
	FeatureModuleInfo  = test.FeatureModuleInfo
	Inference          = test.Inference
	TestReference      = test.TestReference
)

// Registry holds test runner registrations. Use NewRegistry() to create an
// isolated instance for testing, or use the package-level functions which
// delegate to the default registry.
type Registry struct {
	mu          sync.RWMutex
	runners     map[string]TestTypeRunner
	descriptors map[string]*TestTypeDescriptor
	fallback    TestTypeRunner
}

// NewRegistry creates a new, empty Registry with initialized maps.
func NewRegistry() *Registry {
	return &Registry{
		runners:     make(map[string]TestTypeRunner),
		descriptors: make(map[string]*TestTypeDescriptor),
	}
}

// defaultRegistry is the package-level registry instance.
// All package-level functions delegate to this instance.
var defaultRegistry = NewRegistry()

// --- Registry methods ---

// Register registers a test type runner for one or more test types.
func (r *Registry) Register(runner TestTypeRunner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, testType := range runner.TestTypes() {
		r.runners[testType] = runner
	}
}

// RegisterFallback registers a fallback runner for unknown test types.
func (r *Registry) RegisterFallback(runner TestTypeRunner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = runner
}

// Get returns the runner for a specific test type.
// Returns the fallback runner (or nil) if no runner is registered for that type.
func (r *Registry) Get(testType string) TestTypeRunner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if runner, ok := r.runners[testType]; ok {
		return runner
	}
	return r.fallback
}

// GetAll returns all registered runners.
func (r *Registry) GetAll() map[string]TestTypeRunner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]TestTypeRunner, len(r.runners))
	for k, v := range r.runners {
		result[k] = v
	}
	return result
}

// SupportedTypes returns a list of all registered test types.
func (r *Registry) SupportedTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.runners))
	for t := range r.runners {
		types = append(types, t)
	}
	return types
}

// RegisterDescriptor registers a test type descriptor.
func (r *Registry) RegisterDescriptor(desc *TestTypeDescriptor) {
	if desc == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.descriptors[desc.TestType] = desc
}

// GetDescriptor returns the descriptor for a test type.
func (r *Registry) GetDescriptor(testType string) *TestTypeDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.descriptors[testType]
}

// AllDescriptors returns all registered descriptors (deduplicated).
func (r *Registry) AllDescriptors() []*TestTypeDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[*TestTypeDescriptor]bool)
	var result []*TestTypeDescriptor
	for _, d := range r.descriptors {
		if !seen[d] {
			result = append(result, d)
			seen[d] = true
		}
	}
	return result
}

// GetComponentType returns the component type for a test type from the registry.
// Returns empty string if not registered.
func (r *Registry) GetComponentType(testType string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.descriptors[testType]; ok {
		return d.ComponentType
	}
	return ""
}

// GetRunnerFileConventions returns all runner file conventions as a set.
// Used by discovery to know which files to skip.
func (r *Registry) GetRunnerFileConventions() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]bool)
	for _, d := range r.descriptors {
		if d.RunnerFileConvention != "" {
			result[d.RunnerFileConvention] = true
		}
	}
	return result
}

// ResolveFeatureTestType determines which BDD test type should own
// .feature files for a module with the given characteristics.
// Returns "godog" as ultimate fallback if no resolver matches.
func (r *Registry) ResolveFeatureTestType(info FeatureModuleInfo) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.descriptors {
		if d.IsBDD && d.FeatureTestTypeResolver != nil {
			if d.FeatureTestTypeResolver(info) {
				return d.TestType
			}
		}
	}
	// Default to first registered BDD type, or "godog" as ultimate fallback
	for _, d := range r.descriptors {
		if d.IsBDD {
			return d.TestType
		}
	}
	return "godog"
}

// GetMonikerStyle returns the moniker generation style for a test type.
// Returns "file" if not registered (default for unit tests).
func (r *Registry) GetMonikerStyle(testType string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.descriptors[testType]; ok {
		return d.MonikerStyle
	}
	return "file"
}

// CollectInferences returns all default inferences from all registered types.
func (r *Registry) CollectInferences() []Inference {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var all []Inference
	seen := make(map[*TestTypeDescriptor]bool)
	for _, d := range r.descriptors {
		if !seen[d] {
			all = append(all, d.DefaultInferences...)
			seen[d] = true
		}
	}
	return all
}

// BDDComponentNames returns the component names used by BDD test types.
// Used to find test implementation paths without hardcoding "godog".
func (r *Registry) BDDComponentNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]bool)
	var names []string
	for _, d := range r.descriptors {
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
func (r *Registry) ResetForTesting() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runners = make(map[string]TestTypeRunner)
	r.descriptors = make(map[string]*TestTypeDescriptor)
	r.fallback = nil
}

// DefaultRegistry returns the package-level default registry instance.
// This allows callers to pass the registry explicitly (e.g., RegisterWith)
// rather than relying on package-level functions.
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// --- Package-level functions (delegate to defaultRegistry) ---

// Register registers a test type runner for one or more test types.
// Call this from init() in your runner implementation file.
func Register(runner TestTypeRunner) {
	defaultRegistry.Register(runner)
}

// RegisterFallback registers a fallback runner for unknown test types.
func RegisterFallback(runner TestTypeRunner) {
	defaultRegistry.RegisterFallback(runner)
}

// Get returns the runner for a specific test type.
// Returns nil if no runner is registered for that type.
func Get(testType string) TestTypeRunner {
	return defaultRegistry.Get(testType)
}

// GetAll returns all registered runners.
func GetAll() map[string]TestTypeRunner {
	return defaultRegistry.GetAll()
}

// SupportedTypes returns a list of all registered test types.
func SupportedTypes() []string {
	return defaultRegistry.SupportedTypes()
}

// RegisterDescriptor registers a test type descriptor.
// Call this from init() alongside Register().
func RegisterDescriptor(desc *TestTypeDescriptor) {
	defaultRegistry.RegisterDescriptor(desc)
}

// GetDescriptor returns the descriptor for a test type.
func GetDescriptor(testType string) *TestTypeDescriptor {
	return defaultRegistry.GetDescriptor(testType)
}

// AllDescriptors returns all registered descriptors (deduplicated).
func AllDescriptors() []*TestTypeDescriptor {
	return defaultRegistry.AllDescriptors()
}

// GetComponentType returns the component type for a test type from the registry.
// Returns empty string if not registered.
func GetComponentType(testType string) string {
	return defaultRegistry.GetComponentType(testType)
}

// GetRunnerFileConventions returns all runner file conventions as a set.
// Used by discovery to know which files to skip.
func GetRunnerFileConventions() map[string]bool {
	return defaultRegistry.GetRunnerFileConventions()
}

// ResolveFeatureTestType determines which BDD test type should own
// .feature files for a module with the given characteristics.
// Returns "godog" as ultimate fallback if no resolver matches.
func ResolveFeatureTestType(info FeatureModuleInfo) string {
	return defaultRegistry.ResolveFeatureTestType(info)
}

// GetMonikerStyle returns the moniker generation style for a test type.
// Returns "file" if not registered (default for unit tests).
func GetMonikerStyle(testType string) string {
	return defaultRegistry.GetMonikerStyle(testType)
}

// CollectInferences returns all default inferences from all registered types.
func CollectInferences() []Inference {
	return defaultRegistry.CollectInferences()
}

// BDDComponentNames returns the component names used by BDD test types.
// Used to find test implementation paths without hardcoding "godog".
func BDDComponentNames() []string {
	return defaultRegistry.BDDComponentNames()
}

// ResetForTesting clears all registrations in the default registry. Use only in tests.
func ResetForTesting() {
	defaultRegistry.ResetForTesting()
}

func init() {
	// Wire testrunners registry as provider for go/core/testing.
	// This bridges the dependency gap: core defines the provider interface,
	// clibase/testrunners implements it with the actual registry.
	testing.SetSupportedTypesProvider(SupportedTypes)
	testing.SetComponentTypeProvider(GetComponentType)
	testing.SetMonikerStyleProvider(GetMonikerStyle)
	testing.SetRunnerFileConventionsProvider(GetRunnerFileConventions)
	testing.SetFeatureTestTypeProvider(func(hasTypeScript, hasGo, hasPython, hasDotnet bool) string {
		return ResolveFeatureTestType(FeatureModuleInfo{
			HasTypeScript: hasTypeScript,
			HasGo:         hasGo,
			HasPython:     hasPython,
			HasDotnet:     hasDotnet,
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
