# npm

NPM isolation adapter providing safe, parallel-friendly npm environments
for TypeScript test execution by copying source files to isolated work
directories.

## Key Types

- **`NpmIsolation`** -- Manages isolated npm work directories
- **`IsolatedEnv`** -- Represents a prepared isolated environment
- **`NpmInstallMu`** -- Global mutex serializing npm install calls

## Patterns

- Directory isolation: Copies source to `.cache/eac/npm/work/{moniker}/`
- Incremental sync: Only copies files changed by mtime and size
- Shared npm cache: All modules share `.cache/eac/npm/cache/` for downloads
- Serialized installs: `NpmInstallMu` prevents concurrent npm cache contention
- Stale detection: Resets work directory when package.json changes

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| isolation.go | `NpmIsolation`, `IsolatedEnv`, file sync, and cache management |

## Dependencies

- `core/fileutil` -- `RemoveAllWithRetry` for Windows file lock handling
- `core/paths` -- npm cache directory path resolution

## Role in System

The `npm-eac` module solves Windows EPERM errors and parallel test
interference by providing isolated npm environments. Both the
`cucumber-eac` and `mocha-eac` depend on it to prepare safe work
directories before running TypeScript tests, ensuring that concurrent
npm installs and test runs do not conflict.

## Code Health

### Tech Debt
- `NpmInstallMu` in isolation.go is an exported package-level `sync.Mutex` used by both `cucumber` and `mocha` adapters; wrapping it behind a function or moving serialization into `NpmIsolation` would encapsulate the locking strategy
- No unit tests exist for the package; `PrepareIsolatedEnv`, `syncDirectory`, `copyFileIfChanged`, and `packageFilesChanged` are all untested

### Pain Points
- `packageFilesChanged` in isolation.go uses mtime and size comparison, which can miss content changes when file size is unchanged; a content hash check would be more reliable
- `syncDirectory` in isolation.go swallows `filepath.Walk` errors on both source and destination (lines 218-220, 229-230), making failures silent and hard to diagnose

### Optimization Opportunities
- `syncDirectory` performs two full `filepath.Walk` passes (one for destination tracking, one for source copying); building both maps in a single combined pass or using `os.ReadDir` for shallow directories would reduce I/O; low effort
