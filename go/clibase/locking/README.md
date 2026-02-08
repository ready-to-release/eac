# locking

File-based distributed locking with lock tracking integration.
Provides cross-process mutual exclusion for operations that must not run concurrently.

## Key Types

- `Config` -- lock configuration: file path, timeout, poll interval, and lock tracker metadata
- `TrackedLock` -- an acquired lock handle with automatic tracker registration and release
- `WaitConfig` -- configuration for blocking lock acquisition with timeout

## Key Functions

- `Acquire` -- acquires a file lock with timeout, returning a `TrackedLock`
- `AcquireTracked` -- acquires a lock and registers it with the `locktracker.Registry`
- `AcquireWithWait` -- blocking acquisition that retries until timeout with configurable polling
- `BuildConfig` -- convenience config for build locks
- `TestConfig` -- convenience config for test locks
- `ScanConfig` -- convenience config for scan locks

## Patterns

- **Named lock configs**: convenience functions produce pre-configured `Config` values for common operations, ensuring consistent lock file paths
- **Tracked lifecycle**: locks automatically register with `locktracker.Registry` on acquisition and deregister on release
- **Timeout-based polling**: `AcquireWithWait` polls at configurable intervals until the lock is obtained or the timeout expires

## Internal Structure

| File | Purpose |
|---|---|
| `locking.go` | `Config`, `TrackedLock`, acquisition functions, and convenience configs |

## Dependencies

- `clibase/locktracker` -- lock event registration and tracking

## Role in System

Prevents data corruption when multiple CLI processes operate on shared resources (build outputs, test results, cache files). Commands acquire named locks before writing to shared directories, and the lock tracker makes contention visible in the TUI.

## Code Health

### Tech Debt
- `locking.go:142` `AcquireWithWait` is ~114 lines with deeply nested select/ticker logic; extracting the polling loop would aid readability
- Nine near-identical `*Config()` convenience functions (`BuildConfig`, `TestConfig`, etc.) could be collapsed into a data-driven factory

### Pain Points
- `locking.go:224` prints directly to `fmt.Printf` for "Waiting for lock" messages, bypassing structured logging and the display layer

### Optimization Opportunities
- Replace `fmt.Printf` in the wait loop with a writer parameter or log call to avoid stdout pollution in non-interactive contexts (low effort)
- Consolidate `*Config()` functions into a single `NewConfig(action, identifier, baseDir)` with action-specific defaults (low effort)
