# selector

Bubbletea-based subcommand selector TUI for interactive command picking with keyboard navigation and argument input.

## Key Types

- **`Model`** -- Bubbletea model for the selector with command list, text input, and navigation state
- **`Console`** -- Implements `tui.SelectorConsole` with `SetCommands` and `Run` methods

## Key Functions

- `New` -- Creates a new Selector Console
- `Factory` -- Returns a `tui.SelectorFactory` for creating selector instances
- `RunSelector` -- Convenience function to run the selector directly
- `RunSelectorWithSubcommands` -- Convenience function using `SubcommandInfo`
- `ShouldUseSelector` -- Checks if the selector should be shown (interactive terminal, not CI)
- `PrintSelection` -- Formats a message about the selection result

## Patterns

- Factory registration: `init()` registers the selector factory with the TUI system
- Vim navigation: Supports j/k keys in addition to arrow keys
- Dual focus: Tab toggles between command list and argument text input
- Alt-screen: Runs in alternate screen buffer with mouse support
- Context cancellation: Supports graceful shutdown via context

## Internal Structure

| File | Responsibility |
| --- | --- |
| selector.go | `Model` bubbletea implementation, `Console` implementing `SelectorConsole`, and convenience functions |
| init.go | `init()` registering selector factory with the TUI system |

## Dependencies

- `adapters/tui` -- `SelectorConsole` interface, `CommandOption`, and `ShouldUseTUI`

## Role in System

The selector sub-package provides the interactive command picker TUI that appears when users run `eac` without specifying a subcommand in an interactive terminal. It displays available commands with descriptions, supports keyboard navigation and argument entry, and returns the user's selection to the caller for execution.

## Code Health

### Tech Debt
- None

### Pain Points
- None identified

### Optimization Opportunities
- None identified
