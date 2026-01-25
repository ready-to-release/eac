# Quality Gates Reference

CLI commands and configuration for CD Model quality gates.

## Quick Reference

```bash
# Pre-commit validation
r2r eac validate specs
r2r eac validate test-tags
r2r eac test <module> --suite unit

# Security scanning
r2r eac scan --scanner sast
r2r eac scan --scanner vuln
r2r eac scan --scanner secrets

# Coverage and test execution
r2r eac test <module> --coverage
r2r eac show test-summary <module>
```

---

## In This Section

| Topic | Description |
|-------|-------------|
| [Pre-commit Setup](./precommit-setup.md) | Hook configuration, tool setup, time budgets |
| [Pre-commit Checks](./precommit-checks.md) | Check categories, tools, execution environments |
| [Evidence Collection](./evidence-collection.md) | Test and security scan evidence formats |

---

## Quality Gate Stages

| Stage | Gate Type | Time Budget | Reference |
|-------|-----------|-------------|-----------|
| Stage 2 | Pre-commit | 5-10 min | [Pre-commit Setup](./precommit-setup.md) |
| Stage 3 | Merge Request | Minutes-hours | [Merge Request (Conceptual)](../../../explanation/continuous-delivery/quality-gates/merge-request-gates.md) |
| Stage 9 | Release | Seconds (CDe) or hours (RA) | [Release (Conceptual)](../../../explanation/continuous-delivery/quality-gates/release-gates.md) |

---

## Related Documentation

- [Quality Gates (Conceptual)](../../../explanation/continuous-delivery/quality-gates/index.md) - What are quality gates and why
- [Test Suites](../testing/test-suites.md) - Test execution commands
- [Security Scanning](../security/index.md) - SAST, DAST, supply chain commands
