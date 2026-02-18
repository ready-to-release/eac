package testing

import "sync"

// TestTypeProvider supplies test type metadata to core from the adapter layer.
// This avoids circular imports: adapters/clibase register providers, core queries them.
var (
	providerMu sync.RWMutex

	// supportedTypesProvider returns all registered test type names.
	// Set by testrunners registry at init time.
	supportedTypesProvider func() []string

	// componentTypeProvider returns the component type for a test type.
	// Set by testrunners registry at init time.
	componentTypeProvider func(testType string) string

	// monikerStyleProvider returns the moniker style for a test type.
	// "feature" for BDD, "file" for unit tests.
	monikerStyleProvider func(testType string) string

	// runnerFileConventionsProvider returns runner files to skip during discovery.
	runnerFileConventionsProvider func() map[string]bool

	// featureTestTypeProvider resolves which BDD test type owns features for a module.
	featureTestTypeProvider func(hasTypeScript, hasGo, hasPython, hasDotnet bool) string

	// inferenceProvider collects inference rules from all registered adapters.
	inferenceProvider func() []Inference

	// bddComponentNamesProvider returns component names for BDD test types.
	bddComponentNamesProvider func() []string

	// testFrameworkProvider maps component type to its first tester tool.
	// Set by the config/tool layer at init time.
	testFrameworkProvider func(componentType string) string
)

// SetSupportedTypesProvider sets the provider for supported test types.
func SetSupportedTypesProvider(fn func() []string) {
	providerMu.Lock()
	defer providerMu.Unlock()
	supportedTypesProvider = fn
}

// SetComponentTypeProvider sets the provider for test type -> component type mapping.
func SetComponentTypeProvider(fn func(testType string) string) {
	providerMu.Lock()
	defer providerMu.Unlock()
	componentTypeProvider = fn
}

// SetMonikerStyleProvider sets the provider for test type -> moniker style mapping.
func SetMonikerStyleProvider(fn func(testType string) string) {
	providerMu.Lock()
	defer providerMu.Unlock()
	monikerStyleProvider = fn
}

// SetRunnerFileConventionsProvider sets the provider for runner file conventions.
func SetRunnerFileConventionsProvider(fn func() map[string]bool) {
	providerMu.Lock()
	defer providerMu.Unlock()
	runnerFileConventionsProvider = fn
}

// SetFeatureTestTypeProvider sets the provider for BDD test type resolution.
func SetFeatureTestTypeProvider(fn func(hasTypeScript, hasGo, hasPython, hasDotnet bool) string) {
	providerMu.Lock()
	defer providerMu.Unlock()
	featureTestTypeProvider = fn
}

// SetInferenceProvider sets the provider for inference rules from adapters.
func SetInferenceProvider(fn func() []Inference) {
	providerMu.Lock()
	defer providerMu.Unlock()
	inferenceProvider = fn
}

// SetBDDComponentNamesProvider sets the provider for BDD component names.
func SetBDDComponentNamesProvider(fn func() []string) {
	providerMu.Lock()
	defer providerMu.Unlock()
	bddComponentNamesProvider = fn
}

// getSupportedTypes returns all registered test types from the provider.
// Returns nil if no provider is set (adapters must register).
func getSupportedTypes() []string {
	providerMu.RLock()
	fn := supportedTypesProvider
	providerMu.RUnlock()
	if fn != nil {
		return fn()
	}
	return nil
}

// getComponentTypeFromProvider returns the component type for a test type.
// Returns empty string if no provider is set.
func getComponentTypeFromProvider(testType string) string {
	providerMu.RLock()
	fn := componentTypeProvider
	providerMu.RUnlock()
	if fn != nil {
		return fn(testType)
	}
	return ""
}

// getMonikerStyle returns the moniker style for a test type.
// Returns "file" if no provider is set (default for unit tests).
func getMonikerStyle(testType string) string {
	providerMu.RLock()
	fn := monikerStyleProvider
	providerMu.RUnlock()
	if fn != nil {
		return fn(testType)
	}
	return "file"
}

// getRunnerFileConventions returns runner file conventions from the provider.
// Returns empty map if no provider is set (adapters must register).
func getRunnerFileConventions() map[string]bool {
	providerMu.RLock()
	fn := runnerFileConventionsProvider
	providerMu.RUnlock()
	if fn != nil {
		return fn()
	}
	return nil
}

// resolveFeatureTestType determines which BDD test type owns features.
// Delegates to the provider (adapter registry). Returns empty string if no provider is set.
func resolveFeatureTestType(hasTypeScript, hasGo, hasPython, hasDotnet bool) string {
	providerMu.RLock()
	fn := featureTestTypeProvider
	providerMu.RUnlock()
	if fn != nil {
		return fn(hasTypeScript, hasGo, hasPython, hasDotnet)
	}
	return ""
}

// getAdapterInferences returns inference rules from registered adapters.
// Returns nil if no provider is set.
func getAdapterInferences() []Inference {
	providerMu.RLock()
	fn := inferenceProvider
	providerMu.RUnlock()
	if fn != nil {
		return fn()
	}
	return nil
}

// getBDDComponentNames returns component names for BDD test types.
// Returns nil if no provider is set (adapters must register).
func getBDDComponentNames() []string {
	providerMu.RLock()
	fn := bddComponentNamesProvider
	providerMu.RUnlock()
	if fn != nil {
		return fn()
	}
	return nil
}

// GetComponentTypeFromRegistry returns the component type for a test type from the adapter registry.
// This is the public API for cross-package callers (e.g., go/core/tool).
func GetComponentTypeFromRegistry(testType string) string {
	return getComponentTypeFromProvider(testType)
}

// GetBDDComponentNames returns component names for BDD test types from the adapter registry.
// This is the public API for cross-package callers.
func GetBDDComponentNames() []string {
	return getBDDComponentNames()
}

// SetTestFrameworkProvider sets the provider for component type -> test framework mapping.
func SetTestFrameworkProvider(fn func(componentType string) string) {
	providerMu.Lock()
	defer providerMu.Unlock()
	testFrameworkProvider = fn
}

// ResetProvidersForTesting clears all providers. Use only in tests.
func ResetProvidersForTesting() {
	providerMu.Lock()
	defer providerMu.Unlock()
	supportedTypesProvider = nil
	componentTypeProvider = nil
	monikerStyleProvider = nil
	runnerFileConventionsProvider = nil
	featureTestTypeProvider = nil
	inferenceProvider = nil
	bddComponentNamesProvider = nil
	testFrameworkProvider = nil
}
