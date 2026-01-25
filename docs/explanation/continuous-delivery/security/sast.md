# SAST (Static Application Security Testing)

Static analysis examines source code without executing it (white-box testing).

---

## Overview

SAST analyzes code structure and patterns to identify potential vulnerabilities before the code runs.

### What It Detects

| Vulnerability              | Description                          |
| -------------------------- | ------------------------------------ |
| SQL Injection              | Unsanitized database queries         |
| Cross-Site Scripting (XSS) | Unescaped user input in HTML         |
| Hardcoded Secrets          | API keys, passwords in code          |
| Insecure Cryptography      | Weak algorithms, bad implementations |
| Path Traversal             | Unsafe file path handling            |
| Command Injection          | Unsafe system command execution      |

### Benefits

- Catches issues before code runs
- Fast feedback (seconds to minutes)
- Identifies exact code location
- No runtime environment needed

### Limitations

- False positives require tuning
- Can't detect runtime-only issues
- Requires language-specific analyzers
- May miss business logic flaws

---

## Reference Documentation

For CLI commands and scanner configuration, see:

**[SAST Reference](../../../reference/eac/security/sast.md)** - Complete implementation guide including:

- `eac scan --scanner sast/secrets/vuln/iac` commands
- Evidence output locations
- Stage-specific scanning guidance

---

## Related Documentation

- [DAST](./dast.md) - Dynamic Application Security Testing
- [Supply Chain Security](./supply-chain.md) - Dependency and container scanning
- [Shift-Left Security](./shift-left.md) - Security integration principles
