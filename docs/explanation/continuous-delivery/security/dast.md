# DAST (Dynamic Application Security Testing)

Dynamic analysis tests running applications (black-box testing).

---

## Overview

DAST interacts with your application via APIs or UI to simulate attacker behavior and test actual runtime security.

### What It Detects

| Vulnerability              | Description                             |
| -------------------------- | --------------------------------------- |
| Authentication Flaws       | Weak auth, session issues               |
| Authorization Bypasses     | Privilege escalation                    |
| Injection Attacks          | SQL, command, LDAP injection at runtime |
| Security Misconfigurations | Default credentials, open ports         |
| Sensitive Data Exposure    | Unencrypted data, information leakage   |
| Business Logic Flaws       | Workflow bypasses                       |

### Benefits

- Tests actual application behavior
- Finds runtime-only vulnerabilities
- Language-agnostic
- Tests real attack scenarios

### Limitations

- Slower than SAST (minutes to hours)
- Requires running application
- May miss code-level issues
- Can have false positives

---

## Reference Documentation

For CLI commands and scanner configuration, see:

**[DAST Reference](../../../reference/eac/security/dast.md)** - Complete implementation guide including:

- `eac scan --scanner zap` commands
- Scan modes (Baseline, Full, API)
- Evidence output locations

---

## Related Documentation

- [SAST](./sast.md) - Static Application Security Testing
- [Supply Chain Security](./supply-chain.md) - Dependency and container scanning
- [Shift-Left Security](./shift-left.md) - Security integration principles
