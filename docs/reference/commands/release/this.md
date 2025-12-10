# release this

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac release this [module]`
**Purpose**: Finalize changelog and prepare module for release
**Category**: [release](../categories/release.md)

## Syntax

```bash
r2r eac release this [module]
```

## Examples

```bash
# Release current module
r2r eac release this

# Release specific module
r2r eac release this r2r-cli

# Full release workflow
r2r eac release changelog
r2r eac validate release
r2r eac release check-ci $(git rev-parse HEAD)
r2r eac release this
```

## See Also

- [release changelog](./changelog.md)
- [release check-ci](./check-ci.md)
- [validate release](./../validate/release.md)
- [release Commands](../categories/release.md)

{{ diataxis_footer() }}
