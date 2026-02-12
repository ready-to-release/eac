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
| model.go | `Model` struct, `NewModel`, `Init`, and channel listeners |
| model_boot.go | `ModelBootState`, `BootState`, and `ConfigMeta` types |
| model_display.go | `DisplayConfig`, `PaneHeights`, and pane height calculations |
| model_execution.go | `ExecutionState`, UoW types/methods, phase helpers, lock formatting |
| model_interaction.go | `InteractionState`, `ResourceState`, tab/pane sizing, cached metrics |
| update.go | `Update` method handling all message types |
| mouse.go | Mouse event handling: tab detection, scroll, text selection, and hover tracking |
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
- update.go (891 lines) exceeds 300 lines; candidate for splitting into separate message handler files
- derived_test.go (635 lines) exceeds 300 lines; candidate for splitting into focused test suites
- model_execution.go (533 lines) exceeds 300 lines; candidate for splitting execution state and UoW management
- view_logs.go (503 lines) exceeds 300 lines; candidate for splitting log rendering logic
- catalog_test.go (469 lines) exceeds 300 lines; candidate for splitting into focused widget test files
- model_test.go (467 lines) exceeds 300 lines; candidate for splitting into focused test suites
- view_helpers_test.go (423 lines) exceeds 300 lines; candidate for splitting helper test coverage
- catalog_widgets.go (417 lines) exceeds 300 lines; candidate for splitting individual widgets into separate files
- mouse.go (389 lines) exceeds 300 lines; candidate for splitting click, scroll, and hover handling
- view_detail_test.go (333 lines) exceeds 300 lines; candidate for splitting detail view test coverage

### Pain Points
- None identified

### Optimization Opportunities
- None identified
