# work

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac work <subcommand>`
**Purpose**: Workspace management for parallel development using git worktrees
**Category**: [work](../categories/work.md)

## Subcommands

| Command | Purpose |
|---------|---------|
| [create](./create.md) | Create new workspace |
| [commit](./commit.md) | Commit with AI messages |
| [pull](./pull.md) | Sync with main |
| [merge](./merge.md) | Merge to main |
| [remove](./remove.md) | Remove workspace |

## Examples

```bash
# Complete workflow
r2r eac work create feature/auth
cd ../work/feature-auth

# Develop...
r2r eac work commit --all

# Sync and merge
r2r eac work pull
r2r eac work merge
r2r eac work remove feature/auth --delete-branch
```

## See Also

- [show workspaces](../show/workspaces.md) - List workspaces
- [work Commands Category](../categories/work.md) - Full documentation

{{ diataxis_footer() }}
