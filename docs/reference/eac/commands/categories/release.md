# release Commands

## Overview

The **release** category contains commands for release management and version control.

## Commands

<!-- book:category-commands release -->

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

## Key Features

- Automated changelog generation
- CI validation before release
- CalVer and SemVer support
- Tag management
- Version validation

## See Also

- [pipeline status](../pipeline/status.md)
- [validate release](../validate/release.md)
- [get squash-message](../get/squash-message.md)
