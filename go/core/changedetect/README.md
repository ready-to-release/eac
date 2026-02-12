# changedetect

Unified change detection for incremental build, lint, and test operations,
using a hybrid git state plus file hash approach to determine which modules
need processing.

## Key Types

- **`Detector`** -- Compares current workspace state against recorded state to find changed modules
- **`GitStateProvider`** -- Port for HEAD commit and uncommitted file queries
- **`FileHasher`** -- Port for deterministic file content hashing
- **`ModuleFileResolver`** -- Port for expanding module monikers to file lists
- **`WorkspaceState`** -- Recorded snapshot of git commit, uncommitted hash, and per-module hashes
- **`ModuleState`** -- Source hash and operation-specific extras for one module
- **`ChangeResult`** -- Lists of changed and up-to-date modules with per-module reasons
- **`DetectOptions`** -- Configuration for detection including dependency propagation
- **`GitRepositoryAdapter`** -- Adapts `git.GitRepository` to `GitStateProvider`
- **`ContractResolverAdapter`** -- Resolves module glob patterns to files via registry

## Patterns

- Two-level detection: fast git-state check skips hashing when nothing changed
- Parallel hashing: module hashes computed concurrently with bounded worker pool
- Dependency propagation: optional transitive dependency checking for test operations
- Port composition: `Detector` composes three interfaces for testability
- Adapter bridge: `GitRepositoryAdapter`, `HashAdapter`, `RegistryAdapter` bridge external packages
- State conversion: `ConvertToWorkspaceState` bridges `workunit.StateManager` format

## Internal Structure

| File            | Responsibility                                                                      |
| --------------- | ----------------------------------------------------------------------------------- |
| doc.go          | Package documentation with architecture and usage examples                          |
| changedetect.go | `Detector` and detection algorithm                                                  |
| types.go        | `WorkspaceState`, `ModuleState`, `ChangeResult`, `DetectOptions` and related types  |
| state.go        | State computation, conversion, and workspace state management                       |
| adapters.go     | `GitRepositoryAdapter`, `HashAdapter`, `ContractResolverAdapter`, `RegistryAdapter` |

## Dependencies

- `core/environments` -- `EnvDebugChangeDetect` for debug output toggle
- `core/git` -- `GitRepository` interface for git adapter
- `core/hash` -- `Files`, `UncommittedState`, `ExpandGlobPatterns` for hashing

## Role in System

This package is the incremental execution engine's decision layer. Before any
build, lint, or test command processes a module, the `Detector` compares
current source hashes against stored state to skip unchanged modules. The
parallel hashing and git-state fast path ensure detection remains fast even
in large monorepos with many modules.

## Code Health

### Tech Debt
- None identified

### Pain Points
- Heavy use of goroutines and `sync.WaitGroup` in both `DetectChanges` and `ComputeCurrentState` makes debugging concurrency issues difficult

### Optimization Opportunities
- Extract the parallel hashing goroutine pool into a reusable helper to reduce duplication between `DetectChanges` and `ComputeCurrentState` (low effort, ~40 lines saved)
- Good test coverage with changedetect_test.go, helpers_test.go, state_test.go, and types_test.go
- All files are reasonably sized (largest is changedetect.go at ~200 lines)
