# Get specs

<!-- book:cmd get specs -->

## Bundle Modules

For **container/bundle modules** with dependencies, specs are **aggregated from all dependent modules**.

**Example:** `ext-eac` depends on `eac-commands` and `r2r-cli`:

```bash
r2r eac get specs ext-eac --as-json
```

Returns specs from:

- `specs/eac-commands/` (dependency)
- `specs/r2r-cli/` (dependency)
- `specs/ext-eac/` (if any)

This provides a **complete view** of all specifications included in the release bundle.

**Regular modules** (without dependencies) only return specs from their own `specs/<module>/` directory.

## File Location

Specifications are read from: `specs/<module>/`

**Example:** For module `eac-commands`, reads from `specs/eac-commands/`

## See Also

- [show specs](../show/specs.md) - Human-readable markdown output
- [get changelog](./changelog.md) - Changelog data for same version
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-specifications.md)
