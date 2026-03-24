# get changelog

<!-- book:cmd get changelog -->

## Output Fields

| Field | Description |
|-------|-------------|
| `module` | Module moniker |
| `title` | Changelog title |
| `version_type` | `semver` or `calver` |
| `unreleased` | Unreleased changes (if any) |
| `versions` | Array of version entries with dates and changes |

When a specific version is requested, only that version's entry is returned. Use `unreleased` as the version to get unreleased changes.

## See Also

- [show changelog](../show/changelog.md) - Human-readable output
- [get release-notes](./release-notes.md) - Release notes data
- [How-To Guide](../../../../how-to-guides/eac/commands/release-management/view-changelog-release-notes.md)
