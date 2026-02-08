package tool

import (
	"os"
	"path/filepath"
	"sync"

	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

var (
	globalToolConfig     *ToolConfig
	globalToolConfigOnce sync.Once
	globalToolConfigErr  error
	globalToolConfigMu   sync.Mutex
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
	globalToolConfigMu.Lock()
	defer globalToolConfigMu.Unlock()
	globalToolConfig = cfg
	globalToolConfigErr = nil
	// Reset Once so it doesn't try to reload
	globalToolConfigOnce = sync.Once{}
	globalToolConfigOnce.Do(func() {}) // Mark as done
}

// ResetGlobalToolConfigForTesting resets the global tool config singleton.
// Use only in tests to restore normal behavior.
func ResetGlobalToolConfigForTesting() {
	globalToolConfigMu.Lock()
	defer globalToolConfigMu.Unlock()
	globalToolConfig = nil
	globalToolConfigErr = nil
	globalToolConfigOnce = sync.Once{}
}

// builtinTestTypeMapping provides fallback mappings when tool config and
// adapter registry are unavailable (e.g., in unit tests without adapter init).
var builtinTestTypeMapping = map[string]string{
	"gotest":     "go",
	"godog":      "gherkin",
	"mocha":      "typescript",
	"tscucumber": "gherkin",
}

// GetTestTypeComponentType returns the component type for a test type.
// Uses the global tool config's test-type-mapping first, then falls back
// to the adapter registry, then to built-in defaults.
func GetTestTypeComponentType(testType string) string {
	// 1. Try tool config (data-driven from YAML)
	cfg := GlobalToolConfig()
	if cfg != nil && cfg.TestTypeMapping != nil {
		if compType, ok := cfg.TestTypeMapping[testType]; ok {
			return compType
		}
	}

	// 2. Try adapter registry via provider
	if compType := coretesting.GetComponentTypeFromRegistry(testType); compType != "" {
		return compType
	}

	// 3. Built-in fallback for well-known types
	if compType, ok := builtinTestTypeMapping[testType]; ok {
		return compType
	}

	// 4. Ultimate fallback
	return "go"
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
