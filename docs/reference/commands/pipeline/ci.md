# pipeline ci

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac pipeline-ci <subcommand>`
**Purpose**: CI orchestration and diagnostics
**Category**: [pipeline](../categories/pipeline.md)

## Subcommands

| Command | Purpose |
|---------|---------|
| pipeline ci-dispatch-and-wait | Dispatch workflow and wait |
| pipeline ci-summary-link | Generate CI summary markdown |

## Examples

```bash
# Dispatch and wait for CI
r2r eac pipeline ci-dispatch-and-wait

# Generate summary for GH Actions
r2r eac pipeline ci-summary-link >> $GITHUB_STEP_SUMMARY
```

## See Also

- [pipeline wait](./wait.md)
- [pipeline status](./status.md)
- [pipeline Commands](../categories/pipeline.md)

{{ diataxis_footer() }}
