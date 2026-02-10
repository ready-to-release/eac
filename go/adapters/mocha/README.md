# mocha

Mocha adapter providing a test runner for TypeScript mocha unit tests,
with npm isolation and CTRF report generation.

## Key Types

- **`MochaRunner`** -- Implements `TestTypeRunner` for mocha execution

## Patterns

- Test type registration: `init()` registers runner and unit test descriptor
- NPM isolation: Uses `npm.NpmIsolation` for safe parallel execution
- CTRF conversion: Transforms mocha JSON output to standardized CTRF format
- JSON parsing: Reads mocha's native JSON reporter output for results

## Internal Structure

| File | Responsibility |
| --- | --- |
| runner.go | `MochaRunner` implementing `TestTypeRunner` with CTRF conversion |

## Dependencies

- `adapters/npm` -- NPM isolation for safe parallel test execution
- `clibase/testrunners` -- `TestTypeRunner` interface and registration
- `core/ctrf` -- CTRF report format types
- `core/config` -- EAC configuration for module resolution
- `core/testing` -- `TestReference` type
- `core/tool` -- tool registry and command building

## Role in System

The `mocha-eac` module enables TypeScript unit test execution within
the unified test framework. It handles npm dependency installation in
isolated environments, runs mocha with the JSON reporter, and converts
results to CTRF format for standardized test result aggregation.

## Code Health

### Tech Debt
- `Execute()` in runner.go (~124 lines) interleaves npm isolation, dependency installation, command building, piped I/O, CTRF conversion, and result tallying; extracting `installDependencies` and `runMocha` helpers would improve clarity
- `init()` in runner.go performs global registration; a factory function accepting a registry would be more testable

### Pain Points
- `Execute()` uses `len(tests)` as `TestsFailed` on error and `TestsPassed` on success without consulting the parsed mocha JSON stats, so per-test pass/fail counts are approximate even when detailed results are available
- No unit tests exist for any function in the package; `Execute()`, `convertMochaJSONToCTRF`, and `GetTestInfo` are all untested

### Optimization Opportunities
- The `npm ci` / `npm install` step runs synchronously within Execute; for modules with unchanged lockfiles, caching node_modules from a previous run could skip reinstallation entirely; moderate effort
