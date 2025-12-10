# create risk-profile

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create risk-profile <assessment-file>`
**Purpose**: Create OSCAL profile from risk assessment using AI
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create risk-profile <assessment-file>
```

## Examples

```bash
# Generate security profile
r2r eac create risk-profile risk-assessment.md

# Validate generated profile
r2r eac validate risk-profile
```

## See Also

- [create risk-assess](./risk-assess.md)
- [validate risk-profile](../validate/risk-profile.md)
- [scan](../categories/scan.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
