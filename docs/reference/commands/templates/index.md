# templates Commands

{{ page_breadcrumb() }}

Manage project templates for documentation and specifications.

## Commands in this Category

| Command | Purpose |
|---------|---------|
| [templates](./templates.md) | Base templates command |
| [templates apply](./apply.md) | Apply templates with value replacements |
| [templates apply-docs](./apply-docs.md) | Apply documentation templates |
| [templates install](./install.md) | Install templates without value replacements |
| [templates install-reports](./install-reports.md) | Install report templates |
| [templates list](./list.md) | List template placeholder variables |
| [templates tags](./tags.md) | Extract template tags |

## Quick Examples

```bash
# Install templates
r2r eac templates install

# Apply with values
r2r eac templates apply --module src-auth
```

## See Also

- [Category Overview](../categories/templates.md)
- [create spec](../create/spec.md)

{{ diataxis_footer() }}
