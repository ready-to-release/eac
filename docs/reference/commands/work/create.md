# work create

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac work create <branch> [--from <base>]`
**Purpose**: Create a new workspace for parallel development
**Category**: [work](../categories/work.md)

## Syntax

```bash
r2r eac work create <branch> [--from <base-branch>]
```

## Examples

```bash
# Create workspace from main
r2r eac work create feature/authentication
cd ../work/feature-authentication

# Create from specific branch
r2r eac work create hotfix/security --from release/v1.2.0

# Create and start working
r2r eac work create feature/api && cd ../work/feature-api
```

## See Also

- [work commit](./commit.md) - Commit with AI messages
- [work remove](./remove.md) - Remove workspace
- [show workspaces](../show/workspaces.md)
- [work Commands](../categories/work.md)

{{ diataxis_footer() }}
