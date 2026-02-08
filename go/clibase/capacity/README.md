# capacity

Cross-process capacity management using weighted semaphores.
Controls concurrent resource usage across parallel command executions within a workspace.

## Key Types

- `GlobalSemaphore` -- cross-process weighted semaphore backed by file-based locking; limits total concurrency across host and Docker pools
- `DualPoolSemaphore` -- coordinates separate host and Docker capacity pools with independent limits and shared tracking
- `State` -- snapshot of current semaphore state: total capacity, available slots, active allocations
- `Allocation` -- represents an acquired capacity slot with weight, pool type, and release callback

## Patterns

- **Weighted acquisition**: slots have configurable weights, allowing heavy operations (e.g., Docker builds) to consume more capacity than lightweight ones
- **Dual-pool separation**: host and Docker workloads use independent pools to prevent one type from starving the other
- **File-based coordination**: uses file locks for cross-process synchronization, enabling capacity limits across multiple CLI invocations
- **Tracked allocations**: integrates with `locktracker.Registry` for visibility into active capacity usage

## Internal Structure

| File | Purpose |
|---|---|
| `global.go` | `GlobalSemaphore` with weighted acquisition, release, and state reporting |
| `dual_pool.go` | `DualPoolSemaphore` coordinating independent host and Docker pools |

## Dependencies

- `clibase/locktracker` -- lock tracking and event publication
- `core/resource` -- resource type definitions
- `core/logging` -- structured logging

## Role in System

Prevents resource exhaustion when multiple commands run concurrently in the same workspace. The orchestrator acquires capacity slots before dispatching work units, ensuring total parallelism stays within configured limits regardless of how many CLI processes are active.

## Code Health

### Tech Debt
- `global.go:21` package-level mutable `log` var; prefer injecting the logger through constructors
- `global.go:107` signal handler goroutine has no shutdown channel -- relies on process exit rather than graceful cleanup

### Pain Points
- `global_test.go` (128 lines) has limited coverage relative to `global.go` (401 lines); file-lock and stale-process-cleanup paths are under-tested
- Cross-process coordination via file locks is inherently platform-sensitive; no Windows-specific integration tests exist

### Optimization Opportunities
- Add a test that exercises `cleanStaleFromState` with simulated dead-PID entries (low effort)
- Consider replacing busy-wait in `Acquire` with OS-level file-lock notification to reduce CPU spin (high effort, platform-dependent)
