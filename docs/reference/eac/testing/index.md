# Testing Reference

Technical reference for EAC test execution, suite configuration, and test runners.

## Test Suites

EAC uses tag-based test suite selection defined in `contracts/core/0.1.0/schemas/defaults/test-suites.yml`:

| Suite                       | Tags       | Purpose                          | Stage    |
| --------------------------- | ---------- | -------------------------------- | -------- |
| **unit**                    | @L0, @L1   | Fast module tests (2-5 min)      | 2-4      |
| **integration**             | @L2        | Docker-based integration (5-15m) | 5        |
| **acceptance**              | @L3        | Production-like PLTE (1-2h)      | 6        |
| **production-verification** | @L4 + @piv | Production smoke tests           | 11-12    |
| **manual**                  | @Manual    | Human-executed tests             | Releases |

**Commands**:

```bash
eac test <module> --suite unit           # Run unit tests
eac test <module> --suite integration    # Run integration tests
eac test <module> --suite acceptance     # Run acceptance tests
```

**See**: [Test Suites Detail](./test-suites.md)

---

## Manual Testing

Manual tests are Gherkin scenarios tagged with `@Manual` that require human execution.

**Workflow**:

1. Export scenarios: `eac test-export-manual <module>`
2. Execute tests manually (human)
3. Import results: `eac test-import-manual <module> --results results.json`

**Schema**: Manual test results use JSON schema defined in `contracts/core/0.1.0/schemas/manual-test-results.schema.json`

**See**: [Manual Testing Detail](./manual-tests.md)

---

## Test Runners

EAC supports multiple test runners via adapters:

**Go**:

- **gotest** - Standard Go unit tests (`go test`)
- **godog** - BDD/Gherkin tests for Go

**TypeScript/JavaScript**:

- **mocha** - Unit tests via npm
- **cucumber-js** - BDD/Gherkin tests

**Python**:

- **pytest** - Unit tests
- **behave** - BDD/Gherkin tests

**Other**:

- **reqnroll** - .NET BDD tests
- **cucumber** - Ruby BDD tests

**See**: [Adapters](../modules/adapters.md)

---

## Test Levels

Tests are categorized by environment requirements:

| Level  | Tag | Environment | Examples                            |
| ------ | --- | ----------- | ----------------------------------- |
| **L0** | @L0 | DevBox      | Pure unit tests, no I/O             |
| **L1** | @L1 | DevBox      | Unit tests with minimal I/O         |
| **L2** | @L2 | Build Agent | Docker-based integration tests      |
| **L3** | @L3 | PLTE        | Production-like system tests        |
| **L4** | @L4 | Production  | Production verification smoke tests |

**See**: [Test Levels (Explanation)](../../../explanation/specifications/taxonomy/test-levels.md)

---

## Related Documentation

- **[Test Commands](../commands/test/index.md)** - CLI command reference
- **[Specifications](../specifications/index.md)** - Gherkin specification reference
- **[Test Levels Explanation](../../../explanation/specifications/taxonomy/test-levels.md)** - Conceptual overview
