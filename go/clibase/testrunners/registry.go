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
