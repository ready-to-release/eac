# DAST (Dynamic Application Security Testing)

Dynamic analysis tests running applications (black-box testing).

## What DAST Does

- Interacts with application via APIs/UI
- Simulates attacker behavior
- Tests actual runtime behavior
- No source code access needed

## What It Detects

| Vulnerability              | Description                             |
| -------------------------- | --------------------------------------- |
| Authentication Flaws       | Weak auth, session issues               |
| Authorization Bypasses     | Privilege escalation                    |
| Injection Attacks          | SQL, command, LDAP injection at runtime |
| Security Misconfigurations | Default credentials, open ports         |
| Sensitive Data Exposure    | Unencrypted data, information leakage   |
| Business Logic Flaws       | Workflow bypasses                       |

## Running DAST Scans

Use the `scan` command with the `zap` scanner:

```bash
# DAST scan using OWASP ZAP
eac scan --scanner zap

# Scan specific module
eac scan eac-core --scanner zap
```

Evidence is written to `out/scan/<module>/zap/`.

## Scan Modes

| Mode     | Duration  | Use Case         | Environment |
| -------- | --------- | ---------------- | ----------- |
| Baseline | 5-10 min  | Quick validation | Any (safe)  |
| Full     | 1-4 hours | Comprehensive    | Test only   |
| API      | 10-30 min | API-focused      | Test only   |

## Benefits

- Tests actual application behavior
- Finds runtime-only vulnerabilities
- Language-agnostic
- Tests real attack scenarios

## Limitations

- Slower than SAST (minutes to hours)
- Requires running application
- May miss code-level issues
- Can have false positives

## When to Use

| Stage           | Scan Type | Purpose               |
| --------------- | --------- | --------------------- |
| Acceptance (5)  | Baseline  | Quick validation      |
| Extended (6)    | Full scan | Comprehensive testing |
| Production (11) | Baseline  | Continuous monitoring |

## Related Documentation

- [Shift-Left Security (Conceptual)](../../../explanation/continuous-delivery/security/shift-left.md) - Security integration principles
- [Scan Command Reference](../commands/scan/index.md) - Full scan command options
