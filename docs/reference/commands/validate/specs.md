# validate specs

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate specs [module]`
**Purpose**: Validate Gherkin specifications against quality contracts
**Category**: [validate](../categories/validate.md)

## Syntax

```bash
r2r eac validate specs [module]
```

## Examples

```bash
# Validate all specs
r2r eac validate specs

# Validate specific module
r2r eac validate specs src-auth
```

## See Also

- [validate](./validate.md)
- [create spec](../create/spec.md)
- [get specs unused-steps](../get/specs-unused-steps.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
