# Specifications Concepts

Understanding the three-layer testing approach, executable specifications, and test-driven development practices.

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

- Why three layers?
- How the layers interact
- Traceability across layers
- Example workflow from requirement to code

### [ATDD and BDD with Gherkin](atdd-bdd-with-gherkin.md)

Understanding how Acceptance Test-Driven Development and Behavior-Driven Development work together using unified Gherkin format.

**Topics covered:**

- What is ATDD and BDD?
- The unified approach with Rule blocks
- Gherkin and Ubiquitous Language
- Key principles and best practices
- Specs vs Implementation architecture

### [Ubiquitous Language](ubiquitous-language.md)

Building shared domain vocabulary that bridges business and technical communication.

**Topics covered:**

- The translation problem in software teams
- Domain-Driven Design and Ubiquitous Language
- From shared language to executable specifications
- Bounded contexts and language boundaries
- Continuous evolution of domain vocabulary

### [Event Storming](event-storming.md)

Collaborative workshop technique for discovering domain language and business processes.

**Topics covered:**

- Three Event Storming formats (Big Picture, Process Modeling, Software Design)
- Workshop facilitation and sticky note grammar
- Key outputs: domain events, actors, commands, policies
- From Event Storming to specifications
- Workshop best practices

### [Example Mapping](example-mapping.md)

Collaborative workshop technique for discovering requirements and creating acceptance criteria.

**Topics covered:**

- The four card colors and their purpose
- Workshop process and facilitation
- Converting cards to Gherkin specifications
- Readiness assessment criteria
- Best practices and common pitfalls

### [Review and Iterate](review-and-iterate.md)

Maintaining living specifications through continuous feedback and refinement.

**Topics covered:**

- When to review specifications
- What to do after Example Mapping sessions
- Continuous iteration practices (weekly, monthly, quarterly)
- Feedback loops from implementation and production
- Specification refactoring strategies
- Review ceremonies and health metrics
- Handling specification changes

### [Risk Controls](risk-controls.md)

Understanding risk-based testing and how to define executable risk controls for compliance.

**Topics covered:**

- What are risk controls and why executable?
- When you need risk controls (regulated domains)
- Identifying relevant controls for your context
- The traceability chain from risk to evidence
- Common compliance frameworks (FDA, ISO, PCI-DSS, GDPR)
- Best practices for defining controls

---

## Quick Comparison

| Layer    | Focus                | Representation         | Tool    | Specification                    | Implementation            |
| -------- | -------------------- | ---------------------- | ------- | -------------------------------- | ------------------------- |
| **ATDD** | Business requirements | `Rule:` blocks         | Godog   | `specs/.../specification.feature` | `src/.../steps_test.go`   |
| **BDD**  | User-facing behavior | `Scenario:` under Rule | Godog   | `specs/.../specification.feature` | `src/.../steps_test.go`   |
| **TDD**  | Implementation testing | Go test functions      | Go test | N/A                              | `src/.../*_test.go`       |

---

## Related Documentation

- **Looking for technical details?** See [Specifications Reference](../../reference/specifications/index.md)
- **Need to perform tasks?** See [Specifications How-to Guides](../../how-to-guides/specifications/index.md)
