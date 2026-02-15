# internal

Provides shared infrastructure for the `work` command group including git worktree management, git operations interface, dependency injection, and common configuration parsing.

## Key Types

- **`WorkGitOperations`** -- Interface defining all git operations needed by work commands (worktree, branch, status, remote, merge/rebase, stash, commit operations)
- **`defaultGitOps`** -- Production implementation of `WorkGitOperations` that shells out to real git commands
- **`Worktree`** -- Represents a git worktree with path, branch, SHA, and clean status
- **`BaseConfig`** -- Common configuration for all work commands (debug mode, repo root, logger, git ops)
- **`Deps`** -- Injectable dependencies container for work commands

## Key Functions

- **`NewDefaultGitOps()`** -- Create production git operations implementation rooted at a repo path
- **`ParseBaseConfig()`** -- Parse common configuration from command-line arguments (debug flag, repo root, logging)
- **`DefaultDeps()`** -- Create a `Deps` populated with production defaults
- **`GetWorktrees()`** -- List all worktrees in the repository
- **`FindWorktreeByBranch()`** -- Find a worktree by its branch name
- **`GetMainWorktree()`** -- Get the main worktree (main or master branch)
- **`GenerateWorktreePath()`** -- Create a standard worktree path from repo name and branch
- **`IsWorktreeClean()`** -- Check if a worktree has uncommitted changes
- **`GetRepoName()`** -- Extract repository name from repo root path
- **`WorktreeExists()`** -- Check if a worktree exists for a branch
- **`BranchExists()`** -- Check if a branch exists in the repository
- **`GetCurrentBranch()`** -- Get the current branch name
- **`EnsureInGitRepo()`** -- Verify the current directory is inside a git repository
- **`GetAbsolutePath()`** -- Convert a relative path to absolute
- **`parseWorktreeList()`** -- Parse `git worktree list --porcelain` output into `Worktree` structs

## Patterns

- Interface-based git operations: `WorkGitOperations` allows mock injection for testing all 15+ git operations
- Debug logging: every git operation logs its parameters, duration, and result at debug level
- Convenience wrappers: top-level functions (e.g., `GetWorktrees`, `BranchExists`) delegate to `defaultGitOps`
- Porcelain parsing: parses `--porcelain` git output for machine-readable worktree listing

## Internal Structure

| File | Responsibility |
| --- | --- |
| config.go | `BaseConfig` parsing with debug flag, repo root, logging, and git ops initialization |
| deps.go | `Deps` injectable dependency container with production defaults |
| git_ops.go | `WorkGitOperations` interface and `defaultGitOps` production implementation (382 lines) |
| worktree.go | Worktree discovery, parsing, lookup, path generation, and convenience functions (180 lines) |
| timing.go | Timing wrapper for git operations with debug logging |

## Dependencies

- `clibase/flags` -- debug flag parsing
- `clibase/gitexec` -- git command execution
- `core/logging` -- structured logging with debug-level operation tracing
- `core/repository` -- repository root discovery

## Role in System

The `work/internal` package is the foundation for all `work` commands (create, commit, merge, pull, remove). It provides the git abstraction layer that all work commands use to manage worktrees, branches, and merges. The `WorkGitOperations` interface enables comprehensive testing of work command logic without executing real git operations.

## Code Health

### Tech Debt
- None identified

### Pain Points
- worktree_test.go is 499 lines (significantly exceeds 300-line threshold)
- git_ops.go is 359 lines (exceeds 300-line threshold)

### Optimization Opportunities
- None identified
