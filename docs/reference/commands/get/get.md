# get

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get <subcommand>`
**Purpose**: Retrieve repository data in structured format
**Category**: [get](../categories/get.md)

## Subcommands

All get commands return JSON output for automation and scripting. Process with `jq` or pipe to other tools.

See the [get Commands Category](../categories/get.md) for complete list of subcommands.

## Examples

```bash
# Get modules
r2r eac get modules | jq '.modules[].moniker'

# Get dependencies
r2r eac get dependencies | jq '.dependencies["r2r-cli"]'

# Get changed modules (local)
r2r eac get changed-modules | jq -r '.changed_modules[]'

# Get changed modules (CI)
r2r eac get changed-modules-ci | jq '.changed_modules'
```

## Common Patterns

```bash
# Cache expensive queries
r2r eac get files > files.json
jq '.files[] | select(.module == "src-auth")' files.json

# Build changed modules
CHANGED=$(r2r eac get changed-modules | jq -r '.changed_modules[]')
r2r eac build $CHANGED
```

## See Also

- [show Commands](../categories/show.md) - Human-readable output
- [get Commands Category](../categories/get.md) - All get commands

{{ diataxis_footer() }}
