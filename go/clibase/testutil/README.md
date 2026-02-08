# testutil

Test utilities and helpers for the CLI test suite. Provides assertion functions,
output capture, fixture configuration, git repository helpers, and golden file testing.

## Key Types

- `CaptureOutput` -- captures stdout and stderr during test execution for assertion
- `GitRepo` -- creates and manages a temporary git repository for integration tests
- `GoldenFile` -- compares test output against stored golden files with automatic update support
- `ModuleSpec` -- describes a module fixture for `SetupMinimalFixture`

## Key Functions

- `AssertContains` -- asserts a string contains a substring
- `AssertEqual` -- asserts two values are equal
- `AssertNoError` -- asserts an error is nil
- `Capture` -- captures stdout/stderr from a function call, including `os.Exit` interception
- `CaptureNoExit` -- captures output without exit interception
- `FixtureConfig` -- returns a minimal `EACConfig` suitable for testing
- `SetupMinimalFixture` -- creates a temporary workspace with module contracts for integration tests

## Patterns

- **Fixture factories**: `FixtureConfig` and `SetupMinimalFixture` produce minimal but valid configurations, reducing test boilerplate
- **Golden file testing**: `GoldenFile` enables snapshot testing with `-update` flag support for regenerating expected output
- **Exit capture**: `Capture` intercepts `os.Exit` calls to test CLI exit behavior without terminating the test process

## Internal Structure

| File | Purpose |
|---|---|
| `assertions.go` | Assertion helper functions for common test patterns |
| `capture.go` | `CaptureOutput` and `Capture`/`CaptureNoExit` for output interception |
| `config.go` | `FixtureConfig`, `ModuleSpec`, `SetupMinimalFixture` for test fixtures |
| `git.go` | `GitRepo` for temporary git repository creation and management |
| `golden.go` | `GoldenFile` for golden file comparison testing |

## Dependencies

- `core/config` -- EAC config types for fixture generation

## Role in System

Provides shared test infrastructure used across the CLI test suite. Test files in other packages import these utilities to reduce duplication of common patterns like output capture, config fixture creation, and assertion helpers.

## Code Health

### Tech Debt
- `golden.go:14` package-level mutable `UpdateGolden` flag; safe in practice since `flag.Parse` runs once, but not idiomatic for library code
- `testutil_test.go` (128 lines) covers only a subset of the utility functions; `capture.go` and `git.go` lack dedicated tests

### Pain Points
- None identified; files are small and focused

### Optimization Opportunities
- Add tests for `Capture` exit-interception edge cases (e.g., double exit, panic recovery) to avoid silent regressions (low effort)
- Consider using `testing.TB` parameter consistently across all helpers for uniform error reporting (low effort)
