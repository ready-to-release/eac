# Three-Layer Testing Approach

This project uses **a three layered approach to bind specifications to tests**:

- **Acceptance Criteria** Business requirements and customer value
- **Scenario** User-facing behavior specifications
- **Unit tests** Implementation correctness and code quality

Each layer serves a distinct purpose, uses different tools, and addresses different stakeholders' needs.

---

## The Three Layers

| Layer                  | Question | Stakeholders | Tool | Representation | Location |
|------------------------|----------|--------------|------|----------------|----------|
| **Acceptace Criteria** | "What business value?" | Product Owner, Business | Godog | `Rule:` blocks | `specs/` |
| **Scenario**           | "How does user interact?" | QA, Developers, Product | Godog | `Scenario:` under Rules | `specs/` + `src/tests/` |
| **Unit test**          | "Does code work?" | Developers | Go test | Unit test functions | `src/*_test.go` |

### Layer 1: Acceptance Criteria

**Purpose**: Define business requirements before development

**Format**: `Rule:` blocks in Gherkin

```gherkin
Feature: Init CLI project

  As a developer
  I want to initialize a CLI project
  So that I can quickly start development

  Rule: Creates project directory structure

  Rule: Generates valid configuration file

  Rule: Command completes in under 2 seconds
```

**Origin**: 🔵 Blue cards from Example Mapping
**Location**: `specs/<module>/<command or function>/<specification-area>.feature`

### Layer 2: Scenario

**Purpose**: Specify observable behavior through concrete examples

**Format**: `Scenario:` blocks nested under `Rule:` blocks

```gherkin
Rule: Creates project directory structure

  @ov
  Scenario: Initialize in empty directory
    Given I am in an empty folder
    When I run "simply init"
    Then a file named "simply.yaml" should be created
    And a directory named "src/" should exist

  @ov
  Scenario: Initialize in existing project
    Given I am in a directory with "simply.yaml"
    When I run "simply init"
    Then the command should fail
    And stderr should contain "already initialized"
```

**Origin**: 🟢 Green cards from Example Mapping
**Location**: `specs/<module>/<command or function>/<specification-area>.feature`
**Implementation**: `src/<module>/tests/steps_test.go`

### Layer 3: Unit Tests

**Purpose**: Ensure code correctness through systematic test-first development

**Format**: Unit tests in `*_test.go` files

**Tool**: Go test framework
