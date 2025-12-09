# validate artifacts

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac validate artifacts <module>`
**Purpose**: Validate that build artifacts exist for a module and its dependencies
**Category**: [validate](../categories/validate.md)

## Syntax

```bash
r2r eac validate artifacts <module>
```

## Examples

```bash
# Validate artifacts exist
r2r eac validate artifacts r2r-cli

# After building
r2r eac build r2r-cli
r2r eac validate artifacts r2r-cli
```

## See Also

- [validate](./validate.md)
- [get artifacts](../get/artifacts.md)
- [build](../other/build.md)
- [validate Commands](../categories/validate.md)

{{ diataxis_footer() }}
