# Cache System

The cache system enables incremental builds and CI optimization by tracking
what changed between invocations. It uses a **2D taxonomy** (Level x Type)
to classify caches, input hashing and UoW manifests for local builds, and
CI run status for remote build avoidance.

## 2D Cache Taxonomy

All caches in the system are classified along two dimensions:

- **Level** — where the cache lives: `local` (on the developer machine) or `remote` (on the network)
- **Type** — what kind of cache it is: `registry`, `state`, `asset`, `layer`, `work`, or `ci`

```text
           registry  state  asset  layer  work  ci
  local       ✓        ✓      ✓      ✓     ✓
  remote      ✓                      ✓           ✓
```

| Level:Type        | What It Caches                                         |
|-------------------|--------------------------------------------------------|
| `local:registry`  | Container/package registry images (Docker pull cache)  |
| `local:state`     | Incremental build state (UoW manifests in `out/`)      |
| `local:asset`     | Rendered assets (Mermaid SVGs, Structurizr PNGs)       |
| `local:layer`     | Docker BuildKit layer cache                            |
| `local:work`      | Ephemeral work directories                             |
| `remote:registry` | Remote container registry (GHCR)                       |
| `remote:layer`    | Remote BuildKit cache (GitHub Actions cache)           |
| `remote:ci`       | CI build status — last successful GitHub Actions run   |

The `--skip-cache` flag accepts any combination of these specs:

```bash
# Skip a specific cache
eac build --skip-cache=local:state       # Force rebuild (ignore manifests)
eac build --skip-cache=local:asset       # Re-render all diagrams
eac get ci-dispatch --skip-cache=remote:ci  # Dispatch all modules regardless of CI status

# Skip by level (all types at that level)
eac build --skip-cache=local             # Skip all local caches
eac build --skip-cache=remote            # Skip all remote caches

# Skip by type (all levels for that type)
eac build --skip-cache=registry          # Skip registry caches at both levels

# Skip everything
eac build --skip-cache=all

# Combine multiple specs
eac build --skip-cache=local:state,local:asset
```

**Source**: `go/core/cache/cache.go` (taxonomy), `go/core/cache/config.go` (skip logic)

---

## Overview

The cache system operates at two levels:

1. **Local build cache** — input hashing and UoW manifests for `build`, `test`, `lint`, `scan`
2. **CI cache** — GitHub Actions run status for CI dispatch optimization

### Local Build Cache

```text
Source Files → Input Hash → Compare with Manifest → Build or Skip
```

For each UoW, the cache system:

1. Computes an **input hash** from the module's source files
2. Loads the existing **UoW manifest** (if any) from `out/`
3. Compares the hashes — if identical, the UoW is **cached** (skipped)
4. If different, the UoW executes and writes a new manifest

### CI Cache

```text
Module → CICacheChecker.Check(module, headSHA) → Dispatch or Skip
```

For each module in CI dispatch, the checker:

1. Queries GitHub Actions for the last successful run of `ci-{module}.yaml`
2. Compares the run's HEAD SHA against the current HEAD SHA
3. If they match, the module is **CI-cached** (skip dispatch)
4. If they differ (or no run exists), the module needs CI dispatch

**Source**: `go/core/cache/ci.go` (CICacheChecker, CIRunQuerier port)

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

## CI Cache Architecture

The CI cache (`remote:ci`) determines whether a module's CI workflow
needs to be dispatched or can be skipped because a successful build
already exists at the current HEAD SHA.

### Port/Adapter Pattern

The cache package defines a **port interface** — it never imports
GitHub or other infrastructure. Command code injects a concrete adapter.

```text
┌─────────────────────────────┐     ┌──────────────────────────┐
│  go/core/cache              │     │  go/cli/eac/impl/get     │
│                             │     │                          │
│  CIRunQuerier (interface)   │◄────│  ghCIRunQuerier (adapter) │
│  CICacheChecker (logic)     │     │  wraps github.API        │
│  MockCIRunQuerier (testing) │     │                          │
└─────────────────────────────┘     └──────────────────────────┘
```

**CIRunQuerier** is the port:

```go
type CIRunQuerier interface {
    LastSuccessfulRunSHA(workflowName string) (sha string, err error)
}
```

**CICacheChecker** owns the decision logic:

```go
checker := cache.NewCICacheChecker(querier, cacheConfig)
result := checker.Check("core", headSHA)
// result.Cached == true  → skip dispatch
// result.Cached == false → dispatch CI workflow
// result.Reason explains: "valid_ci_at_head", "no_ci_run",
//     "ci_at_different_sha:abc1234", "query_failed: ...", "cache_bypassed"
```

### Decision Flow

```text
1. config.ShouldSkipCI() == true?  → uncached (bypass, reason: "cache_bypassed")
2. Query LastSuccessfulRunSHA("ci-{module}.yaml")
3. Error?                          → uncached (fail-open, reason: "query_failed: ...")
4. No runs exist?                  → uncached (reason: "no_ci_run")
5. Last run SHA == HEAD SHA?       → cached   (reason: "valid_ci_at_head")
6. Different SHA?                  → uncached (reason: "ci_at_different_sha:{sha7}")
```

Fail-open semantics: on querier error, the module is dispatched for safety.

### Bypass

CI cache checking is bypassed when `--skip-cache` includes `remote:ci` or `all`:

```bash
# Force dispatch all modules regardless of CI status
eac get ci-dispatch --skip-cache=remote:ci

# Also bypasses CI cache (and everything else)
eac get ci-dispatch --skip-cache=all
```

CI cache is **not** included in `DefaultSkipSpecs()` — it is never skipped by default.

### Consumers

The CI cache is consumed by two commands:

- **`get ci-dispatch`** — uses `CICacheChecker` to filter which modules need CI dispatch
- **`get changed-modules-ci`** — uses `CIRunQuerier` to check per-module CI status

**Source**: `go/core/cache/ci.go`, `go/cli/eac/impl/get/ci_run_querier.go`

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
# Force rebuild (ignore local state cache)
eac build --force
eac build --skip-cache=local:state    # Equivalent

# Skip specific cache types
eac build --skip-cache=local:asset    # Re-render diagrams
eac build --skip-cache=local:layer    # Docker --no-cache
eac get ci-dispatch --skip-cache=remote:ci  # Force dispatch all

# Skip all caches
eac build --skip-cache=all

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
