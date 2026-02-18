# Security Scanning Reference

Technical reference for EAC security scanning commands and risk configuration.

---

## Scan Commands

```bash
# Supply chain security
eac scan --scanner vuln        # Vulnerability scanning (Trivy)
eac scan --scanner sbom        # Software Bill of Materials (Trivy)
eac scan --scanner compliance  # Compliance checking (Trivy)

# Static analysis
eac scan --scanner sast        # Static code analysis (Semgrep)
eac scan --scanner secrets     # Secret detection (Trivy)
eac scan --scanner iac         # Infrastructure as Code (Trivy)

# Dynamic analysis
eac scan zap --url <url>       # DAST with OWASP ZAP

# Combined scanning
eac scan --scanner sast,secrets,vuln
```

**See**: [Scan Commands](../commands/scan/index.md)

---

## Scanner Types

| Scanner        | Tool    | Purpose                         | Category     |
| -------------- | ------- | ------------------------------- | ------------ |
| **vuln**       | Trivy   | Dependency vulnerabilities      | Supply Chain |
| **sbom**       | Trivy   | Software Bill of Materials      | Supply Chain |
| **compliance** | Trivy   | License compliance              | Supply Chain |
| **sast**       | Semgrep | Static code analysis            | Static       |
| **secrets**    | Trivy   | Hardcoded secrets detection     | Static       |
| **iac**        | Trivy   | Infrastructure as Code scanning | Static       |
| **zap**        | ZAP     | Dynamic application testing     | Dynamic      |

---

## Risk Configuration

Risk scoring configured in `contracts/scanner/0.1.0/schemas/defaults/risk-config.yml`.

**User override**: `.eac/risk-config.yml`

**Key settings**:

```yaml
scoring:
  impact: { api: 4, service: 4, library: 3, cli: 2 }
  criticality: { api: high, service: high, core: medium, cli: low }
```

**See**: [Risk Configuration](./risk-config.md)

---

## OSCAL Risk Assessment

```bash
# Generate OSCAL risk profile
eac create risk-profile assessment.md

# Create risk assessment results
eac create risk-assess --profile profile.json

# Validate OSCAL profile
eac validate risk-profile profile.json
```

**See**: [Risk Configuration](./risk-config.md), [Create Commands](../commands/create/index.md)

---

## Related Documentation

- **[Scan Commands](../commands/scan/index.md)** - Full scan command reference
- **[Supply Chain](./supply-chain.md)** - Dependency scanning details
- **[SAST](./sast.md)** - Static analysis details
- **[DAST](./dast.md)** - Dynamic analysis details
- **[Risk Configuration](./risk-config.md)** - Complete risk config reference
