# behave

Behave adapter providing a test runner for Python behave BDD tests, with pip isolation, tag filter translation, and CTRF report generation.

## Key Types

- **`BehaveRunner`** -- Implements `TestTypeRunner` for Python behave BDD test execution
- **`BehaveTagTranslator`** -- Translates `TagFilter` to behave CLI tag syntax

## Key Functions

- `convertBehaveJSONToCTRF` -- Converts behave JSON output to standardized CTRF format

## Patterns

- Test type registration: `init()` registers runner and BDD descriptor with `@deps:python` inference
- Feature ownership: Resolves Python modules via `FeatureTestTypeResolver`
- Tag translation: Converts `TagFilter` to behave `--tags` syntax with `~@` for exclusions
- Pip isolation: Uses `pip.PipIsolation` for safe parallel execution with venv creation

## Internal Structure

| File | Responsibility |
| --- | --- |
| runner.go | `BehaveRunner` implementing `TestTypeRunner` with venv setup and execution |
| tags.go | `BehaveTagTranslator` for behave tag expression syntax |
| ctrf.go | Behave JSON to CTRF report conversion |

## Dependencies

- `adapters/pip` -- Pip isolation for safe parallel Python test execution
- `contracts/core/0.1.0` -- `TagFilter` for tag translation
- `clibase/testrunners` -- `TestTypeRunner` interface and registration
- `core/config` -- EAC configuration for module resolution
- `core/testing` -- `TestReference` type
- `core/tool` -- tool registry and command building
- `core/ctrf` -- CTRF report format types

## Role in System

The behave adapter enables Python BDD test execution within the unified test framework. It owns feature files for Python modules, translates tag filters to behave syntax, manages pip dependency isolation with virtual environments, and converts behave JSON output to CTRF format for standardized test result aggregation.

## Code Health

### Tech Debt
- `Execute()` in runner.go (~170 lines) interleaves pip isolation, venv creation, dependency installation, command building, and result parsing; extracting helpers would improve clarity
- `init()` in runner.go performs global registration; a factory function accepting a registry would be more testable

### Pain Points
- `Execute()` falls back to `len(tests)` as `TestsFailed` on error when JSON results are unavailable, so per-test pass/fail counts are approximate in that case
- No unit tests for `Execute()` or `GetTestInfo()`; only ctrf_test.go and tags_test.go exist

### Optimization Opportunities
- The venv creation and `pip install` steps run synchronously within Execute; caching the venv from a previous run when pyproject.toml is unchanged could skip reinstallation entirely
