# Dynamic Application Security Testing (DAST)

Runtime security testing using OWASP ZAP (black-box testing).

---

## Command

```bash
# DAST scan with OWASP ZAP
eac scan zap --url <target-url>

# Scan specific module
eac scan zap --url http://localhost:8080 --module api-service
```

**Tool**: OWASP ZAP

**Output**: `out/scan/<module>/zap/`

---

## What It Detects

Tests running application for:

- **Authentication flaws**: Weak auth, session management issues
- **Authorization bypasses**: Privilege escalation, access control
- **Injection attacks**: SQL, command, LDAP injection at runtime
- **Security misconfigurations**: Default credentials, open ports, headers
- **Sensitive data exposure**: Unencrypted data, information leakage
- **Business logic flaws**: Workflow bypasses, race conditions

---

## Scan Modes

| Mode     | Duration  | Use Case         | Environment |
| -------- | --------- | ---------------- | ----------- |
| Baseline | 5-10 min  | Quick validation | Test (safe) |
| Full     | 1-4 hours | Comprehensive    | Test only   |
| API      | 10-30 min | API-focused      | Test only   |

**Baseline mode**: Passive scanning, no active attacks (safe for production-like environments)

**Full mode**: Active scanning with simulated attacks (test environments only)

---

## Requirements

- **Running application**: Target must be deployed and accessible
- **Test data**: Application should be seeded with test data
- **Network access**: Scanner must reach target URL
- **Safe environment**: Active scans should only run in test environments

---

## Related Documentation

- **[Security Index](./index.md)** - Security scanning overview
- **[Scan Commands](../commands/scan/index.md)** - Full scan command reference
