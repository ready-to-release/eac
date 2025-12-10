# get test-timings

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get test-timings [module]`
**Purpose**: Get test timing information from test logs
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get test-timings [module]
```

## Examples

```bash
# Get all test timings
r2r eac get test-timings | jq '.'

# Find slowest tests
r2r eac get test-timings | jq '[.tests[]] | sort_by(.duration) | reverse | .[0:10]'

# Average test time
r2r eac get test-timings | jq '.average'
```

## See Also

- [show test-timings](../show/test-timings.md) - Formatted table
- [get build-times](./build-times.md)

{{ diataxis_footer() }}
