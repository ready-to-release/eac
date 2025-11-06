# Three-Layer Testing Approach

Understanding how ATDD, BDD, and TDD work together to deliver quality software.

---

## Overview

This project uses a **three-layer testing approach** that separates concerns across three complementary testing methodologies:

- **ATDD** (Acceptance Test-Driven Development) - Business requirements and customer value
- **BDD** (Behavior-Driven Development) - User-facing behavior specifications
- **TDD** (Test-Driven Development) - Implementation correctness and code quality

Each layer serves a distinct purpose, uses different tools, and addresses different stakeholders' needs.

---

## Why Three Layers?

### Separation of Concerns

Each layer focuses on a specific aspect of quality:

| Layer | Answers | Stakeholders | Perspective |
|-------|---------|--------------|-------------|
| **ATDD** | "What business value does this deliver?" | Product Owner, Business Stakeholders | Business perspective |
| **BDD** | "How does the user interact with this?" | QA, Developers, Product Owner | User perspective |
| **TDD** | "Does the code work correctly?" | Developers | Implementation perspective |

### Different Questions, Unified Format

**ATDD asks**: "Are we building the right thing?"

- Focus: Business requirements, acceptance criteria
- Representation: `Rule:` blocks in Gherkin
- Tool: Godog
- Location: `specs/<module>/<feature>/specification.feature`
- Output: Validated business value

**BDD asks**: "Does it behave as expected?"

- Focus: Observable user-facing behavior
- Representation: `Scenario:` blocks under Rules
- Tool: Godog
- Location: `specs/<module>/<feature>/specification.feature` (same file as ATDD)
- Implementation: `src/<module>/tests/steps_test.go`
- Output: Executable living documentation

**TDD asks**: "Is the implementation correct?"

- Focus: Code correctness, internal logic
- Tool: Go test framework
- Location: `src/<module>/*_test.go`
- Output: Verified implementation quality

---

## The Three Layers

### Layer 1: ATDD (Acceptance Test-Driven Development)

**Purpose**: Define and validate business requirements before development begins.

**Representation**: `Rule:` blocks in Gherkin

**Example**:

```gherkin
@cli @critical @init
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project with a single command
  So that I can quickly start development with proper structure

  Rule: Creates project directory structure

  Rule: Generates valid configuration file

  Rule: Command completes in under 2 seconds
```

**Architectural placement**:

- **Specification**: `specs/<module>/<feature>/specification.feature` (business-readable WHAT)
- **Tool**: Godog (executes Rules through their nested Scenarios)
- **Origin**: Blue cards from Example Mapping workshops
- **Stakeholders**: Product Owner, Business, QA

**See**: [ATDD and BDD with Gherkin](./atdd-bdd-with-gherkin.md) for detailed explanation of ATDD concepts.

### Layer 2: BDD (Behavior-Driven Development)

**Purpose**: Specify observable user-facing behavior through concrete examples.

**Representation**: `Scenario:` blocks nested under `Rule:` blocks

**Example**:

```gherkin
Rule: Creates project directory structure

  @success @ac1
  Scenario: Initialize in empty directory
    Given I am in an empty folder
    When I run "rr init"
    Then a file named "rr.yaml" should be created
    And a directory named "src/" should exist
    And the command should exit with code 0

  @error @ac1
  Scenario: Initialize in existing project shows error
    Given I am in a directory with "rr.yaml"
    When I run "rr init"
    Then the command should fail
    And stderr should contain "already initialized"
```

**Architectural placement**:

- **Specification**: `specs/<module>/<feature>/specification.feature` (same file as ATDD Rules)
- **Implementation**: `src/<module>/tests/steps_test.go` (separate location - technical HOW)
- **Tool**: Godog (executes scenarios through Go step definitions)
- **Origin**: Green cards from Example Mapping workshops
- **Stakeholders**: QA, Developers, Product Owner

**See**: [ATDD and BDD with Gherkin](./atdd-bdd-with-gherkin.md) for detailed explanation of BDD concepts.

### Layer 3: TDD (Test-Driven Development)

**Purpose**: Ensure code correctness and enable safe refactoring

**Format**: Unit tests in `*_test.go` files

**Tool**: Go test framework

**Who writes it**: Developers

**Content**:

- Unit tests for functions and methods
- Edge case and boundary testing
- Error handling validation
- Internal logic verification

**Example**:

```go
// Feature: cli_init_project
func TestCreateConfig(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "r2r.yaml")

    // Act
    err := CreateConfig(configPath)

    // Assert
    if err != nil {
        t.Fatalf("CreateConfig failed: %v", err)
    }
}
```

**Key Characteristics**:

- Written before implementation (Red-Green-Refactor)
- Tests internal implementation details
- Supports refactoring with confidence
- Executed frequently during development

---

## How the Layers Interact

### The Flow: From Requirements to Code

```text
1. Business Discussion
   ↓
2. Example Mapping Workshop
   ↓ (produces colored cards)
3. Create specification.feature in specs/
   ├─ Yellow Card → Feature description
   ├─ Blue Cards → Rule blocks (ATDD Layer)
   └─ Green Cards → Scenario blocks under Rules (BDD Layer)
   ↓
4. Create step definitions in src/
   └─ Implement Go functions for scenarios
   ↓
5. TDD Layer (unit tests in src/)
   └─ Implementation testing
   ↓
6. Implementation Code (in src/)
```

### From Discovery to Specification

Requirements are discovered through collaborative workshops that establish shared vocabulary and specific requirements:

**1. Establish Ubiquitous Language**:

- Build shared domain vocabulary through Domain-Driven Design
- Ensure business and technical teams speak the same language
- **See**: [Ubiquitous Language](./ubiquitous-language.md) for DDD foundation

**2. Event Storming** - Discover domain vocabulary:

- Collaborative workshop using sticky notes to map business events
- Surfaces domain events, actors, commands, and policies
- Most importantly: discovers the Ubiquitous Language itself
- **See**: [Event Storming](./event-storming.md) for workshop guide

**3. Example Mapping** - Apply vocabulary to features:

- Time-boxed workshops using colored cards
- Produces cards that map directly to the three layers:
  - 🟡 **Yellow Card** (User Story) → Feature description
  - 🔵 **Blue Cards** (Acceptance Criteria) → `Rule:` blocks (ATDD Layer)
  - 🟢 **Green Cards** (Concrete Examples) → `Scenario:` blocks (BDD Layer)
  - 🔴 **Red Cards** (Questions) → issues.md
- **See**: [Example Mapping](./example-mapping.md) for workshop process

### Traceability Across Layers

All layers are linked through **Feature ID**:

```text
Feature ID: cli_init-project

Links across specs/ and src/:
  ├─ specs/cli/init-project/specification.feature
  │    # Feature ID: cli_init-project
  │    (Contains both ATDD Rules and BDD Scenarios)
  │
  ├─ src/cli/tests/steps_test.go
  │    // Feature: cli_init-project
  │    (Implements steps for specification)
  │
  └─ src/cli/*_test.go
       // Feature: cli_init-project
       (Unit tests)
```

### Acceptance Criteria to Scenario Linkage

```text
specification.feature (in specs/):
  Rule: Creates project directory structure    ← ATDD Layer
    |
    +-- Scenarios nested under Rule
          |
          v
    @success @ac1                               ← BDD Layer
    Scenario: Initialize in empty directory
      Given I am in an empty folder
      When I run "r2r init"
      Then...

Implemented by (in src/):
  steps_test.go:
    func iAmInAnEmptyFolder() { ... }
    func iRun(command string) { ... }
    func... (various step definitions)
```
