# show build-times

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show build-times [module]`
**Purpose**: Show build timing analysis in a human-readable format
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show build-times [module]
```

## Examples

```bash
# Show all build times
r2r eac show build-times

# Filter by module
r2r eac show build-times src-auth

# Find slowest builds
r2r eac show build-times | sort -k2 -nr
```

## See Also

- [get build-times](../get/build-times.md) - JSON output
- [show build-summary](./build-summary.md)
- [build](../other/build.md)

{{ diataxis_footer() }}
