# caching

Content-addressable caching for expensive documentation preprocessing operations, including mermaid SVG rendering, DrawIO PNG optimization, and file-level change detection.

## Key Types

- **`AssetCache`** -- Content-addressable cache for mermaid and DrawIO rendered assets, keyed by SHA256 of inputs
- **`MermaidCacheKey`** -- Hash inputs for mermaid rendering (code content, dimensions, theme)
- **`DrawioCacheKey`** -- Hash inputs for DrawIO optimization (source hash, max width)
- **`CacheStats`** -- Hit/miss counters for asset cache reporting
- **`FileHashCache`** -- Per-book file content hash tracker for incremental processing across builds
- **`CacheHitStats`** -- Hit/miss counters for file hash cache

## Patterns

- Content-addressable storage: SHA256 of inputs determines cache file path; identical inputs always produce cache hits regardless of build order
- Atomic writes: Rendered assets are written to `.tmp` then renamed to prevent partial reads by concurrent builds
- Graceful degradation: Missing or corrupt cache files produce empty caches rather than errors, allowing builds to proceed
- Thread-safe access: `FileHashCache` uses `sync.RWMutex` for concurrent read/write safety across parallel pipeline phases
- Skip-cache support: Respects `--skip-cache=asset` flag via `cache.Config.ShouldSkipAsset()` to force re-rendering

## Internal Structure

| File | Responsibility |
| --- | --- |
| assetcache.go | `AssetCache` with get/put for mermaid SVGs and DrawIO PNGs, hash key computation, and cache stats tracking |
| filehash.go | `FileHashCache` for per-file SHA256 change detection, with JSON persistence across builds |

## Dependencies

- `core/cache` -- cache configuration and skip-cache flag support
- `core/paths` -- cache directory and file path conventions
- `docprep/staging` -- `CopyFile` utility for atomic cache writes

## Role in System

The caching package provides the performance optimization layer for documentation builds in `eac`. Diagram rendering (mermaid, DrawIO) is expensive and container-based; the `AssetCache` avoids re-rendering unchanged diagrams by storing outputs keyed by content hash. The `FileHashCache` enables incremental markdown processing by tracking which source files have changed between builds, persisting hash state as JSON.

Both caches are consulted by the diagram processing and content phases in the docprep pipeline. The `AssetCache` is also used directly by the DrawIO and mermaid builder handlers in the builders package for standalone diagram rendering outside the pipeline.

Cache statistics (hit/miss counts via `CacheStats` and `CacheHitStats`) are reported at the end of each build for observability.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified -- both files are focused, well-tested, and use thread-safe patterns

### Optimization Opportunities
- None identified -- package is small (323 lines of source) and already uses content-addressable hashing for efficiency
