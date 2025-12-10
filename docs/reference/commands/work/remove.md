# work remove

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac work remove <workspace> [--delete-branch]`
**Purpose**: Remove workspace and optionally delete associated branches
**Category**: [work](../categories/work.md)

## Syntax

```bash
r2r eac work remove <workspace> [--delete-branch]
```

## Options

| Flag | Description |
|------|-------------|
| `--delete-branch` | Also delete the git branch |

## Examples

```bash
# Remove workspace only
r2r eac work remove feature/auth

# Remove workspace and branch
r2r eac work remove feature/auth --delete-branch

# After merge
r2r eac work merge
r2r eac work remove feature/auth --delete-branch
```

## See Also

- [work create](./create.md)
- [show workspaces](../show/workspaces.md)
- [work Commands](../categories/work.md)

{{ diataxis_footer() }}
