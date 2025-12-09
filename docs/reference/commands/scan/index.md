# scan Commands

{{ page_breadcrumb() }}

Security scanning and evidence collection for audit compliance.

## Commands in this Category

| Command | Purpose |
|---------|---------|
| [scan](./scan.md) | Run all security scans |
| [scan vuln](./vuln.md) | Scan for vulnerabilities using Trivy |
| [scan sast](./sast.md) | Static Application Security Testing using Semgrep |
| [scan secrets](./secrets.md) | Detect secrets and credentials |
| [scan iac](./iac.md) | Scan Infrastructure as Code for misconfigurations |
| [scan sbom](./sbom.md) | Generate Software Bill of Materials |
| [scan compliance](./compliance.md) | Check compliance with security standards |
| [scan zap](./zap.md) | Dynamic Application Security Testing using OWASP ZAP |

## Quick Examples

```bash
# Run all security scans
r2r eac scan

# Scan for vulnerabilities
r2r eac scan vuln

# Detect secrets
r2r eac scan secrets
```

## See Also

- [Category Overview](../categories/scan.md)
- [validate Commands](../validate/index.md)

{{ diataxis_footer() }}
