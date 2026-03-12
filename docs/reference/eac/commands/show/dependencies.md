# Show dependencies

<!-- book:cmd show dependencies -->

Displays the module dependency graph in a human-readable table format, showing which modules depend on which other modules.

## Usage

```bash
eac show dependencies
```

This command takes no flags.

## Output Sections

The report includes:

1. **Statistics** - Summary metrics: total modules, total dependencies, root modules (no dependencies), leaf modules (no dependents), max dependencies, and max dependents.
2. **Module Dependencies** - Table with columns: Module, Depends On, Used By. Shows both forward and reverse dependency relationships for every module.
3. **Display Order** - Modules ordered by dependency depth, group, and declaration order. Baseline modules (those with no dependencies) are marked. This ordering is used by other commands when grouping output by module.

## Examples

```bash
# Show full dependency graph
eac show dependencies
```

## See Also

- [show modules](./modules.md) - Module overview
- [show components](./components.md) - Component-level detail
