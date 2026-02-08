# display

Interface definitions for TUI console implementations. Defines the contract between
the command framework and concrete TUI renderers without depending on specific rendering technology.

## Key Types

- `Console` -- primary interface for TUI lifecycle: start, stop, send events, update status
- `Config` -- TUI configuration: height, ASCII mode, debug mode, command path
- `ConsoleFactory` -- function type that creates `Console` implementations
- `Phase` -- execution phase enumeration for TUI state transitions
- `SummaryData` -- structured summary data sent to TUI for final display
- `InitSummary` -- init-phase summary data for TUI boot panel
- `TUIBootstrap` -- configuration bundle for TUI startup: module list, component types, planned work
- `ExecutionModule` -- module entry for TUI display with status and component list
- `UoWEntry` -- unit of work entry for TUI progress tracking
- `PlannedWorkItem` -- predicted work item sent before tool resolution for skeleton tab display
- `UoWEnrichment` -- enriches a planned work item with resolved tool metadata
- `CacheHit` -- notifies TUI that a work item was satisfied from cache
- `PlannedTool` -- tool metadata for TUI display
- `Line` -- formatted line for TUI output panels
- `Status` -- execution status enumeration for TUI indicators

## Patterns

- **Interface-based decoupling**: `Console` defines the TUI contract without importing any rendering library, enabling multiple TUI implementations
- **Factory pattern**: `ConsoleFactory` allows the framework to create TUI instances without knowing the concrete type
- **Progressive enrichment**: planned work items are sent as grey skeletons, then enriched with tool metadata as resolution completes

## Internal Structure

| File | Purpose |
|---|---|
| `interfaces.go` | All interface and type definitions for TUI abstraction |

## Dependencies

- `contracts/core` -- action types and work unit port definitions
- `core/workunit` -- unit spec definitions for work item display

## Role in System

Defines the boundary between the command framework and TUI rendering. The framework sends events through the `Console` interface, and concrete implementations (in the orchestrator and render packages) translate those events into visual output. This separation allows TUI technology to change without affecting command logic.

## Code Health

### Tech Debt
- `interfaces.go:22` `Console` interface has ~22 methods; this is a very large interface that forces implementors to satisfy all methods even if only a subset is needed
- No test file exists; while this is a types-only package, interface compliance tests would catch contract drift

### Pain Points
- Adding a new TUI feature requires modifying the `Console` interface and updating every implementation, creating high coupling
- The interface mixes lifecycle (`Start`/`Stop`), output (`SendLine`/`WriteResult`), phase management, UoW tracking, and summary concerns

### Optimization Opportunities
- Decompose `Console` into smaller role interfaces (e.g., `PhaseManager`, `UoWTracker`, `SummaryConsole`) composed by implementors (medium effort)
- Add interface compliance test (`var _ Console = (*mockConsole)(nil)`) to catch signature drift early (low effort)
