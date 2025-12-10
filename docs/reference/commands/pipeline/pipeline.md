# pipeline

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac pipeline <subcommand>`
**Purpose**: Pipeline orchestration and CI/CD integration
**Category**: [pipeline](../categories/pipeline.md)

## Subcommands

| Command | Purpose |
|---------|---------|
| [pipeline run](./run.md) | Execute module pipelines |
| [pipeline status](./status.md) | Check CI status |
| [pipeline wait](./wait.md) | Wait for workflows |
| [pipeline ci](./ci.md) | CI orchestration |

## Examples

```bash
# Run pipeline for modules
r2r eac pipeline-run src-auth src-api

# Check CI status
r2r eac pipeline-status

# Wait for CI completion
r2r eac pipeline-wait
```

## See Also

- [get changed-modules-ci](../get/changed-modules-ci.md)
- [build](../other/build.md)
- [test](../test/test.md)
- [pipeline Commands Category](../categories/pipeline.md)

{{ diataxis_footer() }}
