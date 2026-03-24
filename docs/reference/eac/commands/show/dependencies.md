# show dependencies

<!-- book:cmd show dependencies -->

## Output Sections

The report includes:

1. **Statistics** - Summary metrics: total modules, total dependencies, root modules (no dependencies), leaf modules (no dependents), max dependencies, and max dependents.
2. **Module Dependencies** - Table with columns: Module, Depends On, Used By. Shows both forward and reverse dependency relationships for every module.
3. **Display Order** - Modules ordered by dependency depth, group, and declaration order. Baseline modules (those with no dependencies) are marked. This ordering is used by other commands when grouping output by module.

## See Also

- [show modules](./modules.md) - Module overview
- [show components](./components.md) - Component-level detail
