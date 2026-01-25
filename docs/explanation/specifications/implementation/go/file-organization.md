# File Organization

> **Directory structure and file naming conventions for Go/Godog projects**

Overview of how to organize specification files, step definitions, and unit tests in Go projects.

---

## Directory Structure Overview

```text
project/
├── specs/
│   └── <module>/
│       └── <feature>/
│           └── specification.feature    # Gherkin specs
└── src/
    └── <module>/
        ├── *.go                          # Production code
        ├── *_test.go                     # L1 unit tests (default)
        └── tests/
            └── steps_test.go             # Godog step definitions
```

---

## Key Concepts

| Component | Location | Purpose |
|-----------|----------|---------|
| Specification Files | `specs/<module>/<feature>/` | Gherkin scenarios |
| Step Definitions | `src/<module>/tests/` | Connect Gherkin to Go |
| Unit Tests | `src/<module>/` | Fast isolated tests |
| Integration Tests | `src/<module>/tests/` | L2 tests with build tags |

---

## Reference Documentation

For complete file naming conventions and test execution commands, see:

**[File Organization Reference](../../../../reference/eac/testing/go/file-organization.md)** - Complete guide including:

- Full directory structure
- File naming conventions (production, unit, integration, L0)
- Test execution commands (unit, integration, BDD)
- Combined test suite execution

---

## Related Documentation

- [Test Levels](./test-levels.md) - Build tags and test isolation (L0-L4)
- [Step Definitions](./step-definitions.md) - Writing Godog step definitions
- [Best Practices](./best-practices.md) - Go testing best practices
