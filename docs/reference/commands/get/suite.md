# get suite

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get suite <name>`
**Purpose**: Get test suite information as structured data
**Category**: [get](../categories/get.md)

## Syntax

```bash
r2r eac get suite <name>
```

## Examples

```bash
# Get suite info
r2r eac get suite integration | jq '.'

# Extract modules
r2r eac get suite integration | jq -r '.modules[]'

# Get configuration
r2r eac get suite integration | jq '.configuration'
```

## See Also

- [show suite](../show/suite.md) - Formatted display
- [test suite](../test/suite.md) - Run suite
- [test list-suites](../test/list-suites.md)

{{ diataxis_footer() }}
