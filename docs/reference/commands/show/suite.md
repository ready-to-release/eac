# show suite

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac show suite <name>`
**Purpose**: Display detailed information about a test suite
**Category**: [show](../categories/show.md)

## Syntax

```bash
r2r eac show suite <name>
```

## Examples

```bash
# Show suite details
r2r eac show suite integration

# List available suites first
r2r eac test list-suites

# Then show specific suite
r2r eac show suite unit
```

## See Also

- [get suite](../get/suite.md) - JSON output
- [test suite](../test/suite.md) - Run suite
- [test list-suites](../test/list-suites.md)

{{ diataxis_footer() }}
