# validate risk-profile

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate risk-profile [file]`
**Purpose**: Validate OSCAL profile documents against OSCAL 1.1.2 schema
**Category**: [validate](../categories/validate.md)

## Syntax

```bash
r2r eac validate risk-profile [file]
```

## Examples

```bash
# Validate profile
r2r eac validate risk-profile profile.json

# Validate after creating
r2r eac create risk-profile assessment.md
r2r eac validate risk-profile
```

## See Also

- [validate](./validate.md)
- [create risk-profile](../create/risk-profile.md)
- [validate risk-catalog](./risk-catalog.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
