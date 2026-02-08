package output

import (
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// AggregatedChangeResult extends UoWChangeResult with module-level aggregation.
// It bridges UoW-level change detection with module-level TUI display.
type AggregatedChangeResult struct {
	// UoWResult contains the underlying UoW-level change detection results.
	UoWResult *UoWChangeResult

	// CachedUoWs maps UoW longname to whether it's cached.
	CachedUoWs map[string]bool
	// UoWCacheTimes maps UoW longname to when it was last executed.
	UoWCacheTimes map[string]time.Time

	// CachedModules maps module name to whether ALL its UoWs are cached.
	CachedModules map[string]bool
	// ModuleCacheTimes maps module name to its earliest UoW cache time.
	ModuleCacheTimes map[string]time.Time

	// ChangedModules contains modules that need (re)execution.
	ChangedModules []string
	// UpToDateModules contains modules that are fully cached.
	UpToDateModules []string
}

// AggregateUoWChanges performs UoW-level change detection and aggregates to module level.
// A module is considered cached only if ALL its UoWs are cached.
//
// This helper combines:
// 1. UoW-level change detection via DetectUoWChanges
// 2. Module-level aggregation via UoWAggregator
// 3. Cache time collection for TUI display
func AggregateUoWChanges(
	reader *DiskOutputReader,
	ctx core.ActionType,
	expectedUoWs []workunit.UnitID,
	getInputHash InputHashProvider,
) (*AggregatedChangeResult, error) {
	result := &AggregatedChangeResult{
		CachedUoWs:       make(map[string]bool),
		UoWCacheTimes:    make(map[string]time.Time),
		CachedModules:    make(map[string]bool),
		ModuleCacheTimes: make(map[string]time.Time),
	}

	if len(expectedUoWs) == 0 {
		result.UoWResult = &UoWChangeResult{
			Changed:       []workunit.UnitID{},
			UpToDate:      []workunit.UnitID{},
			ChangeReasons: make(map[string]string),
		}
		return result, nil
	}

	// Perform UoW-level change detection
	// nil depResolver: AggregateUoWChanges is used for test/lint/scan which
	// already have same-module build invalidation. Cross-module dependency
	// invalidation is only needed for build UoWs (handled by build command).
	uowResult, err := reader.DetectUoWChanges(ctx, expectedUoWs, getInputHash, nil)
	if err != nil {
		return nil, err
	}
	result.UoWResult = uowResult

	// Handle fresh run - no aggregation needed
	if uowResult.FreshRun {
		return result, nil
	}

	// Use aggregator for module-level rollup
	agg := workunit.NewUoWAggregator(expectedUoWs)

	// Process up-to-date UoWs
	for _, id := range uowResult.UpToDate {
		longname := id.Longname()
		result.CachedUoWs[longname] = true
		agg.MarkCached(id)

		// Load cache time from manifest
		if manifest, err := reader.GetUoW(id); err == nil {
			result.UoWCacheTimes[longname] = manifest.ExecutedAt
		}
	}

	// Aggregate to module level
	result.ChangedModules, result.UpToDateModules = agg.GetModuleLists(expectedUoWs)

	// Mark cached modules and compute cache times
	for _, module := range result.UpToDateModules {
		result.CachedModules[module] = true

		// Set earliest cache time from any UoW in the module
		for _, id := range expectedUoWs {
			if id.Module == module {
				if t, ok := result.UoWCacheTimes[id.Longname()]; ok {
					if existing, exists := result.ModuleCacheTimes[module]; !exists || t.Before(existing) {
						result.ModuleCacheTimes[module] = t
					}
				}
			}
		}
	}

	return result, nil
}
