# work pull

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac work pull`
**Purpose**: Sync workspace with latest main via rebase
**Category**: [work](../categories/work.md)

## Syntax

```bash
r2r eac work pull
```

## Examples

```bash
# Sync with main
r2r eac work pull

# If conflicts occur:
# 1. Resolve conflicts in files
# 2. git add <resolved-files>
# 3. git rebase --continue
```

## See Also

- [work create](./create.md)
- [work merge](./merge.md)
- [work Commands](../categories/work.md)

{{ diataxis_footer() }}
