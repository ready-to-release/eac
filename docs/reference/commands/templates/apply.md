# templates apply

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac templates apply <template> <output>`
**Purpose**: Apply templates with value replacements
**Category**: [templates](../categories/templates.md)

## Syntax

```bash
r2r eac templates apply <template> <output> [--values <file>]
```

## Examples

```bash
# Apply template with values
r2r eac templates apply module.yml.tmpl .eac/contracts/modules/new-module.yml --values values.json

# List template variables first
r2r eac templates list module.yml.tmpl
```

## See Also

- [templates install](./install.md)
- [templates list](./list.md)
- [templates Commands](../categories/templates.md)

{{ diataxis_footer() }}
