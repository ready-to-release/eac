# show

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show <subcommand>`
**Purpose**: Display repository information in human-readable format
**Category**: [show](../categories/show.md)

## Subcommands

All show commands return formatted tables and text for interactive terminal use.

See the [show Commands Category](../categories/show.md) for complete list of subcommands.

## Examples

```bash
# Show modules
r2r eac show modules

# Show dependencies
r2r eac show dependencies

# Show changed files
r2r eac show files-changed

# Show test results
r2r eac show test-summary
```

## Common Patterns

```bash
# Filter output with grep
r2r eac show modules | grep "src-auth"

# Count items
r2r eac show modules | wc -l

# View with pager
r2r eac show dependencies | less
```

## See Also

- [get Commands](../categories/get.md) - JSON output for automation
- [show Commands Category](../categories/show.md) - All show commands

{{ diataxis_footer() }}
