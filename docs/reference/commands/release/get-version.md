# release get-version

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac release-get-version [module]`
**Purpose**: Extract latest version from changelog
**Category**: [release](../categories/release.md)

## Syntax

```bash
r2r eac release-get-version [module]
```

## Examples

```bash
# Get version from changelog
VERSION=$(r2r eac release-get-version)
echo "Release version: $VERSION"

# For specific module
r2r eac release-get-version r2r-cli
```

## See Also

- [release changelog](./changelog.md)
- [release this](./this.md)
- [release Commands](../categories/release.md)

{{ diataxis_footer() }}
