<!-- EDITOR
# Editor: how-to-guides/commands/areas/security-overview.md

## Soul

Comprehensive defense-in-depth security scanning with SAST, vulnerability detection, secrets scanning, IaC analysis, compliance checking, SBOM generation, and DAST using Semgrep, Trivy, and OWASP ZAP.

## Sections

1. What is Security Scanning?
2. When to Use Security Commands
3. Key Concepts
4. Workflow Overview
5. Scan Details
6. Integration Points
7. Best Practices
8. Troubleshooting
9. Next Steps
10. Related Areas
-->

# Security Scanning

Security scanning in EAC provides comprehensive application security testing including SAST, vulnerability scanning, secrets detection, and compliance checking.

## What is Security Scanning?

EAC's security system enables you to:

- **Scan code for vulnerabilities** using static analysis (SAST)
- **Detect dependency vulnerabilities** with Trivy
- **Find exposed secrets** in code and configuration
- **Check compliance** against security frameworks
- **Generate SBOMs** for supply chain security
- **Run dynamic testing** with OWASP ZAP

The system integrates multiple security tools to provide defense-in-depth scanning.

## When to Use Security Commands

Use security commands when you need:

| Scenario                        | Commands              |
| ------------------------------- | --------------------- |
| Static code analysis            | `scan sast`       |
| Dependency vulnerabilities      | `scan vuln`       |
| Infrastructure as Code scanning | `scan iac`        |
| Secrets detection               | `scan secrets`    |
| Compliance checking             | `scan compliance` |
| Generate SBOM                   | `scan sbom`       |
| Dynamic testing                 | `scan zap`        |
| Run all scans                   | `security`            |

### Common Use Cases

- **Pre-commit checks** - Catch issues before code review
- **CI/CD gates** - Block merges with security issues
- **Compliance audits** - Generate evidence for auditors
- **Supply chain security** - Track dependencies and vulnerabilities
- **Penetration testing** - Automated security assessment

## Key Concepts

### Security Scan Types

| Type              | Tool      | Finds                              |
| ----------------- | --------- | ---------------------------------- |
| **SAST**          | Semgrep   | Code vulnerabilities, bad patterns |
| **Vulnerability** | Trivy     | Dependency CVEs                    |
| **IaC**           | Trivy     | Infrastructure misconfigurations   |
| **Secrets**       | Trivy     | Exposed credentials                |
| **Compliance**    | Trivy     | Framework violations               |
| **DAST**          | OWASP ZAP | Runtime vulnerabilities            |
| **SBOM**          | Trivy     | Dependency inventory               |

### Severity Levels

| Level        | Description              | Action              |
| ------------ | ------------------------ | ------------------- |
| **Critical** | Exploitable, high impact | Fix immediately     |
| **High**     | Serious vulnerability    | Fix before release  |
| **Medium**   | Moderate risk            | Fix in next sprint  |
| **Low**      | Minor issue              | Fix when convenient |
| **Info**     | Informational            | Review and document |

### Compliance Frameworks

Supported frameworks for compliance checking:

- **CIS Benchmarks** - Center for Internet Security
- **NIST 800-53** - Federal security controls
- **SOC 2** - Service organization controls
- **PCI DSS** - Payment card security
- **HIPAA** - Healthcare data protection

### SBOM (Software Bill of Materials)

SBOM provides:

- Complete dependency inventory
- License information
- Vulnerability mapping
- Supply chain transparency

Formats: SPDX, CycloneDX

## Workflow Overview

### Development Workflow

```bash
# 1. Run quick security check before commit
r2r eac scan secrets
r2r eac scan sast

# 2. Fix any issues found
# ... make fixes ...

# 3. Verify fixes
r2r eac scan secrets
r2r eac scan sast
```

### CI/CD Workflow

```bash
# Full security scan in CI
r2r eac security

# Or run individual scans
r2r eac scan sast
r2r eac scan vuln
r2r eac scan secrets
r2r eac scan compliance
```

### Compliance Workflow

```bash
# 1. Generate SBOM for audit
r2r eac scan sbom --format spdx

# 2. Run compliance checks
r2r eac scan compliance --framework soc2

# 3. Generate evidence for risk assessment
r2r eac create risk-assess
```

## Scan Details

### SAST with Semgrep

Static Application Security Testing analyzes source code:

```bash
r2r eac scan sast

# Output:
# Scanning with Semgrep...
#
# ❌ HIGH: SQL Injection vulnerability
#    File: src/api/users.go:45
#    Rule: go.lang.security.sql-injection
#
# ⚠️ MEDIUM: Weak cryptographic algorithm
#    File: src/auth/crypto.go:23
#    Rule: go.lang.security.weak-crypto
```

Detects:

- Injection vulnerabilities (SQL, command, etc.)
- Insecure cryptography
- Hardcoded secrets
- Authentication issues
- Configuration problems

### Vulnerability Scanning with Trivy

Scans dependencies for known CVEs:

```bash
r2r eac scan vuln

# Output:
# Scanning dependencies...
#
# go.mod (go)
# ├── github.com/example/lib v1.2.3
# │   └── CVE-2024-1234 (HIGH) - Remote code execution
# │       Fixed in: v1.2.4
# │
# └── golang.org/x/net v0.17.0
#     └── CVE-2024-5678 (MEDIUM) - DoS vulnerability
#         Fixed in: v0.18.0
```

### Infrastructure as Code Scanning

Checks IaC files for misconfigurations:

```bash
r2r eac scan iac

# Scans:
# - Dockerfiles
# - Kubernetes manifests
# - Terraform files
# - GitHub Actions workflows
```

### Secrets Detection

Finds exposed credentials:

```bash
r2r eac scan secrets

# Output:
# Scanning for secrets...
#
# ❌ AWS Access Key found
#    File: config/prod.yaml:12
#    Type: aws-access-key-id
#
# ❌ Private key found
#    File: certs/server.key:1
#    Type: private-key
```

### Compliance Checking

Validates against security frameworks:

```bash
r2r eac scan compliance --framework cis

# Output:
# CIS Benchmark Compliance
#
# ✅ 4.1 Ensure container images are signed
# ❌ 4.2 Ensure containers run as non-root
# ⚠️ 4.3 Ensure container healthchecks are enabled
#
# Compliance: 67% (2/3 controls passing)
```

### SBOM Generation

Creates software bill of materials:

```bash
r2r eac scan sbom --format cyclonedx

# Output: out/sbom/sbom.json
# Contains:
# - All dependencies
# - Versions
# - Licenses
# - Checksums
```

### Dynamic Testing with OWASP ZAP

Runs against live application:

```bash
# Start your application (replace with your actual start command)
./your-api-server &

# Run ZAP scan
r2r eac scan zap --target http://localhost:8080

# Output:
# OWASP ZAP Scan Results
#
# ❌ HIGH: Cross-Site Scripting (XSS)
#    URL: /api/search?q=<script>
#
# ⚠️ MEDIUM: Missing security headers
#    URL: /api/*
```

## Integration Points

### With Risk Management

Security scan results feed into risk assessment:

```bash
# Run security scans
r2r eac security

# Update risk assessment with findings
r2r eac create risk-assess

# View risk report
r2r eac show risk-report
```

### With CI/CD

```yaml
name: Security Gate

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: SAST Scan
        run: r2r eac scan sast

      - name: Vulnerability Scan
        run: r2r eac scan vuln

      - name: Secrets Check
        run: r2r eac scan secrets

      - name: Generate SBOM
        run: r2r eac scan sbom

      - name: Upload SBOM
        uses: actions/upload-artifact@v3
        with:
          name: sbom
          path: out/sbom/
```

### With Pre-commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running security checks..."

# Quick checks only
r2r eac scan secrets || exit 1
r2r eac scan sast --quick || exit 1

echo "✅ Security checks passed"
```

## Best Practices

### Do's

- **Scan early** - Run in development, not just CI
- **Fix critical/high first** - Prioritize by severity
- **Track over time** - Monitor vulnerability trends
- **Generate SBOMs** - For supply chain transparency
- **Automate** - Make security scans part of CI/CD

### Don'ts

- **Don't ignore findings** - Suppressions need justification
- **Don't skip scans** - Security debt compounds
- **Don't expose reports** - Security findings are sensitive
- **Don't rely on one tool** - Defense in depth

## Troubleshooting

| Problem           | Solution                                        |
| ----------------- | ----------------------------------------------- |
| Scan times out    | Increase timeout, scan smaller scope            |
| False positives   | Add to ignore list with justification           |
| Missing tool      | Install Trivy, Semgrep, or ZAP                  |
| Permission denied | Check file permissions, run as appropriate user |

## Next Steps

- [Security Configuration](security-configuration.md) - Configure scanners and thresholds
- [Security Commands](security-commands.md) - Full command reference

## Related Areas

- [Risk Management](risks-overview.md) - Security findings feed into risk assessment
- [Pipeline](pipeline-overview.md) - CI/CD integration for security gates
- [Validate Commands](validate-overview.md) - Contract and configuration validation
