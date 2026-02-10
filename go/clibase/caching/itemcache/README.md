# itemcache

Content-addressable per-item cache for individual work item results.
Stores and retrieves build outputs keyed by input hash, enabling fine-grained
cache hits at the item level rather than the module level.

## Key Types

- `Cache` -- manages a cache directory and manifest file; provides `Execute()` to classify items as hits or misses, build only misses, and copy all results to output
- `Item` -- describes a cacheable work item: key, content hash, cache filename, and output relative path
- `Result` -- outcome of a cache operation: total items, cache hits, misses, built count, and hit rate percentage
- `BuildFunc` -- function signature for producing cacheable results from a list of cache-miss items
- `Manifest` -- persistent JSON manifest mapping item keys to their cached state on disk
- `CachedItem` -- single entry in the manifest: content hash, cache filename, and timestamp

## Patterns

- **Content-addressable caching**: items are keyed by a hash of their inputs; if the hash matches a previous run, the cached file is reused without rebuilding
- **Manifest persistence**: `Manifest` is stored as JSON in the cache directory, surviving across CLI invocations
- **Stale pruning**: after each execution, manifest entries not in the current item set are pruned along with their cached files

## Internal Structure

| File | Purpose |
|---|---|
| `types.go` | `Item`, `BuildFunc`, `Result` type definitions |
| `cache.go` | `Cache` struct with `New()`, `Execute()`, `prune()`, and `copyFile()` |
| `manifest.go` | `Manifest` and `CachedItem` types with `loadManifest()` and `saveManifest()` |

## Dependencies

- `core/logging` -- structured logging

## Role in System

Provides fine-grained caching for individual build artifacts within a module. While the parent `caching` package handles module-level incremental change detection, `itemcache` operates at the individual item level (e.g., per-file or per-component outputs). The build command uses this to skip rebuilding unchanged items even when the module as a whole has changes.

## Code Health

### Tech Debt
- `cache.go:13` package-level mutable `log` var; prefer constructor-injected logger
- `manifest.go` silently returns an empty manifest on load errors; this is intentional for first-run scenarios but could mask corruption

### Pain Points
- None identified; strong test coverage (649 test lines across the sub-package)

### Optimization Opportunities
- Inject logger via `Cache` constructor rather than package-level var to improve testability (low effort)
