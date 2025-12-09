# get changed-modules

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get changed-modules`
**Purpose**: Get modules affected by changed files
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get changed-modules
```

## Examples

```bash
# Get changed modules
r2r eac get changed-modules | jq '.'

# Extract module list
r2r eac get changed-modules | jq -r '.changed_modules[]'

# Build changed modules
CHANGED=$(r2r eac get changed-modules | jq -r '.changed_modules[]')
r2r eac build $CHANGED
```

## See Also

- [get changed-modules-ci](./changed-modules-ci.md) - For CI pipelines
- [show files-changed](../show/files-changed.md)
- [build](../other/build.md)

{{ diataxis_footer() }}
