# Release this

<!-- book:cmd release this -->

## How It Works

Finalizes and creates a release for a module:

- **Changelog Validation**: Verifies CHANGELOG.md has unreleased version entry
- **Version Extraction**: Extracts version from changelog
- **CI Verification**: Ensures CI passed for current commit
- **Tag Creation**: Creates git tag with version (semver or calver)
- **Release Artifacts**: Prepares release bundle and metadata

The command will fail if CI hasn't passed or changelog validation fails.

## See Also

- [release changelog](./changelog.md)
- [release check-ci](./check-ci.md)
- [validate release](./../validate/release.md)
- [release Commands](../categories/release.md)
