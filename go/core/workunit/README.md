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
| state_manager.go | `StateManager` for state persistence and core change detection |
| uow_state.go | UoW-level state management and change detection |
| module_state.go | Module-level state management and change detection |
| test_state.go | Test-module-level state management with test-set-specific invalidation |
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
- `test_state.go` is 342 lines
- `uow_state.go` is 245 lines
- `lock.go` is 229 lines
- `module_state.go` is 140 lines
- `unit_state.go` is 126 lines with mutable package-level `var DefaultRules` and `var IntegrationTestRule` maps that could be overwritten at runtime

### Pain Points
- `unit_spec.go` contains four `var` aliases (`NewBuildSpec`, `NewTestSpec`, `NewLintSpec`, `NewScanSpec`) that re-export contract constructors; changes in contracts silently alter this package's API

### Optimization Opportunities
- Consider making `DefaultRules` and `IntegrationTestRule` unexported or using functions to prevent runtime mutation
