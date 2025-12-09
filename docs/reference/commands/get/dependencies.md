# get dependencies

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get dependencies [module]`
**Purpose**: Get module dependency graph in structured format
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get dependencies [module]
```

## Examples

```bash
# Get dependency graph
r2r eac get dependencies | jq '.'

# Get dependencies for module
r2r eac get dependencies | jq '.dependencies["r2r-cli"]'

# Find modules with no dependencies
r2r eac get dependencies | jq 'to_entries[] | select(.value | length == 0) | .key'
```

## See Also

- [show dependencies](../show/dependencies.md) - Formatted table
- [get execution-order](./execution-order.md)

{{ diataxis_footer() }}
