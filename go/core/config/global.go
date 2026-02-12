package config

// Global returns the global EACConfig singleton.
// Thread-safe, lazy-loaded on first call.
// Returns nil if config cannot be loaded.
func Global() *EACConfig {
	return defaultManager.Global()
}

// GlobalWithError returns the global EACConfig singleton and any load error.
func GlobalWithError() (*EACConfig, error) {
	return defaultManager.GlobalWithError()
}

// SetGlobalForTesting allows tests to inject a mock config.
// This resets the singleton - use only in tests.
func SetGlobalForTesting(cfg *EACConfig) {
	defaultManager.SetGlobalForTesting(cfg)
}

// ResetGlobalForTesting resets the global config singleton.
// Use only in tests to restore normal behavior.
func ResetGlobalForTesting() {
	defaultManager.ResetGlobalForTesting()
}
