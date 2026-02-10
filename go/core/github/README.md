# github

Abstractions for GitHub API interactions including workflow runs, releases,
container package management, and package version safety assessment.

## Key Types

| Type | Purpose |
|------|---------|
| `API` | Interface for workflow runs, releases, and tree queries |
| `CLIExecutor` | Interface for raw `gh` CLI command execution |
| `GHClient` | Production `API` and `CLIExecutor` implementation |
| `GHPackagesClient` | Container package listing and deletion via API |
| `PackagesAPI` | Interface for GitHub Packages (GHCR) operations |
| `PackageSafetyChecker` | Assesses package versions for safe deletion |
| `MockAPI` | Test double for `API` with call tracking |
| `CLIMock` | Fixture-based mock for `CLIExecutor` |

## Patterns

- Interface abstraction: `API` and `CLIExecutor` decouple from `gh` binary
- Constructor injection: callers create or receive `API` instances directly
- Builder pattern: `CLIMock` uses `With*` methods for fixture setup
- Safety-first deletion: `PackageSafetyChecker` enforces release protection

## Internal Structure

| File | Purpose |
|------|---------|
| `interfaces.go` | `API`, `CLIExecutor` interfaces, `WorkflowRun`, `Release` types |
| `gh_client.go` | `GHClient` production implementation via CLI executor |
| `packages.go` | `GHPackagesClient`, `PackageVersion`, `Package` types |
| `package_safety.go` | `PackageSafetyChecker`, version assessment, bundle extraction |
| `mock.go` | `MockAPI` with call tracking and builder helpers |
| `cli_mock.go` | `CLIMock` with command routing and fixture loading |

## Dependencies

_No internal repository imports -- this is a leaf package._

## Role in System

This package provides the GitHub integration layer for the `core` module,
used by pipeline, release, and scan commands. The `PackageSafetyChecker`
implements policy-based protection to ensure released container images are
never accidentally deleted during cleanup operations.

## Code Health

- **Tech Debt**: `package_safety.go:213,220`: compiled regexps at package level are fine, but `bundleTagPattern` and `moduleVersionPattern` lack doc comments explaining expected formats.
- **Pain Points**: `cli_mock.go` (319 lines) duplicates significant CLI output parsing logic from `gh_client.go`; changes to JSON parsing must be mirrored in both. No clear boundary between "workflow" and "release" concerns within the `API` interface -- callers that need only releases still import workflow types.
- **Optimization Opportunities**: No TODO/FIXME markers found -- no deferred work items in the codebase.
