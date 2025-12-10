# create design

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create design <module>`
**Purpose**: Generate workspace.dsl for a module using AI
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create design <module>
```

## Examples

```bash
# Generate architecture diagram
r2r eac create design src-auth

# Validate generated design
r2r eac validate design src-auth

# View in browser
r2r eac serve design
```

## See Also

- [update design](../update/design.md)
- [validate design](../validate/design.md)
- [serve design](../serve/design.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
