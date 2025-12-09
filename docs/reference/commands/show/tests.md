# show tests

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show tests [module]`
**Purpose**: Show all tests in the repository in a human-readable table
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show tests [module]
```

## Examples

```bash
# Show all tests
r2r eac show tests

# Filter by module
r2r eac show tests src-auth

# Find BDD tests
r2r eac show tests | grep "BDD"
```

## See Also

- [get tests](../get/tests.md) - JSON output
- [test](../test/test.md) - Run tests
- [show suite](./suite.md) - Test suite details

{{ diataxis_footer() }}
