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

| File/Sub-package | Responsibility                                                        |
| ---------------- | --------------------------------------------------------------------- |
| build.go         | Entry point, flag parsing, module build orchestration                 |
| framework.go     | Framework hooks, `RunBuildWithFramework`, incremental detection setup |
| buildflags.go    | `ParseBuildSpecificFlags` for build-only flags                        |
| unit_work.go     | `ResolveUnitSpecs` converts modules to component work specs           |
| unit_worker.go   | `buildUnitWorker` executes a single component build with caching      |
| compression.go   | `ProcessArtifactDerivations` for strip/UPX/copy derivations           |
| cache_checker.go | `detectUoWIncrementalChanges` for source-hash change detection        |
| uow_cache.go     | `UoWBuildCacheVerifier` for TUI cache verification                    |
| builders/        | Builder registry and handler implementations (Go, MkDocs, etc.)       |
| docprep/         | Document preprocessing pipeline for MkDocs builds                     |

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

- `build.go` (755 lines), `unit_worker.go` (482 lines), `framework.go` (420 lines) exceed 300 lines
- No test coverage for `unit_worker.go`, `compression.go`, `framework.go`, `uow_cache.go`

### Pain Points

- None identified

### Optimization Opportunities

- None identified
