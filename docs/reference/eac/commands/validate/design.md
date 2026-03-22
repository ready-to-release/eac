# Validate design

<!-- book:cmd validate design -->

Validates Structurizr DSL files for syntax errors and structural issues using the official Structurizr CLI running in Docker.

## Usage

```bash
eac validate design <module> [flags]
eac validate design --all [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module moniker to validate (required unless `--all`) |

## Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--all` | `-a` | bool | Validate all DSL files in `specs/*/.design/` |
| `--file` | `-f` | string | Validate a specific DSL file (e.g., `--file=landscape`) |
| `--debug` | `-d` | bool | Save intermediate outputs and detailed logs |
| `--verbose` | `-v` | bool | Show Docker command and raw Structurizr CLI output |

## Prerequisites

Docker must be running. The command uses the Structurizr CLI Docker image.

## Multi-DSL Support

Files in `specs/<module>/.design/`:

- `workspace.dsl`, `landscape.dsl` -- Validated as standalone workspaces.
- `_model.dsl`, `_styles.dsl` -- Files starting with `_` are fragments (for `!include`) and are skipped.

## Output

- Console: Human-readable validation summary.
- File: `out/logs/design/validation-results.json`.

## Examples

```bash
# Validate all DSL files in a module
eac validate design clie

# Validate a specific DSL file
eac validate design clie --file=workspace

# Validate all modules
eac validate design --all

# Verbose output
eac validate design eac-cli --verbose
```

## Common Errors

- **Docker is not running** -- Start Docker Desktop or the Docker daemon.
- **Module not found** -- The moniker does not exist. Check `eac show modules`.
- **DSL syntax error** -- Structurizr CLI found a syntax error. Check the reported line number.

## See Also

- [validate](./validate.md)
- [create design](../create/design.md)
- [serve design](../serve/design.md)
- [validate Commands](../../categories/validate.md)
