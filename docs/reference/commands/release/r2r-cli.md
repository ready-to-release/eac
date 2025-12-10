# release r2r-cli

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac release-r2r-cli <version>`
**Purpose**: Create a git tag for releasing r2r-cli using semver format
**Category**: [release](../categories/release.md)

## Syntax

```bash
r2r eac release-r2r-cli <version>
```

## Examples

```bash
# Release r2r-cli
r2r eac release-r2r-cli v1.2.3

# Full workflow
r2r eac release changelog r2r-cli
r2r eac validate release
r2r eac release check-ci $(git rev-parse HEAD)
r2r eac release-r2r-cli v1.2.3
```

## See Also

- [release this](./this.md)
- [release generate-module-calver](./generate-module-calver.md)
- [release Commands](../categories/release.md)

{{ diataxis_footer() }}
