# paths

Centralized path constants and builder functions for the EAC repository.
This package has zero internal dependencies (stdlib only) to prevent import cycles.

## Key Types

| Type | Purpose |
|------|---------|
| `OutDir` | Root output directory constant for all generated artifacts |
| `EACCacheRoot` | Root cache directory constant for all transient caches |
| `EACDir` | Configuration directory constant (.eac) |
| `CacheHashLength` | Character count for truncated hashes in cache filenames |

## Patterns

- Zero-dependency leaf: Only imports stdlib to serve as the foundation for all path resolution
- Constant + builder pairs: Directory name constants paired with functions that join them with repo root
- Forward-slash normalization: Cache path helpers normalize separators for cross-platform use
- Traceable cache names: `SanitizeForCacheName` produces human-readable cache filenames from source paths
- Container awareness: `CommandsBinaryPath` checks `CLIE_CONTAINER_ROOT` for container execution

## Internal Structure

| File | Purpose |
|------|---------|
| `paths.go` | Core directory constants, fundamental path builder functions |
| `paths_output.go` | Output path builders for build, test, lint, scan artifacts |
| `paths_cache.go` | Cache path helpers, cache filename sanitization |
| `paths_config.go` | Configuration path builders for .eac directory structure |
| `paths_builders.go` | Composite path builders for specs, docs, design, templates |

## Dependencies

No internal dependencies. This package depends only on the Go standard library.

## Role in System

`paths` is the lowest-level package in the `core` module, providing the single
source of truth for every directory convention (out/, .cache/eac/, specs/, etc.)
and path derivation function used across the codebase. Nearly every other `core`
package depends on it for consistent path resolution. Path builders cover build
outputs, test outputs, staging areas, cache directories, spec locations,
container paths, and documentation assets. The package also provides traceable
cache filename generation used by the drawio and mermaid acceleration caches,
and Windows-safe moniker sanitization for output paths.

## Code Health

- **Tech Debt**: `paths.go`: `hasMainWorkspace` is tracked but immediately discarded (`_ = hasMainWorkspace`); either use it or remove the variable. Many path builder functions repeat `filepath.Join(repoRoot, OutDir, ...)` inline rather than calling a shared helper.
- **Pain Points**: No validation that `repoRoot` is non-empty or absolute in any builder function.
- **Optimization Opportunities**: Extract a common `outSubPath(repoRoot string, segments ...string) string` helper to reduce duplication across the ~30 `OutDir`-based builders (low risk, mechanical refactor). `SanitizeForCacheName` allocates multiple intermediate strings; a `strings.Builder` approach could reduce allocations for hot-path cache lookups (low priority, measure first).
