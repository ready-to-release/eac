// registry.go - Self-registration system for test type runners
package runners

import (
	"sync"
)

var (
	mu       sync.RWMutex
	runners  = make(map[string]TestTypeRunner)
	fallback TestTypeRunner
)

// Register registers a test type runner for one or more test types.
// Call this from init() in your runner implementation file.
// The runner's TestTypes() method determines which types it handles.
func Register(runner TestTypeRunner) {
	mu.Lock()
	defer mu.Unlock()

	for _, testType := range runner.TestTypes() {
		runners[testType] = runner
	}
}

// RegisterFallback registers a fallback runner for unknown test types.
// This is typically the Go test runner which handles default cases.
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
// Useful for debugging and listing available test types.
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
