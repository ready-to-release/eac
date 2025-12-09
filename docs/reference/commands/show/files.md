# show files

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show files [module]`
**Purpose**: Show repository files with their module ownership
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show files [module]
```

## Examples

```bash
# Show all files
r2r eac show files

# Filter by module
r2r eac show files src-auth

# Find test files
r2r eac show files | grep "_test.go"
```

**Note**: Can be slow for large repositories (~2,690 files). Consider using `show files-changed` or `show files-staged`.

## See Also

- [show files-changed](./files-changed.md) - Unstaged changes
- [show files-staged](./files-staged.md) - Staged files
- [get files](../get/files.md) - JSON output (cacheable)

{{ diataxis_footer() }}
