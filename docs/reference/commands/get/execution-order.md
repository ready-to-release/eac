# get execution-order

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get execution-order <module...>`
**Purpose**: Get execution order for specific modules based on dependencies
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get execution-order <module...>
```

## Examples

```bash
# Get build order
r2r eac get execution-order r2r-cli | jq '.'

# Extract ordered list
r2r eac get execution-order r2r-cli | jq -r '.execution_order[]'

# Build in order
for module in $(r2r eac get execution-order r2r-cli | jq -r '.execution_order[]'); do
  r2r eac build $module
done
```

## See Also

- [get dependencies](./dependencies.md)
- [get build-deps](./build-deps.md)
- [build](../other/build.md)

{{ diataxis_footer() }}
