# pdf-screenshots

<!-- book:cmd update pdf-screenshots -->

Scans `out/build/` for generated PDF books and extracts each page as a PNG image. Images are stored in `.cache/eac/pdf-screenshots/` organized by book name with hash markers for cache invalidation.

Requires Docker (uses the `pdf-cli-oci` image for extraction).

## Usage

```bash
eac update pdf-screenshots [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--dry-run` | | Show what would be done without making changes |
| `--force` | `-f` | Regenerate all images ignoring cache |
| `--verbose` | `-v` | Show detailed progress |
| `--dpi` | | Image resolution, 72-300 (default: 150) |
| `--module` | `-m` | Process only a specific module's PDFs |

## Examples

```bash
eac update pdf-screenshots                       # Extract all PDFs
eac update pdf-screenshots --module=eac          # Single module only
eac update pdf-screenshots --force --dpi=300     # High-res regeneration
eac update pdf-screenshots --dry-run             # Preview extraction
```

## See Also

- [update docs](./docs.md)
- [update structurizr](./structurizr.md)
- [update Commands](../categories/update.md)
