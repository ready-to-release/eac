# test suite

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac test suite <name>`
**Purpose**: Run tests for a specific test suite (parallel by default)
**Category**: [test](../categories/test.md)

## Syntax

```bash
r2r eac test suite <name> [options]
```

## Options

| Flag | Description |
|------|-------------|
| `--parallel` | Run tests in parallel (default) |
| `--sequential` | Run tests sequentially |
| `--verbose` | Verbose output |
| `--tags` | Filter by tags |

## Examples

```bash
# Run suite
r2r eac test suite integration

# Run with specific tags
r2r eac test suite integration --tags @critical

# Run sequentially
r2r eac test suite unit --sequential
```

## See Also

- [test](./test.md) - Test modules directly
- [test list-suites](./list-suites.md) - List available suites
- [show suite](../show/suite.md) - Suite details
- [test Commands](../categories/test.md)

{{ diataxis_footer() }}
