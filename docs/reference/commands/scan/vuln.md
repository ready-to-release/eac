# scan vuln

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac scan-vuln [target]`
**Purpose**: Scan for vulnerabilities using Trivy
**Category**: [scan](../categories/scan.md)

## Syntax

```bash
r2r eac scan-vuln [target]
```

## Examples

```bash
# Scan filesystem
r2r eac scan-vuln .

# Scan container image
r2r eac scan-vuln docker.io/myapp:latest

# Scan with severity filter
r2r eac scan-vuln . --severity HIGH,CRITICAL
```

## See Also

- [scan sast](./sast.md)
- [scan iac](./iac.md)
- [scan Commands](../categories/scan.md)

{{ diataxis_footer() }}
