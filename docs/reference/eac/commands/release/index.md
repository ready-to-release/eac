# Release Commands

The **release** category contains commands for release management and version control.

**Key Features**:

- Automated changelog generation
- CI validation before release
- CalVer and SemVer support
- Tag management
- Version validation

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

## Common Use Cases

### Complete Release Workflow

```bash
eac release changelog
eac validate release
eac release check-ci $(git rev-parse HEAD)
eac release this
```

### Version Management

```bash
eac release pending
eac release tag-pending
eac release get-version
```

### Module Release

```bash
TAG=$(eac release get-module-calver src-auth)
git tag -a $TAG -m "Release $TAG"
```

## See Also

- [pipeline status](../pipeline/status.md)
- [validate release](../validate/release.md)
- [get squash-message](../get/squash-message.md)
