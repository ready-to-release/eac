# Test Suites

> **Unit, integration, acceptance, and production verification suites**

Test suites select tests by tags for execution at specific CD Model stages.

---

## Overview

Test suites provide a way to run specific subsets of tests based on the testing stage:

| Suite                       | Tags Selected  | Purpose                 | Environment    |
| --------------------------- | -------------- | ----------------------- | -------------- |
| **unit**                    | `@L0`, `@L1`   | Fast module-level tests | DevBox/Agent   |
| **integration**             | `@L2`          | Emulated system tests   | Agent + Docker |
| **acceptance**              | `@L3`          | Production-like tests   | PLTE           |
| **production-verification** | `@L4` + `@piv` | Production smoke tests  | Production     |

All test suites automatically exclude tests tagged with `@ignore`.

---

## Conceptual Understanding

### Why Test Suites?

Test suites serve several purposes:

1. **Fast feedback** - Unit suites run in minutes, catching issues early
2. **Environment matching** - Tests run where they're designed to execute
3. **Progressive validation** - Each stage adds confidence before deployment
4. **Resource optimization** - Expensive tests run only when needed

### CD Model Integration

Test suites map directly to CD Model stages:

- **Pre-commit/MR/Commit** → unit suite (fast gate)
- **Integration** → integration suite (Docker validation)
- **Acceptance** → acceptance suite (production-like verification)
- **Production** → production-verification suite (smoke tests)

### Tag-Based Selection

Suites select tests automatically based on tags. You don't manually assign tests to suites - you tag tests with appropriate levels and verification types, and suites select them.

---

## Reference Documentation

For complete CLI commands, examples, and configuration details, see:

**[Test Suites Reference](../../../reference/eac/testing/test-suites.md)** - Complete implementation guide including:

- CLI commands (`eac test <module> --suite <suite>`)
- Suite selection logic with examples
- Execution time guidelines
- Debugging test selection
- Best practices

---

## Related Documentation

- [Test Levels](./test-levels.md) - L0-L4 execution environments
- [Verification Tags](./verification-tags.md) - @ov, @iv, @pv, @piv, @ppv
- [Tag Inheritance](./tag-inheritance.md) - How tags accumulate
- [Execution Control](./execution-control-tags.md) - @ignore and @Manual
