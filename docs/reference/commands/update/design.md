# update design

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac update design <module>`
**Purpose**: Update existing workspace.dsl for a module using AI
**Category**: [update](../categories/update.md)

## Syntax

```bash
r2r eac update design <module>
```

## Examples

```bash
# Update architecture diagram
r2r eac update design src-auth

# Validate updated design
r2r eac validate design src-auth

# View changes
r2r eac serve design
```

## See Also

- [create design](../create/design.md)
- [validate design](../validate/design.md)
- [serve design](../serve/design.md)

{{ diataxis_footer() }}
