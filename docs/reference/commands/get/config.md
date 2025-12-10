# get config

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get config`
**Purpose**: Get all EAC configuration in structured format
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get config
```

## Examples

```bash
# Get full configuration
r2r eac get config | jq '.'

# Extract specific values
r2r eac get config | jq -r '.ai.provider'
r2r eac get config | jq -r '.repository.main_branch'

# Check if configured
r2r eac get config | jq -e '.ai.api_key' > /dev/null
```

## See Also

- [show config](../show/config.md) - Formatted display
- [init](../other/init.md) - Configure AI provider

{{ diataxis_footer() }}
