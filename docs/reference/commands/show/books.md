# show books

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show books`
**Purpose**: Display all book configurations in a human-readable table
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show books
```

## Examples

```bash
# Show all books
r2r eac show books

# Find stale books
r2r eac show books | grep "Stale"
```

## See Also

- [serve-docs](../serve/docs.md) - Serve documentation
- [validate books](../validate/books.md)

{{ diataxis_footer() }}
