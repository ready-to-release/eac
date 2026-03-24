# Scan Commands

The **scan** category provides security scanning with multiple scanner types via the `--scanner` flag, plus a dedicated subcommand for dynamic testing.

**Key Features**:

- Multi-tool security scanning (Trivy, Semgrep, OWASP ZAP)
- SBOM generation for supply chain security
- Compliance validation (CIS, NIST)
- Secret detection
- Infrastructure as Code scanning
- Static and dynamic analysis

## Commands in this Category

| Command           | Purpose                                |
| ----------------- | -------------------------------------- |
| [scan](./scan.md) | Run security scans with --scanner flag |

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
eac scan
```

### Vulnerability Assessment

```bash
eac scan --scanner vuln,secrets
```

### Compliance Checking

```bash
eac scan --scanner compliance,sbom
```

### Application Security Testing

```bash
# Static analysis
eac scan --scanner sast

# Dynamic testing (requires running application)
eac scan zap eac-api --target http://localhost:8080
```

## See Also

- [validate Commands](../validate/index.md)
- [create risk-assess](../create/risk-assess.md)
- [validate control-tags](../validate/control-tags.md)
