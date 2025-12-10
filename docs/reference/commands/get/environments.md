# get environments

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get environments`
**Purpose**: Get all environment contracts
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get environments
```

## Examples

```bash
# Get all environments
r2r eac get environments | jq '.'

# Filter by type
r2r eac get environments | jq '.environments[] | select(.type == "production")'

# Extract configuration
r2r eac get environments | jq -r '.environments[] | select(.name == "production") | .configuration.api_endpoint'
```

## See Also

- [show environments](../show/environments.md) - Formatted table

{{ diataxis_footer() }}
