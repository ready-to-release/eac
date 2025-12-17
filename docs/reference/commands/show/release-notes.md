# Show release-notes

<!-- book:cmd show release-notes -->

## Special Keywords

| Keyword  | Description                                             |
| -------- | ------------------------------------------------------- |
| `latest` | Show the most recent released version (same as default) |
| _(omit)_ | Show latest version (default behavior)                  |

## Output Format

Markdown-formatted release notes with:

- Version header (`## [version] - date`)
- Section headers (`### Conclusion on Fitness`, `### Impact on Business Process`, etc.)
- Section content as formatted markdown

## File Location

The command reads from: `release/<module>/RELEASE-NOTES.md`

**Example:** For module `ext-eac`, reads from `release/ext-eac/RELEASE-NOTES.md`

## Error Handling

| Error                           | Cause                                 | Solution                                                                             |
| ------------------------------- | ------------------------------------- | ------------------------------------------------------------------------------------ |
| `module not found`              | Invalid module moniker                | Check `show modules` for valid names                                                 |
| `failed to parse release notes` | RELEASE-NOTES.md missing or malformed | Verify file exists and has proper structure                                          |
| `version not found`             | Version doesn't exist                 | List versions with `get release-notes <module> --as-json \| jq '.versions[].number'` |

## See Also

- [get release-notes](../get/release-notes.md) - Structured JSON/YAML output
- [show changelog](./changelog.md) - Changelog
- [How-To Guide](../../../how-to-guides/eac/commands/release-management/view-changelog-release-notes.md)
