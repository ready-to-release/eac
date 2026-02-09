# cucumber

Cucumber adapter providing a test runner for TypeScript cucumber-js BDD
tests, with npm isolation and tag filter translation.

## Key Types

- **`TsCucumberRunner`** -- Implements `TestTypeRunner` for cucumber-js
- **`CucumberTagTranslator`** -- Translates `TagFilter` to cucumber-js syntax

## Patterns

- Test type registration: `init()` registers runner and BDD descriptor
- Feature ownership: Resolves TypeScript modules via `FeatureTestTypeResolver`
- Tag translation: Converts `TagFilter` to cucumber-js `and`/`or`/`not` syntax
- NPM isolation: Uses `npm.NpmIsolation` for safe parallel execution

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| runner.go | `TsCucumberRunner` implementing `TestTypeRunner` |
| tag_translator.go | `CucumberTagTranslator` for cucumber-js tag syntax |

## Dependencies

- `adapters/npm` -- NPM isolation for safe parallel test execution
- `contracts/core/0.1.0` -- `TagFilter` for tag translation
- `clibase/testrunners` -- `TestTypeRunner` interface and registration
- `core/config` -- EAC configuration for module resolution
- `core/testing` -- `TestReference` type
- `core/tool` -- tool registry and command building

## Role in System

The `cucumber-eac` module enables TypeScript BDD test execution within
the unified test framework. It owns feature files for TypeScript modules,
translates tag filters to cucumber-js syntax, and manages npm dependency
isolation to prevent file lock conflicts during parallel test runs.

## Code Health

### Tech Debt
- `Execute()` in runner.go (~130 lines) interleaves npm isolation, dependency installation, command building, and result parsing; extracting `installDependencies` and `runCucumber` helpers would improve clarity
- `init()` in runner.go performs global registration; a factory function accepting a registry would be more testable

### Pain Points
- `Execute()` uses `len(tests)` as `TestsFailed` on error and `TestsPassed` on success without parsing actual cucumber-js output, so per-test pass/fail counts are approximate
- No unit tests for `Execute()` or `GetTestInfo()`; only tag_translator_test.go exists

### Optimization Opportunities
- The `npm ci` / `npm install` step runs synchronously within Execute; for modules with unchanged lockfiles, caching node_modules from a previous run could skip reinstallation entirely; moderate effort
