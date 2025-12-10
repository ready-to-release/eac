# templates list

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac templates list <template>`
**Purpose**: List template placeholder variables
**Category**: [templates](../categories/templates.md)

## Syntax

```bash
r2r eac templates list <template>
```

## Examples

```bash
# List template variables
r2r eac templates list module.yml.tmpl

# Shows placeholders like:
# - {{MODULE_NAME}}
# - {{MODULE_TYPE}}
# - {{MODULE_PATH}}
```

## See Also

- [templates apply](./apply.md)
- [templates tags](./tags.md)
- [templates Commands](../categories/templates.md)

{{ diataxis_footer() }}
