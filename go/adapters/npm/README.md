# npm

NPM isolation adapter providing safe, parallel-friendly npm environments
for TypeScript test execution by copying source files to isolated work
directories.

## Key Types

- **`NpmIsolation`** -- Manages isolated npm work directories
- **`IsolatedEnv`** -- Represents a prepared isolated environment
- **`WithInstallLock`** -- Serializes npm install calls via unexported mutex

## Patterns

- Directory isolation: Copies source to `.cache/eac/npm/work/{moniker}/`
- Incremental sync: Only copies files changed by mtime and size
- Shared npm cache: All modules share `.cache/eac/npm/cache/` for downloads
- Serialized installs: `WithInstallLock()` prevents concurrent npm cache contention
- Stale detection: Resets work directory when package.json changes

## Internal Structure

| File         | Responsibility                                                 |
| ------------ | -------------------------------------------------------------- |
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
- None identified

### Pain Points
- isolation.go is 332 lines; candidate for splitting (extract file sync logic and change detection into separate files)
- packageFilesChanged in isolation.go uses mtime and size comparison, which can miss content changes when file size is unchanged; a content hash check would be more reliable

### Optimization Opportunities
- None identified
