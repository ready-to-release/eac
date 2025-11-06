# ATDD and BDD with Gherkin Rule Blocks

Understanding how Acceptance Test-Driven Development and Behavior-Driven Development work together using unified Gherkin format.

---

## The Unified Approach

This project maintains the conceptual distinction between ATDD (acceptance criteria) and BDD (behavior scenarios) while using a single unified format: Gherkin.

### ATDD Layer: Rule Blocks

- **Purpose**: Define acceptance criteria
- **Format**: `Rule:` blocks in Gherkin
- **Location**: `specs/<module>/<feature>/specification.feature`
- **Origin**: Blue cards from Example Mapping
- **Audience**: Product owners, stakeholders, QA
- **Focus**: WHAT the system must do

### BDD Layer: Scenario Blocks

- **Purpose**: Executable behavior examples
- **Format**: `Scenario:` blocks under Rules
- **Location**: `specs/<module>/<feature>/specification.feature` (same file as Rules)
- **Implementation**: `src/<module>/tests/steps_test.go` (separate from specification)
- **Origin**: Green cards from Example Mapping
- **Audience**: Developers, QA, automation engineers
- **Focus**: HOW the system behaves (described in specs/, implemented in src/)

### Architectural Principle: Specs vs Implementation

**IMPORTANT**: This documentation emphasizes the separation between:

- **Specifications** (in `specs/`): Business-readable Gherkin describing WHAT
- **Implementations** (in `src/`): Technical Go code describing HOW

```text
specs/cli/login/specification.feature     ← Business reviews this
    Rule: User must provide valid credentials
        Scenario: Valid login succeeds

src/cli/tests/steps_test.go              ← Developers implement this
    func userProvidesCredentials() { ... }
    func loginSucceeds() { ... }
```

---

## What is ATDD?

**Acceptance Test-Driven Development (ATDD)** is a collaborative approach where business stakeholders, developers, and testers define acceptance criteria **before** development begins. It focuses on capturing business value and measurable success criteria from the customer's perspective.

### Core Purpose

ATDD answers the question: **"What does 'done' mean for this feature?"**

By defining acceptance criteria upfront, all stakeholders agree on:

- What business value the feature delivers
- How success will be measured
- What conditions must be satisfied for acceptance

### Why Use ATDD?

- **Business Alignment**
  - **Problem:** Developers build features that don't meet business needs
  - **Solution:** ATDD ensures everyone agrees on requirements before coding starts
- **Reduced Rework**
  - **Problem:** Discovering missing requirements after implementation
  - **Solution:** ATDD catches misunderstandings early through collaborative discussion
- **Measurable Success**
  - **Problem:** Subjective acceptance ("Does this look good?")
  - **Solution:** ATDD requires measurable criteria (e.g., "Creates 3 directories", "Completes in <2s")
- **Stakeholder Collaboration**
  - **Problem:** Product owners can't review technical test code
  - **Solution:** ATDD uses natural language (Gherkin) that stakeholders can read and validate

---

## What is BDD?

**Behavior-Driven Development (BDD)** is a specification technique that describes user-facing behavior through concrete examples. BDD focuses on **observable behavior** - what users can see and interact with, not internal implementation details.

### Core Purpose

BDD answers the question: **"How does the system behave from the user's perspective?"**

By writing scenarios in natural language (Given/When/Then), teams create:

- Shared understanding of expected behavior
- Executable specifications that become automated tests
- Living documentation that stays synchronized with code

### Why Use BDD?

- **Common Language**
  - **Problem:** Developers, testers, and product owners speak different languages
  - **Solution:** BDD uses Gherkin (Given/When/Then), which is readable by all stakeholders
- **Focus on Behavior**
  - **Problem:** Tests focus on implementation details that change frequently
  - **Solution:** BDD tests describe **what the system does**, not **how it does it**
- **Living Documentation**
  - **Problem:** Documentation becomes outdated
  - **Solution:** BDD scenarios are executable — if they pass, the documentation is accurate

---

## Gherkin and the Ubiquitous Language

BDD scenarios are most effective when written using the **Ubiquitous Language** from Domain-Driven Design. The shared vocabulary that both business and technical teams understand.

### Why This Matters

**Using technical language** (harder for business to validate):

```gherkin
Given the database record exists
When the API endpoint is called
Then the response code should be 200
```

**Using Ubiquitous Language** (clear to all stakeholders):

```gherkin
Given an order awaiting approval
When the manager approves the order
Then the order status should be "Approved"
```

The second version uses domain terms that:

- Business stakeholders recognize and can validate
- Developers implement using the domain model
- QA tests reflect actual business rules

**Best practice**: Before writing specifications, participate in Event Storming and Example Mapping workshops to establish the shared language.

See: [Ubiquitous Language](./ubiquitous-language.md) for DDD foundation and [Event Storming](./event-storming.md) for domain discovery workshops

---

## Requirements Discovery with Example Mapping

Requirements for features are discovered through **Example Mapping**, a collaborative workshop technique that uses colored index cards to surface acceptance criteria (Blue cards → Rules) and concrete examples (Green cards → Scenarios).

### Card to Gherkin Mapping

| Card Color | Represents | Maps To | Location |
|-----------|------------|---------|----------|
| 🟡 **Yellow** | User Story | Feature description | `specs/` |
| 🔵 **Blue** | Acceptance Criteria | `Rule:` blocks (ATDD) | `specs/` |
| 🟢 **Green** | Concrete Examples | `Scenario:` blocks (BDD) | `specs/` |
| 🔴 **Red** | Questions/Unknowns | issues.md | `specs/` |
| N/A | Step Implementation | Go functions | `src/` |

**Workshop Format**:

- 15-25 minutes, time-boxed
- Collaborative: Product Owner + Developer + Tester
- Produces cards that map directly to Gherkin elements
- Surfaces questions and risks early

**Result**: A feature is ready to implement when you have:

- 1 Yellow Card (user story)
- 2-6 Blue Cards (acceptance criteria)
- 2-4 Green Cards per Blue Card (examples)
- Few or no Red Cards (questions resolved)

**See**: [Example Mapping Guide](./example-mapping.md) for complete workshop process, best practices, and detailed examples.

---

## From Discovery to Implementation

The complete workflow proceeds through these phases:

1. **Discovery**: Example Mapping workshop → produces colored cards
2. **Specification**: Convert cards → Gherkin in `specs/`
3. **Implementation**: Write step definitions in `src/` → implement features
4. **Validation**: All scenarios pass → feature complete

**Key Points**:

- **Before development**: Run Example Mapping, write specifications in `specs/`
- **During development**: Implement steps in `src/`, write unit tests, implement features
- **After development**: All scenarios pass = acceptance criteria met

**See**: [Three-Layer Approach](./three-layer-approach.md) for detailed workflow showing how ATDD, BDD, and TDD integrate throughout the development lifecycle.

---

## Key Principles

### Measurable Acceptance Criteria

**Bad** (subjective):

```gherkin
Rule: The interface is user-friendly
Rule: Performance is good
Rule: Error messages are helpful
```

**Good** (measurable):

```gherkin
Rule: Creates 3 directories (src/, tests/, docs/)
Rule: Command completes in under 2 seconds
Rule: Error message contains "already initialized" text
```

### Collaboration Before Code

ATDD and BDD are **collaborative** - they require:

- Product Owner (business perspective)
- Developer (technical perspective)
- Tester (quality perspective)

**Don't**: Have developers write specifications alone
**Do**: Run Example Mapping workshop with all roles present

### Acceptance Criteria Drive Development

Acceptance criteria define "done":

- Development starts when criteria are clear
- Development ends when all criteria pass
- No "scope creep" mid-implementation

### Living Documentation

Gherkin specifications serve as:

- **Requirements documentation** (what the feature does)
- **Automated tests** (validation that it works)
- **Audit trail** (proof of testing for compliance)

All in one place, always up to date.

### Behavior Over Implementation

Focus on **what** the system does, not **how** it does it:

**Bad**:

```gherkin
When the ConfigManager loads the file
And the YAML parser deserializes the content
Then the Config struct is populated
```

**Good**:

```gherkin
When I run "r2r init"
Then a file named "r2r.yaml" should be created
And the configuration should contain default values
```

---

## Related Documentation

- [Three-Layer Testing Approach](./three-layer-approach.md) - How ATDD/BDD/TDD work together
- [Ubiquitous Language](./ubiquitous-language.md) - DDD and shared vocabulary foundation
- [Event Storming](./event-storming.md) - Domain discovery workshops
- [Example Mapping](./example-mapping.md) - Requirements discovery workshops
- [Gherkin Format Reference](../../reference/specifications/gherkin-format.md) - Detailed syntax guide
- [Create Feature Spec](../../how-to-guides/specifications/create-specifications.md) - Step-by-step guide
