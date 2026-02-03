package cache

// Config holds parsed cache control configuration.
type Config struct {
	// SkipSpecs are cache specifications to skip
	SkipSpecs []Spec
}

// NewConfig creates an empty cache config.
func NewConfig() *Config {
	return &Config{}
}

// Skip adds a spec to the skip list.
func (c *Config) Skip(spec Spec) {
	c.SkipSpecs = append(c.SkipSpecs, spec)
}

// SkipAll adds a skip-all spec.
func (c *Config) SkipAll() {
	c.SkipSpecs = append(c.SkipSpecs, Spec{Level: LevelAll, Type: TypeAll})
}

// ShouldSkip returns true if the given level/type combination should be skipped.
func (c *Config) ShouldSkip(level Level, typ Type) bool {
	if c == nil {
		return false
	}
	for _, spec := range c.SkipSpecs {
		if spec.Matches(level, typ) {
			return true
		}
	}
	return false
}

// Convenience methods for common cache checks

// ShouldSkipState returns true if state caches should be skipped.
func (c *Config) ShouldSkipState() bool {
	return c.ShouldSkip(LevelLocal, TypeState)
}

// ShouldSkipLocalRegistry returns true if local registry cache should be skipped.
func (c *Config) ShouldSkipLocalRegistry() bool {
	return c.ShouldSkip(LevelLocal, TypeRegistry)
}

// ShouldSkipLocalLayer returns true if local layer cache should be skipped.
func (c *Config) ShouldSkipLocalLayer() bool {
	return c.ShouldSkip(LevelLocal, TypeLayer)
}

// ShouldSkipAsset returns true if asset caches should be skipped.
func (c *Config) ShouldSkipAsset() bool {
	return c.ShouldSkip(LevelLocal, TypeAsset)
}

// ShouldSkipWork returns true if work directories should be skipped.
func (c *Config) ShouldSkipWork() bool {
	return c.ShouldSkip(LevelLocal, TypeWork)
}

// ShouldSkipRemoteLayer returns true if remote layer cache should be skipped.
func (c *Config) ShouldSkipRemoteLayer() bool {
	return c.ShouldSkip(LevelRemote, TypeLayer)
}

// ShouldForcePull returns true if Docker images should be force-pulled.
// This is triggered when local:registry is skipped.
func (c *Config) ShouldForcePull() bool {
	return c.ShouldSkipLocalRegistry()
}

// ShouldForceNoCacheDocker returns true if Docker build should use --no-cache.
// This is triggered when local:layer is skipped.
func (c *Config) ShouldForceNoCacheDocker() bool {
	return c.ShouldSkipLocalLayer()
}
