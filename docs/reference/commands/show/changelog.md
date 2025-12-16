# show changelog

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show changelog <module> [version]`
**Purpose**: Display changelog entries in human-readable markdown format
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show changelog <module>
r2r eac show changelog <module> <version>
r2r eac show changelog <module> latest
r2r eac show changelog <module> unreleased
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module` | Yes | Module moniker (e.g., `ext-eac`, `eac-commands`) |
| `version` | No | Version number, or special keyword (`latest`, `unreleased`) |

## Special Keywords

| Keyword | Description |
|---------|-------------|
| `latest` | Show the most recent released version only |
| `unreleased` | Show pending changes not yet released |
| *(omit)* | Show all versions (unreleased + all releases) |

## Output Format

Markdown-formatted changelog with:

- Version headers (`## [version] - date`)
- Change categories (`### Added`, `### Changed`, `### Fixed`, etc.)
- Tables with columns: Description, Type, Scope, Commit
- Breaking changes marked with ⚠️ prefix

## Examples

```bash
# Show all versions
r2r eac show changelog ext-eac

# Show latest release
r2r eac show changelog ext-eac latest

# Show unreleased changes
r2r eac show changelog ext-eac unreleased

# Show specific version
r2r eac show changelog ext-eac 0.0.7

# Save to file
r2r eac show changelog eac-commands > changelog-output.md
```

## File Location

The command reads from: `release/<module>/CHANGELOG.md`

**Example:** For module `ext-eac`, reads from `release/ext-eac/CHANGELOG.md`

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| `module not found` | Invalid module moniker | Check `show modules` for valid names |
| `failed to parse changelog` | CHANGELOG.md missing or malformed | Verify file exists and follows Keep a Changelog format |
| `version not found` | Version doesn't exist | List versions with `get changelog <module> --as-json \| jq '.versions[].number'` |

## See Also

- [get changelog](../get/changelog.md) - Structured JSON/YAML output
- [show release-notes](./release-notes.md) - Release notes
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-changelog-release-notes.md)

{{ diataxis_footer() }}
