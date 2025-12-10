# scan Commands

{{ page_breadcrumb() }}

## Overview

The **scan** category contains 8 commands for security scanning and evidence collection for audit compliance.

## Commands

| Command | Purpose |
|---------|---------|
| [scan](../scan/scan.md) | Run all security scans |
| [scan vuln](../scan/vuln.md) | Scan for vulnerabilities using Trivy |
| [scan sast](../scan/sast.md) | Static Application Security Testing using Semgrep |
| [scan secrets](../scan/secrets.md) | Detect secrets and credentials using Trivy |
| [scan iac](../scan/iac.md) | Scan Infrastructure as Code for misconfigurations |
| [scan sbom](../scan/sbom.md) | Generate Software Bill of Materials |
| [scan compliance](../scan/compliance.md) | Check compliance with security standards |
| [scan zap](../scan/zap.md) | Dynamic Application Security Testing using OWASP ZAP |

## Common Use Cases

### Complete Security Scan

```bash
r2r eac scan
```

### Vulnerability Assessment

```bash
r2r eac scan vuln
r2r eac scan secrets
```

### Compliance Checking

```bash
r2r eac scan compliance
r2r eac scan sbom
```

### Application Security Testing

```bash
r2r eac scan sast
r2r eac scan zap http://localhost:8080
```

## Key Features

- Multi-tool security scanning (Trivy, Semgrep, OWASP ZAP)
- SBOM generation for supply chain security
- Compliance validation (CIS, NIST)
- Secret detection
- Infrastructure as Code scanning
- Static and dynamic analysis

## See Also

- [validate Commands](./validate.md)
- [create risk-assess](../create/risk-assess.md)
- [validate control-tags](../validate/control-tags.md)

{{ diataxis_footer() }}
