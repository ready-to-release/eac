# Configure Security Scanning

{{ page_breadcrumb() }}

## What You'll Accomplish

Set up security scanning tools (Trivy, Semgrep, OWASP ZAP) for automated vulnerability detection.

## Prerequisites

- Security scanners installed (Trivy, Semgrep)
- Docker installed (for ZAP)

## Steps

### 1. Initialize Security Configuration

```bash
r2r eac init --security
```

**What happens**: Sets up security scanner configuration

### 2. Configure Trivy

Create `.trivyignore` for false positives:

```bash
# .trivyignore
CVE-2023-12345  # Known safe in our context
```

### 3. Configure Semgrep

Create `.semgrep.yml` for rules:

```yaml
rules:
  - id: no-hardcoded-secrets
    patterns:
      - pattern: password = "..."
    severity: ERROR
```

### 4. Test Configuration

```bash
r2r eac scan
```

**What happens**: Runs all security scans with configuration

## Scanner Configuration

### Trivy Configuration

```yaml
# .trivy.yaml
severity: HIGH,CRITICAL
ignore-unfixed: true
timeout: 10m
```

### Semgrep Configuration

```yaml
# .semgrep.yml
rules:
  - id: custom-rule
    languages: [go]
    message: Custom security rule
    severity: ERROR
    pattern: |
      dangerous_function(...)
```

### ZAP Configuration

```bash
# Configure ZAP for DAST
r2r eac scan zap http://localhost:8080 --config zap-config.yaml
```

## Example Scenario

Setting up security scanning for new project:

```bash
# Initialize security
r2r eac init --security

# Output:
# ✓ Created .trivyignore
# ✓ Created .semgrep.yml
# ✓ Security scanning configured

# Run initial scan
r2r eac scan

# Output:
# Running Trivy... ✓ No critical vulnerabilities
# Running Semgrep... ✗ Found 2 issues
#   - Hardcoded password in config.go:23
#   - SQL injection risk in db.go:45
#
# Running secret detection... ✓ No secrets

# Fix issues and rescan
# ... fix code ...

r2r eac scan
# ✓ All scans passed
```

## CI Integration

```yaml
# .github/workflows/security.yml
name: Security Scan

on: [push, pull_request]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Security Scan
        run: r2r eac scan || exit 1
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "Scanner not found" | Install Trivy/Semgrep |
| Too many false positives | Configure ignore lists |
| Slow scans | Limit severity levels |

## Next Steps

- [Scan for Security Issues](../quality-and-validation/scan-for-security-issues.md) → Run scans

## Related Commands

- [`init`](../../../reference/commands/other/init.md) - Initialize configuration
- [`scan`](../../../reference/commands/scan/scan.md) - Run security scans

{{ diataxis_footer() }}
