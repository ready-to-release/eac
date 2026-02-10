# testutil

Provides test fixtures and assertion helpers for GitHub API mocking.

Wraps `github.MockAPI` with convenient builder functions for setting up workflow runs, releases, and tree files.

## Key Functions

| Function               | Purpose                                                |
| ---------------------- | ------------------------------------------------------ |
| `NewMockGitHub`        | Creates a pre-configured `github.MockAPI` instance     |
| `WithSuccessfulRun`    | Adds a successful workflow run to a mock               |
| `WithFailedRun`        | Adds a failed workflow run to a mock                   |
| `WithInProgressRun`    | Adds an in-progress workflow run to a mock             |
| `WithOldSuccessfulRun` | Adds a successful run with a specified age in the past |
| `WithRelease`          | Adds a release to a mock                               |
| `WithTreeFiles`        | Adds tree files for a SHA to a mock                    |
| `AssertCalled`         | Panics if a method was not called on the mock          |
| `AssertNotCalled`      | Panics if a method was called on the mock              |
| `AssertCallCount`      | Panics if a method was not called exactly N times      |

## Patterns

- **Builder helpers**: `With*` functions chain onto `MockAPI` for fluent test setup
- **Panic-based assertions**: `Assert*` functions panic on failure for use in test contexts

## Internal Structure

| File          | Purpose                                          |
| ------------- | ------------------------------------------------ |
| `fixtures.go` | All mock builder functions and assertion helpers |

## Dependencies

| Package       | Purpose                                               |
| ------------- | ----------------------------------------------------- |
| `core/github` | `MockAPI`, `WorkflowRun` types for GitHub API mocking |

## Role in System

Shared test utility package providing pre-built GitHub mock scenarios. Used by tests that need to simulate CI workflow results, releases, or git tree state without making real GitHub API calls.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: The `AssertCallCount` function (line 88 of `fixtures.go`) uses `rune` conversion for number formatting which only works for single-digit counts.
- **Optimization Opportunities**: Replace panic-based assertions with `testing.T`-based assertions that report file/line correctly and use `fmt.Sprintf` for numbers.
