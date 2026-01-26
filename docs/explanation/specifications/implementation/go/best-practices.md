# Best Practices

> **Testing patterns and conventions for Go/Godog**

Follow these best practices to write clear, maintainable, and reliable tests.

---

## Key Principles

| Principle          | Description                                  |
| ------------------ | -------------------------------------------- |
| Clear naming       | `Test<Function>_<Scenario>_<ExpectedResult>` |
| Table-driven tests | Multiple variants in one test                |
| Test isolation     | Each test is independent                     |
| AAA pattern        | Arrange, Act, Assert                         |
| Descriptive errors | Show expected vs actual                      |

---

## Test Naming Convention

**Pattern**: `Test<Function>_<Scenario>_<ExpectedResult>`

| Good                                                 | Bad         |
| ---------------------------------------------------- | ----------- |
| `TestParseConfig_WithValidYAML_ShouldSucceed`        | `TestParse` |
| `TestCreateUser_WithExistingEmail_ShouldReturnError` | `Test1`     |

---

## Core Patterns

### Table-Driven Tests

Use when testing multiple variants of the same behavior.

### Test Isolation

- Use `t.Run()` for subtests
- Use `t.TempDir()` for filesystem tests
- Clean up resources with `defer`

### Arrange-Act-Assert

Structure tests clearly with setup, execution, and verification phases.

---

## Reference Documentation

For complete code examples, patterns, and anti-patterns, see:

**[Best Practices Reference](../../../../reference/eac/testing/go/best-practices.md)** - Complete guide including:

- Table-driven test examples
- Test isolation patterns
- AAA pattern examples
- Error message formatting
- Test helpers with `t.Helper()`
- Parallel test execution
- Coverage analysis commands
- Benchmarking
- Common pitfalls to avoid

---

## Related Documentation

- [File Organization](./file-organization.md) - Test file structure
- [Test Levels](./test-levels.md) - Build tags and isolation
- [Step Definitions](./step-definitions.md) - Godog patterns
