# get specs unused-steps

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get specs unused-steps [module]`
**Purpose**: Detect unused godog step definitions
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get specs unused-steps [module]
```

## Examples

```bash
# Get unused steps
r2r eac get specs unused-steps | jq '.'

# List step names
r2r eac get specs unused-steps | jq -r '.unused_steps[].step'

# Count by module
r2r eac get specs unused-steps | jq '.by_module'
```

## See Also

- [validate specs](../validate/specs.md)
- [get tests](./tests.md)

{{ diataxis_footer() }}
