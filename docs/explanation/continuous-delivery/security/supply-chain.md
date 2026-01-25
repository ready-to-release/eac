# Supply Chain Security

Protecting against vulnerabilities in dependencies and container images.

---

## Overview

Supply chain security focuses on managing risks from third-party code and container images that your application depends on.

### Key Risks

| Risk                  | Description                               |
| --------------------- | ----------------------------------------- |
| Known CVEs            | Published vulnerabilities in dependencies |
| Outdated Dependencies | Old versions with known issues            |
| License Compliance    | Incompatible or problematic licenses      |
| Supply Chain Attacks  | Compromised or malicious packages         |

### Defense Strategy

1. **Scan dependencies** - Identify known vulnerabilities
2. **Generate SBOM** - Track all components in your software
3. **Check licenses** - Ensure compliance with licensing requirements
4. **Scan containers** - Multi-layer image vulnerability analysis

---

## Reference Documentation

For CLI commands and scanning configuration, see:

**[Supply Chain Security Reference](../../../reference/eac/security/supply-chain.md)** - Complete implementation guide including:

- `eac scan --scanner vuln/sbom/compliance` commands
- Container scanning details
- Evidence output locations
- Stage-specific scanning guidance

---

## Related Documentation

- [SAST](./sast.md) - Static Application Security Testing
- [DAST](./dast.md) - Dynamic Application Security Testing
- [Shift-Left Security](./shift-left.md) - Security integration principles
- [Remediation](./remediation.md) - Handling security findings
