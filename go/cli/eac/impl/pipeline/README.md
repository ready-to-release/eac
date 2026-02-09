# pipeline

Orchestrates CI/CD pipeline operations including workflow dispatch, status polling, artifact retrieval, and evidence download. Executes module pipelines in dependency order with wait/timeout support via the GitHub CLI.

## Key Types

- **`PipelineRunner`** -- Orchestrates pipeline execution with GitHub CLI, wait mode, and timeout settings
- **`GitHubCLI`** -- Interface for GitHub workflow operations (trigger, watch, SHA lookup, run listing)
- **`GitHubCLIImpl`** -- Production implementation of `GitHubCLI` using the `gh` CLI tool
- **`MockGitHubCLI`** -- Mock implementation controlled by environment variables for testing
- **`WorkflowRunSummary`** -- Status and conclusion of a single GitHub Actions workflow run
- **`runInfo`** -- Workflow run status from GitHub for SHA-based polling
- **`runByIDInfo`** -- Workflow run info for run-ID-based polling

## Patterns

- Table-driven command registration: `commands.go` registers all subcommands via `RegisterAll()`
- Dependency-ordered execution: processes modules sequentially in dependency order with per-module wait
- Dual polling modes: wait by workflow pattern+SHA or by specific run ID
- Mock injection via environment variables: `CLIE_MOCK_GITHUB_CLI` switches to mock implementation
- Changed-module detection: uses git diff to identify modules needing pipeline execution

## Internal Structure

| File | Responsibility |
| --- | --- |
| run.go | Execute module pipelines with dependency ordering and changed-only filtering |
| status.go | Show CI pipeline status for commits by ref or SHA |
| await_ci.go | Wait for CI workflows to complete by pattern+SHA or run ID with timeout |
| await_release.go | Wait for release workflows to complete |
| wait.go | General workflow wait operations |
| find_run_id.go | Find workflow run ID for a given workflow and commit |
| check_recent_run.go | Check for recent workflow runs |
| get_artifact_id.go | Retrieve artifact IDs from workflow runs |
| get_tree_files.go | Get file tree from a specific commit |
| download_evidence_artifacts.go | Download evidence artifacts from CI runs |
| ci/ | CI orchestration sub-commands (dispatch-and-wait, get-run-id, summary-link) |
| helper/github.go | `GitHubCLI` interface and implementations (production and mock) |
| helper/runner.go | `PipelineRunner` with sequential execution and workflow file filtering |

## Dependencies

- `cli/eac/impl/get` -- SHA detection for await-ci command
- `clibase/flags` -- flag validation from registry metadata
- `clibase/ghexec` -- GitHub CLI command execution
- `clibase/gitexec` -- git command execution
- `clibase/registry` -- command registration
- `core/config` -- CI dispatch settle time configuration
- `core/domain/modules` -- module registry loading for pipeline validation
- `core/environments` -- environment variable constants for mock control
- `core/logging` -- structured logging
- `core/paths` -- workflow directory and file path resolution
- `core/repository` -- repository root discovery and changed module detection

## Role in System

The `pipeline` package provides CI/CD orchestration commands for `eac`, bridging local development with GitHub Actions workflows. It enables dependency-aware pipeline execution, CI status monitoring, and evidence artifact retrieval, serving as the primary interface for both interactive developer-triggered pipelines and automated CI/CD coordination flows.

## Code Health

### Tech Debt
- TODO stubs in steps_helper_test.go:32 and :42 for module creation and changed-module marking remain unimplemented
- `PipelineDownloadEvidenceArtifacts` (download_evidence_artifacts.go:48, ~122 lines) handles argument parsing, evidence lookup, artifact download, and directory flattening
- download_evidence_artifacts.go (403 lines) is the largest file with no dedicated unit tests

### Pain Points
- ~~Many subcommand files each with their own `init()` registration follow the same boilerplate pattern~~ (resolved: table-driven `commands.go`)
- No unit tests exist outside of BDD test files; all testing relies on integration-level BDD scenarios

### Optimization Opportunities
- Implement the TODO stubs in steps_helper_test.go to complete BDD test coverage for module-level scenarios (high feasibility, test infrastructure exists)
- Extract artifact download and directory flattening from download_evidence_artifacts.go into a reusable helper (moderate feasibility, would also benefit CI artifact workflows)
