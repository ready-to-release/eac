# scan compliance

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac scan-compliance [target]`
**Purpose**: Check compliance with security standards using Trivy
**Category**: [scan](../categories/scan.md)

## Syntax

```bash
r2r eac scan-compliance [target]
```

## Examples

```bash
# Check compliance
r2r eac scan-compliance .

# Specific standard
r2r eac scan-compliance . --compliance cis

# Multiple standards
r2r eac scan-compliance . --compliance nist,pci
```

## See Also

- [scan vuln](./vuln.md)
- [scan iac](./iac.md)
- [scan Commands](../categories/scan.md)

{{ diataxis_footer() }}
