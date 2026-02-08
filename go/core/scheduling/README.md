# scheduling

Pull-based work unit scheduling with dependency resolution and LPT
(Longest Processing Time First) ordering for parallel execution.

## Key Types

- **`WorkScheduler`** -- Interface for pull-based scheduling with deps
- **`DependencyScheduler`** -- Concrete scheduler with heap-based LPT ordering
- **`DependencyTracker`** -- Bidirectional dependency graph with cascade failure
- **`Stats`** -- Scheduling statistics (total, pending, ready, blocked)
- **`CircularDependencyError`** -- Reports cycles in the dependency graph
- **`DuplicateUnitIDError`** -- Reports duplicate unit IDs in work slice

## Patterns

- Pull-based scheduling: workers call `WaitForReady` to get next item
- LPT ordering: heaviest ready item is returned first via max-heap
- Implicit in-flight: items removed from queue until `MarkComplete`/`MarkFailed`
- Cascade failure: `MarkFailedCascade` proactively removes dependent items
- Lazy reverse map: `DependencyTracker.blocks` built on first cascade use

## Internal Structure

| File | Responsibility |
| --- | --- |
| doc.go | Package documentation and usage patterns |
| scheduler.go | `WorkScheduler` interface and `Stats` type |
| dependency_scheduler.go | `DependencyScheduler` implementation with heap and cond |
| dependency_tracker.go | `DependencyTracker` forward/reverse dependency maps |
| heap.go | `unitHeap` max-heap implementation for LPT |
| validation.go | Cycle detection (Kahn's algorithm) and duplicate checks |

## Dependencies

- `core/workunit` -- `UnitSpec` and `UnitID` types

## Role in System

This package sits between the command layer (which resolves what to build)
and the orchestrator (which manages concurrency). Commands create
`UnitSpec` slices with dependency edges, the scheduler topologically
orders them, and worker goroutines pull ready items while the scheduler
ensures dependencies are respected and failures cascade correctly.

## Code Health

### Tech Debt
- `scheduler.go`: `WorkScheduler` interface has 8 methods -- on the edge of interface bloat; `MarkFailed` vs `MarkFailedCascade` could be unified with an options parameter
- `dependency_tracker.go:92`: lazy reverse-map (`blocks`) is built on first cascade call; a subtle performance cliff for the first failure in large graphs

### Pain Points
- Only 2 test files for 6 source files; `dependency_tracker.go` and `heap.go` lack dedicated unit tests (tested indirectly through scheduler tests)

### Optimization Opportunities
- Add direct unit tests for `DependencyTracker` and `unitHeap` to improve fault localization when scheduling bugs arise (low effort, high debug-time savings)
- No TODO/FIXME markers found -- codebase is clean of deferred work items
