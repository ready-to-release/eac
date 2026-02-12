# parallel

Parallel task TUI factory and registration, wrapping `ParallelConsole` to implement the `tui.Console` interface.

## Key Types

- **`Console`** -- Wraps `tui.ParallelConsole` to implement the `tui.Console` interface

## Key Functions

- `New` -- Creates a new parallel Console with the given configuration
- `Factory` -- Returns a `ConsoleFactory` for the parallel TUI

## Patterns

- Delegation: All `Console` methods delegate to the inner `ParallelConsole`
- Factory registration: `init()` registers the parallel factory for build, test, lint, scan, and update ai-summary commands
- Interface compliance: Compile-time check ensures `Console` implements `tui.Console`

## Internal Structure

| File | Responsibility |
| --- | --- |
| console.go | `Console` wrapper delegating all methods to `ParallelConsole` |
| init.go | `init()` registering parallel TUI for build, test, lint, scan, and update ai-summary commands |

## Dependencies

- `adapters/tui` -- `ParallelConsole`, `Console` interface, and type definitions

## Role in System

The parallel sub-package provides the concrete TUI implementation used by all parallel-execution commands (build, test, lint, scan). It registers itself during `init()` so that the TUI registry can create parallel consoles for these commands without the caller knowing the concrete type.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
