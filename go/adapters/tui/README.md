# tui

TUI adapter providing terminal user interface implementations for
interactive CLI commands, including parallel task visualization and
subcommand selection.

## Key Types

- **`ParallelConsole`** -- Rich multi-pane TUI for build/test output
- **`Console`** -- Primary interface (re-exported from contracts)
- **`ConsoleFactory`** -- Creates `Console` instances from configuration
- **`TUIObserver`** -- Translates execution events to TUI operations
- **`TUIHooksImpl`** -- Bridges command framework to TUI components
- **`SelectorConsole`** -- Minimal TUI for subcommand selection
- **`ExitHoldController`** -- Controlled exit timing for user interaction
- **`CommandOption`** -- Selectable command in the selector TUI

## Patterns

- Registry pattern: Command-to-TUI bindings with exact, prefix, and default matching
- Observer pattern: `TUIObserver` converts `ExecutionEvent` to `Console` calls
- Bootstrap injection: `init()` registers factory functions into `clibase/display`
- Adapter pattern: `consoleAdapter` wraps contract `Console` for `display.Console`
- Async message pump: Two-channel priority system (critical vs regular messages)

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| interfaces.go | Re-exported contract types and constants |
| console.go | `ParallelConsole` implementation with bubbletea |
| adapter.go | `consoleAdapter` bridging contracts to display |
| bootstrap.go | `init()` wiring TUI factories into clibase/display |
| registry.go | Command-to-TUI registry with pattern matching |
| selector.go | `SelectorConsole` interface and factory registry |
| observer.go | `TUIObserver` translating events to TUI calls |
| hooks.go | `TUIHooksImpl` for command selection and UoW data |
| env.go | `IsInteractive()` and `ShouldUseTUI()` detection |
| exit_hold.go | `ExitHoldController` for delayed exit on interaction |
| console/ | Bubbletea model, update, view, and widget rendering |
| stream/ | Output stream filtering and multi-writer utilities |
| parallel/ | Parallel task TUI factory and registration |
| selector/ | Bubbletea-based selector implementation |
| demo/ | Experimental tui3 layout prototype |

## Dependencies

- `contracts/tui/0.1.0` -- `Console`, `ConsoleFactory`, and type definitions
- `contracts/core/0.1.0` -- `ExecutionEvent`, `TUIHooks` interfaces
- `clibase/display` -- `Console` display interface and bootstrap
- `clibase/registry` -- subcommand registry for selector population
- `core/logging` -- structured logging

## Role in System

The `tui-adapter` module implements the TUI contract interfaces, providing
the interactive terminal experience for build, test, lint, and scan
commands. It bridges between the core orchestrator's execution events and
the bubbletea-based rendering, handling parallel task visualization,
phase management, and post-execution summaries within the broader
dependency graph.

## Code Health

### Tech Debt
- `ParallelConsole.Start()` in console.go (~200 lines) handles signal setup, channel wiring, model creation, and the message pump in one method
- `SetInitSummary()` and `UpdateStatus()` in console.go perform verbose field-by-field struct mapping (~30 fields each); a code-gen or reflection approach could reduce boilerplate
- `bootstrap.go` uses `init()` for factory registration, creating implicit import-order coupling with `clibase/display`

### Pain Points
- `globalRegistry` in registry.go is package-level mutable state; tests must be careful about cross-test leakage
- `observer.go` duplicates the struct-mapping pattern (contract types to TUI types) seen in console.go

### Optimization Opportunities
- Extract the async message pump goroutine in `Start()` into a dedicated `messagePump` type for testability; low risk, structural refactor
- The demo/tui3 code path shares branching logic throughout console.go; extracting a `demoConsole` implementing `Console` would isolate experimental code
