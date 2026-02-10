# cells

Modular TUI cell components for the tui3 grid layout, each representing a discrete visual element.

## Key Types

- **`Cell`** -- Interface for renderable TUI components with `Render` and `ZoneID`
- **`DataSetter`** -- Interface for cells that accept external data updates
- **`TimerCell`** -- Displays elapsed time in MM:SS format
- **`CPUCell`** -- Displays per-core CPU usage as dots
- **`MemCell`** -- Displays memory usage as an 11-dot bar
- **`DockerMemCell`** -- Displays Docker memory pool usage as an 11-dot bar
- **`ContainersCell`** -- Displays container status as active/planned dots
- **`UoWStatsCell`** -- Displays unit of work statistics with lamps and counts
- **`ToolsCell`** -- Displays tool status as container and system dots
- **`CommandCell`** -- Displays the synthetic command line with module list
- **`SelectorCell`** -- Displays a grid of selectable units with weight badges
- **`SelectedCell`** -- Displays information about the currently selected unit
- **`DepsCell`** -- Displays dependencies for the selected unit
- **`CacheCell`** -- Displays cache hit/miss status
- **`ArtifactsCell`** -- Displays build artifacts with formatted sizes
- **`OutputCell`** -- Displays log output with scrolling support
- **`SummaryCell`** -- Displays final execution summary statistics
- **`SummaryData`** -- Summary statistics (total, succeeded, failed, skipped, duration)
- **`SelectorUnit`** -- Unit entry in the selector grid with moniker, status, and weight
- **`UnitStatus`** -- Enum for unit execution states (Pending, Queued, Running, Complete, Skipped, Failed)
- **`OutputLine`** -- Single output line with text, source, and severity level

## Patterns

- Cell interface: Each component implements `Render(width, height)` for dimension-aware rendering
- ASCII fallback: Cells that use Unicode symbols support an `asciiMode` toggle
- Stateless rendering: Cells receive data via setters and render based on current state
- Zone-based interaction: Each cell declares a mouse zone ID for click handling

## Internal Structure

| File | Responsibility |
| --- | --- |
| cell.go | `Cell` and `DataSetter` interfaces |
| types.go | Shared types: `UnitStatus`, `SelectorUnit`, `OutputLine`, `DependencyInfo`, `ArtifactInfo` |
| timer.go | `TimerCell` for elapsed time display |
| cpu.go | `CPUCell` for per-core CPU usage dots |
| mem.go | `MemCell` for memory usage bar |
| docker_mem.go | `DockerMemCell` for Docker memory usage bar |
| containers.go | `ContainersCell` for container status dots |
| uow_stats.go | `UoWStatsCell` for work queue lamps and counts |
| tools.go | `ToolsCell` for container and system tool dots |
| command.go | `CommandCell` for command line with module list |
| selector.go | `SelectorCell` for unit grid with weight badges and tab wrapping |
| selected.go | `SelectedCell` for selected unit details |
| deps.go | `DepsCell` for dependency information |
| cache.go | `CacheCell` for cache hit/miss status |
| artifacts.go | `ArtifactsCell` for build artifact listing |
| output.go | `OutputCell` for scrollable log output |
| summary.go | `SummaryCell` for execution summary with `SummaryData` |

## Dependencies

- None (leaf package; imports only lipgloss for styling in selector.go)

## Role in System

The cells sub-package provides the building blocks for the tui3 demo layout. Each cell is a self-contained visual component that can be tested independently. The parent demo package composes these cells into the full grid layout and drives their state through the bubbletea update cycle.

## Code Health

### Tech Debt
- None

### Pain Points
- None identified

### Optimization Opportunities
- None identified
