# workunit

Unified types for work unit identification, state management, locking,
and cache invalidation across build, test, lint, and scan commands.

## Key Types

- **`UnitID`** -- Uniquely identifies a unit of work (alias from contracts)
- **`UnitSpec`** -- Complete specification for executing a work unit
- **`UnitResult`** -- Outcome of a work unit execution
- **`UnitState`** -- Cached state for cache invalidation decisions
- **`StateManager`** -- Persists and queries unit state on disk
- **`InvalidationRule`** -- Defines when cached units should re-execute
- **`UoWAggregator`** -- Rolls up UoW-level caching to module level
- **`LockInfo`** -- Process-level exclusive lock on a work unit
- **`DisplayNameResolver`** -- Computes shortest unique display names
- **`TagSummary`** -- Classified tag data for test work units

## Patterns

- Type aliasing: `UnitID`, `UnitSpec`, `TagSummary` alias from contracts
- Rule-based invalidation: `DefaultRules` maps actions to invalidation logic
- Test set classification: L-tags drive unit vs integration invalidation
- File-based locking: `Lock`/`Unlock` with stale lock detection

## Internal Structure

| File | Responsibility |
| --- | --- |
| doc.go | Package documentation and naming conventions |
| unit_id.go | `UnitID` alias from contracts |
| unit_spec.go | `UnitSpec` alias and factory functions |
| unit_result.go | `UnitResult` with success/cached/failed helpers |
| unit_state.go | `UnitState`, `InvalidationRule`, test set classification |
| state_manager.go | `StateManager` for state persistence and change detection |
| aggregator.go | `UoWAggregator` for module-level cache rollup |
| lock.go | File-based locking with wait and stale detection |
| display.go | `DisplayNameResolver` alias from contracts |
| tags.go | `TagSummary` alias from contracts |
| action.go | `ActionType` constants re-exported from contracts |

## Dependencies

- `contracts/core` -- canonical `UnitID`, `UnitSpec`, `ActionType` definitions
- `core/cache` -- `Config` for skip-cache flags
- `core/paths` -- incremental cache directory paths

## Role in System

This package is the identity and state layer for all work units in the
`core` module. Every command (build, test, lint, scan) creates `UnitSpec`
values, uses `StateManager` for incremental caching, and reports results
via `UnitResult`. The `UoWAggregator` bridges per-unit caching with the
module-level granularity expected by the TUI and summary displays.

## Code Health

### Tech Debt
- `state_manager.go` (557 lines): `DetectTestModuleChanges` (lines 410-604, ~195 lines) is the largest function in the package -- complex branching for test-set-specific invalidation
- `unit_state.go:33,43`: mutable package-level `var DefaultRules` and `var IntegrationTestRule` maps could be overwritten at runtime; consider making them unexported or using functions

### Pain Points
- `state_manager.go` mixes UoW-level, module-level, and test-module-level change detection in one struct -- three different granularities with overlapping but distinct logic
- `unit_spec.go`: four `var` aliases (`NewBuildSpec`, `NewTestSpec`, `NewLintSpec`, `NewScanSpec`) re-export contract constructors; changes in contracts silently alter this package's API

### Optimization Opportunities
- Split `StateManager` into focused managers per granularity (UoW, module, test-module) to reduce file size and simplify testing (medium effort, high readability gain)
- No TODO/FIXME markers found -- codebase is clean of deferred work items
