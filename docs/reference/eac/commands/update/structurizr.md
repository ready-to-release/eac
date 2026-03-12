# structurizr

<!-- book:cmd update structurizr -->

Exports all Structurizr views from `workspace.dsl` files to SVG format and stores them in the `.cache/eac/structurizr/` acceleration cache.

Uses Docker to run the Structurizr CLI exporter. Cache is keyed by DSL file content hash.

## Usage

```bash
eac update structurizr [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--module` | `-m` | Export specific module only |
| `--force` | `-f` | Force re-export even if cache is current |
| `--verbose` | `-v` | Show detailed progress for each module |
| `--dry-run` | | Show what would be exported without exporting |

## Examples

```bash
eac update structurizr                       # Export all modules
eac update structurizr -m clie               # Export single module
eac update structurizr --force               # Force re-export all
eac update structurizr --dry-run             # Preview what needs export
```

## See Also

- [update](../update/index.md)
- [serve design](../serve/design.md)
- [Module Architecture](../../architecture/index.md)
