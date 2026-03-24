# show approve-summary

<!-- book:cmd show approve-summary -->

## Output Sections

- **Header**: "Release Approval: `<module>`"
- **Check table**: version (with type), tag, commit (shortened), and on success:
  - Changelog status (updated for semver, N/A for calver)
  - Existing Release check
  - CI Check (passed or skipped warning)

## See Also

- [show](show.md)
- [release](../release/index.md)
