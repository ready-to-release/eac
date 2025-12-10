# pipeline run

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac pipeline-run [module...]`
**Purpose**: Execute module pipelines respecting dependencies
**Category**: [pipeline](../categories/pipeline.md)

## Syntax

```bash
r2r eac pipeline-run [module...] [--all]
```

## Examples

```bash
# Run pipeline for module
r2r eac pipeline-run src-auth

# Run for all modules
r2r eac pipeline-run --all

# Run for changed modules
CHANGED=$(r2r eac get changed-modules-ci | jq -r '.changed_modules[]')
r2r eac pipeline-run $CHANGED
```

## See Also

- [pipeline status](./status.md)
- [pipeline ci](./ci.md)
- [pipeline Commands](../categories/pipeline.md)

{{ diataxis_footer() }}
