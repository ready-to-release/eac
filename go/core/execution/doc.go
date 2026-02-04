// Package execution provides the core domain for work unit execution.
//
// # Dependency-Based Execution
//
// Work units are scheduled based on their explicit dependencies (DependsOn).
// A unit is ready to execute when all its dependencies have completed.
// This provides simple, predictable scheduling without layer abstractions.
//
// # CacheVerifier Interface
//
// The [CacheVerifier] interface provides cache status verification for work units.
// Commands implement this interface to define their own caching logic:
//
//   - Build: Verifies source hash + artifact integrity
//   - Test: Checks source hash + build artifact hash
//   - Lint/Scan: Checks source hash only
//
// The orchestrator uses CacheVerifier for background cache detection, allowing
// TUI tabs to progressively "light up" as cache hits are detected.
//
// Example implementation:
//
//	type BuildCacheVerifier struct {
//	    workspaceRoot string
//	    cachedModules map[string]bool
//	    cacheTimes    map[string]time.Time
//	}
//
//	func (v *BuildCacheVerifier) Verify(ctx context.Context, unit workunit.UnitSpec) (CacheResult, error) {
//	    select {
//	    case <-ctx.Done():
//	        return CacheResult{}, ctx.Err()
//	    default:
//	    }
//	    // Check if module is in pre-computed cached set
//	    if v.cachedModules[unit.ID.Module] {
//	        return CacheResult{Cached: true, CacheTime: v.cacheTimes[unit.ID.Module]}, nil
//	    }
//	    return CacheResult{}, nil
//	}
//
// # Usage
//
// Commands typically use this package through the orchestrator adapter:
//
//  1. Create UnitSpecs via command-specific ResolveUnitSpecs functions
//  2. Pass specs to orchestrator which schedules based on DependsOn
//  3. Optionally provide a CacheVerifier for incremental execution
//
// The domain (this package) owns the "what" of execution.
// The orchestrator owns the "how" (parallelism, TUI, logging).
package execution
