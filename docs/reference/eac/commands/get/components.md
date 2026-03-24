# get components

<!-- book:cmd get components -->

## Output Structure

Each component entry contains:

- `moniker` - Parent module identifier
- `component` - Component name
- `type` - Component type (go, typescript, book, dockerfile, etc.)
- `root` - Root path relative to repository
- `phases` - Phase support (build, lint, test, scan) with enabled status
- `dependencies.depends_on` - Components this component depends on
- `dependencies.depended_by` - Components that depend on this component

## See Also

- [show components](../show/components.md) - Human-readable markdown report
