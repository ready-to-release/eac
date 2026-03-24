# get specs

<!-- book:cmd get specs -->

## Bundle Modules

For **container/bundle modules** with dependencies, specs are **aggregated from all dependent modules**.

Example: `eac-ext` depends on `eac-commands` and `clie`:

```bash
eac get specs eac-ext --as-json
```

Returns specs from:

- `specs/eac-commands/` (dependency)
- `specs/clie/` (dependency)
- `specs/eac-ext/` (if any)

## File Location

Specifications are read from: `specs/<module>/`

## See Also

- [show specs](../show/specs.md) - Human-readable output
- [get changelog](./changelog.md) - Changelog data
- [get Commands](../get/index.md)
