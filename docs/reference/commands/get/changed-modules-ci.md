# get changed-modules-ci

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get changed-modules-ci [--base <commit>]`
**Purpose**: Get modules requiring rebuild since last successful CI run
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get changed-modules-ci [--base <commit>]
```

## Examples

```bash
# Get changed modules in CI
r2r eac get changed-modules-ci | jq '.'

# Build changed modules with dependencies
CHANGED=$(r2r eac get changed-modules-ci | jq -r '.changed_modules[]')
for module in $CHANGED; do
  ORDER=$(r2r eac get execution-order $module | jq -r '.execution_order[]')
  r2r eac build $ORDER
done
```

## See Also

- [get changed-modules](./changed-modules.md) - Local changes
- [get execution-order](./execution-order.md)

{{ diataxis_footer() }}
