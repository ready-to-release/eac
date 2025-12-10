# scan

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac scan <subcommand>`
**Purpose**: Security scanning and evidence collection for audit compliance
**Category**: [scan](../categories/scan.md)

## Subcommands

| Command | Purpose |
|---------|---------|
| [scan vuln](./vuln.md) | Vulnerability scanning |
| [scan sast](./sast.md) | Static code analysis |
| [scan secrets](./secrets.md) | Secret detection |
| [scan iac](./iac.md) | IaC misconfiguration |
| [scan sbom](./sbom.md) | Generate SBOM |
| [scan compliance](./compliance.md) | Compliance checking |
| [scan zap](./zap.md) | Dynamic testing (DAST) |

## Examples

```bash
# Run all security scans
r2r eac scan vuln .
r2r eac scan sast .
r2r eac scan secrets .

# Generate compliance evidence
r2r eac create risk-assess
```

## See Also

- [create risk-assess](../create/risk-assess.md)
- [scan Commands Category](../categories/scan.md)

{{ diataxis_footer() }}
