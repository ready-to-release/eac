# Go Implementation Guide

> **Implementation-specific guide for Go/Godog BDD and testing**

This section provides an overview of implementing BDD specifications with Go and Godog.

## Overview

This project uses:

- **Go** for production code and unit tests
- **Godog** for executing Gherkin specifications
- **Go test framework** for unit and integration tests
- **Build tags** for test level isolation (L0-L4)

## Why Go for BDD?

| Advantage | Description |
|-----------|-------------|
| Performance | Compiled nature makes tests fast and efficient |
| Simplicity | Simple syntax and minimal abstraction |
| Tooling | Excellent support with `go test`, coverage, profiling |
| Concurrency | Built-in support for parallel test execution |
| Standard Library | Rich library reduces external dependencies |

---

## Reference Documentation

For complete implementation details, code templates, and commands, see:

**[Go Testing Reference](../../../../reference/eac/testing/go/index.md)** - Complete implementation guide including:

- [Overview](../../../../reference/eac/testing/go/overview.md) - Godog setup and installation
- [File Organization](../../../../reference/eac/testing/go/file-organization.md) - Directory structure and naming
- [Test Levels](../../../../reference/eac/testing/go/test-levels.md) - Build tags (L0-L4)
- [Step Definitions](../../../../reference/eac/testing/go/step-definitions.md) - Writing Godog steps
- [Best Practices](../../../../reference/eac/testing/go/best-practices.md) - Testing patterns

---

## Related Documentation

### Conceptual Understanding

- [Three-Layer Testing Approach](../../concepts/three-layer-approach.md) - Testing philosophy
- [BDD Fundamentals](../../concepts/bdd-fundamentals.md) - BDD principles
- [Testing Taxonomy](../../taxonomy/) - Tag taxonomy concepts

### Organizational

- [Organizing Specifications](../../organization/) - Specification structure
- [Example Mapping](../../discovery/example-mapping.md) - Requirements discovery
