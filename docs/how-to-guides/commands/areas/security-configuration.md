<!-- EDITOR
# Editor: how-to-guides/commands/areas/security-configuration.md

## Soul

Configuration reference for security scanners including Semgrep rules, Trivy settings, secrets detection, IaC policies, compliance frameworks, OWASP ZAP, severity thresholds, SARIF output, and CI integration.

## Sections

1. Configuration Files
2. Security Settings
3. SAST Configuration
4. Vulnerability Scanning
5. Secrets Detection
6. IaC Scanning
7. Compliance Scanning
8. SBOM Generation
9. DAST Configuration
10. Severity Thresholds
11. Output Configuration
12. CI Integration
13. Environment Variables
14. Example Configurations
15. Troubleshooting
16. Related Documentation
-->

# Security Configuration

This guide covers configuration options for EAC's security scanning system, including scanner settings, thresholds, and compliance frameworks.

## Configuration Files

| File                           | Purpose            |
| ------------------------------ | ------------------ |
| `.r2r/eac/security/config.yml` | Security settings  |
| `.semgrep.yml`                 | Semgrep rules      |
| `.trivyignore`                 | Trivy ignore rules |
| `zap-config.yml`               | OWASP ZAP settings |

## Security Settings

### Basic Configuration

`.r2r/eac/security/config.yml`:

```yaml
# Scanner settings
scanners:
  sast:
    enabled: true
    tool: semgrep

  vulnerability:
    enabled: true
    tool: trivy

  secrets:
    enabled: true
    tool: trivy

  iac:
    enabled: true
    tool: trivy

  compliance:
    enabled: true
    tool: trivy

  dast:
    enabled: false  # Requires running application
    tool: zap

  sbom:
    enabled: true
    tool: trivy

# Severity thresholds
thresholds:
  # Fail on these severities
  fail_on:
    - CRITICAL
    - HIGH

  # Warn on these severities
  warn_on:
    - MEDIUM

  # Ignore these severities
  ignore:
    - LOW
    - UNKNOWN

# Output settings
output:
  format: json  # json, table, sarif
  path: out/security/
  save_reports: true
```

## SAST Configuration

### Semgrep Settings

```yaml
sast:
  enabled: true
  tool: semgrep

  semgrep:
    # Rule sets
    rules:
      - p/golang
      - p/security-audit
      - p/owasp-top-ten

    # Custom rules directory
    custom_rules: .r2r/security/semgrep/

    # Exclude patterns
    exclude:
      - "*_test.go"
      - "vendor/"
      - "testdata/"

    # Severity mapping
    severity_map:
      ERROR: CRITICAL
      WARNING: HIGH
      INFO: MEDIUM

    # Timeout per file
    timeout: 30s

    # Max memory
    max_memory: 2048
```

### Custom Semgrep Rules

`.r2r/security/semgrep/custom-rules.yml`:

```yaml
rules:
  - id: custom-sql-injection
    pattern: |
      db.Query($QUERY, ...)
    message: "Potential SQL injection"
    severity: ERROR
    languages:
      - go

  - id: hardcoded-secret
    pattern-regex: "(api_key|password|secret)\\s*=\\s*['\"][^'\"]+['\"]"
    message: "Hardcoded secret detected"
    severity: ERROR
    languages:
      - go
```

### Semgrep Ignore

`.semgrepignore`:

```text
# Test files
*_test.go
test/
testdata/

# Generated code
*.gen.go
*_generated.go

# Vendor
vendor/

# Documentation
docs/
*.md
```

## Vulnerability Scanning

### Trivy Settings

```yaml
vulnerability:
  enabled: true
  tool: trivy

  trivy:
    # Scan types
    scan_types:
      - vuln    # Vulnerabilities
      - config  # Misconfigurations
      - secret  # Secrets
      - license # License compliance

    # Vulnerability database
    db:
      # Skip update (use cached)
      skip_update: false
      # Update interval
      update_interval: 24h

    # Severity filter
    severity:
      - CRITICAL
      - HIGH
      - MEDIUM

    # Ignore unfixed vulnerabilities
    ignore_unfixed: false

    # Timeout
    timeout: 15m

    # Cache directory
    cache_dir: ~/.cache/trivy
```

### Trivy Ignore

`.trivyignore`:

```text
# Ignore specific CVEs
CVE-2023-12345
CVE-2023-67890

# Ignore by package
pkg:golang/github.com/example/lib@1.0.0

# Ignore with expiry
# exp:2024-12-31 CVE-2024-11111
```

### Ignore with Reasons

`.trivyignore.yaml`:

```yaml
vulnerabilities:
  - id: CVE-2023-12345
    reason: "False positive - not exploitable in our context"
    expires: 2024-12-31

  - id: CVE-2023-67890
    reason: "Accepted risk - mitigated by network controls"
    approved_by: "security-team"
```

## Secrets Detection

### Secrets Settings

```yaml
secrets:
  enabled: true
  tool: trivy

  trivy:
    # Built-in detectors
    detectors:
      - aws
      - gcp
      - azure
      - github
      - private-key
      - generic

    # Custom patterns
    custom_patterns:
      - name: internal-api-key
        pattern: "INTERNAL_[A-Z0-9]{32}"

    # Exclude paths
    exclude:
      - "*.md"
      - "docs/"
      - "test/fixtures/"
```

### Allow List

```yaml
secrets:
  allow_list:
    # Known false positives
    - path: "test/fixtures/fake-key.txt"
      reason: "Test fixture with fake key"

    - pattern: "EXAMPLE_.*"
      reason: "Example placeholders"
```

## IaC Scanning

### Infrastructure as Code

```yaml
iac:
  enabled: true
  tool: trivy

  trivy:
    # File types
    file_types:
      - dockerfile
      - kubernetes
      - terraform
      - cloudformation
      - helm

    # Policies
    policies:
      - path: .r2r/security/policies/
      - builtin: true

    # Severity filter
    severity:
      - CRITICAL
      - HIGH
      - MEDIUM
```

### Custom Policies

`.r2r/security/policies/dockerfile.rego`:

```rego
package dockerfile

deny[msg] {
    input.Stages[_].Commands[_].Cmd == "run"
    contains(input.Stages[_].Commands[_].Value[_], "curl")
    msg = "Avoid using curl in Dockerfile"
}

deny[msg] {
    input.Stages[_].Commands[_].Cmd == "user"
    input.Stages[_].Commands[_].Value[0] == "root"
    msg = "Container should not run as root"
}
```

## Compliance Scanning

### Compliance Settings

```yaml
compliance:
  enabled: true
  tool: trivy

  trivy:
    # Compliance frameworks
    frameworks:
      - cis-1.23  # CIS Kubernetes Benchmark
      - nsa       # NSA Kubernetes Hardening
      - pss       # Pod Security Standards

    # Report format
    report_format: all  # all, summary, failed

    # Custom compliance
    custom:
      path: .r2r/security/compliance/
```

### Framework Options

| Framework        | Description                       |
| ---------------- | --------------------------------- |
| `cis-1.23`       | CIS Kubernetes Benchmark v1.23    |
| `cis-docker`     | CIS Docker Benchmark              |
| `nsa`            | NSA Kubernetes Hardening          |
| `pss-baseline`   | Pod Security Standards Baseline   |
| `pss-restricted` | Pod Security Standards Restricted |
| `aws-cis-1.2`    | AWS CIS Benchmark v1.2            |
| `aws-cis-1.4`    | AWS CIS Benchmark v1.4            |

## SBOM Generation

### SBOM Settings

```yaml
sbom:
  enabled: true
  tool: trivy

  trivy:
    # Output formats
    formats:
      - cyclonedx  # CycloneDX format
      - spdx       # SPDX format

    # Include
    include:
      - dependencies
      - licenses
      - vulnerabilities

    # Output path
    output: out/sbom/
```

### SBOM Output

```bash
# Generate CycloneDX SBOM
r2r eac scan sbom --format cyclonedx

# Output: out/sbom/sbom.cdx.json
```

## DAST Configuration

### OWASP ZAP Settings

```yaml
dast:
  enabled: false
  tool: zap

  zap:
    # Target URL
    target: http://localhost:8080

    # Scan type
    type: baseline  # baseline, full, api

    # API specification
    api_spec: openapi.yaml

    # Authentication
    auth:
      type: none  # none, basic, bearer, form
      # credentials in environment variables

    # Scan settings
    settings:
      ajax_spider: true
      active_scan: false  # Only in full scan

    # Rules
    rules:
      # Disable specific rules
      disable:
        - 10096  # Timestamp Disclosure
        - 10027  # Information Disclosure - Suspicious Comments

    # Thresholds
    thresholds:
      fail_on_risk:
        - High
        - Medium
```

### ZAP Configuration File

`zap-config.yml`:

```yaml
env:
  contexts:
    - name: "Default Context"
      urls:
        - "http://localhost:8080"
      includePaths:
        - "http://localhost:8080/api/.*"
      excludePaths:
        - "http://localhost:8080/health"

jobs:
  - type: spider
    parameters:
      maxDuration: 5

  - type: passiveScan-wait
    parameters:
      maxDuration: 10

  - type: report
    parameters:
      template: traditional-html
      reportDir: out/security/dast/
```

## Severity Thresholds

### Threshold Configuration

```yaml
thresholds:
  # Global thresholds
  global:
    fail_on:
      - CRITICAL
    warn_on:
      - HIGH
      - MEDIUM

  # Per-scanner thresholds
  sast:
    fail_on:
      - CRITICAL
      - HIGH

  vulnerability:
    fail_on:
      - CRITICAL
    ignore_unfixed: true

  secrets:
    fail_on:
      - CRITICAL
      - HIGH
      - MEDIUM
      - LOW  # Always fail on secrets

  compliance:
    fail_on:
      - CRITICAL
      - HIGH
    min_score: 80  # Minimum compliance score
```

## Output Configuration

### Report Settings

```yaml
output:
  # Output directory
  path: out/security/

  # Report formats
  formats:
    - json    # Machine-readable
    - sarif   # GitHub integration
    - table   # Human-readable

  # Save individual reports
  save_reports: true

  # Combined report
  combined_report: true
  combined_format: json

  # Report naming
  naming:
    pattern: "{scanner}-{date}"
    date_format: "2006-01-02"
```

### SARIF Output

For GitHub Security integration:

```yaml
output:
  formats:
    - sarif

  sarif:
    # Upload to GitHub
    upload: true
    # Include code snippets
    include_code: true
```

## CI Integration

### GitHub Actions

```yaml
# .github/workflows/security.yml
security:
  fail_on:
    - CRITICAL
    - HIGH

  upload_sarif: true
  create_issues: true

  schedule:
    # Full scan weekly
    full_scan: "0 0 * * 0"
    # Quick scan on PR
    pr_scan: "sast,secrets"
```

### Pre-commit

```yaml
pre_commit:
  enabled: true
  scanners:
    - secrets   # Always check secrets
    - sast      # Quick SAST check

  timeout: 60s
  fail_on:
    - CRITICAL
    - HIGH
```

## Environment Variables

| Variable           | Description      | Default                        |
| ------------------ | ---------------- | ------------------------------ |
| `SECURITY_CONFIG`  | Config file path | `.r2r/eac/security/config.yml` |
| `TRIVY_CACHE_DIR`  | Trivy cache      | `~/.cache/trivy`               |
| `SEMGREP_RULES`    | Semgrep rules    | `p/golang`                     |
| `ZAP_TARGET`       | DAST target URL  | -                              |
| `SECURITY_FAIL_ON` | Fail severities  | `CRITICAL,HIGH`                |

## Example Configurations

### Minimal Configuration

```yaml
scanners:
  sast:
    enabled: true
  vulnerability:
    enabled: true
  secrets:
    enabled: true

thresholds:
  fail_on:
    - CRITICAL
    - HIGH
```

### CI/CD Configuration

```yaml
scanners:
  sast:
    enabled: true
    tool: semgrep
    semgrep:
      rules:
        - p/golang
        - p/security-audit

  vulnerability:
    enabled: true
    tool: trivy
    trivy:
      severity:
        - CRITICAL
        - HIGH
      ignore_unfixed: true

  secrets:
    enabled: true

  sbom:
    enabled: true
    trivy:
      formats:
        - cyclonedx

thresholds:
  fail_on:
    - CRITICAL
    - HIGH

output:
  formats:
    - json
    - sarif
  save_reports: true
```

### Enterprise Configuration

```yaml
scanners:
  sast:
    enabled: true
    tool: semgrep
    semgrep:
      rules:
        - p/golang
        - p/security-audit
        - p/owasp-top-ten
      custom_rules: .r2r/security/semgrep/

  vulnerability:
    enabled: true
    tool: trivy
    trivy:
      severity:
        - CRITICAL
        - HIGH
        - MEDIUM

  secrets:
    enabled: true

  iac:
    enabled: true

  compliance:
    enabled: true
    trivy:
      frameworks:
        - cis-docker
        - pss-restricted

  dast:
    enabled: true
    tool: zap
    zap:
      type: full

  sbom:
    enabled: true
    trivy:
      formats:
        - cyclonedx
        - spdx

thresholds:
  fail_on:
    - CRITICAL
    - HIGH
  compliance:
    min_score: 85

output:
  formats:
    - json
    - sarif
    - table
  save_reports: true
  combined_report: true
```

## Troubleshooting

| Issue                 | Cause                | Solution                        |
| --------------------- | -------------------- | ------------------------------- |
| Scan timeout          | Large codebase       | Increase timeout, exclude paths |
| False positives       | Overly broad rules   | Add to ignore list              |
| Missing CVE data      | Outdated database    | Update Trivy DB                 |
| ZAP connection failed | App not running      | Start target application        |
| SBOM incomplete       | Missing dependencies | Run `go mod tidy`               |

## Related Documentation

- [Security Overview](security-overview.md) - Concepts and workflows
- [Security Commands](security-commands.md) - Command reference
- [Risk Configuration](risks-configuration.md) - Risk assessment settings
