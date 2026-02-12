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

// ShouldSkipCI returns true if CI cache should be skipped (forces dispatch).
func (c *Config) ShouldSkipCI() bool {
	return c.ShouldSkip(LevelRemote, TypeCI)
}

// ShouldForcePull returns true when Docker should pull fresh images instead of using
// locally cached registry layers. This is the Docker-specific alias for ShouldSkipLocalRegistry,
// translating the abstract cache concept (skip local:registry) into the concrete Docker
// flag (--pull). Callers building Docker commands should use this method for clarity.
func (c *Config) ShouldForcePull() bool {
	return c.ShouldSkipLocalRegistry()
}

// ShouldForceNoCacheDocker returns true when Docker build should use --no-cache,
// discarding all locally cached build layers. This is the Docker-specific alias for
// ShouldSkipLocalLayer, translating the abstract cache concept (skip local:layer) into the
// concrete Docker flag (--no-cache). Callers building Docker commands should use this method
// for clarity.
func (c *Config) ShouldForceNoCacheDocker() bool {
	return c.ShouldSkipLocalLayer()
}
