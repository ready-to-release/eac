# cache

TTL-based caching for GitHub Container Registry data and extension metadata.

## Key Types

- **`RegistryCache`** -- Top-level cache with version, extension map, and updated-at timestamp
- **`ExtensionCache`** -- Per-extension cached data: latest SHA, tags, image digests
- **`MetadataCache`** -- Extension metadata cache with digest-based invalidation
- **`ExtensionMeta`** -- Parsed extension metadata (name, version, volumes, env, requirements)
- **`VolumeRequest`** -- Volume mount requested by an extension (name, target, type)
- **`MetaRequirements`** -- Extension requirements (CLI version, runtime, memory, CPU)
- **`MetaEnvVar`** -- Environment variable declared in extension metadata

## Patterns

- Atomic file writes: Cache data written to `.tmp` file then renamed for crash safety
- Digest-based invalidation: Metadata cache is invalidated when image digest changes
- Version compatibility: Registry cache checks version field and rebuilds on mismatch
- Graceful degradation: Load functions return empty cache (not error) on missing or corrupt files

## Internal Structure

| File              | Responsibility                                                  |
| ----------------- | --------------------------------------------------------------- |
| registry_cache.go | RegistryCache load/save, TTL check, extension/digest getters    |
| metadata.go       | MetadataCache load/save, digest validation, ExtensionMeta types |

## Dependencies

- `internal/logging` -- Debug and warning log output

## Role in System

The cache package provides two caching layers: registry cache (tag/SHA data from GHCR with TTL expiration) and metadata cache (extension metadata with digest-based invalidation). Both are stored as JSON files under `.clie/cache/` in the repository root. The registry cache is used by conf, docker, and cmd packages to avoid redundant GHCR API calls. The metadata cache is used by docker's hosting-metadata to avoid redundant container exec calls for extension metadata.

## Code Health

### Tech Debt

_None identified._

### Pain Points

_None identified._

### Optimization Opportunities

_None identified._
