# Cache System

The cache system enables incremental builds by tracking what changed
between invocations. It uses input hashing, UoW manifests, and timestamp
comparison to skip work that has already been done.

## Overview

```text
Source Files → Input Hash → Compare with Manifest → Build or Skip
```

For each UoW, the cache system:

1. Computes an **input hash** from the module's source files
2. Loads the existing **UoW manifest** (if any) from `out/`
3. Compares the hashes — if identical, the UoW is **cached** (skipped)
4. If different, the UoW executes and writes a new manifest

---

## Input Hashing

### What Gets Hashed

The input hash covers all files matched by the module's component patterns:

| Pattern Category | Example Patterns                         |
|-----------------|------------------------------------------|
| `source`        | `**/*.go`, `Dockerfile`                  |
| `tests`         | `**/*_test.go`                           |
| `config`        | `go.mod`, `go.sum`, `package.json`       |
| `data`          | `testdata/**/*`, `assets/**/*.json`      |

Patterns are defined per-component in `repository.yml` or inherited from
blueprint templates via `component-types.yml`.

### How It Works

1. `ModuleContract.GetGlobPatterns()` collects all patterns across all
   components, prefixed with each component's root path
2. `hash.LoadCache().GetOrCompute()` expands the globs and computes a
   SHA256 over the concatenated file contents
3. The hash is computed **per module** — all components in a module share
   the same input hash

**Source**: `go/cli/eac/impl/build/framework.go` (hash computation),
`go/core/domain/modules/types.go` (`GetGlobPatterns()`),
`go/core/hash/cache.go` (hash cache with mtime optimization)

### Mtime Optimization

To avoid reading every file on every invocation, the hash cache uses file
modification times as a fast-path check:

1. Load cached hash entries from `.cache/eac/build/input-hashes.json`
2. For each file, compare current mtime against cached mtime
3. If all mtimes match, reuse the cached hash (no file reads needed)
4. If any mtime differs, recompute the full hash

**Source**: `go/core/hash/mtime.go`, `go/core/hash/cache.go`

---

## UoW Manifest

Each completed UoW writes a manifest to `{outdir}/uow.manifest.json`:

```json
{
  "id": "build:eac-cli:go:go",
  "input_hash": "sha256:abc123...",
  "output_hash": "sha256:def456...",
  "executed_at": "2025-02-07T15:30:00Z",
  "duration_ms": 12345,
  "status": "completed",
  "artifacts": [...]
}
```

| Field          | Purpose                                        |
|----------------|------------------------------------------------|
| `input_hash`   | Hash of source files at build time             |
| `output_hash`  | Hash of produced artifacts                     |
| `executed_at`  | Timestamp for cross-UoW invalidation           |
| `duration_ms`  | Execution time for performance tracking        |
| `artifacts`    | List of produced files with per-artifact SHA256|

**Source**: `go/core/output/manifest.go`, `go/core/output/types.go`

---

## Incremental Change Detection

The cache detector compares the current state against stored manifests to
determine which UoWs need to re-execute.

### Detection Algorithm

For each expected UoW:

```text
1. Load manifest from out/{action}/{module}/{component}-{tool}/uow.manifest.json
2. If no manifest exists → CHANGED (first build)
3. Compute current input hash
4. If input_hash differs from manifest → CHANGED (source modified)
5. Validate all artifacts still exist with correct SHA256
6. If any artifact missing/changed → CHANGED (output corrupted)
7. Cross-context checks (see below)
8. Otherwise → CACHED (skip)
```

**Source**: `go/core/output/cache_detector.go` (`checkUoWChanged()`)

### Cross-Context Invalidation

Some UoWs depend on other UoWs completing, not just source file changes:

**Test/Lint/Scan after Build**: If a module was rebuilt, its tests must
re-run even if the test files haven't changed. The cache detector checks
whether any build manifest for the **same module** has a newer
`executed_at` timestamp:

```text
If action is test/lint/scan:
  Load build manifests for this module
  If any build manifest.executed_at > this manifest.executed_at:
    → CHANGED ("build invalidated")
```

This ensures that `test:core:go:gotest` re-runs after `build:core:go:go`
produces new output, even if no test files changed.

**Source**: `go/core/output/cache_detector.go`, lines 146-155

### Change Result

The detection produces a `UoWChangeResult`:

```go
type UoWChangeResult struct {
    Changed   []UnitID   // UoWs that need to execute
    Unchanged []UnitID   // UoWs that can be skipped (cached)
    Reasons   map[string]string  // Change reason per UoW
}
```

Changed UoWs execute normally. Unchanged UoWs are marked `Cached` in the
state manager and skipped by the scheduler.

---

## Cache File Locations

| Path                                    | Content                           |
|-----------------------------------------|-----------------------------------|
| `out/build/{module}/{comp}-{tool}/`     | Build output + manifest           |
| `out/test/{module}/{comp}-{tool}/`      | Test results + manifest           |
| `out/lint/{module}/{comp}-{tool}/`      | Lint results + manifest           |
| `out/scan/{module}/{comp}-{tool}/`      | Scan results + manifest           |
| `.cache/eac/build/input-hashes.json`    | Mtime-optimized hash cache        |

---

## Cache Commands

```bash
# Force rebuild (ignore cache)
eac build --force

# Clear all caches
eac update cache-clear

# Show build summary (shows cached vs rebuilt)
eac show build-summary
```

---

## Related Documentation

- [Build Execution System](./build-execution.md) — UoW lifecycle and orchestrator
- [Component Resolution](./component-resolution.md) — How components create UoWs
- [Dependencies](./dependencies.md) — Cross-module dependency ordering
