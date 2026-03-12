# Serve design

<!-- book:cmd serve design -->

Launches a Docker container running Structurizr Lite and opens your default browser to view interactive C4 model diagrams defined in `workspace.dsl`.

The viewer runs on a dynamically allocated port (9000-9999) and updates automatically when the DSL file changes. Multiple instances can run simultaneously for different modules.

## Usage

```bash
eac serve design <module>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module moniker whose workspace.dsl to serve (required) |

The command reads from `specs/<module>/.design/workspace.dsl` and generates `workspace.json` and `.structurizr/` files (git-ignored) alongside it.

## Requirements

Docker must be running.

## Examples

```bash
eac serve design clie
eac serve design eac
```

## See Also

- [create design](../create/design.md)
- [update design](../update/design.md)
- [validate design](../validate/design.md)
