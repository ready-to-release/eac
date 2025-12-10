# show dependencies

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show dependencies [module]`
**Purpose**: Show module dependency graph in a human-readable table
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show dependencies [module]
```

## Examples

```bash
# Show all dependencies
r2r eac show dependencies

# Show specific module dependencies
r2r eac show dependencies r2r-cli

# Find leaf modules
r2r eac show dependencies | grep "(none)"
```

## See Also

- [get dependencies](../get/dependencies.md) - JSON dependency graph
- [get execution-order](../get/execution-order.md) - Build order
- [validate dependencies](../validate/dependencies.md)

{{ diataxis_footer() }}
