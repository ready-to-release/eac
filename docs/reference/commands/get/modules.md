# get modules

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get modules`
**Purpose**: Get all module contracts in the repository
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get modules
```

## Examples

```bash
# Get modules as JSON
r2r eac get modules | jq '.'

# Extract module names
r2r eac get modules | jq -r '.modules[].moniker'

# Filter by type
r2r eac get modules | jq '.modules[] | select(.type == "go-library")'
```

## See Also

- [show modules](../show/modules.md) - Formatted table
- [get dependencies](./dependencies.md)

{{ diataxis_footer() }}
