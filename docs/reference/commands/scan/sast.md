# scan sast

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac scan-sast [path]`
**Purpose**: Static Application Security Testing using Semgrep
**Category**: [scan](../categories/scan.md)

## Syntax

```bash
r2r eac scan-sast [path]
```

## Examples

```bash
# Scan current directory
r2r eac scan-sast .

# Scan specific path
r2r eac scan-sast go/src/auth

# Output to file
r2r eac scan-sast . --output sast-report.json
```

## See Also

- [scan vuln](./vuln.md)
- [scan secrets](./secrets.md)
- [scan Commands](../categories/scan.md)

{{ diataxis_footer() }}
