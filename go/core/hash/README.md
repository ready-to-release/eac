# hash

Deterministic file content hashing for change detection, with an mtime-based
cache layer for incremental build acceleration.

## Key Types

| Type | Purpose |
|------|---------|
| `HashCache` | Persistent mtime-aware cache for computed file hashes |
| `ModuleCache` | Per-module cached hash with file mtime metadata |
| `ParallelOptions` | Configures worker count for parallel hashing |
| `GlobPatternGetter` | Interface for types that provide glob patterns |
| `ModuleInputHashProvider` | Function type computing input hash for a UnitID |

## Patterns

- Mtime short-circuit: If all file mtimes match cached values, returns hash without reading content
- Parallel with determinism: Files hashed concurrently but combined in sorted order
- Atomic persistence: Cache saved via temp file + rename to prevent corruption
- Glob expansion: `ExpandGlobPatterns` supports `**` patterns via doublestar library

## Internal Structure

| File | Purpose |
|------|---------|
| `hash.go` | Core SHA-256 hashing, glob expansion, `Files` and `ComputeFromPatterns` |
| `parallel.go` | Parallel file hashing with worker pool and context cancellation |
| `cache.go` | `HashCache` with `GetOrCompute` fast/slow path |
| `mtime.go` | Mtime comparison and collection helpers |
| `provider.go` | `ModuleInputHashProvider` factory for UoW change detection |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/workunit` | UnitID type for module input hash provider |

## Role in System

`hash` is the change-detection engine for the `core` module's incremental build
system. It computes content hashes consumed by `workunit.StateManager` for cache
invalidation, and its mtime cache avoids redundant I/O on unchanged source files.
The `HashCache` persists to `.cache/eac/build/input-hashes.json` and provides a
fast path that skips content reads when file modification times are unchanged.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: `parallel.go`: `FilesParallel` is ~108 lines with three separate context-cancellation checks; extracting the fan-out loop into a helper would improve readability.
- **Optimization Opportunities**: `mtimesUnchanged` stats files sequentially; for modules with thousands of source files, parallel stat calls would reduce the fast-path latency (low priority, profile first).
