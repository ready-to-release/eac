# Testing Concepts

Understanding the three-layer testing approach and test-driven development practices.

---

## Overview

This project uses a **three-layer testing approach** that separates business requirements, user-facing behavior, and implementation testing into distinct layers with specialized tools.

Each layer serves a specific purpose and uses different tools and formats:

- **ATDD** (Acceptance Test-Driven Development) - Business requirements and acceptance criteria
- **BDD** (Behavior-Driven Development) - User-facing behavior and observable outcomes
- **TDD** (Test-Driven Development) - Implementation testing and code correctness

---

## Available Explanations

### [Three-Layer Testing Approach](three-layer-approach.md)

Understanding how ATDD, BDD, and TDD work together to deliver quality software.

**Topics covered:**

- Why three separate layers?
- How the layers interact
- When to use each layer
- Traceability across layers
- Example workflow from requirement to code

### [ATDD Concepts](atdd-concepts.md)

Understanding Acceptance Test-Driven Development with Gauge.

**Topics covered:**

- What is ATDD and why use it?
- Example Mapping workshop technique
- User stories and acceptance criteria
- Gauge specifications and executable tests
- Stakeholder collaboration

### [BDD Concepts](bdd-concepts.md)

Understanding Behavior-Driven Development with Godog.

**Topics covered:**

- What is BDD and why use it?
- Gherkin language (Given/When/Then)
- Observable behavior vs implementation
- Converting examples to scenarios
- Living documentation

---

## Quick Comparison

| Layer | Focus | Format | Tool | File |
|-------|-------|--------|------|------|
| **ATDD** | Business requirements | Markdown specifications | Gauge | `acceptance.spec` |
| **BDD** | User-facing behavior | Gherkin scenarios | Godog | `behavior.feature` |
| **TDD** | Implementation testing | Unit tests | Go test | `*_test.go` |

---

**Looking for technical details?** See [Testing Reference](../../reference/testing/index.md)

**Need to perform tasks?** See [Testing How-to Guides](../../how-to-guides/testing/index.md)
