# get specs

<!-- book:cmd get specs -->

## Bundle Modules

For **container/bundle modules** with dependencies, specs are **aggregated from all dependent modules**.

Example: `ext-eac` depends on `eac-commands` and `r2r-cli`:

```bash
eac get specs ext-eac --as-json
```

Returns specs from:

- `specs/eac-commands/` (dependency)
- `specs/r2r-cli/` (dependency)
- `specs/ext-eac/` (if any)

## File Location

Specifications are read from: `specs/<module>/`

## See Also

- [show specs](../show/specs.md) - Human-readable output
- [get changelog](./changelog.md) - Changelog data
- [get Commands](../categories/get.md)
