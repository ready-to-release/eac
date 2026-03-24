# validate docs

<!-- book:cmd validate docs -->

## What It Checks

Scans all `.md` files for references to known obsolete files:

- `module-types.yml` (deleted)
- `system-dependencies.yml` (deleted)

Automatically excludes `assets/cache/`, `.git/`, `node_modules/` directories and deprecation notice files.

## Common Errors

- **references obsolete file** -- A markdown file mentions a deleted config file. Update the documentation to use the current file name.

## See Also

- [validate](./validate.md)
