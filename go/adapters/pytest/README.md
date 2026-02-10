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
| runner.go | `PytestRunner` implementing `TestTypeRunner` with venv setup and execution |
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
- `Execute()` in runner.go (~150 lines) interleaves pip isolation, venv creation, dependency installation, command building, and result parsing; extracting helpers would improve clarity
- `init()` in runner.go performs global registration; a factory function accepting a registry would be more testable

### Pain Points
- `Execute()` falls back to `len(tests)` as `TestsFailed` on error when JSON results are unavailable, so per-test pass/fail counts are approximate in that case
- No unit tests for `Execute()` or `GetTestInfo()`; only ctrf_test.go exists

### Optimization Opportunities
- The venv creation and `pip install` steps run synchronously within Execute; caching the venv from a previous run when pyproject.toml is unchanged could skip reinstallation entirely
