# Go/Godog Overview

> **Introduction to Go/Godog BDD testing**

This guide provides an overview of Go-specific implementation for the BDD testing approach.

---

## What This Project Uses

- **Go** for production code and unit tests
- **Godog** for executing Gherkin specifications
- **Go test framework** for unit and integration tests
- **Build tags** for test level isolation

---

## Why Go for BDD?

### Advantages

| Advantage | Description |
|-----------|-------------|
| Performance | Go's compiled nature makes tests fast and efficient |
| Simplicity | Simple syntax and minimal abstraction make tests easy to understand |
| Tooling | Excellent tooling support with `go test`, coverage analysis, and profiling |
| Concurrency | Built-in concurrency support for parallel test execution |
| Standard Library | Rich standard library reduces external dependencies |

---

## Why Godog?

**Godog** is the Go implementation of Cucumber, enabling BDD with Gherkin syntax.

| Feature | Description |
|---------|-------------|
| Native Go Integration | Works seamlessly with `go test` framework |
| Gherkin Support | Full support for Feature, Rule, Scenario, Examples |
| Table Tests | Natural mapping from Gherkin examples to Go table-driven tests |
| Parallel Execution | Built-in support for concurrent scenario execution |
| Output Formats | Multiple formats (pretty, progress, JSON, JUnit) |

---

## Reference Documentation

For installation, setup, and code examples, see:

**[Go Testing Reference](../../../../reference/eac/testing/go/overview.md)** - Complete setup guide including:

- Godog installation
- Feature file creation
- Step definition implementation
- Test execution commands

---

## Related Documentation

- [File Organization](./file-organization.md) - Directory structure overview
- [Test Levels](./test-levels.md) - L0-L4 test isolation concept
- [Step Definitions](./step-definitions.md) - Step definition patterns
- [Best Practices](./best-practices.md) - Go testing best practices
