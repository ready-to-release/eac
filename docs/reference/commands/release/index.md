# release Commands
Release management and version control for changelogs and tagging.

## Commands in this Category

| Command | Purpose |
|---------|---------|
| [release](./release.md) | Base release command |
| [release this](./this.md) | Finalize changelog and prepare module for release |
| [release changelog](./changelog.md) | Generate/update changelog from commits |
| [release check-ci](./check-ci.md) | Check CI status before releasing |
| [release get-version](./get-version.md) | Extract latest version from changelog |
| [release pending](./pending.md) | Check for pending changes |
| [release tag-pending](./tag-pending.md) | Check for missing git tags |
| [release generate-module-calver](./generate-module-calver.md) | Generate calver tag for module |
| [release r2r-cli](./r2r-cli.md) | Release r2r-cli with semver |
| [validate release](./../validate/release.md) | Validate changelog format |
| [validate release-version](../validate/release-version.md) | Validate version format |

## Quick Examples

```bash
# Generate changelog
r2r eac release changelog

# Check what needs releasing
r2r eac release pending

# Create release
r2r eac release this
```

## See Also

- [Category Overview](../categories/release.md)
- [pipeline status](../pipeline/status.md)
