# git

Git-related utility functions. Provides high-level convenience wrappers
over raw git CLI execution.

## Key Functions

- `GetCommitSHA` -- returns the current HEAD commit SHA by delegating to `gitexec`

## Patterns

- **Thin wrapper**: this package provides domain-level functions that internally call `gitexec` for actual git CLI invocation, keeping callers insulated from execution details

## Internal Structure

| File | Purpose |
|---|---|
| `git.go` | `GetCommitSHA` utility function |

## Dependencies

- `clibase/gitexec` -- tool-routed git CLI execution

## Role in System

Provides a convenient API for git operations used by commands that need repository metadata (e.g., commit SHA for output tagging, release versioning). Sits above `gitexec` in the abstraction stack.

## Code Health

### Tech Debt
- None identified; single-function leaf package

### Pain Points
- None identified

### Optimization Opportunities
- None identified; minimal surface area
