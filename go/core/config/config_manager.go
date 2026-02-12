package config

import "sync"

// ConfigManager consolidates the three independent global singletons
// (globalConfig, globalConfigCache, gitRemoteProvider) into a single
// struct with unified lifecycle management.
//
// The package-level functions in global.go, cache.go, and git_provider.go
// delegate to the defaultManager instance for backward compatibility.
//
// For testing, use NewConfigManager() to create isolated instances that
// do not share state with the package-level defaults.
type ConfigManager struct {
	mu         sync.Mutex
	config     *EACConfig
	configErr  error
	configOnce sync.Once

	cache     *ConfigCache
	cacheOnce sync.Once

	gitProvider GitRemoteProvider
}

// defaultManager is the package-level singleton used by all existing
// package-level functions (Global, ClearCache, SetGitRemoteProvider, etc.).
var defaultManager = &ConfigManager{}

// NewConfigManager creates a new isolated ConfigManager.
// Useful in tests to avoid sharing state with the package-level defaults.
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// Global returns the global EACConfig singleton from this manager.
// Thread-safe, lazy-loaded on first call.
// Returns nil if config cannot be loaded.
func (m *ConfigManager) Global() *EACConfig {
	m.configOnce.Do(func() {
		m.config, m.configErr = Load(DefaultLoadOptions())
	})
	return m.config
}

// GlobalWithError returns the global EACConfig singleton and any load error.
func (m *ConfigManager) GlobalWithError() (*EACConfig, error) {
	m.configOnce.Do(func() {
		m.config, m.configErr = Load(DefaultLoadOptions())
	})
	return m.config, m.configErr
}

// SetGlobalForTesting allows tests to inject a mock config.
// This resets the singleton - use only in tests.
func (m *ConfigManager) SetGlobalForTesting(cfg *EACConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
	m.configErr = nil
	// Reset Once so it doesn't try to reload
	m.configOnce = sync.Once{}
	m.configOnce.Do(func() {}) // Mark as done
}

// ResetGlobalForTesting resets the global config singleton.
// Use only in tests to restore normal behavior.
func (m *ConfigManager) ResetGlobalForTesting() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = nil
	m.configErr = nil
	m.configOnce = sync.Once{}
}

// ClearCache clears the config cache managed by this manager.
// This is primarily used by tests that need to reload config with different files.
func (m *ConfigManager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cache != nil {
		m.cache.mu.Lock()
		m.cache.cache = make(map[string]map[bool]*EACConfig)
		m.cache.mu.Unlock()
	}
}

// Cache returns the config cache, initializing it on first call.
// Thread-safe via sync.Once.
func (m *ConfigManager) Cache() *ConfigCache {
	m.cacheOnce.Do(func() {
		m.cache = NewConfigCache()
	})
	return m.cache
}

// SetGitRemoteProvider configures the function used to resolve git remote URLs.
// Thread-safe: protected by the manager's mutex.
func (m *ConfigManager) SetGitRemoteProvider(provider GitRemoteProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gitProvider = provider
}

// GetGitRemoteProvider returns the currently configured git remote provider.
// Thread-safe: protected by the manager's mutex.
func (m *ConfigManager) GetGitRemoteProvider() GitRemoteProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.gitProvider
}

// resolveGitRemoteURL returns the remote URL using the configured provider.
// Returns empty string if no provider is set (caller handles fallback).
func (m *ConfigManager) resolveGitRemoteURL(repoRoot, remoteName string) (string, error) {
	provider := m.GetGitRemoteProvider()
	if provider == nil {
		return "", nil
	}
	return provider(repoRoot, remoteName)
}
