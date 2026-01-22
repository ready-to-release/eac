# Scan Commands

## Overview

The **scan** category provides security scanning with multiple scanner types via the `--scanner` flag, plus a dedicated subcommand for dynamic testing.

## Commands

<!-- book:category-commands scan -->

## Scanner Types

| Type         | Description                         | Tool    |
| ------------ | ----------------------------------- | ------- |
| `sbom`       | Software Bill of Materials          | Trivy   |
| `vuln`       | Vulnerability scanning              | Trivy   |
| `secrets`    | Secret detection                    | Trivy   |
| `iac`        | Infrastructure as Code scanning     | Trivy   |
| `compliance` | Compliance checking                 | Trivy   |
| `sast`       | Static Application Security Testing | Semgrep |

## Common Use Cases

### Complete Security Scan

```bash
r2r eac scan
```

### Vulnerability Assessment

```bash
r2r eac scan --scanner vuln,secrets
```

### Compliance Checking

```bash
r2r eac scan --scanner compliance,sbom
```

### Application Security Testing

```bash
# Static analysis
r2r eac scan --scanner sast

# Dynamic testing (requires running application)
r2r eac scan zap eac-api --target http://localhost:8080
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
