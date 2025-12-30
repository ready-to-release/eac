# SAST (Static Application Security Testing)

Static analysis examines source code without executing it (white-box testing).

## What SAST Does

- Analyzes code structure and patterns
- Identifies potential vulnerabilities
- Detects insecure coding practices
- Runs before or during build

## What It Detects

| Vulnerability              | Description                          |
| -------------------------- | ------------------------------------ |
| SQL Injection              | Unsanitized database queries         |
| Cross-Site Scripting (XSS) | Unescaped user input in HTML         |
| Hardcoded Secrets          | API keys, passwords in code          |
| Insecure Cryptography      | Weak algorithms, bad implementations |
| Path Traversal             | Unsafe file path handling            |
| Command Injection          | Unsafe system command execution      |

## Running SAST Scans

Use the `scan` command with appropriate scanner types:

```bash
# SAST scan (Semgrep)
eac scan --scanner sast

# Secret detection
eac scan --scanner secrets

# Vulnerability scan
eac scan --scanner vuln

# IaC scanning (Terraform, Kubernetes)
eac scan --scanner iac

# Multiple scanners
eac scan --scanner sast,secrets,vuln
```

Evidence is written to `out/scan/<module>/<scanner>/`.

## Benefits

- Catches issues before code runs
- Fast feedback (seconds to minutes)
- Identifies exact code location
- No runtime environment needed

## Limitations

- False positives require tuning
- Can't detect runtime-only issues
- Requires language-specific analyzers
- May miss business logic flaws

## When to Use

| Stage             | Scan Type               | Purpose             |
| ----------------- | ----------------------- | ------------------- |
| Pre-commit (2)    | Secrets, critical vulns | Fast gate           |
| Merge Request (3) | Full SAST               | PR validation       |
| Commit (4)        | Full codebase           | Comprehensive check |

See [Shift-Left Security](shift-left.md) for the complete stage matrix.
