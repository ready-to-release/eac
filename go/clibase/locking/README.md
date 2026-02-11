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
- `NewConfig` -- creates a module-level lock config for a given `Action`
- `NewUnitConfig` -- creates a component-level lock config for a given `Action`
- `NewFileConfig` -- creates a file-level lock config for a given `Action`

## Patterns

- **Action-driven config factory**: `NewConfig`, `NewUnitConfig`, and `NewFileConfig` accept an `Action` constant and produce pre-configured `Config` values, ensuring consistent lock file paths
- **Tracked lifecycle**: locks automatically register with `locktracker.Registry` on acquisition and deregister on release
- **Timeout-based polling**: `AcquireWithWait` polls at configurable intervals until the lock is obtained or the timeout expires
- **Writer-based messaging**: `WaitConfig.Writer` controls where "Waiting for lock" messages are sent (defaults to `os.Stderr`; set to `nil` to discard)

## Internal Structure

| File         | Purpose                                                                 |
| ------------ | ----------------------------------------------------------------------- |
| `locking.go` | `Config`, `TrackedLock`, acquisition functions, and convenience configs |

## Dependencies

- `clibase/locktracker` -- lock event registration and tracking

## Role in System

Prevents data corruption when multiple CLI processes operate on shared resources (build outputs, test results, cache files).

Commands acquire named locks before writing to shared directories, and the lock tracker makes contention visible in the TUI.

## Code Health

### Tech Debt

- `locking.go:142` `AcquireWithWait` is ~114 lines with deeply nested select/ticker logic; extracting the polling loop would aid readability

### Pain Points

- None identified

### Optimization Opportunities

- None identified
