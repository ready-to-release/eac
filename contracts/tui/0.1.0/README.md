# tui

TUI port interfaces defining the contract between command orchestration
and terminal user interface implementations.

## Key Types

- **`Console`** -- Primary TUI interface (lifecycle, output, status, phases)
- **`ConsoleFactory`** -- Creates `Console` instances from runtime config
- **`Registry`** -- Command-to-TUI binding registry
- **`Config`** -- TUI display, behavior, and command context configuration
- **`TUIConfig`** -- Timeout and layout tuning parameters
- **`Status`** -- Orchestrator status update (capacity, locks, tools)
- **`InitSummary`** -- Structured init phase data for display
- **`SummaryData`** -- Final execution summary for the summary pane
- **`PlannedWorkItem`** -- Predicted work item for skeleton tab display
- **`UoWEnrichment`** -- Incremental enrichment for planned tabs

## Patterns

- Hexagonal ports: `Console` is implemented by bubbletea-based TUI adapters
- Progressive display: `SendPlannedWork` then `EnrichUoW` for streaming updates
- Three-phase lifecycle: Init, Run, Summary with `SetPhase`/`CompletePhase`
- Factory pattern: `ConsoleFactory` enables lazy instantiation with runtime config
- Cache status: `CacheHit` enum for incremental build visualization

## Internal Structure

| File | Responsibility |
| --- | --- |
| ports.go | All interfaces, config structs, enums, and display types |

## Dependencies

- `contracts/core` -- `ActionType` and `TagSummary` value types

## Role in System

The `tui` package (moniker: contracts-tui) is the only contract module that
depends on contracts-core. It defines the display contract that bubbletea
console implementations satisfy, allowing the orchestrator and command
framework to drive TUI output without knowing the rendering details.

## Code Health

### Tech Debt
- `Console` interface in ports.go has 20 methods -- decompose into role interfaces (Lifecycle, Output, PhaseManager, UoWTracker) so adapters can implement subsets
- `InitSummary` struct (~30 fields) is a data gravity well; consider grouping into sub-structs (DepmStatus, IncrementalStatus, TestStatus)

### Pain Points
- ports.go is 429 lines in a single file; splitting types into ports.go, status_types.go, and summary_types.go would improve navigability
- Test coverage only validates `DefaultTUIConfig` defaults -- no compile-time `var _ Console = ...` check to catch interface drift

### Optimization Opportunities
- Split `Console` into 4 role interfaces and compose via embedding -- high impact, moderate effort, unblocks partial adapters
- Add compile-time interface satisfaction check in ports_test.go -- trivial effort, catches breakage early
