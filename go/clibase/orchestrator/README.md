# orchestrator

Parallel execution engine for build, test, lint, and scan commands. Manages
worker goroutines, weighted capacity scheduling, TUI integration, and
serialized console output to prevent interleaved terminal writes.

## Key Types

- **`Orchestrator`** -- Top-level coordinator for parallel module execution; manages observers, TUI console, and delegates to `UnitScheduler`
- **`UnitScheduler`** -- Component-level scheduler with dual-pool weighted semaphore, LPT dispatch, background cache detection, and tool tracking
- **`WeightedSemaphore`** -- Capacity-based concurrency primitive with lock-tracker integration
- **`Config`** -- Execution settings (concurrency, turbo, TUI, dry-run, action type)
- **`ConfigUpdate`** -- Partial config update for deferred configuration after orchestrator creation
- **`WorkerFunc`** -- Signature for module-level workers
- **`UnitWorkerFunc`** -- Signature for component-level workers
- **`WorkItem`** -- Module-level work item with moniker and index
- **`WorkResult`** -- Module-level execution outcome with exit code, duration, warnings, errors
- **`UnitResult`** -- Component-level execution outcome with test-specific fields
- **`UnitExtras`** -- Additional data (test counts) passed from workers to unit results
- **`ModuleResultSet`** -- Aggregated unit results per module with derived status
- **`ModuleStatus`** -- Execution status enumeration (Pending, Running, Success, Skipped, Failed)
- **`SummaryBuilder`** -- Interface for incremental summary computation
- **`CacheVerifier`** -- Alias for `execution.CacheVerifier` for background cache detection
- **`EarlyCacheInfo`** -- Cache verification result stored for worker short-circuit
- **`LogEvent`** -- Structured log event compatible with `go test -json`
- **`JSONLogWriter`** -- Structured JSON log output for non-test workers

## Patterns

- Single-writer display: `displayManager` is the only goroutine that writes
  to stdout, preventing interleaved output on Windows
- Observer pattern: `Orchestrator` emits `core.ExecutionEvent` to registered
  `core.ExecutionObserver` instances for TUI and summary updates
- Dynamic capacity: a ticker recalculates semaphore capacity every 2s from
  live Docker/WSL/host memory stats
- LPT scheduling: `UnitScheduler` dispatches heaviest-first via
  `scheduling.DependencyScheduler` with dependency-aware ordering
- Dual-pool semaphore: separate host and Docker capacity pools via
  `capacity.DualPoolSemaphore` for cross-process coordination
- Background cache detection: parallel goroutines verify cache status and
  short-circuit workers that find early-cached items
- Cascade failure: when a unit fails, all transitive dependents are immediately
  marked as failed without execution

## Internal Structure

| File                        | Responsibility                                   |
| ---                         | ---                                              |
| `types.go`                  | Core type definitions, status enum, `ModuleResultSet`, `WorkResult` |
| `orchestrator_core.go`      | `Orchestrator` struct, `New`, `Run`, `RunLayered`, `RunUnitsParallel`, result aggregation |
| `orchestrator_display.go`   | TUI lifecycle, observer dispatch, `SetConsole`, `StartTUI`, `StopTUI`, `AddObserver` |
| `orchestrator_phases.go`    | Phase start/complete/write event emission         |
| `display.go`                | `displayManager` single-writer console loop, status formatting |
| `unit_scheduler_core.go`    | `UnitScheduler`, worker pool, LPT dispatch, `RunUnits`, result aggregation helpers |
| `unit_scheduler_capacity.go`| Dynamic capacity calculation ticker               |
| `unit_scheduler_display.go` | TUI status tracking, tool/container lamp state, resource status emission |
| `unit_scheduler_execution.go`| `executeWorker`, background cache detection, log file management |
| `weighted_semaphore.go`     | `WeightedSemaphore` with lock-tracker integration |
| `memory_detection.go`       | Docker/WSL/host memory and CPU detection          |
| `parser.go`                 | Log parsing (JSON events, Cucumber, CTRF)         |
| `newline_unix.go`           | Platform line-ending constant for Unix            |
| `newline_windows.go`        | Platform line-ending constant for Windows         |

## Dependencies

- `contracts/core` -- `ExecutionObserver`, `ExecutionEvent`, `ActionType`, `WriterFactory`
- `clibase/capacity` -- `DualPoolSemaphore` for cross-process scheduling
- `clibase/display` -- `Console` interface, `Phase` enum, `SummaryData`, `PlannedWorkItem`
- `clibase/locktracker` -- `Registry` for semaphore visualization in TUI
- `clibase/output` -- Display-name formatting and section headers
- `clibase/ansi` -- Bad-ANSI filter for log file sanitization
- `core/config` -- Timeout, capacity interval, Docker query context
- `core/workunit` -- `UnitSpec`, `UnitID`, `TagSummary`
- `core/scheduling` -- `DependencyScheduler` (LPT + dependency graph)
- `core/execution` -- `CacheVerifier` interface
- `core/logging` -- Structured logger
- `core/resource` -- `PoolAllocation` for dual-pool weights
- `gopsutil/v3/mem` -- Host memory stats fallback

## Role in System

This package is the execution backbone of the `clibase` module. Every
parallel command (build, test, lint, scan) delegates to `Orchestrator` for
worker management, capacity control, and TUI display. It sits between the
command framework (`cmdframework`) and the per-language adapters, keeping
scheduling and resource concerns out of both layers.

## Code Health

### Tech Debt
- `unit_scheduler_core.go:261` `RunUnits` is ~180 lines; extract dependency-failure cascade and over-capacity handling into helpers
- `orchestrator_core.go:314` `processWorkItem` is ~108 lines; repeated error-result boilerplate could be collapsed
- `unit_scheduler_capacity.go:61` mutable package-level `defaultDetector` var; prefer constructor injection

### Pain Points
- Dual-pool semaphore logic mixed with cascade-failure bookkeeping makes `RunUnits` hard to follow
- The defensive `BUG:` warn at `unit_scheduler_core.go:434` suggests scheduler drain is not fully trusted

### Optimization Opportunities
- Break `RunUnits` worker-pool loop into a `dispatchLoop` method to improve testability (low effort)
- Replace `resultsMu` mutex-guarded slice with indexed atomic writes since each worker writes to a unique index (medium effort)
