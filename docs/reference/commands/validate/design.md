# validate design

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate design [module]`
**Purpose**: Check workspace.dsl syntax using Structurizr CLI (requires Docker)
**Category**: [validate](../categories/validate.md)

## Syntax

```bash
r2r eac validate design [module]
```

## Examples

```bash
# Validate all designs
r2r eac validate design

# Validate specific module
r2r eac validate design src-auth
```

## See Also

- [validate](./validate.md)
- [create design](../create/design.md)
- [serve design](../serve/design.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
