# get artifacts

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get artifacts <module>`
**Purpose**: Get resolved artifacts for a module
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get artifacts <module>
```

## Examples

```bash
# Get artifacts
r2r eac get artifacts r2r-cli | jq '.'

# Check if all exist
r2r eac get artifacts r2r-cli | jq -e '.all_exist'

# List missing artifacts
r2r eac get artifacts r2r-cli | jq '.artifacts[] | select(.exists == false) | .path'
```

## See Also

- [show artifacts](../show/artifacts.md) - Formatted table
- [build](../other/build.md)
- [validate artifacts](../validate/artifacts.md)

{{ diataxis_footer() }}
