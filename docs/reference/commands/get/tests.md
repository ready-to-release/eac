# get tests

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get tests [module]`
**Purpose**: Get all tests in the repository
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get tests [module]
```

## Examples

```bash
# Get all tests
r2r eac get tests | jq '.'

# Filter by module
r2r eac get tests | jq '.tests[] | select(.module == "src-auth")'

# Count by type
r2r eac get tests | jq '.by_type'
```

## See Also

- [show tests](../show/tests.md) - Formatted table
- [test](../test/test.md)
- [get suite](./suite.md)

{{ diataxis_footer() }}
