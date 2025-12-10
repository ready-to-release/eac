# test

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac test [module...] [--all]`
**Purpose**: Test one or more modules by moniker
**Category**: [test](../categories/test.md)

## Language Support

The test command supports multiple test frameworks via language-specific runners:

- **Go** - `gotest` (unit tests), `godog` (BDD/Gherkin)
- **TypeScript** - `mocha` (unit tests), `cucumber-js` (BDD/Gherkin)

Tests are discovered by file patterns (`*_test.go`, `*.test.ts`, `*.feature`) and executed by the appropriate runner based on module type and test type. See [Language Support](../language-support.md) for details.

## Syntax

```bash
r2r eac test [module...] [--all] [options]
```

## Options

| Flag | Description |
|------|-------------|
| `--all` | Test all modules |
| `--parallel` | Run tests in parallel (default) |
| `--sequential` | Run tests sequentially |
| `--verbose` | Verbose output |
| `--coverage` | Generate coverage report |

## Examples

```bash
# Test single module
r2r eac test src-auth

# Test multiple modules
r2r eac test src-auth src-api

# Test all modules
r2r eac test --all

# Test with coverage
r2r eac test src-auth --coverage
```

## See Also

- [test suite](./suite.md) - Run test suites
- [test debug](./debug.md) - Debug failures
- [show tests](../show/tests.md)
- [test Commands](../categories/test.md)

{{ diataxis_footer() }}
