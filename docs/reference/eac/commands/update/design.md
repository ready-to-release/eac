# Update design

<!-- book:cmd update design -->

Updates an existing Structurizr DSL workspace file for a module by re-analyzing its source code using AI. The AI preserves the overall structure while incorporating new components, changed relationships, and removed elements. Updated workspaces are validated against Structurizr CLI before saving.

## Usage

```bash
eac update design <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module moniker to update (required) |

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--debug` | `-d` | Save intermediate outputs to `out/commands.log` |
| `--force` | `-f` | Overwrite workspace.dsl even if validation fails |
| `--output` | `-o` | Custom output path (default: `specs/<module>/.design/workspace.dsl`) |
| `--prompt` | | Custom AI prompt file path |

## Examples

```bash
eac update design clie
eac update design eac --debug
eac update design core -o ./custom/workspace.dsl
```

## See Also

- [create design](../create/design.md)
- [validate design](../validate/design.md)
- [serve design](../serve/design.md)
