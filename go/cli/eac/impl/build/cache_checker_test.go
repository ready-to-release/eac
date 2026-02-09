package build

import (
	"context"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
)

func TestUoWBuildCacheVerifier_NotInCachedSet(t *testing.T) {
	verifier := NewUoWBuildCacheVerifier("/tmp/workspace", nil, nil, nil)

	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Action:        core.ActionBuild,
			Module:        "test-module",
			ComponentType: "go",
			ComponentName: "go",
			Tool:          "go",
		},
	}

	result, err := verifier.Verify(context.Background(), spec)
	assert.NoError(t, err)
	assert.False(t, result.Cached)
}

func TestUoWBuildCacheVerifier_EmptyMaps(t *testing.T) {
	verifier := NewUoWBuildCacheVerifier("/tmp/workspace", map[string]bool{}, map[string]time.Time{}, map[string]bool{})

	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Action:        core.ActionBuild,
			Module:        "test-module",
			ComponentType: "go",
			ComponentName: "go",
			Tool:          "go",
		},
	}

	result, err := verifier.Verify(context.Background(), spec)
	assert.NoError(t, err)
	assert.False(t, result.Cached)
}

func TestUoWBuildCacheVerifier_UoWNotCached(t *testing.T) {
	cachedUoWs := map[string]bool{
		"build:other-module:go:go": true,
	}

	verifier := NewUoWBuildCacheVerifier("/tmp/workspace", cachedUoWs, nil, nil)

	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Action:        core.ActionBuild,
			Module:        "test-module",
			ComponentType: "go",
			ComponentName: "go",
			Tool:          "go",
		},
	}

	result, err := verifier.Verify(context.Background(), spec)
	assert.NoError(t, err)
	assert.False(t, result.Cached)
}

func TestUoWBuildCacheVerifier_FallbackToModuleCache(t *testing.T) {
	// UoW not in cache, but module is
	cachedModules := map[string]bool{
		"test-module": true,
	}

	verifier := NewUoWBuildCacheVerifier("/tmp/workspace", nil, nil, cachedModules)

	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Action:        core.ActionBuild,
			Module:        "test-module",
			ComponentType: "go",
			ComponentName: "go",
			Tool:          "go",
		},
	}

	result, err := verifier.Verify(context.Background(), spec)
	assert.NoError(t, err)
	assert.True(t, result.Cached) // Falls back to module-level cache
}

func TestUoWBuildCacheVerifier_CachedNoManifest(t *testing.T) {
	cacheTime := time.Now().Add(-1 * time.Hour)
	cachedUoWs := map[string]bool{
		"build:test-module:go:go:go": true,
	}
	uowCacheTimes := map[string]time.Time{
		"build:test-module:go:go:go": cacheTime,
	}

	// Non-existent workspace path means no manifest
	verifier := NewUoWBuildCacheVerifier("/nonexistent/workspace", cachedUoWs, uowCacheTimes, nil)

	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Action:        core.ActionBuild,
			Module:        "test-module",
			ComponentType: "go",
			ComponentName: "go",
			Tool:          "go",
		},
	}

	result, err := verifier.Verify(context.Background(), spec)
	assert.NoError(t, err)
	// Should be cached (trust source hash when no manifest)
	assert.True(t, result.Cached)
	assert.Equal(t, cacheTime, result.CacheTime)
}

func TestUoWBuildCacheVerifier_ContextCancelled(t *testing.T) {
	verifier := NewUoWBuildCacheVerifier("/tmp/workspace", nil, nil, nil)

	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Action:        core.ActionBuild,
			Module:        "test-module",
			ComponentType: "go",
			ComponentName: "go",
			Tool:          "go",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := verifier.Verify(ctx, spec)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
