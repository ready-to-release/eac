# pipeline ci-dispatch-and-wait

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac pipeline ci-dispatch-and-wait`
**Purpose**: Wait for GitHub workflow runs to complete
**Category**: [pipeline](../categories/pipeline.md)

## Syntax

```bash
r2r eac pipeline ci-dispatch-and-wait
```

## Examples

```bash
# Trigger and wait for CI
r2r eac pipeline ci-dispatch-and-wait

# Use in scripts
if r2r eac pipeline ci-dispatch-and-wait; then
  echo "CI passed"
  r2r eac release this
fi
```

## See Also

- [pipeline wait](./wait.md)
- [pipeline ci](./ci.md)
- [pipeline Commands](../categories/pipeline.md)

{{ diataxis_footer() }}
