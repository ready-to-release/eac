# show artifacts

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show artifacts <module>`
**Purpose**: Display artifacts for a module with status
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show artifacts <module>
```

## Examples

```bash
# Show artifacts
r2r eac show artifacts r2r-cli

# Verify after build
r2r eac build src-auth
r2r eac show artifacts src-auth
```

## See Also

- [get artifacts](../get/artifacts.md) - JSON output
- [build](../other/build.md) - Build modules
- [validate artifacts](../validate/artifacts.md)

{{ diataxis_footer() }}
