# pip

Pip isolation adapter providing safe, parallel-friendly Python virtual environments for test execution by copying source files to isolated work directories.

## Key Types

- **`PipIsolation`** -- Manages isolated Python venv and work directories
- **`IsolatedEnv`** -- Represents a prepared isolated environment with venv paths and env vars
- **`WithInstallLock`** -- Serializes pip install calls via unexported mutex

## Key Functions

- `NewPipIsolation` -- Creates a new isolation instance for a workspace
- `VenvPipPath` -- Returns platform-specific path to pip binary in a venv

## Patterns

- Directory isolation: Copies source to `.cache/eac/python/work/{key}/`
- Virtual environment management: Creates venvs at `.cache/eac/python/venv/{key}/`
- Incremental sync: Only copies files changed by mtime and size
- Shared pip cache: All modules share `.cache/eac/python/pip-cache/` for downloads
- Serialized installs: `WithInstallLock()` prevents concurrent pip cache contention
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
- None identified

### Pain Points
- isolation.go is 375 lines; candidate for splitting (extract venv management, file sync logic, and change detection into separate files)
- requirementsChanged in isolation.go uses mtime and size comparison, which can miss content changes when file size is unchanged; a content hash check would be more reliable

### Optimization Opportunities
- None identified
