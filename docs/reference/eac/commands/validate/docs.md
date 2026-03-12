# Validate docs

<!-- book:cmd validate docs -->

Validates documentation files for references to obsolete or deleted configuration files.

## Usage

```bash
eac validate docs [docs-directory]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `docs-directory` | Directory to scan (default: `docs/`) |

## What It Checks

Scans all `.md` files for references to known obsolete files:

- `module-types.yml` (deleted)
- `system-dependencies.yml` (deleted)

Automatically excludes `assets/cache/`, `.git/`, `node_modules/` directories and deprecation notice files.

## Examples

```bash
# Validate all documentation
eac validate docs

# Validate a specific directory
eac validate docs docs/reference
```

## Common Errors

- **references obsolete file** -- A markdown file mentions a deleted config file. Update the documentation to use the current file name.

## See Also

- [validate](./validate.md)
