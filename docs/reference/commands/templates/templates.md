# templates

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac templates <subcommand>`
**Purpose**: Manage project templates for documentation and specifications
**Category**: [templates](../categories/templates.md)

## Subcommands

| Command | Purpose |
|---------|---------|
| [templates apply](./apply.md) | Apply with value replacements |
| [templates install](./install.md) | Install without replacements |
| [templates list](./list.md) | List template variables |
| [templates tags](./tags.md) | Extract template tags |

## Examples

```bash
# List variables in template
r2r eac templates list module.yml.tmpl

# Apply template with values
r2r eac templates apply module.yml.tmpl output.yml --values values.json

# Install template
r2r eac templates install README.md.tmpl docs/
```

## See Also

- [templates Commands Category](../categories/templates.md)

{{ diataxis_footer() }}
