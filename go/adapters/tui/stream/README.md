# stream

Output stream filtering and multi-writer utilities for the TUI console, handling subprocess output capture and line classification.

## Key Types

- **`Filter`** -- Determines which output lines to display using important/noise patterns and deduplication
- **`OutputPipe`** -- Captures subprocess output via `io.Pipe` and streams classified lines to a channel
- **`MultiWriter`** -- Combines an `OutputPipe` with additional `io.Writer` destinations

## Key Functions

- `NewFilter` -- Creates a filter with default Go build/test patterns
- `NewOutputPipe` -- Creates a pipe that streams filtered lines to a channel
- `NewMultiWriter` -- Creates a writer combining an output pipe with additional writers

## Patterns

- Pattern-based filtering: Important patterns (errors, panics) always shown; noise patterns (downloads, test markers) always hidden
- Deduplication: Tracks recent lines and suppresses after 2 repetitions
- Line classification: `classifyLine` assigns LevelInfo, LevelWarn, or LevelError based on content analysis
- Summary-aware: `isSummaryLine` prevents false error coloring on statistics lines like "0 errors"
- Graceful shutdown: `OutputPipe.Close` switches to blocking sends to flush final error output before completion
- Non-blocking writes: Normal operation uses non-blocking channel sends to avoid stalling workers

## Internal Structure

| File      | Responsibility                                                                                        |
| --------- | ----------------------------------------------------------------------------------------------------- |
| filter.go | `Filter` with important/noise regex patterns and deduplication tracking                               |
| pipe.go   | `OutputPipe` for subprocess capture, `classifyLine` for severity detection, `MultiWriter` for fan-out |

## Dependencies

- `contracts/tui/0.1.0` -- `Line` and `Level` types for output classification

## Role in System

The stream sub-package sits between subprocess execution and the TUI console, filtering noise from build/test output and classifying lines by severity. It ensures that error output is highlighted in red, warnings in orange, and routine output in the default color, while suppressing verbose download messages and test markers that would clutter the display.

## Code Health

### Tech Debt

- pipe.go (321 lines) exceeds 300 lines; candidate for splitting output classification and multiwriter logic

### Pain Points

- None identified

### Optimization Opportunities

- None identified
