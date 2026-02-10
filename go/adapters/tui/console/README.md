# console

Core bubbletea model and rendering logic for the parallel TUI, managing
phase-based layout, tabbed output panes, and status visualization.

## Key Types

- **`Model`** -- Bubbletea model holding all TUI state
- **`Line`** -- Output line with source, level, and timestamp
- **`Status`** -- Orchestrator status update (running, completed, locks)
- **`Phase`** -- Execution phase (Init, Run, Summary)
- **`SummaryData`** -- Structured summary for post-execution display
- **`InitSummary`** -- Init phase metadata for structured rendering
- **`ExecutionModule`** -- Module and its units of work for display
- **`PlannedWorkItem`** -- Predicted work item for skeleton tabs

## Patterns

- Elm architecture: `Model`, `Update`, `View` separation via bubbletea
- Message-driven: All state changes arrive as typed `tea.Msg` values
- Phase-based layout: Three collapsible phases (Init, Run, Summary)
- Ring buffer: Fixed-size output buffers per tab for memory control

## Internal Structure

| File | Responsibility |
| --- | --- |
| model.go | `Model` struct, `NewModel`, and `Init` method |
| update.go | `Update` method handling all message types |
| view.go | Main `View` method composing layout sections |
| view_init.go | Init phase rendering |
| view_tabs.go | Tab bar rendering with status indicators |
| view_logs.go | Log output pane rendering |
| view_detail.go | Detail panel for selected tab |
| view_resources.go | Resource/lock status visualization |
| view_selected.go | Selected tab content rendering |
| messages.go | Message types for state transitions |
| phase.go | Phase state machine and transitions |
| buffer.go | Ring buffer for per-tab output storage |
| styles.go | Lipgloss style definitions |
| catalog.go | Widget catalog for component rendering |
| catalog_widgets.go | Individual widget implementations |
| derived.go | Derived/computed view state from model |
| widget.go | Base widget types and interfaces |
| render/ | Shared rendering utilities (icons, lamps, styles, text, layout) |

## Dependencies

- `contracts/tui/0.1.0` -- shared type definitions

## Role in System

The `console` sub-package is the rendering engine within `tui-eac`,
responsible for all visual output in the parallel TUI. It receives typed
messages from `ParallelConsole` and translates them into terminal output
through the bubbletea framework, maintaining phase state, tab management,
and output buffering.

## Code Health

### Tech Debt
- `Update()` in update.go (~530 lines) is a single switch with 20+ case arms; extracting handler methods per message type would improve readability
- `handleMouse()` in update.go (~310 lines) inlines tab detection, scroll handling, text selection, and hover tracking; splitting into focused helpers would reduce complexity
- `Model` in model.go nests four large state structs (Boot, Display, Execution, Interaction); model.go is ~1160 lines total

### Pain Points
- `NewModel()` in model.go takes 8 positional parameters; a config/options struct would be clearer and more extensible
- Package-level `var log` in model.go couples to `core/logging` at import time

### Optimization Opportunities
- The `calculatePaneHeights()` logic (~70 lines) re-derives structured vs buffer mode each call; caching the result alongside layout metrics would avoid redundant computation
- Extracting mouse hit-testing (detectTabAt, detectResourceZoneAt closures) into reusable methods would allow unit testing without full bubbletea scaffolding
