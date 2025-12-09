# show test-timings

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show test-timings [module]`
**Purpose**: Show test timing analysis in a human-readable format
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show test-timings [module]
```

## Examples

```bash
# Show all test timings
r2r eac show test-timings

# Filter by module
r2r eac show test-timings src-auth

# Find slowest tests
r2r eac show test-timings | sort -k3 -nr | head -10
```

## See Also

- [get test-timings](../get/test-timings.md) - JSON output
- [show test-summary](./test-summary.md)
- [test](../test/test.md)

{{ diataxis_footer() }}
