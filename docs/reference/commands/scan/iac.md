# scan iac

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac scan-iac [path]`
**Purpose**: Scan Infrastructure as Code for misconfigurations using Trivy
**Category**: [scan](../categories/scan.md)

## Syntax

```bash
r2r eac scan-iac [path]
```

## Examples

```bash
# Scan IaC files
r2r eac scan-iac .

# Scan Terraform
r2r eac scan-iac terraform/

# Scan Kubernetes manifests
r2r eac scan-iac k8s/
```

## See Also

- [scan vuln](./vuln.md)
- [scan compliance](./compliance.md)
- [scan Commands](../categories/scan.md)

{{ diataxis_footer() }}
