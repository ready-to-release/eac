# ansi

ANSI escape sequence filtering for clean text output.
Strips unwanted ANSI codes from subprocess output before display or logging.

## Key Types

- `Filter` -- wraps an `io.Writer` and strips ANSI sequences from written data before forwarding
- `FilterMode` -- controls filtering behavior: strip all sequences or strip only problematic ones

## Key Functions

- `NewFilter` -- creates a `Filter` with a specified mode wrapping a target writer
- `NewBadOnlyFilter` -- creates a filter that strips only problematic sequences (cursor movement, screen clearing) while preserving colors
- `NewStripAllFilter` -- creates a filter that strips all ANSI escape sequences
- `StripBad` -- removes problematic ANSI sequences from a byte slice
- `StripAll` -- removes all ANSI escape sequences from a byte slice
- `StripAllString` -- removes all ANSI escape sequences from a string

## Patterns

- **Writer wrapping**: `Filter` implements `io.Writer`, allowing transparent insertion into any output pipeline
- **Selective filtering**: "bad only" mode preserves color codes while removing cursor and screen manipulation sequences that corrupt TUI layouts

## Internal Structure

| File | Purpose |
|---|---|
| `writer.go` | `Filter`, `FilterMode`, constructors, and stripping functions |

## Dependencies

None (leaf package within clibase).

## Role in System

Inserted into the output pipeline between subprocess stdout/stderr and the TUI or log files. Subprocess output often contains ANSI sequences that would corrupt TUI rendering or make log files unreadable; this package removes them while optionally preserving color information for display.

## Code Health

### Tech Debt
- `writer.go:92` `badPatterns` is a package-level mutable slice of compiled regexes; safe in practice but could be wrapped in a `sync.Once` for clarity

### Pain Points
- None identified; compact leaf package with good test coverage (195 test lines vs 221 source lines)

### Optimization Opportunities
- Consider pre-compiling a single combined regex from `badPatterns` to reduce per-line matching overhead on high-throughput streams (medium effort, measure first)
