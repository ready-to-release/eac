# create risk-assess

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac create risk-assess`
**Purpose**: Update OSCAL assessment-results with test and security evidence
**Category**: [create](../categories/create.md)

## Syntax

```bash
r2r eac create risk-assess
```

## Examples

```bash
# Generate assessment with evidence
r2r eac create risk-assess

# Run tests and scans first
r2r eac test --suite unit+integration+acceptance
r2r eac scan vuln
r2r eac create risk-assess
```

## See Also

- [create risk-profile](./risk-profile.md)
- [scan](../categories/scan.md)
- [test](../test/test.md)
- [create Commands](../categories/create.md)

{{ diataxis_footer() }}
