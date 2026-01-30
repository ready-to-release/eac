package build

import (
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/workunit"
	"github.com/stretchr/testify/assert"
)

func TestVerifyComponentCache_NotInCachedSet(t *testing.T) {
	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Module:    "test-module",
			Component: "go",
			Tool:      "go",
		},
	}

	// Module not in cached set
	result := VerifyComponentCache("/tmp/workspace", spec, nil, nil)

	assert.False(t, result.IsCached)
	assert.Equal(t, "test-module:go:go", result.Moniker)
	assert.Equal(t, "test-module", result.Module)
	assert.Equal(t, "go", result.Component)
	assert.Equal(t, "go", result.Handler)
}

func TestVerifyComponentCache_EmptyCachedModules(t *testing.T) {
	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Module:    "test-module",
			Component: "go",
			Tool:      "go",
		},
	}

	// Empty cached modules map
	result := VerifyComponentCache("/tmp/workspace", spec, map[string]bool{}, nil)

	assert.False(t, result.IsCached)
}

func TestVerifyComponentCache_ModuleNotCached(t *testing.T) {
	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Module:    "test-module",
			Component: "go",
			Tool:      "go",
		},
	}

	// Different module is cached, not this one
	cachedModules := map[string]bool{
		"other-module": true,
	}

	result := VerifyComponentCache("/tmp/workspace", spec, cachedModules, nil)

	assert.False(t, result.IsCached)
}

func TestVerifyComponentCache_CachedNoManifest(t *testing.T) {
	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Module:    "test-module",
			Component: "go",
			Tool:      "go",
		},
	}

	cachedModules := map[string]bool{
		"test-module": true,
	}
	cacheTime := time.Now().Add(-1 * time.Hour)
	cacheTimes := map[string]time.Time{
		"test-module": cacheTime,
	}

	// Non-existent workspace path means no manifest
	result := VerifyComponentCache("/nonexistent/workspace", spec, cachedModules, cacheTimes)

	// Should be cached (trust source hash when no manifest)
	assert.True(t, result.IsCached)
	assert.Equal(t, cacheTime, result.CacheTime)
	assert.Equal(t, "test-module:go:go", result.Moniker)
}

func TestVerifyComponentCache_MonikerFormatWithTool(t *testing.T) {
	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Module:    "eac-commands",
			Component: "docs",
			Tool:      "mkdocs",
		},
	}

	result := VerifyComponentCache("/tmp/workspace", spec, nil, nil)

	assert.Equal(t, "eac-commands:docs:mkdocs", result.Moniker)
}

func TestVerifyComponentCache_MonikerFormatWithoutTool(t *testing.T) {
	spec := workunit.UnitSpec{
		ID: workunit.UnitID{
			Module:    "eac-commands",
			Component: "docs",
			Tool:      "",
		},
	}

	result := VerifyComponentCache("/tmp/workspace", spec, nil, nil)

	assert.Equal(t, "eac-commands:docs", result.Moniker)
}
