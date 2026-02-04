// Package build provides the build command implementation using cmdframework.
package build

import (
	"context"
	"time"

	"github.com/ready-to-release/eac/go/core/execution"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// UoWBuildCacheVerifier implements execution.CacheVerifier for UoW-based caching.
// It verifies cache status at the component:tool level instead of module level.
type UoWBuildCacheVerifier struct {
	workspaceRoot string
	cachedUoWs    map[string]bool      // UoW longname -> cached
	uowCacheTimes map[string]time.Time // UoW longname -> cache time
	cachedModules map[string]bool      // Module-level cache (aggregated from UoWs for TUI)
}

// NewUoWBuildCacheVerifier creates a new UoWBuildCacheVerifier.
func NewUoWBuildCacheVerifier(
	workspaceRoot string,
	cachedUoWs map[string]bool,
	uowCacheTimes map[string]time.Time,
	cachedModules map[string]bool,
) *UoWBuildCacheVerifier {
	return &UoWBuildCacheVerifier{
		workspaceRoot: workspaceRoot,
		cachedUoWs:    cachedUoWs,
		uowCacheTimes: uowCacheTimes,
		cachedModules: cachedModules,
	}
}

// Verify implements execution.CacheVerifier for UoW-based caching.
// It checks if a specific component:tool UoW is cached and valid.
func (v *UoWBuildCacheVerifier) Verify(ctx context.Context, unit workunit.UnitSpec) (execution.CacheResult, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return execution.CacheResult{}, ctx.Err()
	default:
	}

	longname := unit.ID.Longname()

	// Check if UoW is in cached set
	if v.cachedUoWs == nil || !v.cachedUoWs[longname] {
		// Fall back to module-level check for backwards compatibility
		// (module cache is aggregated from UoWs but used for TUI display)
		if v.cachedModules != nil && v.cachedModules[unit.ID.Module] {
			// Module is cached, but this specific UoW wasn't tracked
			// Trust module-level cache
			return execution.CacheResult{
				Cached:    true,
				CacheTime: time.Time{},
			}, nil
		}
		return execution.CacheResult{Cached: false}, nil
	}

	// Verify UoW artifact integrity
	reader := coreoutput.NewReader(v.workspaceRoot)
	validationResult := reader.ValidateUoW(
		workunit.ContextBuild,
		unit.ID.Module,
		unit.ID.Component,
		unit.ID.Tool,
	)

	if !validationResult.ManifestExists {
		// No manifest - trust the hash check that already passed
		return execution.CacheResult{
			Cached:    true,
			CacheTime: v.uowCacheTimes[longname],
		}, nil
	}

	if !validationResult.Valid {
		// Artifacts invalid - not cached
		return execution.CacheResult{Cached: false}, nil
	}

	// Confirmed cache hit
	return execution.CacheResult{
		Cached:    true,
		CacheTime: v.uowCacheTimes[longname],
	}, nil
}
