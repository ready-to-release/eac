# ghexec

Tool-routed GitHub CLI (`gh`) execution. Implements the `github.CLIExecutor` interface
by routing `gh` commands through the tool registry for consistent executor-mode behavior.

## Key Types

- `ToolRoutedExecutor` -- implements `github.CLIExecutor` by looking up the `gh` tool in the tool registry and executing commands through it

## Key Functions

- `NewToolRoutedExecutor` -- creates a `ToolRoutedExecutor` backed by the tool registry
- `Run` -- executes a `gh` command, returning combined stdout/stderr output and exit code
- `RunJSON` -- executes a `gh` command and returns the output as raw JSON bytes

## Patterns

- **Tool-routed execution**: instead of invoking `gh` directly, commands are routed through the tool registry, allowing the tool binary path and environment to be configured centrally
- **Thin wrapper**: the executor adds no business logic beyond routing; all `gh` semantics are preserved

## Internal Structure

| File | Purpose |
|---|---|
| `executor.go` | `ToolRoutedExecutor` implementing `github.CLIExecutor` via tool registry |

## Dependencies

- `core/github` -- `CLIExecutor` interface definition
- `core/tool` -- tool registry for resolving the `gh` binary

## Role in System

Provides the concrete `gh` CLI executor used by commands that interact with GitHub (e.g., CI result fetching, PR creation). By routing through the tool registry, the executor respects tool-config.yml overrides and executor-mode settings.

## Code Health

### Tech Debt
- None identified; compact single-file package

### Pain Points
- None identified

### Optimization Opportunities
- None identified; the package is a thin adapter with minimal surface area
