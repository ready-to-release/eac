# Security Commands

Command reference for EAC's security scanning system.

## Quick Reference

| Command               | Description                                          |
| --------------------- | ---------------------------------------------------- |
| `security`            | Run all security scans                               |
| `security-sast`       | Static Application Security Testing using Semgrep    |
| `security-vuln`       | Scan for dependency vulnerabilities using Trivy      |
| `security-iac`        | Scan Infrastructure as Code for misconfigurations    |
| `security-sbom`       | Generate Software Bill of Materials                  |
| `security-secrets`    | Detect secrets and credentials                       |
| `security-compliance` | Check compliance with security standards             |
| `security-zap`        | Dynamic Application Security Testing using OWASP ZAP |

---

## security

Run all security scans.

### Synopsis

```bash
r2r eac security [options]
```

### Description

Executes a comprehensive security scan including SAST, vulnerability scanning, secrets detection, and IaC scanning. Aggregates results from all scanners.

### Flags

| Flag         | Short | Type   | Default         | Description                        |
| ------------ | ----- | ------ | --------------- | ---------------------------------- |
| `--severity` | `-s`  | string | `HIGH,CRITICAL` | Minimum severity to report         |
| `--output`   | `-o`  | string | `out/security/` | Output directory                   |
| `--format`   | `-f`  | string | `table`         | Output format (table, json, sarif) |
| `--fail-on`  |       | string | `CRITICAL`      | Severity that causes exit code 1   |

### Examples

```bash
# Run all security scans
r2r eac security

# Include medium severity
r2r eac security --severity MEDIUM,HIGH,CRITICAL

# JSON output
r2r eac security --format json

# Custom output directory
r2r eac security --output reports/security/

# Fail only on critical issues
r2r eac security --fail-on CRITICAL
```

### Output

```text
Running comprehensive security scan...

SAST Scan (Semgrep):
  ✓ Completed in 45s
  ⚠ 2 findings

Vulnerability Scan (Trivy):
  ✓ Completed in 12s
  ⚠ 5 vulnerabilities

Secrets Scan:
  ✓ Completed in 8s
  ✓ No secrets found

IaC Scan:
  ✓ Completed in 15s
  ⚠ 1 misconfiguration

Summary:
─────────────────────────────────────────────────────
│ Scanner     │ Critical │ High │ Medium │ Low │
├─────────────┼──────────┼──────┼────────┼─────┤
│ SAST        │ 0        │ 1    │ 1      │ 0   │
│ Vuln        │ 1        │ 2    │ 2      │ 0   │
│ Secrets     │ 0        │ 0    │ 0      │ 0   │
│ IaC         │ 0        │ 1    │ 0      │ 0   │
├─────────────┼──────────┼──────┼────────┼─────┤
│ Total       │ 1        │ 4    │ 3      │ 0   │

✗ 1 critical issue found
  See out/security/report.html for details
```

### Exit Codes

| Code | Description                               |
| ---- | ----------------------------------------- |
| 0    | No issues at or above fail-on severity    |
| 1    | Issues found at or above fail-on severity |
| 2    | Scanner error                             |

---

## security-sast

Static Application Security Testing using Semgrep.

### Synopsis

```bash
r2r eac security-sast [options]
```

### Description

Analyzes source code for security vulnerabilities, bad patterns, and potential bugs using Semgrep rules.

### Flags

| Flag         | Short | Type     | Default              | Description                              |
| ------------ | ----- | -------- | -------------------- | ---------------------------------------- |
| `--rules`    | `-r`  | string   | `auto`               | Semgrep ruleset (auto, security, golang) |
| `--severity` | `-s`  | string   | `WARNING,ERROR`      | Severity levels to report                |
| `--path`     | `-p`  | string   | `.`                  | Path to scan                             |
| `--exclude`  | `-e`  | string[] | -                    | Patterns to exclude                      |
| `--output`   | `-o`  | string   | `out/security/sast/` | Output directory                         |
| `--format`   | `-f`  | string   | `table`              | Output format (table, json, sarif)       |
| `--quick`    | `-q`  | bool     | `false`              | Quick scan (fewer rules)                 |

### Examples

```bash
# Full SAST scan
r2r eac security-sast

# Quick scan for pre-commit
r2r eac security-sast --quick

# Scan specific path
r2r eac security-sast --path go/eac/commands/

# Use specific ruleset
r2r eac security-sast --rules golang

# Exclude test files
r2r eac security-sast --exclude "*_test.go"

# SARIF output for GitHub
r2r eac security-sast --format sarif
```

### Output

```text
Scanning with Semgrep...

Rules loaded: 245 (security, golang, best-practices)
Files scanned: 156

Findings:
─────────────────────────────────────────────────────

❌ HIGH: SQL Injection vulnerability
   File: go/eac/api/handlers/users.go:45
   Rule: go.lang.security.sql-injection

   Code:
   44│   query := "SELECT * FROM users WHERE id = " + id
   45│   rows, err := db.Query(query)

   Fix: Use parameterized queries

⚠️ MEDIUM: Weak cryptographic algorithm
   File: go/eac/auth/crypto.go:23
   Rule: go.lang.security.weak-crypto

   Code:
   22│   h := md5.New()
   23│   h.Write([]byte(password))

   Fix: Use SHA-256 or stronger

Summary:
  Files: 156
  Findings: 2 (1 high, 1 medium)
  Time: 45s

✗ Security issues found
```

### Detects

- Injection vulnerabilities (SQL, command, LDAP)
- Insecure cryptography
- Hardcoded secrets
- Authentication issues
- Path traversal
- XML external entities (XXE)
- Cross-site scripting (XSS)

### Exit Codes

| Code | Description                |
| ---- | -------------------------- |
| 0    | No findings at ERROR level |
| 1    | Findings at ERROR level    |
| 2    | Scanner error              |

---

## security-vuln

Scan for dependency vulnerabilities using Trivy.

### Synopsis

```bash
r2r eac security-vuln [options]
```

### Description

Scans project dependencies for known CVEs using Trivy's vulnerability database.

### Flags

| Flag               | Short | Type   | Default              | Description                          |
| ------------------ | ----- | ------ | -------------------- | ------------------------------------ |
| `--severity`       | `-s`  | string | `HIGH,CRITICAL`      | Severity levels                      |
| `--ignore-unfixed` |       | bool   | `false`              | Ignore vulnerabilities without fixes |
| `--output`         | `-o`  | string | `out/security/vuln/` | Output directory                     |
| `--format`         | `-f`  | string | `table`              | Output format (table, json, sarif)   |

### Examples

```bash
# Scan for vulnerabilities
r2r eac security-vuln

# Include medium severity
r2r eac security-vuln --severity MEDIUM,HIGH,CRITICAL

# Ignore unfixed vulnerabilities
r2r eac security-vuln --ignore-unfixed

# JSON output
r2r eac security-vuln --format json
```

### Output

```text
Scanning dependencies for vulnerabilities...

go.mod (go)
─────────────────────────────────────────────────────

github.com/example/lib v1.2.3
│
├── CVE-2024-1234 (CRITICAL)
│   Remote code execution via crafted input
│   Fixed in: v1.2.4
│   https://nvd.nist.gov/vuln/detail/CVE-2024-1234
│
└── CVE-2024-2345 (HIGH)
    Information disclosure vulnerability
    Fixed in: v1.2.5
    https://nvd.nist.gov/vuln/detail/CVE-2024-2345

golang.org/x/net v0.17.0
│
└── CVE-2024-5678 (MEDIUM)
    Denial of service vulnerability
    Fixed in: v0.18.0
    https://nvd.nist.gov/vuln/detail/CVE-2024-5678

Summary:
  Total: 3 vulnerabilities
  Critical: 1
  High: 1
  Medium: 1

✗ Critical vulnerabilities found
  Run 'go get -u' to update dependencies
```

### Exit Codes

| Code | Description                              |
| ---- | ---------------------------------------- |
| 0    | No vulnerabilities at specified severity |
| 1    | Vulnerabilities found                    |
| 2    | Scanner error                            |

---

## security-iac

Scan Infrastructure as Code for misconfigurations.

### Synopsis

```bash
r2r eac security-iac [options]
```

### Description

Scans Dockerfiles, Kubernetes manifests, Terraform files, and GitHub Actions for security misconfigurations.

### Flags

| Flag         | Short | Type     | Default             | Description                                    |
| ------------ | ----- | -------- | ------------------- | ---------------------------------------------- |
| `--severity` | `-s`  | string   | `HIGH,CRITICAL`     | Severity levels                                |
| `--path`     | `-p`  | string   | `.`                 | Path to scan                                   |
| `--type`     | `-t`  | string[] | `all`               | IaC types (dockerfile, k8s, terraform, github) |
| `--output`   | `-o`  | string   | `out/security/iac/` | Output directory                               |
| `--format`   | `-f`  | string   | `table`             | Output format                                  |

### Examples

```bash
# Scan all IaC files
r2r eac security-iac

# Scan specific types
r2r eac security-iac --type dockerfile,k8s

# Scan specific path
r2r eac security-iac --path deploy/

# Include medium severity
r2r eac security-iac --severity MEDIUM,HIGH,CRITICAL
```

### Output

```text
Scanning Infrastructure as Code...

Dockerfiles:
─────────────────────────────────────────────────────

containers/api/Dockerfile
│
├── DS002 (HIGH) - Container running as root
│   Line 1: FROM golang:1.21
│   Fix: Add 'USER nonroot' directive
│
└── DS017 (MEDIUM) - No HEALTHCHECK defined
    Fix: Add HEALTHCHECK instruction

GitHub Actions:
─────────────────────────────────────────────────────

.github/workflows/ci.yml
│
└── GHA001 (HIGH) - Unpinned action version
    Line 12: uses: actions/checkout@v4
    Fix: Pin to specific SHA

Summary:
  Files scanned: 8
  Misconfigurations: 3 (2 high, 1 medium)

✗ IaC security issues found
```

### Scans

- **Dockerfiles** - Best practices, security issues
- **Kubernetes** - Pod security, RBAC, network policies
- **Terraform** - AWS/Azure/GCP misconfigurations
- **GitHub Actions** - Injection risks, permission issues

### Exit Codes

| Code | Description                                |
| ---- | ------------------------------------------ |
| 0    | No misconfigurations at specified severity |
| 1    | Misconfigurations found                    |
| 2    | Scanner error                              |

---

## security-sbom

Generate Software Bill of Materials.

### Synopsis

```bash
r2r eac security-sbom [options]
```

### Description

Creates a comprehensive inventory of all software components including dependencies, versions, and licenses.

### Flags

| Flag       | Short | Type   | Default     | Description                   |
| ---------- | ----- | ------ | ----------- | ----------------------------- |
| `--format` | `-f`  | string | `cyclonedx` | SBOM format (cyclonedx, spdx) |
| `--output` | `-o`  | string | `out/sbom/` | Output directory              |

### Examples

```bash
# Generate CycloneDX SBOM
r2r eac security-sbom

# Generate SPDX SBOM
r2r eac security-sbom --format spdx

# Custom output directory
r2r eac security-sbom --output artifacts/sbom/
```

### Output

```text
Generating Software Bill of Materials...

Scanning dependencies...
  ✓ go.mod: 45 packages
  ✓ package.json: 12 packages

SBOM Generated:
  Format: CycloneDX 1.4
  Components: 57

  Direct dependencies: 23
  Transitive dependencies: 34

License Summary:
  MIT: 32
  Apache-2.0: 18
  BSD-3-Clause: 7

✓ SBOM saved: out/sbom/sbom.json

Upload to dependency tracking:
  gh api -X POST /repos/owner/repo/dependency-graph/snapshots \
    --input out/sbom/sbom.json
```

### Exit Codes

| Code | Description                 |
| ---- | --------------------------- |
| 0    | SBOM generated successfully |
| 1    | Error generating SBOM       |

---

## security-secrets

Detect secrets and credentials using Trivy.

### Synopsis

```bash
r2r eac security-secrets [options]
```

### Description

Scans code and configuration files for exposed secrets, API keys, passwords, and other sensitive credentials.

### Flags

| Flag       | Short | Type   | Default                 | Description      |
| ---------- | ----- | ------ | ----------------------- | ---------------- |
| `--path`   | `-p`  | string | `.`                     | Path to scan     |
| `--output` | `-o`  | string | `out/security/secrets/` | Output directory |
| `--format` | `-f`  | string | `table`                 | Output format    |

### Examples

```bash
# Scan for secrets
r2r eac security-secrets

# Scan specific path
r2r eac security-secrets --path config/

# JSON output
r2r eac security-secrets --format json
```

### Output

```text
Scanning for secrets...

Secrets Found:
─────────────────────────────────────────────────────

❌ AWS Access Key
   File: config/prod.yaml:12
   Type: aws-access-key-id
   Value: AKIA...EXAMPLE (redacted)

   Recommendation: Rotate immediately and use environment variables

❌ Private Key
   File: certs/server.key:1
   Type: private-key

   Recommendation: Move to secure key management

❌ API Token
   File: scripts/deploy.sh:45
   Type: generic-api-key
   Value: sk-...xxx (redacted)

   Recommendation: Use secrets manager

Summary:
  Files scanned: 234
  Secrets found: 3

✗ Secrets detected - immediate action required
```

### Detects

- AWS credentials
- API keys and tokens
- Private keys
- Database passwords
- OAuth secrets
- Generic passwords
- SSH keys

### Exit Codes

| Code | Description      |
| ---- | ---------------- |
| 0    | No secrets found |
| 1    | Secrets detected |
| 2    | Scanner error    |

---

## security-compliance

Check compliance with security standards using Trivy.

### Synopsis

```bash
r2r eac security-compliance [options]
```

### Description

Validates configuration and code against security compliance frameworks.

### Flags

| Flag          | Short | Type   | Default                    | Description          |
| ------------- | ----- | ------ | -------------------------- | -------------------- |
| `--framework` | `-f`  | string | `cis`                      | Compliance framework |
| `--output`    | `-o`  | string | `out/security/compliance/` | Output directory     |
| `--format`    |       | string | `table`                    | Output format        |

### Supported Frameworks

| Framework | Description                |
| --------- | -------------------------- |
| `cis`     | CIS Benchmarks             |
| `nist`    | NIST 800-53                |
| `soc2`    | SOC 2 Trust Criteria       |
| `pci-dss` | PCI Data Security Standard |
| `hipaa`   | HIPAA Security Rule        |

### Examples

```bash
# Check CIS compliance
r2r eac security-compliance --framework cis

# Check SOC 2 compliance
r2r eac security-compliance --framework soc2

# Check multiple frameworks
r2r eac security-compliance --framework nist
```

### Output

```text
CIS Benchmark Compliance Check
═══════════════════════════════════════════════════════

Container Image Security:
─────────────────────────────────────────────────────
✓ 4.1 Ensure container images are signed
✗ 4.2 Ensure containers run as non-root
  Finding: Dockerfile does not specify USER
  Fix: Add 'USER nonroot' directive
⚠ 4.3 Ensure container healthchecks are enabled
  Finding: No HEALTHCHECK instruction
  Fix: Add HEALTHCHECK instruction

CI/CD Security:
─────────────────────────────────────────────────────
✓ 5.1 Ensure branch protection is enabled
✓ 5.2 Ensure code review is required
✗ 5.3 Ensure signed commits are required
  Finding: Branch does not require signed commits
  Fix: Enable commit signing requirement

Summary:
  Controls: 15
  Passing: 11 (73%)
  Failing: 3 (20%)
  Warning: 1 (7%)

✗ Compliance check failed
```

### Exit Codes

| Code | Description                  |
| ---- | ---------------------------- |
| 0    | All controls passing         |
| 1    | One or more controls failing |
| 2    | Check error                  |

---

## security-zap

Dynamic Application Security Testing using OWASP ZAP.

### Synopsis

```bash
r2r eac security-zap [options]
```

### Description

Runs OWASP ZAP against a live application to find runtime vulnerabilities. Requires Docker.

### Flags

| Flag        | Short | Type   | Default             | Description                     |
| ----------- | ----- | ------ | ------------------- | ------------------------------- |
| `--target`  | `-t`  | string | -                   | Target URL (required)           |
| `--mode`    | `-m`  | string | `baseline`          | Scan mode (baseline, full, api) |
| `--output`  | `-o`  | string | `out/security/zap/` | Output directory                |
| `--timeout` |       | int    | `3600`              | Scan timeout in seconds         |

### Examples

```bash
# Baseline scan
r2r eac security-zap --target http://localhost:8080

# API scan
r2r eac security-zap --target http://localhost:8080/api --mode api

# Full scan
r2r eac security-zap --target http://localhost:8080 --mode full

# With custom timeout
r2r eac security-zap --target http://localhost:8080 --timeout 7200
```

### Output

```text
OWASP ZAP Dynamic Security Scan
═══════════════════════════════════════════════════════

Target: http://localhost:8080
Mode: baseline
Status: Running...

Spider Progress: 100% (45 URLs found)
Active Scan Progress: 100%

Findings:
─────────────────────────────────────────────────────

❌ HIGH: Cross-Site Scripting (XSS)
   URL: /api/search?q=<script>alert(1)</script>
   Method: GET
   Evidence: <script>alert(1)</script> reflected in response

   Solution: Encode output and validate input

⚠️ MEDIUM: Missing Security Headers
   URL: /api/*
   Missing: Content-Security-Policy, X-Frame-Options

   Solution: Add security headers to responses

⚠️ MEDIUM: Cookie Without Secure Flag
   URL: /api/auth/login
   Cookie: session_id

   Solution: Set Secure flag on cookies

Summary:
  URLs scanned: 45
  Alerts: 3 (1 high, 2 medium)
  Duration: 5m 32s

✗ Security vulnerabilities found
  Report: out/security/zap/report.html
```

### Scan Modes

| Mode       | Description          | Duration |
| ---------- | -------------------- | -------- |
| `baseline` | Quick passive scan   | ~5 min   |
| `full`     | Complete active scan | ~30 min  |
| `api`      | API-focused scan     | ~15 min  |

### Exit Codes

| Code | Description                            |
| ---- | -------------------------------------- |
| 0    | No high/critical vulnerabilities       |
| 1    | High or critical vulnerabilities found |
| 2    | Scanner error                          |
| 3    | Target unreachable                     |

---

## Common Workflows

### Pre-Commit Security Check

```bash
# Quick checks before commit
r2r eac security-secrets
r2r eac security-sast --quick
```

### CI/CD Security Gate

```bash
# Full security scan in CI
r2r eac security --fail-on HIGH

# Or run individual scans
r2r eac security-sast
r2r eac security-vuln
r2r eac security-secrets
r2r eac security-iac
```

### Compliance Audit

```bash
# Generate SBOM
r2r eac security-sbom --format spdx

# Run compliance checks
r2r eac security-compliance --framework soc2

# Generate evidence for risk assessment
r2r eac create-risk-assess

# View risk report
r2r eac show-risk-report
```

### API Security Testing

```bash
# Start application
r2r eac serve-api &

# Run ZAP API scan
r2r eac security-zap --target http://localhost:8080 --mode api

# Stop application
kill %1
```

---

## Integration Patterns

### GitHub Actions

```yaml
name: Security Gate

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: SAST Scan
        run: r2r eac security-sast

      - name: Vulnerability Scan
        run: r2r eac security-vuln

      - name: Secrets Check
        run: r2r eac security-secrets

      - name: Generate SBOM
        run: r2r eac security-sbom

      - name: Upload SBOM
        uses: actions/upload-artifact@v4
        with:
          name: sbom
          path: out/sbom/
```

### Pre-Commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running security checks..."

r2r eac security-secrets || exit 1
r2r eac security-sast --quick || exit 1

echo "✓ Security checks passed"
```

### Risk Management Integration

```bash
# Run security scans
r2r eac security

# Update risk assessment with findings
r2r eac create-risk-assess

# View risk report
r2r eac show-risk-report
```

---

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

---

## Troubleshooting

| Problem              | Solution                              |
| -------------------- | ------------------------------------- |
| Scan times out       | Increase timeout, scan smaller scope  |
| False positives      | Add to ignore list with justification |
| Missing tool         | Install Trivy, Semgrep, or ZAP        |
| Permission denied    | Check file permissions                |
| Docker not available | Install Docker for ZAP scans          |
| Rate limited         | Reduce scan frequency                 |

---

## Related Documentation

- [Security Overview](security-overview.md) - Concepts and scan types
- [Security Configuration](security-configuration.md) - Configuration reference
- [Risk Commands](risks-commands.md) - Risk assessment integration
