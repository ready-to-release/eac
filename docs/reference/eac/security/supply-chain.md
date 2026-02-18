# Supply Chain Security

Dependency and container vulnerability scanning using Trivy.

---

## Scanners

```bash
# Vulnerability scanning
eac scan --scanner vuln        # CVE detection in dependencies

# Software Bill of Materials
eac scan --scanner sbom        # Generate SBOM

# License compliance
eac scan --scanner compliance  # Check license compatibility

# All supply chain scanners
eac scan --scanner vuln,sbom,compliance
```

**Tool**: Trivy

**Output**: `out/scan/<module>/<scanner>/`

---

## Vulnerability Scanning

Detects known CVEs in:

- **Language dependencies**: Go modules, npm packages, Python packages, etc.
- **OS packages**: Alpine apk, Ubuntu apt, RHEL yum, etc.
- **Container images**: Base image and application layer vulnerabilities

**Severity levels**: Critical, High, Medium, Low, Unknown

---

## SBOM Generation

Generates Software Bill of Materials in:

- **CycloneDX** format (JSON/XML)
- **SPDX** format (JSON/XML)

**Contents**:

- All direct and transitive dependencies
- Version information
- License information
- Package URLs (purl)

---

## Compliance Checking

Validates licenses against policy:

- **Allowed licenses**: MIT, Apache-2.0, BSD-3-Clause, etc.
- **Prohibited licenses**: GPL-3.0, AGPL-3.0, etc.
- **Unknown licenses**: Flagged for review

---

## Related Documentation

- **[Security Index](./index.md)** - Security scanning overview
- **[Scan Commands](../commands/scan/index.md)** - Full scan command reference
