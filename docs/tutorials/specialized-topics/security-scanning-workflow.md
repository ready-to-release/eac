# Security Scanning Workflow
**Status:** Placeholder - Content coming soon

**Prerequisites:** [Your First Module](../getting-started/first-module.md), [CI/CD Integration](../advanced-practices/ci-cd-integration.md)

## Planned Content

This tutorial teaches you how to integrate security scanning into your development workflow using Trivy, Semgrep, and OWASP ZAP.

### What You'll Learn

- Scan for vulnerabilities: `r2r scan vuln`
- Perform SAST (Static Application Security Testing): `r2r scan sast`
- Detect secrets: `r2r scan secrets`
- Scan Infrastructure as Code: `r2r scan iac`
- Check compliance: `r2r scan compliance`
- Generate SBOM (Software Bill of Materials): `r2r scan sbom`
- Integrate security scanning in CI/CD
- Understand scan results and remediation

### Tutorial Structure

1. **Understanding security scanning**
   - Types of scans: SAST, DAST, SCA, secrets
   - When to run each scan
   - Security in the development lifecycle
   - Shift-left security

2. **Vulnerability scanning with Trivy**
   - Scan dependencies: `r2r scan vuln`
   - Understand CVE severity (Critical, High, Medium, Low)
   - Review scan results
   - Remediate vulnerabilities

3. **SAST with Semgrep**
   - Static code analysis: `r2r scan sast`
   - Detect security issues in code
   - Common vulnerability patterns
   - Fix SAST findings

4. **Secret detection**
   - Detect hardcoded credentials: `r2r scan secrets`
   - Prevent secret leaks
   - Use environment variables instead
   - `.gitignore` for sensitive files

5. **Infrastructure as Code scanning**
   - Scan IaC configs: `r2r scan iac`
   - Detect misconfigurations
   - Dockerfile best practices
   - Kubernetes security

6. **Compliance scanning**
   - Check compliance standards: `r2r scan compliance`
   - NIST, CIS benchmarks
   - Generate compliance reports
   - Remediate compliance issues

7. **Software Bill of Materials (SBOM)**
   - Generate SBOM: `r2r scan sbom`
   - SPDX and CycloneDX formats
   - Track dependencies for supply chain security
   - SBOM in compliance and audits

8. **CI/CD integration**
   - Run scans in GitHub Actions
   - Quality gates based on scan results
   - Fail builds on critical vulnerabilities
   - Security scan reports

### Example: Complete Security Workflow

The tutorial demonstrates a comprehensive security workflow:

```bash
# 1. Vulnerability scanning (dependencies)
r2r scan vuln eac-commands
# Scans Go modules, npm packages, container images

# 2. SAST (static analysis)
r2r scan sast eac-commands
# Analyzes Go code for security issues

# 3. Secret detection
r2r scan secrets
# Scans all files for hardcoded credentials

# 4. IaC scanning
r2r scan iac
# Scans Dockerfiles, k8s manifests, Terraform

# 5. Compliance scanning
r2r scan compliance eac-commands --compliance nist-800-53
# Checks against compliance framework

# 6. Generate SBOM
r2r scan sbom eac-commands
# Creates software bill of materials
```

### Key Concepts Covered

- Security scanning types and purposes
- Vulnerability management and CVEs
- SAST and code security
- Secret detection and prevention
- IaC security best practices
- Compliance frameworks
- SBOM generation and use

### Understanding Scan Results

**Vulnerability scan output:**
```text
Scanning eac-commands for vulnerabilities...

┌──────────────────────────────────────────────────────────┐
│ Package         │ CVE ID          │ Severity │ Fix       │
├──────────────────────────────────────────────────────────┤
│ golang.org/x/   │ CVE-2023-xxxxx  │ HIGH     │ 0.14.0    │
│ crypto          │                 │          │           │
└──────────────────────────────────────────────────────────┘

Total: 1 vulnerability found
  - Critical: 0
  - High: 1
  - Medium: 0
  - Low: 0
```

**SAST output:**
```text
Running SAST scan...

┌──────────────────────────────────────────────────────────┐
│ File            │ Line │ Issue                │ Severity │
├──────────────────────────────────────────────────────────┤
│ cmd/auth.go     │ 42   │ SQL Injection risk   │ HIGH     │
│ pkg/token.go    │ 18   │ Weak crypto          │ MEDIUM   │
└──────────────────────────────────────────────────────────┘
```

### Trivy Features

- **Vulnerability scanning**: Dependencies, OS packages, container images
- **Secret detection**: API keys, passwords, tokens
- **IaC scanning**: Dockerfile, Kubernetes, Terraform
- **License scanning**: Detect problematic licenses
- **SBOM generation**: SPDX, CycloneDX formats

### Semgrep Features

- **Pattern-based SAST**: Language-aware security rules
- **Custom rules**: Write project-specific checks
- **Fast performance**: Efficient local scanning
- **Low false positives**: High-quality rules

### CI/CD Integration

```yaml
# .github/workflows/security.yml
name: Security Scans

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Vulnerability Scan
        run: r2r scan vuln
        continue-on-error: false

      - name: SAST
        run: r2r scan sast

      - name: Secret Detection
        run: r2r scan secrets

      - name: IaC Scan
        run: r2r scan iac

      - name: Generate SBOM
        run: r2r scan sbom
```

### Security Quality Gates

Define thresholds for CI/CD:

- **Block on Critical vulnerabilities**: Fail build
- **Warn on High vulnerabilities**: Alert but don't block
- **Report Medium/Low**: Track for remediation
- **Block on secrets detected**: Always fail
- **Advisory on SAST findings**: Review in PR

### Remediation Workflow

1. **Identify**: Run scans to detect issues
2. **Prioritize**: Address Critical/High first
3. **Research**: Understand the vulnerability
4. **Fix**: Update dependencies or code
5. **Verify**: Re-run scans
6. **Document**: Update changelog

### Best Practices

- Run scans locally before committing
- Integrate scans in CI/CD pipeline
- Set quality gates based on severity
- Remediate Critical/High vulnerabilities immediately
- Keep dependencies up to date
- Use environment variables for secrets
- Generate SBOM for every release
- Track vulnerabilities over time

### Compliance Frameworks

Supported compliance standards:

- **NIST 800-53**: Federal security controls
- **CIS Benchmarks**: Configuration standards
- **PCI-DSS**: Payment card security
- **HIPAA**: Healthcare data protection
- **Custom frameworks**: Define your own

### SBOM Use Cases

- **Supply chain security**: Track all dependencies
- **Vulnerability management**: Know what you're running
- **Compliance**: Required by some regulations
- **License compliance**: Identify license issues
- **Procurement**: Share with customers

### Next Steps

After completing this tutorial, you'll have integrated security scanning into your workflow. Explore other specialized topics: [Effective BDD Scenarios](./effective-bdd-scenarios.md), [Architecture Documentation](./architecture-documentation.md), or [TypeScript Setup](./typescript-module-setup.md).
