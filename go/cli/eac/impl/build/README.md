# build

Implements the `build` command, which compiles one or more modules by moniker using the command framework for parallel, dependency-ordered execution with incremental caching.

## Key Types

- **`BuildConfig`** -- Build-specific flags: tidy, version, reproducible, artifacts mode
- **`BuildResult`** -- Outcome of a single module build
- **`BuildSpecificFlags`** -- Parsed build-only CLI flags
- **`UoWBuildCacheVerifier`** -- Verifies unit-of-work cache for incremental builds

## Patterns

- Command framework integration: Registers hooks (`AfterInit`, `AfterResolve`, `AfterExecute`) for build lifecycle
- UoW-level caching: Detects unchanged components via input hash comparison against prior manifests
- Component-level parallelism: `ResolveUnitSpecs` expands modules to buildable components dispatched as work units
- Artifact derivation: Post-build compression (strip, UPX) and cross-platform artifact generation

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| build.go | Entry point, flag parsing, module build orchestration |
| framework.go | Framework hooks, `RunBuildWithFramework`, incremental detection setup |
| buildflags.go | `ParseBuildSpecificFlags` for build-only flags |
| unit_work.go | `ResolveUnitSpecs` converts modules to component work specs |
| unit_worker.go | `buildUnitWorker` executes a single component build with caching |
| compression.go | `ProcessArtifactDerivations` for strip/UPX/copy derivations |
| cache_checker.go | `detectUoWIncrementalChanges` for source-hash change detection |
| uow_cache.go | `UoWBuildCacheVerifier` for TUI cache verification |
| builders/ | Builder registry and handler implementations (Go, MkDocs, etc.) |
| docprep/ | Document preprocessing pipeline for MkDocs builds |

## Dependencies

- `contracts/core` -- action type constants and module port interface
- `clibase/cmdframework` -- command lifecycle framework and hook registration
- `clibase/flags` -- shared flag parsing with environment awareness
- `clibase/initsummary` -- init summary reporting for TUI display
- `core/config` -- repository configuration and artifact definitions
- `core/hash` -- input hash computation for incremental detection
- `core/output` -- UoW manifest tracking and artifact validation
- `core/tool` -- build handler registry and tool execution
- `core/workunit` -- `UnitSpec` and `UnitID` types for component work items

## Role in System

The build package is the largest command implementation in `eac`, orchestrating multi-module parallel builds with per-component granularity. It delegates actual compilation to registered `BuildHandler` implementations in the builders sub-package, while managing incremental caching, artifact derivation, and manifest tracking through the command framework. The docprep sub-package handles MkDocs-specific preprocessing before documentation builds.

## Code Health

### Tech Debt
- `unit_worker.go`: `buildUnitWorker` is ~285 lines (30-315) with deeply nested cache verification, locking, and build execution; extract cache-check and manifest-recording into helpers
- `build.go`: `Build()` is ~156 lines (195-351) with inlined flag extraction; the block at lines 233-263 assigns ~18 locals that could be a config struct
- `framework.go:36`: `cachedUnitWorkMu` is a package-level mutex protecting shared state; consider encapsulating in a type
- No tests for `unit_worker.go`, `unit_work.go`, `compression.go`, or `framework.go` -- only `buildflags`, `cache_checker`, and `build_freshness` have test files

### Pain Points
- ~~`hasExistingArtifacts` reloads module registry on every call despite being invoked per-module in a loop~~ (resolved: `modRegistry` now passed as parameter from caller)
- Artifact derivation runs in both `runModuleBuild` and `processAllArtifactDerivations`, relying on comments to explain why duplication is safe

### Optimization Opportunities
- Split `buildUnitWorker` into a cache-check phase and a build-execute phase to improve readability and testability -- moderate effort
- ~~Hoist config/module loading out of `hasExistingArtifacts`~~ (resolved)
