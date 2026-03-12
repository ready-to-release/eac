# install

<!-- book:cmd templates install -->

Installs template files by copying them as-is to your project without variable substitution. All `{{ .Variable }}` placeholders are preserved for later customization.

## Usage

```bash
eac templates install <template-type> [flags]
```

## Template Types

| Type | Output Directory |
|------|------------------|
| `docs` | `docs/reference/` |
| `ai` | `.eac/templates/ai/` |
| `reports` | `.eac/templates/reports/` |
| `specs` | `specs/risk-controls/` |
| `claude` | `.claude/` |

## Examples

```bash
eac templates install docs
eac templates install ai
eac templates install claude
eac templates install docs --destination ./custom-docs
```

Use `help templates install <template-type>` for detailed information on each type.

## See Also

- [templates install-ai](./install-ai.md)
- [templates install-docs](./install-docs.md)
- [templates Commands](../categories/templates.md)
