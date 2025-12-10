# show files-staged

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show files-staged`
**Purpose**: Show staged files with their module ownership
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show files-staged
```

## Examples

```bash
# Show staged files
r2r eac show files-staged

# Verify before commit
git add go/src/auth/*.go
r2r eac show files-staged
r2r eac work commit
```

## See Also

- [show files-changed](./files-changed.md) - Unstaged changes
- [work commit](../work/commit.md) - Commit staged files
- [create commit-message](../create/commit-message.md)

{{ diataxis_footer() }}
