# caching

Incremental change detection and per-item content-addressable caching.
Determines which modules have changed since the last successful execution and
provides a general-purpose build cache for individual work items.

## Key Types

- `IncrementalResult` -- outcome of change detection: lists of changed, unchanged, and skipped modules with cache hit ratio

## Key Functions

- `DetectIncrementalChanges` -- compares current module state against cached results to identify which modules need re-execution

### Sub-package: itemcache

- `Cache` -- content-addressable cache that stores and retrieves individual work item results by input hash
- `Item` -- describes a cacheable work item: key, input hash, and build function
- `Result` -- outcome of a cache lookup: hit/miss, cached output, and timing
- `BuildFunc` -- function signature for producing a cacheable result
- `Manifest` -- persistent cache manifest mapping item keys to their cached state
- `CachedItem` -- single entry in the manifest: input hash, output hash, timestamp

## Patterns

- **Content-addressable caching**: items are keyed by a hash of their inputs; if the hash matches a previous run, the cached result is reused
- **Module-level change detection**: `DetectIncrementalChanges` checks file hashes, dependency changes, and config changes to determine rebuild scope
- **Manifest persistence**: `Manifest` is stored as JSON alongside build outputs, surviving across CLI invocations

## Internal Structure

| File | Purpose |
|---|---|
| `incremental.go` | `IncrementalResult` and `DetectIncrementalChanges` for module-level change detection |
| `itemcache/types.go` | `Item`, `BuildFunc`, `Result` type definitions |
| `itemcache/cache.go` | `Cache` with content-addressable lookup and storage |
| `itemcache/manifest.go` | `Manifest` and `CachedItem` for persistent cache state |

## Dependencies

- `contracts/core` -- action type definitions
- `clibase/cmdframework` -- execution context for accessing module state
- `clibase/initsummary` -- incremental info types for summary reporting
- `core/hash` -- file and directory hashing
- `core/logging` -- structured logging
- `core/output` -- output directory management
- `core/workunit` -- unit spec definitions

## Role in System

Enables incremental execution across CLI commands. The build command uses module-level change detection to skip unchanged modules, while the item cache provides fine-grained caching for individual components. Both reduce execution time on repeated runs.

## Code Health

### Tech Debt
- `incremental.go:27` and `itemcache/cache.go:13` both declare package-level mutable `log` vars; prefer constructor-injected loggers
- `incremental.go` (193 lines) has no dedicated test file; change-detection logic is a critical path

### Pain Points
- `itemcache/` sub-package has strong test coverage (649 test lines), but `incremental.go` is untested at the unit level

### Optimization Opportunities
- Add unit tests for `DetectIncrementalChanges` with mocked module state to cover edge cases (medium effort)
- Inject logger via constructor rather than package-level var to improve testability (low effort)
