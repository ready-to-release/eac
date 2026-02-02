# Scan for Security Issues

## What You'll Accomplish

Detect security vulnerabilities, secrets, and compliance issues before committing code.

## Prerequisites

- Security scanners installed (Trivy, Semgrep)
- Working in repository root

## Steps

### 1. Run All Security Scans

```bash
eac scan
```

**What happens**: Runs all security scans (vulnerabilities, secrets, SAST, IaC)

### 2. Scan for Vulnerabilities

```bash
eac scan --scanner vuln
```

**What happens**: Uses Trivy to scan for known CVEs in dependencies

### 3. Detect Secrets

```bash
eac scan --scanner secrets
```

**What happens**: Scans for exposed API keys, passwords, tokens

### 4. Static Analysis

```bash
eac scan --scanner sast
```

**What happens**: Uses Semgrep for code security issues

## Targeted Scans

```bash
# Scan infrastructure code
eac scan --scanner iac

# Check compliance standards
eac scan --scanner compliance

# Generate SBOM
eac scan --scanner sbom

# Run multiple scan types together
eac scan --scanner vuln,secrets,sbom
```

## Example Scenario

Pre-commit security check:

```bash
# Run full security scan
eac scan

# Output:
# Running vulnerability scan... ✓ No vulnerabilities
# Detecting secrets... ✗ Found 1 potential secret
#   File: config/dev.yaml
#   Line: 23
#   Type: API Key
#
# Running SAST... ✓ No issues
# Scanning IaC... ✓ No misconfigurations

# Fix the secret
# Remove hardcoded API key, use environment variable

# Scan again
eac scan --scanner secrets
# ✓ No secrets detected
```

## CI Integration

```bash
# In CI pipeline
eac scan || exit 1
```

## Common Issues

| Problem           | Solution                          |
| ----------------- | --------------------------------- |
| False positives   | Add to ignore list in config      |
| Scanner not found | Install Trivy/Semgrep             |
| Slow scans        | Use targeted scans with --scanner |

## Next Steps

- [Validate Before Commit](./validate-before-commit.md) → Complete quality checks

## Related Commands

- [`scan`](../../../../reference/eac/commands/scan/scan.md) - Run all scans with --scanner flag
- [`scan zap`](../../../../reference/eac/commands/scan/zap.md) - Dynamic testing (DAST)
