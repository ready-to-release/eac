// Package execution provides the core domain for work unit execution ordering.
//
// # Three-Layer Execution Model
//
// The execution model organizes work into three nested layers:
//
//	Module Layer → Component Layer → UoW (Unit of Work)
//
// Each layer provides progressively finer-grained parallelism control:
//
//   - Module Layer: Inter-module dependency ordering. Modules in layer N must
//     complete before modules in layer N+1 can start. This respects build
//     dependencies where module A depends on module B's artifacts.
//
//   - Component Layer: Intra-module build_after ordering. Within a module,
//     components can specify build_after constraints (e.g., docs builds after go).
//     Components without dependencies run in parallel.
//
//   - UoW Layer: Explicit DependsOn constraints. Individual work units can
//     declare dependencies on other units via DependsOn field.
//
// # LayerMode
//
// The [LayerMode] type controls which layers are enforced:
//
//   - [LayerModeStrict]: Enforces all three layers. Use for build commands
//     where inter-module dependencies must be respected (module A needs
//     module B's binary before it can build).
//
//   - [LayerModeNone]: Skips module layer enforcement. Units from different
//     modules can run in parallel. Use for lint/scan/test commands that
//     don't depend on other modules' build artifacts.
//
// # LayerPolicy Interface
//
// The [LayerPolicy] interface defines the domain's contract for execution
// scheduling. It provides three core operations:
//
//   - ComputePlan: Organizes work units into a [LayerPlan] with nested layers
//   - IsReady: Checks if a specific unit can execute given current completions
//   - GetReadyUnits: Returns all units that can execute now (batch scheduling)
//
// The orchestrator (adapter layer) queries LayerPolicy to determine what to
// execute and when. This keeps scheduling logic in the domain, not the adapter.
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
//  2. Pass specs to orchestrator which uses LayerPolicy for scheduling
//  3. Optionally provide a CacheVerifier for incremental execution
//
// The domain (this package) owns the "what" and "when" of execution.
// The orchestrator owns the "how" (parallelism, TUI, logging).
package execution
