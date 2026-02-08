package config

import "sync"

var (
	globalConfig     *EACConfig
	globalConfigOnce sync.Once
	globalConfigErr  error
	globalConfigMu   sync.Mutex
)

// Global returns the global EACConfig singleton.
// Thread-safe, lazy-loaded on first call.
// Returns nil if config cannot be loaded.
func Global() *EACConfig {
	globalConfigOnce.Do(func() {
		globalConfig, globalConfigErr = Load(DefaultLoadOptions())
	})
	return globalConfig
}

// GlobalWithError returns the global EACConfig singleton and any load error.
func GlobalWithError() (*EACConfig, error) {
	globalConfigOnce.Do(func() {
		globalConfig, globalConfigErr = Load(DefaultLoadOptions())
	})
	return globalConfig, globalConfigErr
}

// SetGlobalForTesting allows tests to inject a mock config.
// This resets the singleton - use only in tests.
func SetGlobalForTesting(cfg *EACConfig) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = cfg
	globalConfigErr = nil
	// Reset Once so it doesn't try to reload
	globalConfigOnce = sync.Once{}
	globalConfigOnce.Do(func() {}) // Mark as done
}

// ResetGlobalForTesting resets the global config singleton.
// Use only in tests to restore normal behavior.
func ResetGlobalForTesting() {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = nil
	globalConfigErr = nil
	globalConfigOnce = sync.Once{}
}
