# docs

<!-- book:cmd update docs -->

Syncs documentation assets by rendering mermaid diagrams, optimizing drawio images, and syncing command reference docs (creating missing pages, removing orphans).

Results are stored in the `.cache/eac/` directory.

## Usage

```bash
eac update docs [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--area` | | Area to update: `mermaid`, `drawio`, `command-refs`, `all` (repeatable) |
| `--dry-run` | | Preview changes without applying them |
| `--force` | | Force re-render all assets even if cached |
| `--verbose` | `-v` | Show detailed progress |
| `--prune-cache` | | Remove orphaned mermaid/drawio cache files |

## Examples

```bash
eac update docs                          # Update all areas
eac update docs --area=mermaid           # Update mermaid diagrams only
eac update docs --dry-run                # Preview changes
eac update docs --prune-cache            # Remove orphaned cache files
eac update docs --force                  # Regenerate all cached assets
```

## See Also

- [update](../update/index.md)
- [serve docs](../serve/docs.md)
