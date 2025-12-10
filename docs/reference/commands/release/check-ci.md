# release check-ci

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac release check-ci <commit>`
**Purpose**: Check CI status for a commit before releasing
**Category**: [release](../categories/release.md)

## Syntax

```bash
r2r eac release check-ci <commit>
```

## Examples

```bash
# Check current commit
r2r eac release check-ci $(git rev-parse HEAD)

# Check specific commit
r2r eac release check-ci abc123

# In release workflow
r2r eac release check-ci $(git rev-parse HEAD) && r2r eac release this
```

## See Also

- [release this](./this.md)
- [pipeline status](../pipeline/status.md)
- [release Commands](../categories/release.md)

{{ diataxis_footer() }}
