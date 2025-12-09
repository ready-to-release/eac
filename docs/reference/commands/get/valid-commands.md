# get valid-commands

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get valid-commands`
**Purpose**: Get all valid commands in structured format
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get valid-commands
```

## Examples

```bash
# Get all commands
r2r eac get valid-commands | jq '.'

# Extract command names
r2r eac get valid-commands | jq -r '.commands[].command'

# Filter by category
r2r eac get valid-commands | jq '.commands[] | select(.category == "show")'
```

## See Also

- [show valid-commands](../show/valid-commands.md) - Formatted table
- [get commands](./commands.md)

{{ diataxis_footer() }}
