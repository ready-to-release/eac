# validate go-tidy

<!-- book:cmd validate go-tidy -->

## What It Checks

- Discovers all Go modules from module contracts (modules with a `go` component).
- Runs `go mod tidy -diff` on each module directory.
- Reports modules where `go.mod` or `go.sum` are out of sync with source code.

## Common Errors

- **Modules with untidy dependencies** -- The `go.mod` or `go.sum` file is not synchronized. Fix with `eac update go-tidy`.

## See Also

- [validate](./validate.md)
- [validate Commands](../validate/index.md)
