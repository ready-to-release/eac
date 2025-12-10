# validate dependencies

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate dependencies`
**Purpose**: Validate module dependencies from go.mod files against contracts
**Category**: [validate](../categories/validate.md)

## Syntax

```bash
r2r eac validate dependencies
```

## Examples

```bash
# Validate dependencies
r2r eac validate dependencies

# Check specific module
r2r eac validate dependencies src-auth
```

## See Also

- [validate](./validate.md)
- [get dependencies](../get/dependencies.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
