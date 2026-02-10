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
- `framework.go:28` `Run()` is ~218 lines; the phased orchestration would benefit from extracting each phase dispatch into named helpers
- `summary.go:105` `generateComponentTUISummary` is ~258 lines and `printComponentConsoleSummary` ~144 lines; both duplicate `moduleCache` iteration logic
- `framework.go:16` package-level `log` var initialized at import time

### Pain Points
- `summary.go` and `summary_builder.go` together are ~1170 lines; summary logic is the largest area of the package and hard to navigate
- `verify.go` and `resolve.go` have only minimal test coverage (`verify_test.go` 61 lines, no `resolve_test.go`)

### Optimization Opportunities
- Extract shared `moduleCache` iteration in `summary.go` into a single helper used by both TUI and console paths (low effort)
- Add integration tests for `resolve.go` scope-expansion logic (medium effort)
