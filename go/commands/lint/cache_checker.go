package lint

import (
	"context"
	"time"

	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// LintCacheVerifier implements execution.CacheVerifier for lint commands.
// Uses UoW-level cache for fine-grained lint caching at component:provider granularity.
type LintCacheVerifier struct {
	cachedUoWs    map[string]bool      // UoW longname -> cached
	uowCacheTimes map[string]time.Time // UoW longname -> cache time
	cachedModules map[string]bool      // Module-level cache (aggregated from UoWs for TUI)
}

// Verify implements execution.CacheVerifier.
func (v *LintCacheVerifier) Verify(ctx context.Context, unit workunit.UnitSpec) (execution.CacheResult, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return execution.CacheResult{}, ctx.Err()
	default:
	}

	longname := unit.ID.Longname()

	// Check UoW-level cache first
	if v.cachedUoWs != nil && v.cachedUoWs[longname] {
		return execution.CacheResult{
			Cached:    true,
			CacheTime: v.uowCacheTimes[longname],
		}, nil
	}

	// Fall back to module-level check for backwards compatibility
	if v.cachedModules != nil && v.cachedModules[unit.ID.Module] {
		return execution.CacheResult{Cached: true}, nil
	}

	return execution.CacheResult{}, nil
}
