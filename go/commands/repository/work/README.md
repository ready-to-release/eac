# work

Manages parallel development workspaces using git worktrees. Provides a complete lifecycle from workspace creation through committing, pulling, merging, and PR creation, enabling concurrent Claude Code sessions on separate branches.

## Key Types

- **`Worktree`** -- Represents a git worktree with path, branch, SHA, and clean status
- **`BaseConfig`** -- Shared config for all work commands (debug, repo root, logger, git ops)
- **`WorkGitOperations`** -- Interface for git operations covering worktree, branch, status, remote, merge/rebase, stash, and commit operations
- **`createConfig`** -- Configuration for workspace creation (branch, base, path)

## Patterns

- Registry-based subcommand dispatch: each subcommand file calls `registry.Register` in `init()`
- TUI selector fallback: interactive subcommand picker when run without arguments
- Phased execution: create command runs through parse, validate, execute, output phases
- Shared base config: `internal.ParseBaseConfig` extracts common flags and initializes logging
- Testable git operations: git interactions abstracted through `WorkGitOperations` interface

## Internal Structure

| File      | Responsibility                                              |
| --------- | ----------------------------------------------------------- |
| work.go   | Parent command entry point with TUI and subcommand dispatch |
| create.go | Create new worktree for parallel development                |
| commit.go | Commit changes in current worktree                          |
| pull.go   | Pull latest changes from remote                             |
| merge.go  | Merge worktree branch back to main                          |
| pr.go     | Create pull request from worktree branch                    |
| list.go   | List all active worktrees with status                       |
| remove.go | Remove a worktree and clean up branch                       |

## Dependencies

- `cli/eac/help` -- help text rendering for parent command
- `cli/eac/impl/work/internal` -- shared config, git operations interface, and worktree utilities
- `clibase/registry` -- command registration and subcommand discovery
- `clibase/flags` -- flag validation and parsing
- `clibase/gitexec` -- low-level git command execution
- `adapters/tui` -- TUI detection and subcommand options
- `adapters/tui/selector` -- interactive command selector
- `core/logging` -- structured logging with zap
- `core/repository` -- repository root discovery

## Role in System

The `work` package enables parallel feature development in `eac` by managing git worktrees as isolated workspaces.

Each workspace gets its own directory and branch,
allowing developers to run multiple Claude Code sessions simultaneously without branch conflicts or stash juggling.

## Code Health

### Tech Debt
- None identified

### Pain Points
- work_test.go is 717 lines (significantly exceeds 300-line threshold)
- pr.go is 545 lines (significantly exceeds 300-line threshold)
- remove.go is 514 lines (significantly exceeds 300-line threshold)
- merge.go is 447 lines (exceeds 300-line threshold)
- create.go is 359 lines (exceeds 300-line threshold)
- pull.go is 347 lines (exceeds 300-line threshold)
- pr_test.go is 318 lines (exceeds 300-line threshold)

### Optimization Opportunities
- None identified
