# Test Suites Reference

Test suite definitions from `contracts/core/0.1.0/schemas/defaults/test-suites.yml`.

---

## unit

**Tags**: Include `@L0` or `@L1`, exclude `@L2`, `@L3`, `@L4`
**Purpose**: Fast module-level validation
**Time**: 2-5 minutes
**Environment**: DevBox or Build Agent
**Stage**: 2-4 (Pre-commit, MR, Commit)

```bash
eac test <module> --suite unit
```

---

## integration

**Tags**: Include `@L2`, exclude `@L0`, `@L1`, `@L3`, `@L4`
**Purpose**: Emulated system tests with Docker
**Time**: 5-15 minutes
**Environment**: Build Agent with Docker
**Stage**: 5 (Continuous Build)

```bash
eac test <module> --suite integration
```

---

## acceptance

**Tags**: Include `@L3`, exclude `@L0`, `@L1`, `@L2`, `@L4`
**Purpose**: Production-like system tests
**Time**: 1-2 hours
**Environment**: PLTE (Production-Like Test Environment)
**Stage**: 6 (PLTE Deployment)
**Extended Suite**: Yes

```bash
eac test <module> --suite acceptance
```

Acceptance tests verify:

- **@iv** - Installation verification (deployment succeeded)
- **@ov** - Operational verification (features work)
- **@pv** - Performance verification (meets SLA/SLI/SLO)

---

## production-verification

**Tags**: Require `@L4` and `@piv`
**Purpose**: Production smoke tests
**Time**: 5-15 minutes
**Environment**: Production
**Stage**: 11-12 (Production Deployment)
**Extended Suite**: Yes

```bash
eac test <module> --suite production-verification
```

Production Installation Verification (@piv) confirms:

- Deployment succeeded
- Core functionality operational
- System health checks pass

---

## manual

**Tags**: Require `@Manual`, exclude `@L0`, `@L1`, `@L2`, `@L3`, `@L4`
**Purpose**: Human-executed tests requiring manual verification
**Environment**: As required by test
**Extended Suite**: Yes

Manual tests are exported, executed by humans, then results imported:

```bash
eac test-export-manual <module>          # Export scenarios
# ... human execution ...
eac test-import-manual <module> --results results.json
```

---

## Suite Configuration

All suites automatically exclude tests tagged with `@ignore`.

**Extended suites** (`acceptance`, `production-verification`, `manual`) are not run by default in CI/CD pipelines unless explicitly requested.

**YAML Definition**: `contracts/core/0.1.0/schemas/defaults/test-suites.yml`
**JSON Schema**: `contracts/core/0.1.0/schemas/test-suites.schema.json`

---

## Related Documentation

- [Test Levels Explanation](../../../explanation/specifications/taxonomy/test-levels.md) - L0-L4 environment concepts
- [Test Commands](../commands/test/index.md) - CLI command reference
- [Manual Testing](./manual-tests.md) - Manual test workflow
