# go-tidy

<!-- book:cmd update go-tidy -->

Runs `go mod tidy` across all workspace modules to ensure `go.mod` and `go.sum` files are clean. Also runs `go work sync` at the repo root.

This is the fix counterpart to `validate go-tidy` (which only checks).

## Usage

```bash
eac update go-tidy [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--dry-run` | | Show what would change without modifying files |
| `--verbose` | `-v` | Show `go mod tidy` output for each module |

## Examples

```bash
eac update go-tidy                    # Tidy all modules
eac update go-tidy --dry-run          # Preview what would change
eac update go-tidy --verbose          # Show tidy output per module
```

## See Also

- [update go-mod-sums](./go-mod-sums.md)
- [validate go-tidy](../validate/go-tidy.md)
- [update](update.md)
