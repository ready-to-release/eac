# Core Concepts

> **Foundational concepts for Behavior-Driven Development and testing**

Understand the principles and practices that underpin specification-driven development.

---

## Overview

This section covers:

- **Three-Layer Approach** - Rules, Scenarios, and Unit Tests working together
- **BDD Fundamentals** - Behavior-Driven Development principles and Gherkin
- **Executable Specifications** - Living documentation that drives development
- **Specifications Evolution** - How specifications change over time
- **Canon TDD Workflow** - Test-Driven Development cycle
- **Ubiquitous Language** - Shared vocabulary for domain understanding

---

## In This Section

| Topic                                                       | Description                                        |
| ----------------------------------------------------------- | -------------------------------------------------- |
| [Three-Layer Approach](./three-layer-approach.md)           | Rules, Scenarios, and Unit Tests hierarchy         |
| [BDD Fundamentals](./bdd-fundamentals.md)                   | Behavior-Driven Development and Gherkin syntax     |
| [Executable Specifications](./executable-specifications.md) | Living documentation as automated tests            |
| [Specifications Evolution](./specifications-evolution.md)   | How specifications change with understanding       |
| [Canon TDD Workflow](./canon-tdd-workflow.md)               | Test-Driven Development cycle (Red-Green-Refactor) |
| [Ubiquitous Language](./ubiquitous-language.md)             | Domain-specific shared vocabulary                  |

---

## Key Principles

### 1. Specifications are executable

Specifications aren't just documentation - they're automated tests that verify system behavior.

```gherkin
Scenario: User creates a project
  Given I am in an empty directory
  When I run "r2r init my-project"
  Then a file named "my-project/r2r.yaml" should exist
```

This scenario is both documentation AND an automated test.

### 2. Three layers of testing

```text
Rules (Acceptance Criteria)
  └── Scenarios (Examples)
        └── Unit Tests (Implementation)
```

Each layer serves a distinct purpose:

- **Rules** define measurable business requirements
- **Scenarios** provide concrete examples
- **Unit Tests** verify implementation details

### 3. Ubiquitous Language

Use domain terms consistently across specifications, code, and conversations:

- ✅ "release" (domain term)
- ❌ "deployment" (technical term)

### 4. Specifications evolve

Specifications are living documents that improve through:

- Discovery (Event Storming, Example Mapping)
- Implementation (edge cases, refinement)
- Feedback (bugs, production issues)

---

## Quick Start

### New to BDD?

1. **Start here**: [BDD Fundamentals](./bdd-fundamentals.md) - Understand Given/When/Then
2. **Learn the structure**: [Three-Layer Approach](./three-layer-approach.md) - How Rules/Scenarios/Tests work together
3. **Discover requirements**: [Example Mapping](../discovery/example-mapping.md) - Workshop technique for finding acceptance criteria
4. **Write specifications**: [Organizing Specifications](../organization/index.md) - File structure and naming conventions

### Common Workflows

**Test-Driven Development**:

1. Write failing scenario (Red)
2. Write step definitions
3. Implement minimal code (Green)
4. Refactor
5. Commit

See: [Canon TDD Workflow](./canon-tdd-workflow.md)

**Specification-First Development**:

1. Run Example Mapping workshop
2. Write feature with Rules and Scenarios
3. Implement step definitions
4. Write unit tests
5. Implement production code

---

## Related Documentation

### Workshop Techniques

- [Event Storming](../discovery/event-storming-overview.md) - Discover domain events and language
- [Example Mapping](../discovery/example-mapping.md) - Discover acceptance criteria with colored cards

### Organization

- [Organizing Specifications](../organization/index.md) - File structure and naming conventions
- [Testing Taxonomy](../taxonomy/index.md) - Tags, test levels, and execution control

### Implementation

- [Go Implementation Guide](../implementation/go/index.md) - Go/Godog BDD implementation
- [Quality and Maintenance](../quality/index.md) - Review and iterate on specifications
