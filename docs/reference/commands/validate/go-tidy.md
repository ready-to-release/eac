# validate go-tidy

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate go-tidy`
**Purpose**: Validate Go module dependencies are tidy
**Category**: [validate](../categories/validate.md)

## Syntax

```bash
r2r eac validate go-tidy
```

## Examples

```bash
# Validate go.mod/go.sum are tidy
r2r eac validate go-tidy

# Fix if needed
go mod tidy
r2r eac validate go-tidy
```

## See Also

- [validate](./validate.md)
- [validate dependencies](./dependencies.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
