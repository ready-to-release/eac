# gotest

Go test runner adapter handling execution of both `go test` unit tests
and godog BDD tests, with JSON event streaming and CTRF report generation.

## Key Types

- **`GoTestRunner`** -- Implements `TestTypeRunner` for Go test execution
- **`RunResult`** -- Test execution results (pass/fail/skip counts)

## Patterns

- Test type registration: `init()` registers runner and descriptor
- Fallback runner: Serves as the default runner for unknown test types
- JSON streaming: Parses `go test -json` events for real-time results
- CTRF conversion: Transforms Go test events to standardized CTRF format
- Tag translation: Extracts Go build tags from godog suite filter strings

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| runner.go | `GoTestRunner` implementing `TestTypeRunner` |
| runner_helpers.go | Module lookup, feature extraction, `go generate` |
| ctrf.go | Go test event to CTRF JSON report conversion |

## Dependencies

- `adapters/godog` -- `GodogTagTranslator` for tag filter conversion
- `clibase/testrunners` -- `TestTypeRunner` interface and registration
- `core/ctrf` -- CTRF report format types
- `core/config` -- EAC configuration for module resolution
- `core/testing` -- `TestReference` type
- `core/tool` -- tool registry and command building
- `core/logging` -- structured logging

## Role in System

The `gotest-eac` module is registered as both the Go test runner and
the fallback runner in the test execution framework. It handles `go test`
invocation with JSON output parsing, coverage collection, and CTRF
report generation, enabling the orchestrator to run Go unit tests and
godog BDD tests uniformly through the `TestTypeRunner` interface.

## Code Health

### Tech Debt
- `init()` in runner.go registers both runner and descriptor globally; accepting a registry parameter would improve testability
- `Execute()` in runner.go (~120 lines) mixes package-path parsing, env-var construction, command building, and CTRF conversion in one method
- Package-level `var goRunnerLog` in runner.go is global mutable state

### Pain Points
- `extractGoBuildTags` in runner_helpers.go uses manual index arithmetic for tag parsing; a regex-based approach would be more readable and maintainable
- No unit tests for `Execute()` or `convertGoTestEventsToCTRF` (only runner_helpers_test.go exists)

### Optimization Opportunities
- `runGoGenerate` in runner_helpers.go runs `go generate ./...` on every test execution even if no generate directives exist; checking for `//go:generate` directives first could skip unnecessary work
