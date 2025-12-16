# test

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac test [module...] [--suite <name>]`
**Purpose**: Test one or more modules by moniker
**Category**: [test](../categories/test.md)

## Language Support

The test command supports multiple test frameworks via language-specific runners:

- **Go** - `gotest` (unit tests), `godog` (BDD/Gherkin)
- **TypeScript** - `mocha` (unit tests), `cucumber-js` (BDD/Gherkin)

Tests are discovered by file patterns (`*_test.go`, `*.test.ts`, `*.feature`) and executed by the appropriate runner based on module type and test type. See [Language Support](../language-support.md) for details.

## Syntax

```bash
r2r eac test [module...] [--suite <name>] [options]
```

## Options

| Flag | Description |
|------|-------------|
| `--suite <name>` | Run test suite(s) - use `+` for multiple (e.g., `unit+integration`) |
| `--sequential` | Run tests sequentially (parallel is default) |
| `--coverage` | Generate coverage report |
| `--retest` | Force full test run (ignore incremental detection) |
| `--skip-deps` | Skip dependency verification |
| `--list-only` | List tests without running |
| `--timings` | Show detailed timing summary |
| `--no-tui` | Disable TUI console |

## Examples

```bash
# Test single module with default suites (unit+integration)
r2r eac test src-auth

# Test multiple modules
r2r eac test src-auth src-api

# Test with specific suite
r2r eac test src-auth --suite integration

# Run multiple suites together
r2r eac test src-auth --suite unit+integration+acceptance

# Test with coverage
r2r eac test src-auth --coverage
```

## Test Suites

Use composite suites with `+` to run multiple suites in a single pass. Test results are output to the module's test folder:

- `out/test/<module>/` - All test results for the module

This provides a single initialization and summary with organized output per module.

## See Also

- [test suite](./suite.md) - Run test suites
- [test debug](./debug.md) - Debug failures
- [show tests](../show/tests.md)
- [test Commands](../categories/test.md)
- [Run Test Suites](../../../how-to-guides/eac/commands/build-test-validate/run-test-suites.md) - How-to guide

{{ diataxis_footer() }}
