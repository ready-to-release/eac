# pipeline status

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac pipeline-status`
**Purpose**: Show CI status for the head of trunk
**Category**: [pipeline](../categories/pipeline.md)

## Syntax

```bash
r2r eac pipeline-status
```

## Examples

```bash
# Check CI status
r2r eac pipeline-status

# Check before releasing
r2r eac pipeline-status
r2r eac release this
```

## See Also

- [pipeline run](./run.md)
- [release check-ci](../release/check-ci.md)
- [pipeline Commands](../categories/pipeline.md)

{{ diataxis_footer() }}
