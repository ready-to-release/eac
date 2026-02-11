# output

Console output formatting for non-TUI execution mode.
Provides a plain-text observer for execution events and formatting utilities for result display.

## Key Types

- `ConsoleObserver` -- implements the execution observer pattern for plain-text console output; receives start, progress, and completion events from the orchestrator

## Key Functions

- `ResultLine` -- formats a single execution result as a status line with pass/fail indicator, duration, and module name
- `PackageDisplayName` -- extracts a human-readable display name from a full package path
- `PackageDisplayNames` -- batch version of `PackageDisplayName` for multiple monikers
- `FormatComponentName` -- formats a component type for display (e.g., shortening long names)
- `ListFormat` -- formats a list of items with configurable separators and truncation
- `SectionHeader` -- formats a section header string for console output
- `SummaryCount` -- formats a summary count line (e.g., "Modules: 10 passed, 2 failed")
- `TimingLine` -- formats a timing entry for the timing summary
- `TimingTotal` -- formats the total timing line

## Patterns

- **Observer pattern**: `ConsoleObserver` receives orchestrator events without coupling the orchestrator to console output concerns
- **Platform-aware formatting**: uses platform-specific path separators and terminal width detection

## Internal Structure

| File                  | Purpose                                                              |
| --------------------- | -------------------------------------------------------------------- |
| `console_observer.go` | `ConsoleObserver` implementing plain-text event observation          |
| `format.go`           | Section headers, list formatting, component name formatting          |
| `display_name.go`     | `PackageDisplayName` and `PackageDisplayNames` extraction utilities  |
| `result_line.go`      | `ResultLine`, `SummaryCount`, `TimingLine`, `TimingTotal` formatting |

## Dependencies

- `contracts/core` -- action type definitions for status formatting
- `core/platform` -- terminal width detection and platform utilities

## Role in System

Provides the non-TUI output path for command execution. When TUI mode is disabled, the orchestrator sends events to `ConsoleObserver` which renders them as plain-text status lines. The formatting utilities are shared between console and summary output.

## Code Health

### Tech Debt

- None identified

### Pain Points

- None identified

### Optimization Opportunities

- Good test coverage exists (`format_test.go` 307 lines, `console_observer_test.go` 245 lines); no urgent gaps
