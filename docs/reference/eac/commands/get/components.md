# Get Components

<!-- book:cmd get components -->

Returns structured data about all components in the repository, including type, root path, phase support, and bidirectional dependencies.

## Usage

```bash
eac get components [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--module <name>` | string | Filter to a specific module |
| `--type <type>` | string | Filter by component type (go, typescript, book, etc.) |
| `--buildable` | bool | Only components with build phase |
| `--lintable` | bool | Only components with lint phase |
| `--testable` | bool | Only components with test phase |
| `--scannable` | bool | Only components with scan phase |

Phase flags can be combined for intersection filtering (e.g., `--lintable --scannable` returns components that are both lintable and scannable).

## Output Structure

Each component entry contains:

- `moniker` - Parent module identifier
- `component` - Component name
- `type` - Component type (go, typescript, book, dockerfile, etc.)
- `root` - Root path relative to repository
- `phases` - Phase support (build, lint, test, scan) with enabled status
- `dependencies.depends_on` - Components this component depends on
- `dependencies.depended_by` - Components that depend on this component

## Examples

```bash
# All components
eac get components

# Components in a specific module
eac get components --module eac

# Only Go components
eac get components --type go

# Buildable components only
eac get components --buildable
```

## See Also

- [show components](../show/components.md) - Human-readable markdown report
