# release generate-module-calver

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac release-generate-module-calver <module>`
**Purpose**: Generate a calver tag for a module
**Category**: [release](../categories/release.md)

## Syntax

```bash
r2r eac release-generate-module-calver <module>
```

## Examples

```bash
# Generate calver tag
TAG=$(r2r eac release-generate-module-calver r2r-cli)
echo "Generated tag: $TAG"

# Create git tag
git tag -a $TAG -m "Release $TAG"
```

## See Also

- [release this](./this.md)
- [release r2r-cli](./r2r-cli.md)
- [release Commands](../categories/release.md)

{{ diataxis_footer() }}
