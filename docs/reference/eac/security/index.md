# Security Scanning Reference

Technical reference for EAC security scanning commands and configuration.

## In This Section

| Reference                         | Description                                     |
| --------------------------------- | ----------------------------------------------- |
| [Risk Configuration](./risk-config.md) | Risk scoring and OSCAL profile configuration |
| [Supply Chain](./supply-chain.md) | Dependency and container vulnerability scanning |
| [SAST](./sast.md)                 | Static Application Security Testing             |
| [DAST](./dast.md)                 | Dynamic Application Security Testing            |

## Quick Reference

```bash
# Risk assessment
eac create risk-profile assessment.md    # Generate OSCAL profile
eac create risk-assess --profile ...     # Create assessment results
eac validate risk-profile profile.json   # Validate OSCAL profile

# Static analysis
eac scan --scanner sast
eac scan --scanner secrets
eac scan --scanner iac

# Supply chain security
eac scan --scanner vuln
eac scan --scanner sbom
eac scan --scanner compliance

# Dynamic analysis
eac scan --scanner zap

# Combined scanning
eac scan --scanner sast,secrets,vuln
```

## Related Documentation

- [Scan Command Reference](../commands/scan/index.md) - Full scan command options
- [Shift-Left Security (Conceptual)](../../../explanation/continuous-delivery/security/shift-left.md) - Security integration principles
