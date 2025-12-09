# validate contracts

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate contracts`
**Purpose**: Validate repository contracts against JSON schemas
**Category**: [validate](../categories/validate.md)

## Syntax

```bash
r2r eac validate contracts
```

## Examples

```bash
# Validate all contracts
r2r eac validate contracts

# In CI
r2r eac validate contracts || exit 1
```

## See Also

- [validate](./validate.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
