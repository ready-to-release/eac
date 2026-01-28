# Implementation Guides

> **Language-specific guides for implementing BDD specifications**

Learn how to implement executable specifications using your programming language and BDD framework.

---

## Overview

This section provides conceptual guidance for writing step definitions,
organizing test code, and integrating BDD specifications with your test runner.

Each implementation guide covers:

- Framework installation and setup
- File organization and directory structure
- Writing step definitions
- Test level implementation (L0-L4)
- Build tags and test execution
- Best practices and patterns

---

## Available Guides

| Language | Framework | Status   | Link                             |
| -------- | --------- | -------- | -------------------------------- |
| **Go**   | Godog     | Complete | [Go Implementation Guide](./go/index.md) |

---

## Key Concepts

### 1. Separation of WHAT and HOW

**Specifications** (WHAT) are business-readable Gherkin scenarios.

**Step definitions** (HOW) are language-specific code that implements the steps.

This separation allows non-technical stakeholders to understand the behavior while developers implement the automation.

### 2. Test Levels with Build Tags

Different test levels run in different environments:

| Level | Purpose                      | Speed        |
| ----- | ---------------------------- | ------------ |
| L0    | Fast unit tests, no I/O      | Microseconds |
| L1    | Unit tests with minimal deps | Milliseconds |
| L2    | Integration with containers  | Seconds      |
| L3    | Pre-production (PLTE)        | Minutes      |
| L4    | Production verification      | Minutes+     |

### 3. Step Definition Organization

- Group related steps together
- Share common steps across features
- Use context structures for state management

---

## Choosing an Implementation

### Go (Godog) - Available

**Best for**: CLI tools, high-performance applications, concurrent testing

### Python (behave) - Planned

**Best for**: Data science, API testing, web applications

### Java (Cucumber-JVM) - Planned

**Best for**: Enterprise applications, Android, legacy integration

### TypeScript (Cucumber.js) - Planned

**Best for**: Web frontend, Node.js backend, full-stack JavaScript

---

## Reference Documentation

For complete code examples, commands, and templates, see:

**[Go Testing Reference](../../../reference/eac/testing/go/index.md)** - Complete implementation guide including:

- [Overview](../../../reference/eac/testing/go/overview.md) - Godog setup and installation
- [File Organization](../../../reference/eac/testing/go/file-organization.md) - Directory structure and naming
- [Test Levels](../../../reference/eac/testing/go/test-levels.md) - Build tags (L0-L4)
- [Step Definitions](../../../reference/eac/testing/go/step-definitions.md) - Writing Godog steps
- [Best Practices](../../../reference/eac/testing/go/best-practices.md) - Testing patterns

---

## Getting Started

### If you're new to BDD implementation

1. **Understand the concepts first**:
   - [BDD Fundamentals](../concepts/bdd-fundamentals.md) - What is BDD?
   - [Three-Layer Approach](../concepts/three-layer-approach.md) - How layers work together

2. **Discover requirements**:
   - [Example Mapping](../discovery/example-mapping.md) - Workshop to find acceptance criteria
   - [Organizing Rules](../organization/organizing-rules.md) - Structure your specifications

3. **Choose your language guide**:
   - [Go Implementation Guide](./go/index.md) - Currently the only complete guide

---

## Related Documentation

### Core Concepts

- [BDD Fundamentals](../concepts/bdd-fundamentals.md) - BDD principles and Gherkin
- [Three-Layer Approach](../concepts/three-layer-approach.md) - Rules/Scenarios/Unit Tests
- [Executable Specifications](../concepts/executable-specifications.md) - Living documentation

### Organization

- [File Structure](../organization/file-structure.md) - Separation of specs and implementation
- [Organizing Specifications](../organization/index.md) - File and folder structure

### Testing Taxonomy

- [Test Levels](../taxonomy/test-levels.md) - L0-L4 execution environments
- [Verification Tags](../taxonomy/verification-tags.md) - @ov, @iv, @pv tags
- [Test Suites](../taxonomy/test-suites.md) - Pre-commit, acceptance, production
