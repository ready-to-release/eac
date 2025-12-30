# Supply Chain Security

Protecting against vulnerabilities in dependencies and container images.

## Dependency Scanning

Identifies known vulnerabilities in third-party dependencies.

### What It Detects

| Risk                  | Description                               |
| --------------------- | ----------------------------------------- |
| Known CVEs            | Published vulnerabilities in dependencies |
| Outdated Dependencies | Old versions with known issues            |
| License Compliance    | Incompatible or problematic licenses      |
| Supply Chain Attacks  | Compromised or malicious packages         |

### Running Dependency Scans

```bash
# Vulnerability scan
eac scan --scanner vuln

# Software Bill of Materials (SBOM)
eac scan --scanner sbom

# Compliance checking
eac scan --scanner compliance

# All supply chain scanners
eac scan --scanner sbom,vuln,compliance
```

Evidence is written to `out/scan/<module>/<scanner>/`.

## Container Security

Multi-layer scanning of container images.

### What It Scans

| Layer               | Examples                   |
| ------------------- | -------------------------- |
| OS Layer            | Base image vulnerabilities |
| Application Layer   | Application dependencies   |
| Configuration Layer | Misconfigurations, secrets |

### What It Detects

- OS package vulnerabilities
- Application dependency vulnerabilities
- Running as root
- Hardcoded secrets in image layers
- Exposed ports

### Running Container Scans

Container scanning is included in the vulnerability scanner:

```bash
# Scan module containers
eac scan --scanner vuln
```

## Container Best Practices

- Use specific version tags (not `latest`)
- Run containers as non-root user
- Use multi-stage builds for minimal attack surface
- Scan before every deployment
- Remove unnecessary packages

## When to Use

| Stage             | Activity             | Scanner    |
| ----------------- | -------------------- | ---------- |
| Pre-commit (2)    | Scan changed deps    | vuln       |
| Merge Request (3) | Full dependency scan | vuln, sbom |
| Commit (4)        | Container image scan | vuln       |
| Deployment (10)   | Final image scan     | vuln       |

See [Shift-Left Security](shift-left.md) for the complete stage matrix.
