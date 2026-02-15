# pipeline

Orchestrates CI/CD pipeline operations including workflow dispatch, status polling, artifact retrieval, and evidence download. Executes module pipelines in dependency order with wait/timeout support via the GitHub CLI.

## Key Types

- **`PipelineRunner`** -- Orchestrates pipeline execution with GitHub CLI, wait mode, and timeout settings
- **`GitHubCLI`** -- Interface for GitHub workflow operations (trigger, watch, SHA lookup, run listing)
- **`GitHubCLIImpl`** -- Production implementation of `GitHubCLI` using the `gh` CLI tool
- **`MockGitHubCLI`** -- Mock implementation controlled by environment variables for testing
- **`WorkflowRunSummary`** -- Status and conclusion of a single GitHub Actions workflow run
- **`WorkflowRun`** -- Represents a GitHub Actions workflow run with status, conclusion, and timing
- **`ArtifactInfo`** -- Artifact metadata from GitHub API (ID and name)
- **`ArtifactsResponse`** -- GitHub API response containing a list of artifacts
- **`DownloadResult`** -- Tracks what was downloaded per module (test and scan artifacts)
- **`runInfo`** -- Workflow run status from GitHub for SHA-based polling
- **`runByIDInfo`** -- Workflow run info for run-ID-based polling

## Key Functions

- **`PipelineRun()`** -- Execute module pipelines with dependency ordering and changed-only filtering
- **`PipelineStatus()`** -- Show CI pipeline status for commits by ref or SHA
- **`PipelineAwaitCI()`** -- Wait for CI workflows to complete by pattern+SHA or run ID with timeout
- **`PipelineAwaitRelease()`** -- Wait for release workflows to complete
- **`PipelineWait()`** -- General workflow wait with live progress display
- **`PipelineFindRunID()`** -- Find workflow run ID for a given workflow and commit
- **`PipelineCheckRecentRun()`** -- Check for recent successful workflow runs to skip redundant CI
- **`PipelineGetArtifactID()`** -- Retrieve artifact IDs from workflow runs via GitHub API
- **`PipelineGetTreeFiles()`** -- Get file tree from a specific commit using GitHub Trees API
- **`PipelineDownloadEvidenceArtifacts()`** -- Download test/scan evidence artifacts from CI runs
- **`awaitWorkflows()`** -- Poll GitHub for active workflows matching a pattern and SHA
- **`getTransitiveDeps()`** -- Recursively collect module and all transitive dependencies
- **`flattenArtifactDirs()`** -- Flatten gh-created artifact directory structure for evidence loading
- **`mergeDirectories()`** -- Recursively merge source directory contents into destination

## Patterns

- Table-driven command registration: `commands.go` registers all 10 subcommands via `RegisterAll()`
- Dependency-ordered execution: processes modules sequentially in dependency order with per-module wait
- Dual polling modes: wait by workflow pattern+SHA or by specific run ID
- Mock injection via environment variables: `CLIE_MOCK_GITHUB_CLI` switches to mock implementation
- Changed-module detection: uses git diff to identify modules needing pipeline execution
- Most-recent-run-wins: for re-runs, only the latest completed run determines success/failure

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | Table-driven registration of all 10 pipeline subcommands via `RegisterAll()` |
| run.go | Execute module pipelines with dependency ordering and changed-only filtering |
| status.go | Show CI pipeline status for commits by ref or SHA |
| await_ci.go | Wait for CI workflows to complete by pattern+SHA or run ID with timeout |
| await_release.go | Wait for release workflows to complete |
| wait.go | General workflow wait with live progress display and status icons |
| find_run_id.go | Find workflow run ID for a given workflow and commit SHA |
| check_recent_run.go | Check for recent successful workflow runs within a time window |
| get_artifact_id.go | Retrieve artifact IDs from workflow runs via GitHub API |
| get_tree_files.go | Get file tree from a specific commit using GitHub Trees API |
| download_evidence_artifacts.go | Command definition, argument parsing, and orchestration for evidence artifact download |
| download_evidence_ci.go | Evidence CI run lookup, module registry loading, and transitive dependency resolution |
| download_helpers.go | Artifact download execution, directory flattening, and directory merging |

## Dependencies

- `cli/eac/impl/get` -- SHA detection for await-ci and await-release commands
- `cli/eac/impl/pipeline/helper` -- PipelineRunner and GitHubCLI implementations
- `clibase/flags` -- flag validation from registry metadata
- `clibase/ghexec` -- GitHub CLI command execution
- `clibase/gitexec` -- git command execution
- `clibase/registry` -- command registration
- `core/domain/modules` -- module registry loading for dependency resolution
- `core/logging` -- structured logging
- `core/repository` -- repository root discovery

## Role in System

The `pipeline` package provides CI/CD orchestration commands for `eac`, bridging local development with GitHub Actions workflows. It enables dependency-aware pipeline execution, CI status monitoring, and evidence artifact retrieval, serving as the primary interface for both interactive developer-triggered pipelines and automated CI/CD coordination flows.

## Code Health

### Tech Debt
- await_ci.go (303 lines) is the largest file; contains workflow polling and status checking logic
- No unit tests for most pipeline commands; tested via BDD scenarios and helper package unit tests

### Pain Points
- steps_bdd_test.go (449 lines) contains TODO comments: "Implement module creation" and "Implement marking module as changed"

### Optimization Opportunities
- None identified
