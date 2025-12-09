# extension-meta

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac extension-meta`
**Purpose**: Output extension metadata for r2r CLI
**Category**: [other](../categories/other.md)

## Syntax

```bash
r2r eac extension-meta
```

## Output

Returns JSON metadata about the EAC extension for the r2r CLI tool.

```json
{
  "name": "eac",
  "version": "1.0.0",
  "description": "Engineering Automation Commands",
  "commands": [...],
  "author": "..."
}
```

## Use Case

This command is primarily used by the r2r CLI tool to discover and integrate EAC commands as an extension.

## Examples

```bash
# Output metadata
r2r eac extension-meta

# Validate extension format
r2r eac extension-meta | jq '.commands | length'
```

## See Also

- [get commands](../get/commands.md) - List all commands
- [show valid-commands](../show/valid-commands.md)
- [other Commands](../categories/other.md)

{{ diataxis_footer() }}
