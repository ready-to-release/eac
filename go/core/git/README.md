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

## Patterns

- Interface abstraction: `GitRepository` enables mock injection in tests
- Constructor injection: `RepositoryManager` injects logger into all repos
- Builder pattern: `MockRepository` uses `With*` methods for test setup
- Lazy initialization: `LazyRepo` defers repo open until first access

## Internal Structure

| File | Responsibility |
| --- | --- |
| interface.go | `GitRepository` interface definition |
| git.go | `Repository` implementation (status, diff, staging, branch ops) |
| manager.go | `RepositoryManager` for opening and initializing repos |
| history.go | Commit history, tag queries, and ref resolution |
| mock.go | `MockRepository` with builder methods and error injection |
| testutil.go | `LazyRepo` lazy-init helper for consumers |
| time.go | Overridable `timeNow` for deterministic tests |

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
- `interface.go`: `GitRepository` has 28 methods -- consider splitting into focused interfaces (e.g., `Reader`, `Writer`, `HistoryQuerier`) to reduce coupling
- `time.go:7`: package-level mutable `var timeNow = time.Now` used for test overrides; a clock interface injected via constructor would be safer

### Pain Points
- `git.go` (601 lines) covers status, diff, staging, and branch operations -- splitting by responsibility would improve navigability
- `mock.go` (427 lines) must shadow all 28 interface methods, making it expensive to maintain when the interface changes

### Optimization Opportunities
- Extract branch-comparison methods (`GetBranch*`) into a dedicated sub-interface; most consumers only need read-only operations (low effort, high impact on mock size)
- No TODO/FIXME markers found -- codebase is clean of deferred work items
