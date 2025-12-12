# test

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac test [module...] [--suite <name>] [--all]`
**Purpose**: Test one or more modules by moniker
**Category**: [test](../categories/test.md)

## Language Support

The test command supports multiple test frameworks via language-specific runners:

- **Go** - `gotest` (unit tests), `godog` (BDD/Gherkin)
- **TypeScript** - `mocha` (unit tests), `cucumber-js` (BDD/Gherkin)

Tests are discovered by file patterns (`*_test.go`, `*.test.ts`, `*.feature`) and executed by the appropriate runner based on module type and test type. See [Language Support](../language-support.md) for details.

## Syntax

```bash
r2r eac test [module...] [--suite <name>] [--all] [options]
```

## Options

| Flag | Description |
|------|-------------|
| `--suite <name>` | Run a specific test suite (component, integration, acceptance) |
| `--all` | Run all test suites (component + integration + acceptance) in a single pass |
| `--parallel` | Run tests in parallel (default) |
| `--sequential` | Run tests sequentially |
| `--verbose` | Verbose output |
| `--coverage` | Generate coverage report |
| `--retest` | Force full test run (ignore incremental detection) |

## Examples

```bash
# Test single module with default suite (component)
r2r eac test src-auth

# Test multiple modules
r2r eac test src-auth src-api

# Test with specific suite
r2r eac test src-auth --suite integration

# Run all suites (component + integration + acceptance)
r2r eac test src-auth --all

# Test with coverage
r2r eac test src-auth --coverage
```

## Test Suites

When using `--all`, the command runs component, integration, and acceptance suites in a single pass. Test results are routed to their respective output folders based on test level:

- `out/test/component/` - L0, L1 tests
- `out/test/integration/` - L2 tests
- `out/test/acceptance/` - L3 tests

This provides a single initialization and summary while maintaining separate output folders for each suite.

## See Also

- [test suite](./suite.md) - Run test suites
- [test debug](./debug.md) - Debug failures
- [show tests](../show/tests.md)
- [test Commands](../categories/test.md)
- [Run Test Suites](../../../how-to-guides/eac/commands/build-test-validate/run-test-suites.md) - How-to guide

{{ diataxis_footer() }}
