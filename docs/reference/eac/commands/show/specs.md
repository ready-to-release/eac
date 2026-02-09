# show specs

<!-- book:cmd show specs -->

## Bundle Modules

For **container/bundle modules** with dependencies, specs are **aggregated from all dependent modules**.

Example: When querying `eac-ext` (which depends on `eac-commands` and `clie`):

```bash
eac show specs eac-ext
```

Shows specs from:

- `specs/eac-commands/` (dependency)
- `specs/clie/` (dependency)
- `specs/eac-ext/` (if any)

**Regular modules** only show specs from their own `specs/<module>/` directory.

## File Location

Specifications are read from: `specs/<module>/`

## See Also

- [get specs](../get/specs.md) - JSON/YAML output
- [show changelog](./changelog.md) - View changelog
- [show Commands](../categories/show.md)
