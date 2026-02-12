# git

Pure Go git operations using the go-git library, providing repository
access, commit history, tag management, and branch comparison without
shelling out to the git CLI.

## Key Types

- **`GitRepository`** -- Interface for all git operations
- **`Repository`** -- go-git backed implementation of `GitRepository`
- **`RepositoryManager`** -- Factory that opens/inits repos with shared logger
- **`CommitInfo`** -- Commit metadata for changelog generation
- **`MockRepository`** -- Builder-pattern test double with error injection
- **`LazyRepo`** -- Lazy-initialized repo with test injection support
- **`Clock`** -- Function type `func() time.Time` for injectable time source

## Patterns

- Interface abstraction: `GitRepository` enables mock injection in tests
- Constructor injection: `RepositoryManager` injects logger and clock into all repos
- Builder pattern: `MockRepository` uses `With*` methods for test setup
- Lazy initialization: `LazyRepo` defers repo open until first access
- Clock injection: `RepositoryManager.WithClock()` enables deterministic commit timestamps in tests

## Internal Structure

| File | Responsibility |
| --- | --- |
| interface.go | `GitRepository` interface definition |
| git.go | `Repository` struct, basic info methods (RootPath, RemoteURL, CurrentBranch, HeadSHA), and shared helpers (resolveBaseRef, resolveToCommit, findMergeBase) |
| git_status.go | Status and tracking operations (UncommittedFiles, TrackedFiles, StagedFiles, IsFileTracked, IsFileIgnored) |
| git_staging.go | Staging and commit operations (ConfigSet, Add, Commit, AddRemote) |
| git_diff.go | Diff and branch comparison operations (StagedDiff, StagedDiffStats, GetBranchCommits, GetBranchDiff, GetBranchDiffStats, GetBranchFiles, formatDiffStats) |
| manager.go | `RepositoryManager` for opening and initializing repos |
| history.go | Commit history, tag queries, and ref resolution |
| mock.go | `MockRepository` with builder methods and error injection |
| testutil.go | `LazyRepo` lazy-init helper for consumers |
| time.go | `Clock` type definition for injectable time source |

## Dependencies

_No internal repository imports -- this is a leaf package._

## Role in System

This package is a foundational leaf dependency in the `core` module,
providing git operations to higher-level packages like `changedetect`,
`changelog`, and `repository`. The `GitRepository` interface allows
the rest of the system to work against mocks in tests while using
real go-git operations in production.

## Code Health

### Tech Debt
- None -- `time.go` mutable var replaced with `Clock` type injected via `RepositoryManager.WithClock()` (TD-124)

### Assessed and Accepted
- `mock.go` (~490 lines) explicitly implements all 28 `GitRepository` interface methods.
  This was assessed as part of TD-124. Embedding a base type with default no-op
  implementations would reduce line count but would hide missing method implementations
  at compile time when the interface changes. The current explicit approach is verbose
  but provides immediate compiler errors on interface drift, which is the safer trade-off
  for a foundational package with many consumers.

### Optimization Opportunities
- No TODO/FIXME markers found -- codebase is clean of deferred work items
