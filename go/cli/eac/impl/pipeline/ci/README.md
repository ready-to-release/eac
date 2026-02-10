# ci

Provides CI-specific pipeline sub-commands for dispatching, scheduling, and monitoring GitHub Actions workflows across multiple modules with concurrency control and cascade failure handling.

## Key Types

- **`CIWorkflowDispatcher`** -- Interface for dispatching and querying GitHub Actions workflow runs
- **`ghWorkflowDispatcher`** -- Production implementation using `gh` CLI for workflow dispatch and status polling
- **`CIScheduler`** -- Orchestrates concurrent module CI dispatches with dependency ordering and failure cascading
- **`CISchedulerConfig`** -- Configuration for scheduler behavior (concurrency limit, timeouts, polling interval)
- **`CIModuleStatus`** -- Tracks per-module dispatch state (pending, active, completed, failed, skipped)
- **`CIScheduleResult`** -- Aggregated scheduling results across all modules
- **`runStatusInfo`** -- Workflow run status metadata from GitHub API

## Key Functions

- **`PipelineCI()`** -- Parent command entry point that prints CI sub-command usage
- **`PipelineCISchedule()`** -- Schedule CI workflows for modules with concurrency limits and cascade failure
- **`PipelineCIDispatchAndWait()`** -- Dispatch a single workflow and wait for its completion
- **`PipelineCIGetRunID()`** -- Find the run ID for a workflow dispatch by workflow name and SHA
- **`PipelineCISummaryLink()`** -- Generate diagnostic markdown summary with CI links
- **`NewCIScheduler()`** -- Create a scheduler with dispatcher, config, and module list
- **`NewGHWorkflowDispatcher()`** -- Create a production workflow dispatcher for a specific repo
- **`filterModulesForSchedule()`** -- Filter module list to only those with matching CI workflow files

## Patterns

- Table-driven command registration: `commands.go` registers 5 CI sub-commands via `RegisterAll()`
- Interface-based dispatch: `CIWorkflowDispatcher` allows mock injection for testing
- Concurrent scheduling with configurable parallelism: scheduler tracks active/pending/completed states
- Cascade failure: when a module fails, all dependents are automatically skipped
- Poll-based status monitoring with configurable intervals and timeouts

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | Table-driven registration of 5 CI sub-commands via `RegisterAll()` |
| ci.go | Parent `pipeline ci` command entry point with usage display |
| dispatcher.go | `CIWorkflowDispatcher` interface and `gh` CLI implementation |
| schedule.go | `CIScheduler` with concurrency control, dependency ordering, and cascade failure |
| dispatch_and_wait.go | Single workflow dispatch-and-wait with run ID discovery and polling |
| get_run_id.go | Find workflow run ID by workflow file name and commit SHA |
| summary_link.go | Generate diagnostic markdown with CI summary links |

## Dependencies

- `cli/eac/impl/get` -- SHA detection for dispatch commands
- `clibase/flags` -- flag validation from registry metadata
- `clibase/ghexec` -- GitHub CLI command execution
- `clibase/gitexec` -- git command execution
- `clibase/registry` -- command registration and workspace root
- `core/domain/modules` -- module registry for dependency resolution
- `core/logging` -- structured logging
- `core/paths` -- workflow file path resolution

## Role in System

The `ci` sub-package handles the advanced CI orchestration workflows within the pipeline system. While the parent `pipeline` package handles simple run/wait/status operations, `ci` provides the sophisticated scheduling layer needed for multi-module repositories -- dispatching CI workflows in dependency order, respecting concurrency limits, and cascading failures to dependent modules.

## Code Health

### Tech Debt
- `schedule.go` is 572 lines and contains both the scheduler logic and the command entry point (`PipelineCISchedule`)
- `get_run_id.go` defines custom error types (`simpleError`, `newError`, `sprintf`) that duplicate standard library functionality

### Pain Points
- The `CIScheduler` state machine (pending/active/completed/failed/skipped) is implicit rather than explicitly modeled

### Optimization Opportunities
- Extract `PipelineCISchedule()` from `schedule.go` into a separate command file to separate CLI concerns from scheduler logic (moderate feasibility)
- Remove custom error types in `get_run_id.go` in favor of `fmt.Errorf` (low effort)
