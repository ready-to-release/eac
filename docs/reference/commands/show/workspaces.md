# show workspaces

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show workspaces`
**Purpose**: List all workspaces and their status
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show workspaces
```

## Examples

```bash
# List all workspaces
r2r eac show workspaces

# Find dirty workspaces
r2r eac show workspaces | grep "Dirty"

# Count active workspaces
r2r eac show workspaces | grep -c "feature"
```

## See Also

- [work create](../work/create.md) - Create workspace
- [work remove](../work/remove.md) - Remove workspace
- [work Commands](../categories/work.md)

{{ diataxis_footer() }}
