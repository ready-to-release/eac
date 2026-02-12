# locktracker

Lock tracking and visualization registry. Provides observability into active locks,
semaphores, and capacity allocations across the system.

## Key Types

- `Registry` -- thread-safe registry that tracks active locks with pub/sub event notification; global singleton via `sync.Once`
- `TrackedSemaphore` -- channel-based counting semaphore with automatic lock tracking registration
- `LockType` -- categorizes locks: file lock, semaphore, capacity slot
- `LockInfo` -- describes an active lock: type, name, holder, acquired time, metadata
- `LockEvent` -- represents a lock lifecycle event (acquired, released, waiting)
- `EventType` -- enumeration of lock event types
- `LockSummary` -- aggregated snapshot of all active locks grouped by type

## Patterns

- **Pub/sub events**: `Registry` publishes `LockEvent` values to subscribers, enabling real-time lock visualization in the TUI
- **Automatic registration**: `TrackedSemaphore` acquires and releases are automatically registered with the `Registry`
- **Thread-safe snapshots**: `Summary()` and `Snapshot()` return consistent point-in-time views of all active locks
- **Global singleton**: `Get()` returns the shared `Registry` instance, initialized once via `sync.Once`

## Internal Structure

| File | Purpose |
|---|---|
| `types.go` | Type definitions: `LockType`, `EventType`, `LockInfo`, `LockEvent`, `LockSummary` |
| `registry.go` | `Registry` with add/remove, subscribe/unsubscribe, summary, and snapshot operations |
| `semaphore.go` | `TrackedSemaphore` wrapping channel-based semaphore with registry integration |

## Dependencies

None (leaf package within clibase).

## Role in System

Provides the observability layer for concurrent resource management. The capacity package and locking package register their locks here, and the TUI subscribes to events for real-time display of what the system is waiting on during parallel execution.

## Code Health

### Tech Debt

- None identified

### Pain Points

- Test files are significantly larger than implementation: `registry_test.go` (979 lines), `semaphore_test.go` (924 lines), `types_test.go` (477 lines)
- Implementation files are appropriately sized: `registry.go` (183 lines), `semaphore.go` (138 lines), `types.go` (55 lines)

### Optimization Opportunities

- None identified
