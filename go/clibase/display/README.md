# display

Interface definitions for TUI console implementations. Defines the contract between
the command framework and concrete TUI renderers without depending on specific rendering technology.

## Key Types

- `Console` -- primary interface for TUI lifecycle: start, stop, send events, update status (~22 methods covering lifecycle, output, phase management, UoW tracking, and summary)
- `Config` -- TUI configuration: height, buffer size, ASCII mode, TUI3 demo, command context
- `ConsoleFactory` -- function type that creates `Console` implementations
- `Phase` -- execution phase enumeration (Init, Run, Summary) with display-name `String()` method
- `Level` -- output severity level enumeration (Info, Warn, Error)
- `SummaryData` -- structured summary data sent to TUI for final display
- `InitSummary` -- init-phase summary data for TUI boot panel with flags, deps, incremental, parallelism, and test info
- `TUIBootstrap` -- configuration bundle for TUI startup: factory functions for creating Console, Observer, and Hooks instances
- `ExecutionModule` -- module entry for TUI display with status and component list
- `UoWEntry` -- unit of work entry for TUI progress tracking with tags, tool, and container info
- `PlannedWorkItem` -- predicted work item sent before tool resolution for skeleton tab display
- `UoWEnrichment` -- enriches a planned work item with resolved tool metadata and cache status
- `CacheHit` -- cache status enumeration (Unknown, Miss, HitFresh)
- `PlannedTool` -- tool metadata for TUI display
- `Line` -- formatted line for TUI output panels with source, level, and timestamp
- `Status` -- execution status update with running/completed counts, lock states, and tool tracking
- `LockStatus` -- state of a single lock for TUI display

## Key Functions

- `SetTUIBootstrap` -- registers TUI factory functions (called by adapters/tui during initialization)
- `GetTUIBootstrap` -- returns the registered TUI bootstrap, or nil if none set

## Patterns

- **Interface-based decoupling**: `Console` defines the TUI contract without importing any rendering library, enabling multiple TUI implementations
- **Factory pattern**: `ConsoleFactory` and `TUIBootstrap` allow the framework to create TUI instances without knowing the concrete type
- **Progressive enrichment**: planned work items are sent as grey skeletons, then enriched with tool metadata as resolution completes
- **Global bootstrap**: `TUIBootstrap` is registered once at init time by the adapter package, avoiding import cycles

## Internal Structure

| File | Purpose |
|---|---|
| `interfaces.go` | All interface and type definitions for TUI abstraction, plus `TUIBootstrap` global registration |

## Dependencies

- `contracts/core` -- action types and work unit port definitions
- `core/workunit` -- unit spec definitions for work item display

## Role in System

Defines the boundary between the command framework and TUI rendering. The framework sends events through the `Console` interface, and concrete implementations (in the adapters/tui package) translate those events into visual output. This separation allows TUI technology to change without affecting command logic.

## Code Health

### Tech Debt

- None identified

### Pain Points

- `interfaces_test.go` is 614 lines, significantly exceeds 300-line threshold

### Optimization Opportunities

- None identified
