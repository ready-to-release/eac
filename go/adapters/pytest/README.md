# pytest

Pytest adapter providing a test runner for Python pytest unit tests, with pip isolation and CTRF report generation.

## Key Types

- **`PytestRunner`** -- Implements `TestTypeRunner` for Python pytest unit test execution

## Key Functions

- `convertPytestJSONToCTRF` -- Converts pytest-json-report output to standardized CTRF format
- `findPythonModuleForPath` -- Resolves module moniker for a given Python source path

## Patterns

- Test type registration: `init()` registers runner and unit test descriptor with `@deps:python` and `@L1` inferences
- Module resolution: Matches test paths to modules by component root prefix
- Pip isolation: Uses `pip.PipIsolation` for safe parallel execution with venv creation
- JSON reporting: Uses pytest-json-report plugin for structured test results

## Internal Structure

| File | Responsibility |
| --- | --- |
| runner.go | `PytestRunner` implementing `TestTypeRunner` with `RegisterWith()` for explicit registry injection, venv setup, and execution via extracted helpers (`collectPytestResults`, `pytestExecutionFailed`) |
| ctrf.go | Pytest JSON report to CTRF conversion with crash trace extraction |

## Dependencies

- `adapters/pip` -- Pip isolation for safe parallel Python test execution
- `clibase/testrunners` -- `TestTypeRunner` interface and registration
- `core/config` -- EAC configuration for module resolution
- `core/testing` -- `TestReference` type
- `core/tool` -- tool registry and command building
- `core/ctrf` -- CTRF report format types

## Role in System

The pytest adapter enables Python unit test execution within the unified test framework. It handles pip dependency installation in isolated virtual environments, runs pytest with the json-report plugin, and converts results to CTRF format for standardized test result aggregation.

## Code Health

### Tech Debt
- None identified

### Pain Points
- runner.go is 335 lines; candidate for splitting (extract CTRF conversion and execution orchestration into separate files)
- No unit tests for the full Execute() (requires tool registry); runner_test.go covers helpers, module resolution, and public API surface

### Optimization Opportunities
- None identified
