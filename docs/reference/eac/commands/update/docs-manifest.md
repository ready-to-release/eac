<!-- book:cmd update docs-manifest -->

# update docs-manifest

Update the documentation assets manifest.

## Synopsis

```bash
r2r eac update docs-manifest [flags]
```

## Description

Scans `docs/assets/` for documentation assets (drawio diagrams, images, SVGs) and updates `docs/assets/manifest.json` with current usage information.

The manifest tracks:

- **Asset descriptions** - Human/LLM-authored, preserved on update
- **Usage references** - Auto-detected from markdown files
- **File metadata** - Size, hash, last modified
- **Statistics** - Total, used, unused by category

## Flags

| Flag            | Default | Description                                               |
| --------------- | ------- | --------------------------------------------------------- |
| `--check`       | `false` | Validate manifest is up-to-date (exits non-zero if stale) |
| `--dry-run`     | `false` | Show what would change without writing                    |
| `-v, --verbose` | `false` | Show detailed progress                                    |

## Examples

### Update manifest

```bash
r2r eac update docs-manifest
```

### Validate in CI

```bash
r2r eac update docs-manifest --check
```

### Preview changes

```bash
r2r eac update docs-manifest --dry-run
```

## Output

- Updates `docs/assets/manifest.json`
- Reports added/removed/changed assets
- Lists new assets needing descriptions

## See Also

- [update docs](./docs.md)
- [update structurizr](./structurizr.md)
