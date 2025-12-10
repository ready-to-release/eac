# scan secrets

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac scan-secrets [path]`
**Purpose**: Detect secrets and credentials using Trivy
**Category**: [scan](../categories/scan.md)

## Syntax

```bash
r2r eac scan-secrets [path]
```

## Examples

```bash
# Scan for secrets
r2r eac scan-secrets .

# Scan specific files
r2r eac scan-secrets config/

# Output to file
r2r eac scan-secrets . --output secrets-report.json
```

## See Also

- [scan vuln](./vuln.md)
- [scan sast](./sast.md)
- [scan Commands](../categories/scan.md)

{{ diataxis_footer() }}
