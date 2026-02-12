// Package config provides a central configuration loader for all EAC repository configs.
package config

import (
	"sync"
)

// ConfigCache provides cached EAC configuration data.
// Thread-safe for concurrent access by parallel test packages.
// No cache invalidation - data persists for process lifetime.
type ConfigCache struct {
	mu sync.RWMutex

	// Cache key: repoRoot + validateSchemas
	// Map structure: map[repoRoot]map[validateSchemas]*EACConfig
	cache map[string]map[bool]*EACConfig
}

// NewConfigCache creates a new empty cache.
func NewConfigCache() *ConfigCache {
	return &ConfigCache{
		cache: make(map[string]map[bool]*EACConfig),
	}
}

// ClearCache clears the global config cache.
// This is primarily used by tests that need to reload config with different files.
func ClearCache() {
	defaultManager.ClearCache()
}

// Get returns cached config if available.
func (c *ConfigCache) Get(repoRoot string, validateSchemas bool) (*EACConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if repoConfigs, ok := c.cache[repoRoot]; ok {
		if cfg, ok := repoConfigs[validateSchemas]; ok {
			return cfg, true
		}
	}
	return nil, false
}

// Set caches a config after validation.
func (c *ConfigCache) Set(repoRoot string, validateSchemas bool, cfg *EACConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache[repoRoot] == nil {
		c.cache[repoRoot] = make(map[bool]*EACConfig)
	}
	c.cache[repoRoot][validateSchemas] = cfg
}
