# show files-changed

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show files-changed`
**Purpose**: Show changed (modified, unstaged) files with their module ownership
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show files-changed
```

## Examples

```bash
# Show unstaged changes
r2r eac show files-changed

# Count changes per module
r2r eac show files-changed | grep "src-auth" | wc -l

# Build affected modules
CHANGED=$(r2r eac get changed-modules | jq -r '.changed_modules[]')
r2r eac build $CHANGED
```

## See Also

- [show files-staged](./files-staged.md) - Staged files
- [get changed-modules](../get/changed-modules.md) - Modules affected (JSON)
- [work commit](../work/commit.md) - Commit changes

{{ diataxis_footer() }}
