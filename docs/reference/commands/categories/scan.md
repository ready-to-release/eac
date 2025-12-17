# Scan Commands

## Overview

The **scan** category contains 8 commands for security scanning and evidence collection for audit compliance.

## Commands

<!-- book:category-commands scan -->

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
