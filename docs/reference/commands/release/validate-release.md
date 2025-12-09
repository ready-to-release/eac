# validate release

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate release [module]`
**Purpose**: Validate changelog format and structure
**Category**: [release](../categories/release.md)

## Syntax

```bash
r2r eac validate release [module]
```

## Examples

```bash
# Validate changelog
r2r eac validate release

# After generating changelog
r2r eac release changelog
r2r eac validate release
```

## See Also

- [release changelog](./changelog.md)
- [release this](./this.md)
- [validate release-version](./validate-release-version.md)
- [release Commands](../categories/release.md)

{{ diataxis_footer() }}
