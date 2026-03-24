# validate design

<!-- book:cmd validate design -->

## Prerequisites

Docker must be running. The command uses the Structurizr CLI Docker image.

## Multi-DSL Support

Files in `specs/<module>/.design/`:

- `workspace.dsl`, `landscape.dsl` -- Validated as standalone workspaces.
- `_model.dsl`, `_styles.dsl` -- Files starting with `_` are fragments (for `!include`) and are skipped.

## Output

- Console: Human-readable validation summary.
- File: `out/logs/design/validation-results.json`.

## Common Errors

- **Docker is not running** -- Start Docker Desktop or the Docker daemon.
- **Module not found** -- The moniker does not exist. Check `eac show modules`.
- **DSL syntax error** -- Structurizr CLI found a syntax error. Check the reported line number.

## See Also

- [validate](./validate.md)
- [create design](../create/design.md)
- [serve design](../serve/design.md)
- [validate Commands](../validate/index.md)
