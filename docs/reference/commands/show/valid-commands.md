# show valid-commands

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show valid-commands`
**Purpose**: Show all valid commands in a table
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show valid-commands
```

## Examples

```bash
# Show all commands
r2r eac show valid-commands

# Filter by category
r2r eac show valid-commands | grep "show"

# Count commands
r2r eac show valid-commands | grep -c "│"
```

## See Also

- [get valid-commands](../get/valid-commands.md) - JSON output
- [get commands](../get/commands.md) - With tree structure
- [show help](./help.md)

{{ diataxis_footer() }}
