# show release-notes

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show release-notes <module> [version]`
**Purpose**: Display release notes in human-readable markdown format
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show release-notes <module>
r2r eac show release-notes <module> <version>
r2r eac show release-notes <module> latest
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module` | Yes | Module moniker (e.g., `ext-eac`, `eac-commands`) |
| `version` | No | Version number or `latest` keyword |

## Special Keywords

| Keyword | Description |
|---------|-------------|
| `latest` | Show the most recent released version (same as default) |
| *(omit)* | Show latest version (default behavior) |

## Output Format

Markdown-formatted release notes with:

- Version header (`## [version] - date`)
- Section headers (`### Conclusion on Fitness`, `### Impact on Business Process`, etc.)
- Section content as formatted markdown

## Examples

```bash
# Show latest release notes (default)
r2r eac show release-notes ext-eac

# Show latest release notes (explicit)
r2r eac show release-notes ext-eac latest

# Show specific version
r2r eac show release-notes ext-eac 0.0.7

# Save to file
r2r eac show release-notes ext-eac > release-notes.md
```

## File Location

The command reads from: `release/<module>/RELEASE-NOTES.md`

**Example:** For module `ext-eac`, reads from `release/ext-eac/RELEASE-NOTES.md`

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| `module not found` | Invalid module moniker | Check `show modules` for valid names |
| `failed to parse release notes` | RELEASE-NOTES.md missing or malformed | Verify file exists and has proper structure |
| `version not found` | Version doesn't exist | List versions with `get release-notes <module> --as-json \| jq '.versions[].number'` |

## See Also

- [get release-notes](../get/release-notes.md) - Structured JSON/YAML output
- [show changelog](./changelog.md) - Changelog
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-changelog-release-notes.md)

{{ diataxis_footer() }}
