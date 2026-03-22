# Release changelog

<!-- book:cmd release changelog -->

## How It Works

Generates or updates CHANGELOG.md from git commit history:

- **Commit Parsing**: Extracts conventional commit messages since last release
- **Categorization**: Groups changes by type (feat, fix, chore, docs, etc.)
- **Version Detection**: Identifies next version based on change types
- **Format**: Generates markdown following Keep a Changelog format
- **Validation**: Ensures changelog entries are valid for release

Used as part of the release workflow to document changes.

## See Also

- [release this](./this.md)
- [validate release](../validate/release.md)
- [release Commands](../../categories/release.md)
