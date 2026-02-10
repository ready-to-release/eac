# gitexec

Tool-routed git CLI execution. Routes git commands through the tool registry
for consistent executor-mode behavior and configurable binary paths.

## Key Functions

- `Run` -- executes a git command and returns stdout output
- `RunContext` -- executes a git command with a Go context for cancellation/timeout support
- `RunCombined` -- executes a git command and returns combined stdout/stderr output
- `RunSilent` -- executes a git command and discards output, returning only success/failure

## Patterns

- **Tool-routed execution**: git commands are resolved through the tool registry, allowing the binary path and environment to be configured via tool-config.yml
- **Multiple output modes**: separate functions for stdout-only, combined output, and silent execution cover all common use cases

## Internal Structure

| File | Purpose |
|---|---|
| `executor.go` | `Run`, `RunContext`, `RunCombined`, `RunSilent` functions |

## Dependencies

- `core/tool` -- tool registry for resolving the `git` binary

## Role in System

Provides the low-level git CLI execution layer used by `clibase/git` and other packages that need to run git commands. By routing through the tool registry, the executor respects tool-config.yml overrides.

## Code Health

### Tech Debt
- None identified; compact single-file package

### Pain Points
- None identified

### Optimization Opportunities
- None identified; thin adapter with minimal surface area
