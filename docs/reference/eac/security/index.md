# Security Scanning Reference

Technical reference for EAC security scanning commands and configuration.

## In This Section

| Reference                         | Description                                     |
| --------------------------------- | ----------------------------------------------- |
| [Supply Chain](./supply-chain.md) | Dependency and container vulnerability scanning |
| [SAST](./sast.md)                 | Static Application Security Testing             |
| [DAST](./dast.md)                 | Dynamic Application Security Testing            |

## Quick Reference

```bash
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
