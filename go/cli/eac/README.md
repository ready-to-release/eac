# eac

Entry point and command dispatcher for the `eac` CLI. Bootstraps dependency
providers, resolves the command path from `os.Args`, and dispatches to the
matching `registry.CommandFunc`.

## Key Types

- **`CommandFunc`** -- Signature for all registered commands (via `registry`)
- **`InitialWorkingDir`** -- Package-level var capturing launch directory

## Patterns

- Init-based wiring: `container_init.go`, `tui_init.go`, and `imports.go` use
  blank imports and `init()` to register adapters before `main` runs
- Longest-match routing: dispatcher walks `os.Args` from longest to shortest
  to resolve nested commands like `show files changed`
- Panic recovery: top-level `defer/recover` in `main()` prints full stack
  traces with platform-correct line endings

## Internal Structure

| File               | Responsibility                                    |
| ---                | ---                                               |
| main.go            | Arg parsing, longest-match dispatch, panic handler |
| container_init.go  | Wires Docker adapter as default container provider |
| tui_init.go        | Registers parallel and selector TUI factories      |
| imports.go         | Blank-imports all command implementation packages   |
| custom_renderers.go| Blank-imports custom render extensions              |

## Dependencies

- `clibase/registry` -- command registration and lookup
- `clibase/ghexec` -- GitHub CLI executor for report commands
- `core/config` -- repository configuration loading
- `core/git` -- go-git based remote URL resolution
- `core/domain/reports` -- report generation with GitHub data
- `core/environments` -- environment variable constants
- `cli/eac/help` -- unified help printer for commands
- `adapters/docker` -- container provider wired via `init()`
- `adapters/tui/parallel` -- TUI console for build/test/lint/scan
- `adapters/tui/selector` -- TUI selector for interactive commands

## Role in System

This package is the `main` binary of the `eac` module. It contains no
business logic itself; it wires adapters and delegates to command implementations
under `cli/eac/impl/`. All command packages self-register via blank imports in
`imports.go`, keeping the dispatcher decoupled from individual commands.

## Code Health

### Tech Debt
- `main.go`: `main()` is ~147 lines (37-184) with inlined help-flag detection and duplicate longest-match resolution logic
- `main.go:27`: `InitialWorkingDir` is exported mutable global state; consider injecting via a context or config struct

### Pain Points
- Help-flag interception logic (lines 112-138) duplicates the longest-match loop from the main dispatch path
- Test file (`main_test.go`) is gated behind `L1 && ov` build tags, making it easy to skip inadvertently

### Optimization Opportunities
- Extract help resolution and dispatch into separate functions to reduce `main()` complexity -- low effort, high readability gain
