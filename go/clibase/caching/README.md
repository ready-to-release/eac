# caching

Incremental change detection and per-item content-addressable caching.
Determines which modules have changed since the last successful execution and
provides a general-purpose build cache for individual work items.

## Key Types

- `IncrementalResult` -- outcome of change detection: lists of changed, unchanged, and skipped modules with cache hit ratio and pre-computed module input hashes

## Key Functions

- `DetectIncrementalChanges` -- compares current module state against cached results to identify which modules need re-execution; pre-computes module input hashes for worker reuse

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
- **Pre-computed hashes**: module input hashes are computed once during detection and stored in `IncrementalResult` for workers to reuse, ensuring consistency

## Internal Structure

| File                    | Purpose                                                                              |
| ----------------------- | ------------------------------------------------------------------------------------ |
| `incremental.go`        | `IncrementalResult` and `DetectIncrementalChanges` for module-level change detection |
| `itemcache/types.go`    | `Item`, `BuildFunc`, `Result` type definitions                                       |
| `itemcache/cache.go`    | `Cache` with content-addressable lookup and storage                                  |
| `itemcache/manifest.go` | `Manifest` and `CachedItem` for persistent cache state                               |

## Dependencies

- `contracts/core` -- action type definitions
- `clibase/cmdframework` -- execution context for accessing module state
- `clibase/initsummary` -- incremental info types for summary reporting
- `core/hash` -- file and directory hashing
- `core/logging` -- structured logging
- `core/output` -- output directory management and UoW change aggregation
- `core/workunit` -- unit spec and UoW aggregator definitions

## Role in System

Enables incremental execution across CLI commands. The build command uses module-level change detection to skip unchanged modules, while the item cache provides fine-grained caching for individual components. Both reduce execution time on repeated runs.

## Code Health

### Tech Debt

- None identified

### Pain Points

- `incremental_test.go` is 728 lines, significantly exceeds 300-line threshold

### Optimization Opportunities

- None identified
