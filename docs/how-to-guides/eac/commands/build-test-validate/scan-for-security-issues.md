# Scan for Security Issues
## What You'll Accomplish

Detect security vulnerabilities, secrets, and compliance issues before committing code.

## Prerequisites

- Security scanners installed (Trivy, Semgrep)
- Working in repository root

## Steps

### 1. Run All Security Scans

```bash
r2r eac scan
```

**What happens**: Runs all security scans (vulnerabilities, secrets, SAST, IaC)

### 2. Scan for Vulnerabilities

```bash
r2r eac scan vuln
```

**What happens**: Uses Trivy to scan for known CVEs in dependencies

### 3. Detect Secrets

```bash
r2r eac scan secrets
```

**What happens**: Scans for exposed API keys, passwords, tokens

### 4. Static Analysis

```bash
r2r eac scan sast
```

**What happens**: Uses Semgrep for code security issues

## Targeted Scans

```bash
# Scan infrastructure code
r2r eac scan iac

# Check compliance standards
r2r eac scan compliance

# Generate SBOM
r2r eac scan sbom
```

## Example Scenario

Pre-commit security check:

```bash
# Run full security scan
r2r eac scan

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
r2r eac scan secrets
# ✓ No secrets detected
```

## CI Integration

```bash
# In CI pipeline
r2r eac scan || exit 1
```

## Common Issues

| Problem | Solution |
|---------|----------|
| False positives | Add to ignore list in config |
| Scanner not found | Install Trivy/Semgrep |
| Slow scans | Use targeted scans only |

## Next Steps

- [Validate Before Commit](./validate-before-commit.md) → Complete quality checks

## Related Commands

- [`scan`](../../../../reference/commands/scan/scan.md) - Run all scans
- [`scan vuln`](../../../../reference/commands/scan/vuln.md) - Vulnerability scan
- [`scan secrets`](../../../../reference/commands/scan/secrets.md) - Secret detection
