# staging

File staging infrastructure for the docprep pipeline, handling source file copying to the staging directory, file indexing, asset reference scanning, and orphan cleanup.

## Key Types

- **`FileIndex`** -- Pre-built index of all files in staging for efficient iteration
- **`FileMapper`** -- Interface tracking source-to-staging file mappings
- **`SimpleFileMap`** -- Basic `FileMapper` implementation using a map
- **`CopyConfig`** -- Parameters for a copy operation (book, roots, options)
- **`CopyStats`** -- Statistics from a copy operation (copied, skipped, lazy)
- **`MirrorStats`** -- Statistics from a mirror sync (copied, skipped, deleted)

## Patterns

- Lazy asset copying: Pre-scans markdown for asset references, skipping unreferenced images during copy to reduce I/O
- Mtime/size comparison: Files are only copied if modification time or size differs from existing staged versions
- File index sharing: `FileIndex` is built once after copy and shared across all pipeline phases to avoid repeated directory walks
- Orphan cleanup: Files in staging not tracked by the current build's `FileMapper` are removed to prevent stale content
- Mirror sync: `MirrorSync` provides bidirectional directory synchronization with orphan deletion for staging updates

## Internal Structure

| File | Responsibility |
| --- | --- |
| copier.go | `CopyAllSources`, `FileMapper`, `SimpleFileMap`, and glob-based file copying |
| file_index.go | `FileIndex` with pre-built markdown and all-file lists, thread-safe refresh |
| assetref.go | `ScanAssetReferences` for pre-copy asset reference discovery |
| mirror.go | `MirrorSync` for directory synchronization with orphan deletion |
| orphan.go | `CleanOrphanedStagedFiles` removes untracked source files from staging |

## Dependencies

- `core/config` -- book configuration for copy source definitions

## Role in System

The staging package provides the foundational file operations for the docprep pipeline in `eac`. It handles the first three pipeline phases: scanning source markdown for asset references (phase 1), copying source files to the staging directory with lazy asset optimization (phase 2), and building the file index (phase 3).

The `FileIndex` it produces is consumed by nearly every subsequent phase for efficient file iteration without repeated filesystem walks. The `FileMapper` tracks which files were copied, enabling orphan detection in this package and link translation in the linking package.

## Code Health

### Tech Debt
- ~~`assetref.go:11`: `assetRefPattern` duplicates the identical regex in `cleanup/assets.go:24`~~ (resolved: `AssetReferencePattern` is now the canonical source; cleanup imports it)
- `orphan.go` has no corresponding test file

### Pain Points
- None identified -- files are small and well-structured

### Optimization Opportunities
- None identified -- lazy-copy optimization and mtime comparison are already in place for I/O reduction
