# show modules

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show modules`
**Purpose**: Display all module contracts in a human-readable table
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show modules
```

## Examples

```bash
# Show all modules
r2r eac show modules

# Filter specific module
r2r eac show modules | grep "src-auth"

# Count modules
r2r eac show modules | wc -l
```

## See Also

- [get modules](../get/modules.md) - JSON output
- [show moduletypes](./moduletypes.md) - Group by type
- [show dependencies](./dependencies.md) - Dependency graph

{{ diataxis_footer() }}
