# scan Commands

Security scanning and evidence collection for audit compliance.

## Commands in this Category

| Command | Purpose |
|---------|---------|
| [scan](./scan.md) | Run security scans with --scanner flag |
| [scan zap](./zap.md) | Dynamic Application Security Testing (DAST) |

## Scanner Types

The main `scan` command supports these scanner types via `--scanner` flag:

- `sbom` - Software Bill of Materials (Trivy)
- `vuln` - Vulnerability scanning (Trivy)
- `secrets` - Secret detection (Trivy)
- `iac` - Infrastructure as Code scanning (Trivy)
- `compliance` - Compliance checking (Trivy)
- `sast` - Static Application Security Testing (Semgrep)

## Quick Examples

```bash
# Run all default scans
r2r eac scan

# Specific scanner types
r2r eac scan --scanner vuln,secrets

# Multiple modules with specific scanners
r2r eac scan eac-core eac-commands --scanner sbom,vuln

# Dynamic testing (separate subcommand)
r2r eac scan zap eac-api --target http://localhost:8080
```

## See Also

- [Category Overview](../categories/scan.md)
- [validate Commands](../validate/index.md)
