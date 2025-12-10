# release pending

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac release-pending [module]`
**Purpose**: Check if module has pending changes for release
**Category**: [release](../categories/release.md)

## Syntax

```bash
r2r eac release-pending [module]
```

## Examples

```bash
# Check for pending changes
r2r eac release-pending

# Check specific module
r2r eac release-pending r2r-cli

# Use in CI
if r2r eac release-pending; then
  r2r eac release this
fi
```

## See Also

- [release this](./this.md)
- [release tag-pending](./tag-pending.md)
- [release Commands](../categories/release.md)

{{ diataxis_footer() }}
