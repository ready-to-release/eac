# show moduletypes

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show moduletypes`
**Purpose**: Show all module types grouped by count
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show moduletypes
```

## Examples

```bash
# Show module type distribution
r2r eac show moduletypes

# Sort by count
r2r eac show moduletypes | sort -k2 -nr
```

## See Also

- [show modules](./modules.md) - List all modules
- [get modules](../get/modules.md) - JSON output

{{ diataxis_footer() }}
