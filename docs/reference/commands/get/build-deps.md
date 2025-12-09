# get build-deps

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get build-deps <module>`
**Purpose**: Get build dependencies for a module
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get build-deps <module>
```

## Examples

```bash
# Get build dependencies
r2r eac get build-deps r2r-cli | jq '.'

# Extract dependency names
r2r eac get build-deps r2r-cli | jq -r '.dependencies[].moniker'

# Count dependencies
r2r eac get build-deps r2r-cli | jq '.total'
```

## See Also

- [get dependencies](./dependencies.md) - Full dependency graph
- [get execution-order](./execution-order.md)
- [build](../other/build.md)

{{ diataxis_footer() }}
