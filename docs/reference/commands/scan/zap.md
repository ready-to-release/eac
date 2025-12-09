# scan zap

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac scan-zap <url>`
**Purpose**: Dynamic Application Security Testing using OWASP ZAP
**Category**: [scan](../categories/scan.md)

## Syntax

```bash
r2r eac scan-zap <url>
```

## Examples

```bash
# Scan web application
r2r eac scan-zap http://localhost:8080

# Full scan
r2r eac scan-zap http://localhost:8080 --scan-type full

# Baseline scan
r2r eac scan-zap http://localhost:8080 --scan-type baseline
```

## See Also

- [scan vuln](./vuln.md)
- [scan sast](./sast.md)
- [scan Commands](../categories/scan.md)

{{ diataxis_footer() }}
