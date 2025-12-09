# validate release

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate release [module]`
**Purpose**: Validate changelog format and structure
**Category**: [validate](../categories/validate.md)

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

- [release changelog](../release/changelog.md)
- [release this](../release/this.md)
- [validate release-version](./release-version.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
