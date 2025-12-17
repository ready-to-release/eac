# Show changelog

<!-- book:cmd show changelog -->

## Special Keywords

| Keyword      | Description                                   |
| ------------ | --------------------------------------------- |
| `latest`     | Show the most recent released version only    |
| `unreleased` | Show pending changes not yet released         |
| _(omit)_     | Show all versions (unreleased + all releases) |

## Output Format

Markdown-formatted changelog with:

- Version headers (`## [version] - date`)
- Change categories (`### Added`, `### Changed`, `### Fixed`, etc.)
- Tables with columns: Description, Type, Scope, Commit
- Breaking changes marked with ⚠️ prefix

## File Location

The command reads from: `release/<module>/CHANGELOG.md`

**Example:** For module `ext-eac`, reads from `release/ext-eac/CHANGELOG.md`

## Error Handling

| Error                       | Cause                             | Solution                                                                         |
| --------------------------- | --------------------------------- | -------------------------------------------------------------------------------- |
| `module not found`          | Invalid module moniker            | Check `show modules` for valid names                                             |
| `failed to parse changelog` | CHANGELOG.md missing or malformed | Verify file exists and follows Keep a Changelog format                           |
| `version not found`         | Version doesn't exist             | List versions with `get changelog <module> --as-json \| jq '.versions[].number'` |

## See Also

- [get changelog](../get/changelog.md) - Structured JSON/YAML output
- [show release-notes](./release-notes.md) - Release notes
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-changelog-release-notes.md)
