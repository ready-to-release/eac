# update/cache-clear

Clears incremental cache state files, rendered asset caches, ephemeral work directories, and Docker image/builder caches. Supports a two-dimensional cache taxonomy (level x type) with dry-run preview, verbose output, and per-category result summaries.

## Key Types

- **`ClearDir`** -- Cache directory definition with relative path, description, clear mode, cache level, and cache type
- **`CacheTarget`** -- Resolved `ClearDir` with absolute path and spec-matching logic
- **`ClearMode`** -- Enumeration of clearing strategies: `ClearContents`, `ClearDocker`, `ClearSemaphore`
- **`CacheSetResult`** -- Per-directory clearing outcome with deleted count, bytes freed, and item paths
- **`CategoryResult`** -- Aggregated clearing results grouped by cache type
- **`ClearResult`** -- Overall clearing outcome with counts, bytes, targets, and errors

## Patterns

- Two-dimensional cache taxonomy: targets classified by level (local, remote) and type (state, asset, work, registry, layer)
- Spec-based filtering: `--type` flag parsed into `cache.Spec` values that match targets by level and type
- Three clearing modes: directory content deletion, Docker prune commands, and semaphore file cleanup
- Dry-run support: previews all deletions with file sizes without modifying the filesystem
- Ordered type processing: clears state, work, asset, registry, and layer caches in a fixed sequence
- Default clearing scope: state + work caches (matching the `--skip-cache` default elsewhere)

## Internal Structure

| File | Responsibility |
| --- | --- |
| clear.go | Command entry point, flag parsing, target clearing execution, summary display |
| clear_docker.go | Docker cache clearing and space reclamation logic |
| clear_semaphores.go | Semaphore file cleanup logic |
| types.go | `ClearDir`, `CacheTarget`, `ClearMode`, `ClearResult` types, target building/filtering, type flag parsing |

## Dependencies

- `core/fileutil` -- retry-aware file removal for Windows lock handling
- `clibase/registry` -- command registration
- `core/cache` -- cache taxonomy types (`Level`, `Type`, `Spec`), spec parsing, and default skip specs
- `core/logging` -- structured logging
- `core/paths` -- output directory, cache root, and subdirectory path constants
- `core/repository` -- repository root discovery

## Role in System

The `update cache-clear` command provides cache management for `eac-cli`, allowing developers to reset incremental build/test/lint/scan state, clear rendered asset caches, remove ephemeral work directories, and prune Docker images and builder caches. It is the primary tool for resolving stale cache issues, hung parallel execution (via semaphore cleanup), and reclaiming disk space across all cache layers.

## Code Health

### Tech Debt
- Package-level `var semaphoreFiles` in clear_semaphores.go hardcodes known semaphore filenames; adding new semaphores requires editing this list

### Pain Points
- clear.go is 344 lines (exceeds 300-line threshold)
- types.go is 301 lines (exceeds 300-line threshold)
- clear_test.go is 1028 lines (exceeds 300-line threshold)
- No test coverage for clear_docker.go (117 lines) or clear_semaphores.go (59 lines)

### Optimization Opportunities
- Make semaphore file list discoverable via a registry or glob pattern instead of a hardcoded slice in clear_semaphores.go (low effort, removes maintenance burden)
