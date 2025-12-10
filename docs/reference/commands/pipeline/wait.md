# pipeline wait

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac pipeline-wait [run-id]`
**Purpose**: Wait for GitHub workflow runs to complete
**Category**: [pipeline](../categories/pipeline.md)

## Syntax

```bash
r2r eac pipeline-wait [run-id]
```

## Examples

```bash
# Wait for latest run
r2r eac pipeline-wait

# Wait for specific run
r2r eac pipeline-wait 12345678

# Use in CI
r2r eac pipeline ci-dispatch-and-wait
```

## See Also

- [pipeline ci](./ci.md)
- [pipeline status](./status.md)
- [pipeline Commands](../categories/pipeline.md)

{{ diataxis_footer() }}
