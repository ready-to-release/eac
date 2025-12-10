# show environments

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show environments`
**Purpose**: Show all environment contracts in a human-readable table
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show environments
```

## Examples

```bash
# Show all environments
r2r eac show environments

# Filter production environments
r2r eac show environments | grep "production"

# Count environments
r2r eac show environments | tail -n +4 | wc -l
```

## See Also

- [get environments](../get/environments.md) - JSON output

{{ diataxis_footer() }}
