# show test-summary

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show test-summary <module>`
**Purpose**: Generate formatted test summary for a module
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show test-summary <module>
```

## Examples

```bash
# Show test summary
r2r eac show test-summary src-auth

# After testing
r2r eac test src-auth
r2r eac show test-summary src-auth
```

## See Also

- [test](../test/test.md) - Run tests
- [show test-timings](./test-timings.md) - Performance analysis
- [test debug](../test/debug.md) - Debug failures

{{ diataxis_footer() }}
