# release changelog

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac release-changelog [module]`
**Purpose**: Generate or update changelog from commits
**Category**: [release](../categories/release.md)

## Syntax

```bash
r2r eac release-changelog [module]
```

## Examples

```bash
# Generate changelog
r2r eac release-changelog

# For specific module
r2r eac release-changelog r2r-cli

# Validate after generating
r2r eac validate release
```

## See Also

- [release this](./this.md)
- [validate release](./validate-release.md)
- [release Commands](../categories/release.md)

{{ diataxis_footer() }}
