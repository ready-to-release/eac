# get files

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get files [module]`
**Purpose**: Get repository files with their module ownership
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get files [module]
```

## Examples

```bash
# Get all files (cache recommended)
r2r eac get files > files.json

# Query from cache
jq '.files[] | select(.module == "src-auth")' files.json

# Find test files
r2r eac get files | jq '.files[] | select(.path | endswith("_test.go"))'
```

**Note**: Loads ~2,690 files. Consider caching or using `get changed-modules`.

## See Also

- [show files](../show/files.md) - Formatted table
- [get changed-modules](./changed-modules.md)

{{ diataxis_footer() }}
