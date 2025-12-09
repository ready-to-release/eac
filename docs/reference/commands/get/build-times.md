# get build-times

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get build-times [module]`
**Purpose**: Get build timing information from build logs
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get build-times [module]
```

## Examples

```bash
# Get all build times
r2r eac get build-times | jq '.'

# Find slowest builds
r2r eac get build-times | jq '[.builds[]] | sort_by(.duration) | reverse | .[0:10]'

# Average build time
r2r eac get build-times | jq '.average'
```

## See Also

- [show build-times](../show/build-times.md) - Formatted table
- [get test-timings](./test-timings.md)

{{ diataxis_footer() }}
