# update docs-manifest

<!-- book:cmd update docs-manifest -->

## Synopsis

```bash
eac update docs-manifest [flags]
```

## Description

Scans `docs/assets/` for documentation assets (drawio diagrams, images, SVGs) and updates `docs/assets/manifest.json` with current usage information.

The manifest tracks:

- **Asset descriptions** - Human/LLM-authored, preserved on update
- **Usage references** - Auto-detected from markdown files
- **File metadata** - Size, hash, last modified
- **Statistics** - Total, used, unused by category

## Output

- Updates `docs/assets/manifest.json`
- Reports added/removed/changed assets
- Lists new assets needing descriptions

## See Also

- [update docs](./docs.md)
- [update structurizr](./structurizr.md)
