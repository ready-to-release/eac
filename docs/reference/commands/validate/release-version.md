# validate release-version

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate release-version <version>`
**Purpose**: Validate release version format
**Category**: [validate](../categories/validate.md)

## Syntax

```bash
r2r eac validate release-version <version>
```

## Examples

```bash
# Validate version format
r2r eac validate release-version v1.2.3

# Check before releasing
VERSION=$(r2r eac release get-version)
r2r eac validate release-version $VERSION
```

## See Also

- [validate release](./release.md)
- [release get-version](../release/get-version.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
