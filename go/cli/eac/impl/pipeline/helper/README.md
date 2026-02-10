# helper

Provides the core pipeline execution engine and GitHub CLI abstraction layer used by all pipeline commands. Contains the `PipelineRunner` for dependency-ordered module execution and the `GitHubCLI` interface for workflow operations.

## Key Types

- **`PipelineRunner`** -- Orchestrates pipeline execution across modules with dependency ordering, wait modes, and timeout support
- **`GitHubCLI`** -- Interface defining GitHub workflow operations: trigger, watch, get SHA, list runs, get run status
- **`GitHubCLIImpl`** -- Production implementation of `GitHubCLI` that shells out to the `gh` CLI tool
- **`MockGitHubCLI`** -- Mock implementation controlled by environment variables (`CLIE_MOCK_GITHUB_CLI`) for testing
- **`WorkflowRunSummary`** -- Status and conclusion of a single GitHub Actions workflow run

## Key Functions

- **`New()`** -- Create a new `PipelineRunner` with GitHub CLI, repo root, and workflow directory
- **`SetWaitOptions()`** -- Configure wait mode, timeout, and poll interval for pipeline execution
- **`RunPipeline()`** -- Execute a single module's CI workflow by dispatching and optionally waiting
- **`RunPipelines()`** -- Execute pipelines for specific modules in dependency order
- **`RunAllPipelines()`** -- Execute pipelines for all modules in dependency order
- **`RunAllChangedPipelines()`** -- Execute pipelines only for modules with changes (via git diff)
- **`topologicalSort()`** -- Sort modules by dependency order using topological sort for sequential execution
- **`NewGitHubCLI()`** -- Create a production `GitHubCLI` using the `gh` CLI tool

## Patterns

- Interface-based GitHub CLI abstraction: production `gh` CLI and environment-variable-controlled mock
- Topological sort for dependency-ordered execution: ensures dependent modules run after their dependencies
- Sequential pipeline execution with per-module wait: runs one module at a time when wait mode is enabled
- Workflow file filtering: only dispatches modules that have matching `.github/workflows/ci-{module}.yaml` files

## Internal Structure

| File | Responsibility |
| --- | --- |
| github.go | `GitHubCLI` interface, `GitHubCLIImpl` production implementation, and `MockGitHubCLI` test double |
| runner.go | `PipelineRunner` with dependency-ordered execution, topological sort, and changed-module filtering |

## Dependencies

- `clibase/ghexec` -- GitHub CLI command execution
- `clibase/gitexec` -- git command execution
- `core/domain/modules` -- module registry for dependency graph traversal
- `core/environments` -- environment variable constants for mock control
- `core/logging` -- structured logging
- `core/paths` -- workflow file path resolution
- `core/repository` -- changed module detection via git diff

## Role in System

The `helper` package is the execution backbone of the pipeline system. All pipeline commands that interact with GitHub Actions workflows use this package's `GitHubCLI` interface and `PipelineRunner`. It provides the dependency-aware execution ordering that ensures modules are processed in the correct sequence, and its mock implementation enables BDD testing without actual GitHub API calls.

## Code Health

### Tech Debt
- `runner.go` (363 lines) combines pipeline orchestration, topological sort, and changed-module detection in one file
- `github.go` (283 lines) contains both the interface, production implementation, and mock in one file

### Pain Points
- `MockGitHubCLI` relies on environment variables for control, making test setup implicit rather than explicit

### Optimization Opportunities
- Extract `topologicalSort()` into a shared graph utility (moderate feasibility, could benefit other dependency-ordering code)
- Consider constructor injection for `MockGitHubCLI` instead of environment variable control (moderate effort, improves testability)
