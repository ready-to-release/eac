# create squash-message

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create squash-message [branch]`
**Purpose**: Generate squash commit message from branch commits
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create squash-message [branch]
```

## Examples

```bash
# Generate from current branch
r2r eac create squash-message

# Generate from specific branch
r2r eac create squash-message feature/auth

# Use in merge workflow
git checkout main
git merge --squash feature/auth
r2r eac create squash-message feature/auth > .git/SQUASH_MSG
git commit
```

## See Also

- [work merge](../work/merge.md)
- [create pr](./pr.md)
- [create commit-message](./commit-message.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
