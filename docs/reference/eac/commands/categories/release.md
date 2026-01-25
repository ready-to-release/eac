# release Commands

## Overview

The **release** category contains 11 commands for release management and version control.

## Commands

<!-- book:category-commands release -->

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
