# release

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac release <subcommand>`
**Purpose**: Release management and version control
**Category**: [release](../categories/release.md)

## Subcommands

| Command | Purpose |
|---------|---------|
| [release this](./this.md) | Finalize and release module |
| [release changelog](./changelog.md) | Generate/update changelog |
| [release check-ci](./check-ci.md) | Check CI status |
| [release get-version](./get-version.md) | Extract version |
| [release pending](./pending.md) | Check pending changes |
| [release tag-pending](./tag-pending.md) | Check missing tags |
| [release r2r-cli](./r2r-cli.md) | Release r2r-cli |
| [validate release](./validate-release.md) | Validate changelog |
| [validate release-version](./validate-release-version.md) | Validate version |

## Examples

```bash
# Complete release workflow
r2r eac release changelog
r2r eac validate release
r2r eac release check-ci $(git rev-parse HEAD)
r2r eac release this

# Check what needs releasing
r2r eac release pending
r2r eac release tag-pending
```

## See Also

- [release Commands Category](../categories/release.md)
- [pipeline status](../pipeline/status.md)

{{ diataxis_footer() }}
