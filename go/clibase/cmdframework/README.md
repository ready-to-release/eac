# cmdframework

Shared framework for orchestrated CLI commands (build, test, scan, lint).
Extracts common execution patterns into reusable phases: init, resolve, verify, execute, and summary.

## Key Types

- `CommandConfig` -- holds all configuration for a command execution: action type, monikers, concurrency, TUI settings, cache config, and command-specific typed configs
- `ExecutionContext` -- populated by framework phases; provides access to loaded config, module state, orchestrator, results, and cleanup hooks
- `CommandWorkerFunc` -- processes a single module and returns an exit code
- `UnitWorkerFunc` -- processes a single work unit (module + component) and returns an exit code
- `Hooks` -- allows commands to customize framework behavior at key points: `AfterInit`, `AfterResolve`, `BeforeExecute`, `AfterExecute`, `CustomSummary`
- `SummaryBuilder` -- incremental summary builder that receives results as components complete
- `UnitWorkProvider` -- interface for commands that provide component-level work units
- `UnitRegistry` -- maps unit keys to their `UnitSpec` for dependency injection during execution
- `ErrInformationalExit` -- signals exit code 1 without additional error logging (for expected validation failures)
- `InitTimings` -- tracks duration of initialization phases for boot-style output
- `UnitPipeline` -- configures per-unit execution steps: lock style, cache check, manifest recording, parsing

## Patterns

- **Phased execution**: `Run()` orchestrates five sequential phases (init, resolve, verify, execute, summary), each populating `ExecutionContext`
- **Two execution modes**: module-level via `CommandWorkerFunc` or component-level via `UnitWorkerFunc` with dependency-aware scheduling
- **Async dependency verification**: system dependency checks run in background during init, results collected before execution
- **Scope expansion**: requested monikers are expanded to include transitive module dependencies when `IncludeDepm` is set
- **Planned work prediction**: sends predicted work items to TUI before tool resolution completes, enabling skeleton tab display
- **Cleanup registration**: `AddCleanup` registers deferred functions executed in reverse order on exit

## Internal Structure

| File | Purpose |
|---|---|
| `types.go` | Core type definitions: `CommandConfig`, `ExecutionContext`, `Hooks`, worker function signatures |
| `framework.go` | Entry points `Run()` and `RunSimple()` that orchestrate the five phases |
| `init.go` | `phaseInitEarly()` and `phaseInitDeferred()`: config loading, orchestrator setup, TUI bootstrap |
| `resolve.go` | `phaseResolve()`: module discovery, moniker resolution, skip filters, scope expansion |
| `pipeline.go` | `UnitPipeline`, `LockStyle`, `CheckCache`, `AcquireLock`, `RecordManifest`, `ParseUnit`, `UnitDir` |
| `execute.go` | `phaseExecute()`: work unit creation, dependency injection, orchestrator dispatch |
| `verify.go` | `phaseVerify()`: dependency verification, artifact validation, incremental detection |
| `summary.go` | `phaseSummary()`: TUI and console summary generation, exit code calculation |
| `summary_builder.go` | `SummaryBuilder`: incremental summary accumulation during execution |
| `manifest_assert.go` | `AssertManifestsExist()`: validates that required manifest files exist |
| `summary_helpers.go` | Shared `moduleCache` iteration helpers for TUI and console summary paths |

## Dependencies

- `contracts/core` -- action types, TUI hooks, output buffer port
- `clibase/display` -- console interface and planned work types
- `clibase/initsummary` -- summary data structures for init phase reporting
- `clibase/orchestrator` -- work dispatch, result collection, TUI coordination
- `clibase/output` -- console observer for non-TUI mode
- `clibase/render` -- summary rendering for TUI display
- `core/config` -- EAC config, repository config, component types, display order
- `core/domain/modules` -- module registry
- `core/domain/reports` -- module contract loading
- `core/workunit` -- unit spec definitions for component-level execution

## Role in System

Acts as the central execution engine for all orchestrated CLI commands. Commands like build, test, scan, and lint define their specific worker functions and hooks, then delegate to `Run()` which handles the full lifecycle from config loading through result summarization. This eliminates duplicated orchestration logic across command implementations.

## Code Health

### Tech Debt

- None identified

### Pain Points

- `summary_builder.go` is 512 lines, significantly exceeds 300-line threshold
- `summary.go` is 488 lines, significantly exceeds 300-line threshold
- `execute.go` is 449 lines, significantly exceeds 300-line threshold
- `framework.go` is 441 lines, significantly exceeds 300-line threshold
- `verify.go` is 351 lines, exceeds 300-line threshold with minimal test coverage (only 61 lines in `verify_test.go`)
- `resolve.go` is 288 lines with no dedicated test file
- `execute_test.go` is 555 lines and `summary_builder_test.go` is 422 lines, both exceed 300-line threshold

### Optimization Opportunities

- None identified
