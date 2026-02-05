package tool

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	globalToolConfig     *ToolConfig
	globalToolConfigOnce sync.Once
	globalToolConfigErr  error
)

// GlobalToolConfig returns the global ToolConfig singleton.
// Thread-safe, lazy-loaded on first call.
// Returns nil if config cannot be loaded.
func GlobalToolConfig() *ToolConfig {
	globalToolConfigOnce.Do(func() {
		// Find repo root by looking for .eac directory
		repoRoot := findRepoRootForGlobal()
		if repoRoot == "" {
			globalToolConfigErr = ErrNoToolConfig
			return
		}
		configRoot := filepath.Join(repoRoot, ".eac")
		globalToolConfig, globalToolConfigErr = LoadToolConfig(repoRoot, configRoot)
	})
	return globalToolConfig
}

// GlobalToolConfigWithError returns the global ToolConfig singleton and any load error.
func GlobalToolConfigWithError() (*ToolConfig, error) {
	globalToolConfigOnce.Do(func() {
		repoRoot := findRepoRootForGlobal()
		if repoRoot == "" {
			globalToolConfigErr = ErrNoToolConfig
			return
		}
		configRoot := filepath.Join(repoRoot, ".eac")
		globalToolConfig, globalToolConfigErr = LoadToolConfig(repoRoot, configRoot)
	})
	return globalToolConfig, globalToolConfigErr
}

// SetGlobalToolConfigForTesting allows tests to inject a mock config.
// This resets the singleton - use only in tests.
func SetGlobalToolConfigForTesting(cfg *ToolConfig) {
	globalToolConfig = cfg
	globalToolConfigErr = nil
	// Reset Once so it doesn't try to reload
	globalToolConfigOnce = sync.Once{}
	globalToolConfigOnce.Do(func() {}) // Mark as done
}

// ResetGlobalToolConfigForTesting resets the global tool config singleton.
// Use only in tests to restore normal behavior.
func ResetGlobalToolConfigForTesting() {
	globalToolConfig = nil
	globalToolConfigErr = nil
	globalToolConfigOnce = sync.Once{}
}

// GetTestTypeComponentType returns the component type for a test type.
// Uses the global tool config's test-type-mapping.
// Falls back to hardcoded defaults if config is unavailable.
func GetTestTypeComponentType(testType string) string {
	cfg := GlobalToolConfig()
	if cfg != nil && cfg.TestTypeMapping != nil {
		if compType, ok := cfg.TestTypeMapping[testType]; ok {
			return compType
		}
	}

	// Fallback to hardcoded defaults
	switch testType {
	case "gotest":
		return "go"
	case "godog":
		return "gherkin"
	case "mocha":
		return "typescript"
	case "tscucumber":
		return "gherkin"
	default:
		return "go" // Default
	}
}

// findRepoRootForGlobal walks up from cwd looking for .eac directory.
func findRepoRootForGlobal() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".eac")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // Reached root
		}
		dir = parent
	}
}
