# go-mod-sums

<!-- book:cmd update go-mod-sums -->

Downloads all declared dependencies and refreshes `go.sum` checksums across all workspace modules without modifying `go.mod` files.

Parses `go.work` to discover modules, runs `go work sync` at the repo root, then runs `go mod download` in each module directory.

Use `update go-tidy` instead if you need to add or remove dependencies in `go.mod`.

## Usage

```bash
eac update go-mod-sums [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--dry-run` | | Show modules with stale `go.sum` without modifying them |
| `--verbose` | `-v` | Show `go mod download` output for each module |

## Examples

```bash
eac update go-mod-sums                # Sync all go.sum files
eac update go-mod-sums --dry-run      # Check which go.sum files are stale
eac update go-mod-sums --verbose      # Show download output
```

## See Also

- [update go-tidy](./go-tidy.md)
- [validate go-tidy](../validate/go-tidy.md)
- [update](update.md)
