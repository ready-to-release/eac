# pipeline ci-summary-link

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac pipeline ci-summary-link`
**Purpose**: Generate diagnostic markdown for CI summaries
**Category**: [pipeline](../categories/pipeline.md)

## Syntax

```bash
r2r eac pipeline ci-summary-link
```

## Examples

```bash
# Generate summary for GitHub Actions
r2r eac pipeline ci-summary-link >> $GITHUB_STEP_SUMMARY

# In workflow
- name: Generate Summary
  run: r2r eac pipeline ci-summary-link >> $GITHUB_STEP_SUMMARY
```

## See Also

- [pipeline ci](./ci.md)
- [pipeline status](./status.md)
- [pipeline Commands](../categories/pipeline.md)

{{ diataxis_footer() }}
