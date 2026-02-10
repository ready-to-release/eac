# pip

Pip isolation adapter providing safe, parallel-friendly Python virtual environments for test execution by copying source files to isolated work directories.

## Key Types

- **`PipIsolation`** -- Manages isolated Python venv and work directories
- **`IsolatedEnv`** -- Represents a prepared isolated environment with venv paths and env vars
- **`PipInstallMu`** -- Global mutex serializing pip install calls

## Key Functions

- `NewPipIsolation` -- Creates a new isolation instance for a workspace
- `VenvPipPath` -- Returns platform-specific path to pip binary in a venv

## Patterns

- Directory isolation: Copies source to `.cache/eac/python/work/{key}/`
- Virtual environment management: Creates venvs at `.cache/eac/python/venv/{key}/`
- Incremental sync: Only copies files changed by mtime and size
- Shared pip cache: All modules share `.cache/eac/python/pip-cache/` for downloads
- Serialized installs: `PipInstallMu` prevents concurrent pip cache contention
- Stale detection: Resets work directory when pyproject.toml or requirements change
- Cross-platform: Platform-aware venv binary paths (Scripts/ on Windows, bin/ on Unix)

## Internal Structure

| File         | Responsibility                                                                       |
| ------------ | ------------------------------------------------------------------------------------ |
| isolation.go | `PipIsolation`, `IsolatedEnv`, file sync, venv path resolution, and cache management |

## Dependencies

- `core/fileutil` -- `RemoveAllWithRetry` for Windows file lock handling
- `core/paths` -- Pip cache directory path resolution

## Role in System

The pip adapter solves Windows file lock conflicts and parallel test interference by providing isolated Python environments. Both the behave and pytest adapters depend on it to prepare safe work directories with virtual environments before running Python tests, ensuring that concurrent pip installs and test runs do not conflict.

## Code Health

### Tech Debt

- `PipInstallMu` in isolation.go is an exported package-level `sync.Mutex` used by both `behave` and `pytest` adapters; wrapping it behind a function or moving serialization into `PipIsolation` would encapsulate the locking strategy
- No unit tests exist for the package; `PrepareIsolatedEnv`, `syncDirectory`, `copyFileIfChanged`, and `requirementsChanged` are all untested

### Pain Points

- `requirementsChanged` in isolation.go uses mtime and size comparison, which can miss content changes when file size is unchanged; a content hash check would be more reliable
- `syncDirectory` in isolation.go:283 swallows `filepath.Walk` errors on the destination pass, making failures silent and hard to diagnose

### Optimization Opportunities

- `syncDirectory` performs two full `filepath.Walk` passes (one for destination tracking, one for source copying); building both maps in a single combined pass would reduce I/O
