package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Config Creation Tests
// =============================================================================

func TestNewConfig_ReturnsEmptyConfig(t *testing.T) {
	cfg := NewConfig()

	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.SkipSpecs)
}

// =============================================================================
// Config.Skip() Tests
// =============================================================================

func TestConfig_Skip_AddsSpec(t *testing.T) {
	cfg := NewConfig()

	cfg.Skip(Spec{Level: LevelLocal, Type: TypeState})

	assert.Len(t, cfg.SkipSpecs, 1)
	assert.Equal(t, LevelLocal, cfg.SkipSpecs[0].Level)
	assert.Equal(t, TypeState, cfg.SkipSpecs[0].Type)
}

func TestConfig_Skip_AddsMultipleSpecs(t *testing.T) {
	cfg := NewConfig()

	cfg.Skip(Spec{Level: LevelLocal, Type: TypeState})
	cfg.Skip(Spec{Level: LevelLocal, Type: TypeAsset})
	cfg.Skip(Spec{Level: LevelRemote, Type: TypeLayer})

	assert.Len(t, cfg.SkipSpecs, 3)
}

// =============================================================================
// Config.SkipAll() Tests
// =============================================================================

func TestConfig_SkipAll_AddsAllWildcard(t *testing.T) {
	cfg := NewConfig()

	cfg.SkipAll()

	assert.Len(t, cfg.SkipSpecs, 1)
	assert.Equal(t, LevelAll, cfg.SkipSpecs[0].Level)
	assert.Equal(t, TypeAll, cfg.SkipSpecs[0].Type)
}

// =============================================================================
// Config.ShouldSkip() Tests
// =============================================================================

func TestConfig_ShouldSkip_NilConfig(t *testing.T) {
	var cfg *Config

	// Should not panic, should return false
	assert.False(t, cfg.ShouldSkip(LevelLocal, TypeState))
}

func TestConfig_ShouldSkip_EmptyConfig(t *testing.T) {
	cfg := NewConfig()

	assert.False(t, cfg.ShouldSkip(LevelLocal, TypeState))
	assert.False(t, cfg.ShouldSkip(LevelRemote, TypeLayer))
}

func TestConfig_ShouldSkip_ExactMatch(t *testing.T) {
	cfg := NewConfig()
	cfg.Skip(Spec{Level: LevelLocal, Type: TypeState})

	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeState))
	assert.False(t, cfg.ShouldSkip(LevelLocal, TypeAsset))
	assert.False(t, cfg.ShouldSkip(LevelRemote, TypeState))
}

func TestConfig_ShouldSkip_LevelWildcard(t *testing.T) {
	cfg := NewConfig()
	cfg.Skip(Spec{Level: LevelAll, Type: TypeState})

	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeState))
	assert.True(t, cfg.ShouldSkip(LevelRemote, TypeState))
	assert.False(t, cfg.ShouldSkip(LevelLocal, TypeAsset))
}

func TestConfig_ShouldSkip_TypeWildcard(t *testing.T) {
	cfg := NewConfig()
	cfg.Skip(Spec{Level: LevelLocal, Type: TypeAll})

	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeState))
	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeAsset))
	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeLayer))
	assert.False(t, cfg.ShouldSkip(LevelRemote, TypeState))
}

func TestConfig_ShouldSkip_AllWildcard(t *testing.T) {
	cfg := NewConfig()
	cfg.SkipAll()

	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeState))
	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeAsset))
	assert.True(t, cfg.ShouldSkip(LevelRemote, TypeLayer))
	assert.True(t, cfg.ShouldSkip(LevelRemote, TypeRegistry))
}

func TestConfig_ShouldSkip_MultipleSpecs(t *testing.T) {
	cfg := NewConfig()
	cfg.Skip(Spec{Level: LevelLocal, Type: TypeState})
	cfg.Skip(Spec{Level: LevelLocal, Type: TypeAsset})

	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeState))
	assert.True(t, cfg.ShouldSkip(LevelLocal, TypeAsset))
	assert.False(t, cfg.ShouldSkip(LevelLocal, TypeRegistry))
	assert.False(t, cfg.ShouldSkip(LevelRemote, TypeState))
}

// =============================================================================
// Convenience Method Tests
// =============================================================================

func TestConfig_ShouldSkipState(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Config)
		expected bool
	}{
		{
			name:     "nil config",
			setup:    nil,
			expected: false,
		},
		{
			name:     "empty config",
			setup:    func(c *Config) {},
			expected: false,
		},
		{
			name: "local:state skipped",
			setup: func(c *Config) {
				c.Skip(Spec{Level: LevelLocal, Type: TypeState})
			},
			expected: true,
		},
		{
			name: "all:state skipped",
			setup: func(c *Config) {
				c.Skip(Spec{Level: LevelAll, Type: TypeState})
			},
			expected: true,
		},
		{
			name: "local:all skipped",
			setup: func(c *Config) {
				c.Skip(Spec{Level: LevelLocal, Type: TypeAll})
			},
			expected: true,
		},
		{
			name: "all skipped",
			setup: func(c *Config) {
				c.SkipAll()
			},
			expected: true,
		},
		{
			name: "only local:asset skipped",
			setup: func(c *Config) {
				c.Skip(Spec{Level: LevelLocal, Type: TypeAsset})
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *Config
			if tt.setup != nil {
				cfg = NewConfig()
				tt.setup(cfg)
			}
			assert.Equal(t, tt.expected, cfg.ShouldSkipState())
		})
	}
}

func TestConfig_ShouldSkipLocalRegistry(t *testing.T) {
	cfg := NewConfig()
	assert.False(t, cfg.ShouldSkipLocalRegistry())

	cfg.Skip(Spec{Level: LevelLocal, Type: TypeRegistry})
	assert.True(t, cfg.ShouldSkipLocalRegistry())
}

func TestConfig_ShouldSkipLocalLayer(t *testing.T) {
	cfg := NewConfig()
	assert.False(t, cfg.ShouldSkipLocalLayer())

	cfg.Skip(Spec{Level: LevelLocal, Type: TypeLayer})
	assert.True(t, cfg.ShouldSkipLocalLayer())
}

func TestConfig_ShouldSkipAsset(t *testing.T) {
	cfg := NewConfig()
	assert.False(t, cfg.ShouldSkipAsset())

	cfg.Skip(Spec{Level: LevelLocal, Type: TypeAsset})
	assert.True(t, cfg.ShouldSkipAsset())
}

func TestConfig_ShouldSkipWork(t *testing.T) {
	cfg := NewConfig()
	assert.False(t, cfg.ShouldSkipWork())

	cfg.Skip(Spec{Level: LevelLocal, Type: TypeWork})
	assert.True(t, cfg.ShouldSkipWork())
}

func TestConfig_ShouldSkipRemoteLayer(t *testing.T) {
	cfg := NewConfig()
	assert.False(t, cfg.ShouldSkipRemoteLayer())

	cfg.Skip(Spec{Level: LevelRemote, Type: TypeLayer})
	assert.True(t, cfg.ShouldSkipRemoteLayer())
}

func TestConfig_ShouldForcePull(t *testing.T) {
	cfg := NewConfig()
	assert.False(t, cfg.ShouldForcePull())

	// ShouldForcePull is triggered when local:registry is skipped
	cfg.Skip(Spec{Level: LevelLocal, Type: TypeRegistry})
	assert.True(t, cfg.ShouldForcePull())
}

func TestConfig_ShouldForceNoCacheDocker(t *testing.T) {
	cfg := NewConfig()
	assert.False(t, cfg.ShouldForceNoCacheDocker())

	// ShouldForceNoCacheDocker is triggered when local:layer is skipped
	cfg.Skip(Spec{Level: LevelLocal, Type: TypeLayer})
	assert.True(t, cfg.ShouldForceNoCacheDocker())
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestConfig_RealWorldScenarios(t *testing.T) {
	t.Run("legacy --skip-cache (no value) should skip state only", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Skip(Spec{Level: LevelAll, Type: TypeState})

		assert.True(t, cfg.ShouldSkipState(), "should skip state")
		assert.False(t, cfg.ShouldSkipAsset(), "should NOT skip asset")
		assert.False(t, cfg.ShouldSkipLocalRegistry(), "should NOT skip registry")
		assert.False(t, cfg.ShouldSkipLocalLayer(), "should NOT skip layer")
		assert.False(t, cfg.ShouldSkipWork(), "should NOT skip work")
	})

	t.Run("--skip-cache=local should skip all local caches", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Skip(Spec{Level: LevelLocal, Type: TypeAll})

		assert.True(t, cfg.ShouldSkipState())
		assert.True(t, cfg.ShouldSkipAsset())
		assert.True(t, cfg.ShouldSkipLocalRegistry())
		assert.True(t, cfg.ShouldSkipLocalLayer())
		assert.True(t, cfg.ShouldSkipWork())
		assert.False(t, cfg.ShouldSkipRemoteLayer(), "should NOT skip remote layer")
	})

	t.Run("--skip-cache=all should skip everything", func(t *testing.T) {
		cfg := NewConfig()
		cfg.SkipAll()

		assert.True(t, cfg.ShouldSkipState())
		assert.True(t, cfg.ShouldSkipAsset())
		assert.True(t, cfg.ShouldSkipLocalRegistry())
		assert.True(t, cfg.ShouldSkipLocalLayer())
		assert.True(t, cfg.ShouldSkipWork())
		assert.True(t, cfg.ShouldSkipRemoteLayer())
	})

	t.Run("--skip-cache=local:registry,local:layer for Docker rebuild", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Skip(Spec{Level: LevelLocal, Type: TypeRegistry})
		cfg.Skip(Spec{Level: LevelLocal, Type: TypeLayer})

		assert.False(t, cfg.ShouldSkipState(), "should NOT skip state")
		assert.False(t, cfg.ShouldSkipAsset(), "should NOT skip asset")
		assert.True(t, cfg.ShouldSkipLocalRegistry())
		assert.True(t, cfg.ShouldSkipLocalLayer())
		assert.True(t, cfg.ShouldForcePull())
		assert.True(t, cfg.ShouldForceNoCacheDocker())
	})
}
