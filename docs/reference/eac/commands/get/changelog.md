# Get changelog

<!-- book:cmd get changelog -->

Returns parsed changelog data for a module in structured format. Can return the full changelog or a specific version's entry.

## Usage

```bash
eac get changelog <module> [version] [--as-yaml|--as-json|--as-toml]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module` | Yes | Module moniker |
| `version` | No | Specific version to return (omit for full changelog) |

## Flags

| Flag | Description |
|------|-------------|
| `--as-yaml` | Output as YAML (default) |
| `--as-json` | Output as JSON |
| `--as-toml` | Output as TOML |

## Output Fields

| Field | Description |
|-------|-------------|
| `module` | Module moniker |
| `title` | Changelog title |
| `version_type` | `semver` or `calver` |
| `unreleased` | Unreleased changes (if any) |
| `versions` | Array of version entries with dates and changes |

When a specific version is requested, only that version's entry is returned. Use `unreleased` as the version to get unreleased changes.

## Examples

```bash
# Full changelog for a module
eac get changelog core

# Specific version
eac get changelog core 1.2.0

# Unreleased changes as JSON
eac get changelog eac-cli unreleased --as-json
```

## See Also

- [show changelog](../show/changelog.md) - Human-readable output
- [get release-notes](./release-notes.md) - Release notes data
- [How-To Guide](../../../../how-to-guides/eac/commands/release-management/view-changelog-release-notes.md)
