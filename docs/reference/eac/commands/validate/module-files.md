# Validate module-files

<!-- book:cmd validate module-files -->

Validates that all tracked files in the repository have proper module ownership. Every file should belong to exactly one module.

## Usage

```bash
eac validate module-files
```

## What It Checks

- **Unordered files** -- Files that fall into the catch-all "unordered" module because no module contract claims them.
- **Multi-ownership files** -- Files claimed by more than one module due to overlapping glob patterns.

## Examples

```bash
eac validate module-files
```

## Common Errors

- **Files in Unordered Module** -- These files are not claimed by any module. Create or update module contracts to claim them.
- **Files with Multi-Module Ownership** -- A file is matched by glob patterns in multiple modules. Adjust patterns to prevent overlap.

## See Also

- [validate](./validate.md)
- [get files](../get/files.md)
- [validate Commands](../categories/validate.md)
