# cmdframework

Shared framework for orchestrated CLI commands (build, test, scan, lint, deploy).
Extracts common execution patterns into reusable phases: init, resolve, verify, execute, and summary.

## Overview

Commands define their specific worker functions and hooks, then delegate to `Run()` which handles the full lifecycle from config loading through parallel execution to result summarization. This eliminates duplicated orchestration logic across command implementations.

The framework sits above the low-level `clibase/orchestrator` package, which handles goroutine management, capacity-aware scheduling, and terminal output serialization. This package provides the higher-level command lifecycle: configuration, module resolution, dependency verification, and summary generation.

## Execution Phases

`Run()` orchestrates five sequential phases, each populating `ExecutionContext`:

1. **Init** (`phaseInitEarly` + `phaseInitDeferred`) -- TUI bootstrap, config loading, tool system initialization. Early init completes in under 10ms so the TUI appears immediately; deferred init loads config and tools while the TUI shows a loading animation.
2. **Resolve** (`phaseResolve`) -- Module discovery, moniker resolution, skip filters, scope expansion. When `IncludeDepm` is set, requested monikers are expanded to include transitive module dependencies.
3. **Verify** (`phaseVerify`) -- System dependency checks (async, started during init), artifact validation for commands that depend on build output.
4. **Execute** (`phaseExecute`) -- Dispatches work to the orchestrator. Supports two modes: module-level via `CommandWorkerFunc` or component-level via `UnitWorkerFunc` with dependency-aware scheduling.
5. **Summary** (`phaseSummary`) -- TUI and console summary generation, exit code calculation. The `AfterExecute` hook runs concurrently with summary display.

## Key Types

### Configuration

- **`CommandConfig`** -- All configuration for a command execution: action type, monikers, concurrency, TUI settings, cache config, and command-specific typed configs (stored as `interface{}` to avoid import cycles).
- **`CommandConfig.ActionVerb()`** -- Returns present-continuous verb (e.g. "Building") from the `ActionDescriptor` registry.

### Execution Context

- **`ExecutionContext`** -- Populated by framework phases; provides access to loaded config, workspace root, module state, orchestrator, tool system, results, TUI hooks, output buffer, and cleanup hooks.
- **`InitTimings`** -- Tracks duration of initialization phases (config load, tool init, module discovery, execution order, change detection, deps verify) for boot-style status output.

### Worker Functions

- **`CommandWorkerFunc`** -- Processes a single module. Signature: `func(ctx context.Context, ectx *ExecutionContext, moniker string, logWriter io.Writer) int`.
- **`UnitWorkerFunc`** -- Processes a single work unit (module + component + tool). Signature: `func(ctx context.Context, ectx *ExecutionContext, spec core.UnitSpec, logWriter io.Writer) int`.

### Hooks

- **`Hooks`** -- Customization points: `AfterInit`, `AfterResolve`, `BeforeExecute`, `AfterExecute`, `CustomSummary`. Each is a `PhaseHook` (`func(*ExecutionContext) error`).
- **`ErrInformationalExit`** -- Returned from hooks to signal exit code 1 without additional error logging (for expected validation failures like missing dependencies).

### Unit Execution

- **`UnitWorkProvider`** -- Converts `ExecutionContext` into a flat list of `workunit.UnitSpec`; scheduling order is determined by `DependsOn` constraints.
- **`UnitRegistry`** -- Thread-safe registry mapping `ActionType` to providers and workers.
- **`UnitPipeline`** -- Per-unit execution steps shared by all commands: cache checking (with optional artifact integrity validation), lock acquisition, and UoW manifest recording.
- **`LockStyle`** -- Controls lock behavior: `NoLock` (test), `LockUnlessDryRun` (build, lint), `AlwaysLock` (scan).

### Summary

- **`SummaryBuilder`** -- Incremental summary builder that receives results as components complete during execution.

## Concurrency Model

Concurrency is determined by a priority chain:

1. `--sequential` forces concurrency to 1
2. CLI `--concurrency` flag sets an explicit limit
3. Repository config `parallelism` ceiling (CI vs devbox)
4. Dynamic mode (0) -- the orchestrator calculates capacity from CPU and RAM

Turbo mode applies a 1.25x multiplier to the pressure roof capacity.

## Non-Blocking TUI Boot

The init phase is split so the TUI renders immediately:

- `phaseInitEarly` (<10ms): creates TUI console, output buffer, minimal orchestrator, wires observers, starts TUI.
- `phaseInitDeferred` (200-800ms): loads config, initializes tools, creates output directory. Updates the orchestrator via `UpdateConfig()`. If this fails with the TUI running, it sends an error and summary to the TUI before exiting.

Progressive boot states (BootChrome -> BootMetrics -> BootConfig -> BootTabs -> BootRunning) let the TUI show OFF lamps and zero counters during loading, then populate real data as it becomes available.

## Internal Structure

| File | Purpose |
|---|---|
| `types.go` | Core type definitions: `CommandConfig`, `ExecutionContext`, `Hooks`, worker function signatures |
| `framework.go` | Entry points `Run()` and `RunSimple()` that orchestrate the five phases |
| `init.go` | `phaseInitEarly()` and `phaseInitDeferred()`: config loading, orchestrator setup, TUI bootstrap |
| `resolve.go` | `phaseResolve()`: module discovery, moniker resolution, skip filters, scope expansion |
| `pipeline.go` | `UnitPipeline`, `LockStyle`, `CheckCache`, `AcquireLock`, `RecordManifest`, `ParseUnit` |
| `execute.go` | `phaseExecute()`: work unit creation, dependency injection, orchestrator dispatch |
| `verify.go` | `phaseVerify()`: dependency verification, artifact validation, incremental detection |
| `summary.go` | `phaseSummary()`: TUI and console summary generation, exit code calculation |
| `summary_builder.go` | `SummaryBuilder`: incremental summary accumulation during execution |
| `summary_helpers.go` | Shared `moduleCache` iteration helpers for TUI and console summary paths |
| `manifest_assert.go` | `AssertManifestsExist()`: validates that required UoW manifest files exist after execution |
| `orchestrated_command.go` | Registration helpers for the unit provider/worker registry |

## Dependencies

- `contracts/core` -- action types, TUI hooks, output buffer port, `UnitSpec`
- `clibase/display` -- console interface, TUI bootstrap, planned work types
- `clibase/initsummary` -- summary data structures for init phase reporting
- `clibase/locking` -- file-based unit locks with wait/retry
- `clibase/orchestrator` -- parallel work dispatch, capacity scheduling, result collection, TUI coordination
- `clibase/output` -- console observer for non-TUI mode
- `core/config` -- EAC config, repository config, component types
- `core/domain/modules` -- module registry
- `core/domain/reports` -- module contract loading
- `core/output` -- UoW manifest tracker and validation
- `core/tool` -- tool system (registry, executor, bridges)
- `core/workunit` -- unit spec definitions, unit IDs
