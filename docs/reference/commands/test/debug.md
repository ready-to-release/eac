# test debug

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac test debug`
**Purpose**: Parse test results and list all failures
**Category**: [test](../categories/test.md)

## Syntax

```bash
r2r eac test debug
```

## Examples

```bash
# Run tests then debug
r2r eac test src-auth
r2r eac test debug

# Shows:
# - Failed test names
# - Error messages
# - File locations
# - Stack traces
```

## See Also

- [test](./test.md) - Run tests
- [show test-summary](../show/test-summary.md)
- [test Commands](../categories/test.md)

{{ diataxis_footer() }}
