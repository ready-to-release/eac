# release Commands

{{ page_breadcrumb() }}

## Overview

The **release** category contains 11 commands for release management and version control.

## Commands

| Command                                                                | Purpose                                           |
| ---------------------------------------------------------------------- | ------------------------------------------------- |
| [release this](../release/this.md)                                     | Finalize changelog and prepare module for release |
| [release changelog](../release/changelog.md)                           | Generate/update changelog from commits            |
| [release check-ci](../release/check-ci.md)                             | Check CI status before releasing                  |
| [release get-version](../release/get-version.md)                       | Extract latest version from changelog             |
| [release pending](../release/pending.md)                               | Check for pending changes                         |
| [release tag-pending](../release/tag-pending.md)                       | Check for missing git tags                        |
| [release generate-module-calver](../release/generate-module-calver.md) | Generate calver tag for module                    |
| [release r2r-cli](../release/r2r-cli.md)                               | Release r2r-cli with semver                       |
| [validate release](../validate/release.md)                             | Validate changelog format                         |
| [validate release-version](../validate/release-version.md)             | Validate version format                           |

## Common Use Cases

### Complete Release Workflow

```bash
r2r eac release changelog
r2r eac validate release
r2r eac release check-ci $(git rev-parse HEAD)
r2r eac release this
```

### Version Management

```bash
r2r eac release pending
r2r eac release tag-pending
r2r eac release get-version
```

### Module Release

```bash
TAG=$(r2r eac release generate-module-calver src-auth)
git tag -a $TAG -m "Release $TAG"
```

## Key Features

- Automated changelog generation
- CI validation before release
- CalVer and SemVer support
- Tag management
- Version validation

## See Also

- [pipeline status](../pipeline/status.md)
- [validate release](../validate/release.md)
- [create squash-message](../create/squash-message.md)

{{ diataxis_footer() }}
