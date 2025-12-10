# templates install

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac templates install <template> <output>`
**Purpose**: Install templates without value replacements
**Category**: [templates](../categories/templates.md)

## Syntax

```bash
r2r eac templates install <template> <output>
```

## Examples

```bash
# Install template
r2r eac templates install module.yml.tmpl .eac/contracts/modules/

# Install docs template
r2r eac templates install-docs README.md.tmpl docs/
```

## See Also

- [templates apply](./apply.md)
- [templates Commands](../categories/templates.md)

{{ diataxis_footer() }}
