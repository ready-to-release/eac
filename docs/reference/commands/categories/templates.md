# templates Commands

{{ page_breadcrumb() }}

## Overview

The **templates** category contains 5 commands for managing project templates for documentation and specifications.

## Commands

| Command                                      | Purpose                                      |
| -------------------------------------------- | -------------------------------------------- |
| [templates](../templates/templates.md)       | Manage project templates                     |
| [templates apply](../templates/apply.md)     | Apply templates with value replacements      |
| [templates install](../templates/install.md) | Install templates without value replacements |
| [templates list](../templates/list.md)       | List template placeholder variables          |
| [templates tags](../templates/tags.md)       | Extract template tags                        |

## Common Use Cases

### Install Templates

```bash
r2r eac templates install
```

### Apply Templates with Values

```bash
r2r eac templates apply --module src-auth
```

### List Template Variables

```bash
r2r eac templates list
```

## Key Features

- Template installation for documentation
- Variable substitution and value replacement
- Placeholder discovery
- Tag extraction for navigation

## See Also

- [create spec](../create/spec.md)
- [validate markdown](../validate/markdown.md)
- [validate books](../validate/books.md)

{{ diataxis_footer() }}
