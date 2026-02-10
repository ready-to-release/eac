# testutil

Test utilities and helpers for the CLI test suite. Provides assertion functions,
output capture, fixture configuration, git repository helpers, and golden file testing.

## Key Types

- `CaptureOutput` -- captures stdout and stderr during test execution for assertion; provides `AssertSuccess`, `AssertFailure`, `AssertExitCode`, `AssertValidJSON`, `AssertValidYAML`, `DecodeJSON`, `DecodeYAML`
- `GitRepo` -- creates and manages a temporary git repository for integration tests; provides `Git`, `WriteFile`, `ReadFile`, `Add`, `Commit`, `Tag`, `Branch`, `Checkout`, `SetupEACConfig`
- `GoldenFile` -- compares test output against stored golden files with automatic update support via `-update-golden` flag
- `ModuleSpec` -- describes a module fixture for `SetupMinimalFixture` and `SetupFixtureWithModules`

## Key Functions

- `AssertContains`, `AssertNotContains`, `AssertContainsAll`, `AssertContainsAny` -- string containment assertions
- `AssertMatches`, `AssertNotMatches` -- regex-based string assertions
- `AssertEqual`, `AssertNotEqual`, `AssertNil`, `AssertNotNil` -- value equality assertions
- `AssertTrue`, `AssertFalse`, `AssertEmpty`, `AssertNotEmpty` -- boolean and string assertions
- `AssertLineCount`, `AssertMinLineCount` -- line count assertions
- `AssertHasPrefix`, `AssertHasSuffix` -- string prefix/suffix assertions
- `AssertSliceContains`, `AssertSliceNotContains`, `AssertSliceLength` -- generic slice assertions
- `AssertMapHasKey`, `AssertMapNotHasKey` -- generic map assertions
- `AssertError`, `AssertNoError`, `AssertErrorContains` -- error assertions
- `Capture` -- captures stdout/stderr from a function call returning an exit code
- `CaptureNoExit` -- captures output without exit code
- `FixtureConfig` -- loads EAC config from a test fixture directory
- `FixtureConfigValidated` -- loads EAC config with schema validation
- `SetupMinimalFixture` -- creates a temporary workspace with minimal EAC config
- `SetupFixtureWithModules` -- creates a temporary workspace with specified module contracts
- `NewGitRepo` -- creates a temporary git repository with initial config
- `NewGolden` -- creates a `GoldenFile` with sensible defaults
- `AssertGolden`, `AssertGoldenTrimmed` -- convenience functions for golden file assertions
- `WithWorkingDir` -- temporarily changes working directory for test scope
- `TempDir` -- creates a temporary directory cleaned up after test
- `WriteFixtureFile` -- writes a file within a fixture directory
- `ClearConfigCache` -- clears the global config cache for test isolation

## Patterns

- **Fixture factories**: `FixtureConfig`, `SetupMinimalFixture`, and `SetupFixtureWithModules` produce minimal but valid configurations, reducing test boilerplate
- **Golden file testing**: `GoldenFile` enables snapshot testing with `-update-golden` flag support for regenerating expected output
- **Output capture**: `Capture` intercepts stdout/stderr via pipe redirection to test CLI output without running a subprocess
- **Generic assertions**: slice and map assertion functions use Go generics for type-safe testing

## Internal Structure

| File | Purpose |
|---|---|
| `assertions.go` | Assertion helper functions for string, value, slice, map, and error patterns |
| `capture.go` | `CaptureOutput`, `Capture`, `CaptureNoExit`, and output format assertions (JSON/YAML) |
| `config.go` | `FixtureConfig`, `ModuleSpec`, `SetupMinimalFixture`, `SetupFixtureWithModules`, `WithWorkingDir`, `TempDir`, `WriteFixtureFile` |
| `git.go` | `GitRepo` for temporary git repository creation and management with EAC config setup |
| `golden.go` | `GoldenFile` for golden file comparison testing with `-update-golden` flag |

## Dependencies

- `core/config` -- EAC config types for fixture generation and cache clearing

## Role in System

Provides shared test infrastructure used across the CLI test suite. Test files in other packages import these utilities to reduce duplication of common patterns like output capture, config fixture creation, and assertion helpers.

## Code Health

### Tech Debt
- `golden.go:14` package-level mutable `UpdateGolden` flag; safe in practice since `flag.Parse` runs once, but not idiomatic for library code

### Pain Points
- None identified; files are small and focused

### Optimization Opportunities
- Add tests for `Capture` edge cases (e.g., pipe failures, concurrent capture) to avoid silent regressions (low effort)
- Consider using `testing.TB` parameter consistently across all helpers for uniform error reporting (low effort)
