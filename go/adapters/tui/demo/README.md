# demo

Experimental TUI layout prototype (tui3) using modular cell components for a grid-based parallel task visualization.

## Key Types

- **`Model`** -- Bubbletea model holding all TUI state including units, metrics, and output
- **`Cells`** -- Container holding all cell component instances (Timer, CPU, Mem, Selector, Output, etc.)
- **`Layout`** -- Calculated layout dimensions for rendering sections

## Key Functions

- `NewModel` -- Creates a new tui3 model with initial cell configuration

## Patterns

- Elm architecture: `Model`, `Update`, `View` separation via bubbletea
- Cell composition: 15 independent cell components arranged in a grid layout
- Message-driven: All state changes arrive as typed messages (TickMsg, UnitStartMsg, etc.)
- Metrics caching: CPU/memory metrics refreshed every 500ms to avoid expensive gopsutil calls
- Single source of truth: UoW stats derived from unit state during `updateCells()`

## Internal Structure

| File | Responsibility |
| --- | --- |
| model.go | `Model` struct, `Cells` struct, `NewModel`, cell initialization, state setters, metrics refresh |
| update.go | `Update` method handling all message types with helper functions |
| view.go | `View` method composing resources, command, selector, output, and summary panes |
| messages.go | Message types for ticks, unit lifecycle, metrics, tools, summary, and quit |

## Dependencies

- `adapters/tui/demo/cells` -- Cell component implementations

## Role in System

The demo sub-package is an experimental redesign of the parallel TUI using a modular cell-based architecture. Each visual element (timer, CPU, memory, selector, output) is an independent cell component with its own `Render` method, composed into a grid layout. This prototype explores an alternative to the monolithic console model.

## Code Health

### Tech Debt
- None

### Pain Points
- `model.go` (~340 lines) mixes cell initialization, state setters, and metrics refresh logic; a dedicated `state` helper would reduce the file size

### Optimization Opportunities
- None identified
