# show build-summary

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show build-summary <module>`
**Purpose**: Generate formatted build summary for a module
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show build-summary <module>
```

## Examples

```bash
# Show build summary
r2r eac show build-summary src-auth

# After building
r2r eac build src-auth
r2r eac show build-summary src-auth
```

## See Also

- [build](../other/build.md) - Build modules
- [show build-times](./build-times.md) - Performance analysis
- [get build-times](../get/build-times.md) - JSON output

{{ diataxis_footer() }}
