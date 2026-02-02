# show specs

<!-- book:cmd show specs -->

## Bundle Modules

For **container/bundle modules** with dependencies, specs are **aggregated from all dependent modules**.

Example: When querying `ext-eac` (which depends on `eac-commands` and `r2r-cli`):

```bash
eac show specs ext-eac
```

Shows specs from:

- `specs/eac-commands/` (dependency)
- `specs/r2r-cli/` (dependency)
- `specs/ext-eac/` (if any)

**Regular modules** only show specs from their own `specs/<module>/` directory.

## File Location

Specifications are read from: `specs/<module>/`

## See Also

- [get specs](../get/specs.md) - JSON/YAML output
- [show changelog](./changelog.md) - View changelog
- [show Commands](../categories/show.md)
