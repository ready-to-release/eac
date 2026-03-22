# release Commands

Release management and version control for changelogs and tagging.

## Commands in this Category

| Command                                                       | Purpose                                           |
| ------------------------------------------------------------- | ------------------------------------------------- |
| [release](./release.md)                                       | Base release command                              |
| [release this](./this.md)                                     | Finalize changelog and prepare module for release |
| [release changelog](./changelog.md)                           | Generate/update changelog from commits            |
| [release check-ci](./check-ci.md)                             | Check CI status before releasing                  |
| [release get-version](./get-version.md)                       | Extract latest version from changelog             |
| [release pending](./pending.md)                               | Check for pending changes                         |
| [release tag-pending](./tag-pending.md)                       | Check for missing git tags                        |
| [release get-module-calver](./get-module-calver.md)           | Generate calver tag for module                    |
| [release prune-packages](./prune-packages.md)                 | Clean up old container images from GHCR           |
| [release clie](./clie.md)                               | Release clie with semver                       |
| [validate release](./../validate/release.md)                  | Validate changelog format                         |
| [validate release-version](../validate/release-version.md)    | Validate version format                           |

## Quick Examples

```bash
# Generate changelog
eac release changelog

# Check what needs releasing
eac release pending

# Create release
eac release this
```

## See Also

- [Category Overview](../../categories/release.md)
- [pipeline status](../pipeline/status.md)
