# validate module-files

<!-- book:cmd validate module-files -->

## What It Checks

- **Unordered files** -- Files that fall into the catch-all "unordered" module because no module contract claims them.
- **Multi-ownership files** -- Files claimed by more than one module due to overlapping glob patterns.

## Common Errors

- **Files in Unordered Module** -- These files are not claimed by any module. Create or update module contracts to claim them.
- **Files with Multi-Module Ownership** -- A file is matched by glob patterns in multiple modules. Adjust patterns to prevent overlap.

## See Also

- [validate](./validate.md)
- [get files](../get/files.md)
- [validate Commands](../validate/index.md)
