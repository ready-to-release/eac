# validate dependencies

<!-- book:cmd validate dependencies -->

## What It Checks

- Loads module contracts from the workspace.
- Builds a dependency graph from all `go.mod` files (excluding `out/`, `vendor/`, `.git/`, `node_modules/`, `tools/`, `templates/`).
- Compares the actual Go module dependencies against declared contract dependencies.
- Reports extra or missing dependencies.

## Common Errors

- **Extra dependency** -- A `go.mod` file imports a workspace module not declared in the contract. Update the contract to declare it.
- **Missing dependency** -- A contract declares a dependency not present in `go.mod`. Add the import or remove the contract declaration.

## See Also

- [get dependencies](../get/dependencies.md)
- [show dependencies](../show/dependencies.md)
- [validate](./validate.md)
