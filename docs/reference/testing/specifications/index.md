# Testing Strategy Overview

This project uses a **layered testing approach** that separates concerns across three complementary testing methodologies:

- **[ATDD (Acceptance Test-Driven Development)](./atdd.md)** - Business requirements and customer value
- **[BDD (Behavior-Driven Development)](./bdd.md)** - User-facing behavior specifications
- **[TDD (Test-Driven Development)](./tdd.md)** - Implementation correctness and code quality

## Quick Reference

| Layer             | Focus               | Format                             | Location          | Who                  |
|-------------------|---------------------|------------------------------------|-------------------|----------------------|
| [ATDD](./atdd.md) | Business value      | User stories + acceptance criteria | `.feature` files  | Product Owner + Team |
| [BDD](./bdd.md)   | Observable behavior | Gherkin scenarios                  | `.feature` files  | QA + Developers      |
| [TDD](./tdd.md)   | Code correctness    | Unit tests                         | Test files        | Developers           |

## Core Principle

**Feature files contain requirements and behavior (ATDD/BDD). Unit tests contain implementation (TDD).**

Traceability is maintained through feature file names:

- Feature file: `requirements/cli/init_project.feature`
- Unit test comment: `Feature: init_project` (in a comment)

## Example Mapping

**[Example Mapping](https://cucumber.io/blog/bdd/example-mapping-introduction/)** is a collaborative workshop technique used during ATDD to discover requirements and generate BDD scenarios.

### Four Card Colors

| Color  | Represents           | Becomes           | Count per Story |
|--------|----------------------|-------------------|-----------------|
| Yellow | User Story           | ATDD user story   | 1               |
| Blue   | Rules/Criteria       | ATDD acceptance criteria | 2-6  |
| Green  | Concrete Examples    | BDD scenarios     | 2-4 per rule    |
| Red    | Questions/Blockers   | Follow-up items   | Resolve before coding |

### Workshop Flow

```text
1. Place Yellow Card (user story) at top
2. Discuss and create Blue Cards (rules)
3. For each Blue Card, create Green Cards (examples)
4. Capture uncertainties as Red Cards
5. Assess readiness: Ready / Too Large / Too Uncertain
```

### Benefits

- **Collaborative**: Product Owner, Developer, Tester work together
- **Time-boxed**: 15-25 minutes prevents over-analysis
- **Visual**: Physical/digital cards make structure clear
- **Outcome-driven**: Green Cards become executable BDD scenarios

**See**: [ATDD Guide - Example Mapping Workshop](./atdd.md#example-mapping-workshop-collaborative-discovery) for complete details and templates.

## Project Structure

```text
requirements/
+-- cli/                               # CLI module requirements
|   +-- init_project.feature           # ATDD + BDD layers
|   +-- deploy_module.feature
+-- vscode/                            # VS Code extension requirements
|   +-- commit_button.feature
+-- docs/                              # Documentation requirements
|   +-- build_docs.feature
+-- mcp/                               # MCP server requirements
    +-- server_startup.feature

src/**/test files                       # Unit tests with feature references
contracts/testing/0.1.0/
+-- specifications.yml                 # Contract definitions
+-- taxonomy.yml                       # Test level classifications
```

## Decision Tree for AI Assistants

### 1. User asks to CREATE a new feature specification

**Action**: Generate a two-layer feature file

**Template**:

```gherkin
@cli @critical
Feature: [Feature Name]

  # ATDD Layer: Business value
  As a [user role]
  I want [capability]
  So that [business value]

  Acceptance Criteria:
  - [ ] [Measurable criterion 1]
  - [ ] [Measurable criterion 2]

  # BDD Layer: User-facing behavior
  @success
  Scenario: [Happy path scenario name]
    Given [precondition]
    When [action]
    Then [observable outcome]

  @error
  Scenario: [Error scenario name]
    Given [precondition]
    When [invalid action]
    Then [error behavior]
```

**File Storage**:

- Store in `requirements/<module-name>/`
- Module folders: `cli/`, `vscode/`, `docs/`, `mcp/`
- Naming: `<feature-description>.feature`

**Unit Test Linking** (language-agnostic):

```
# Comment in test file linking to feature:
# Feature: init_project

# Examples by language:
# Go:         // Feature: init_project
# C#:         // Feature: init_project
# Python:     # Feature: init_project
# TypeScript: // Feature: init_project
```

### 2. User asks to UPDATE an existing feature file

**Action**:

- Read the existing `.feature` file
- Identify which layer needs updating (ATDD/BDD)
- Preserve existing content in other layers
- Maintain tag consistency per tagging taxonomy

### 3. User asks to VALIDATE feature files

**Action**:

- Check for both layers (ATDD, BDD)
- Verify tags match taxonomy
- Ensure acceptance criteria are measurable
- Confirm scenarios use Given/When/Then format
- Verify file name is descriptive

### 4. User asks to IMPLEMENT a feature

**Workflow**:

1. Read the corresponding `.feature` file
2. Extract acceptance criteria (ATDD) -> business goals
3. Extract scenarios (BDD) -> user-facing behavior
4. Note the feature file name for traceability
5. Write unit tests FIRST, adding feature reference comment (e.g., `Feature: init_project`)
6. Implement to pass unit tests
7. Verify BDD scenarios pass
8. Confirm ATDD acceptance criteria are met

## Layer Selection Guide

| User Request                 | Layer to Use                        | Why                                              |
|------------------------------|-------------------------------------|--------------------------------------------------|
| "Add a feature for X"        | [ATDD](./atdd.md) -> [BDD](./bdd.md) | Start with business value, then specify behavior |
| "Write a scenario for Y"     | [BDD](./bdd.md) only                | User explicitly wants behavioral spec            |
| "What tests do I need?"      | Check unit tests ([TDD](./tdd.md))  | Implementation testing is in code                |
| "Define acceptance criteria" | [ATDD](./atdd.md) only              | Business stakeholder perspective                 |
| "Create complete spec"       | [ATDD](./atdd.md) + [BDD](./bdd.md) | Comprehensive specification                      |

## When to Use Each Layer

### Use [ATDD](./atdd.md) When

- Defining new features with business stakeholders
- Running Example Mapping workshops for requirement discovery
- Creating user stories for sprint planning
- Establishing definition of "done"
- Features require customer sign-off
- Need to communicate business value

### Use [BDD](./bdd.md) When

- Specifying CLI command behavior
- Documenting user-facing interactions
- Creating executable specifications
- Collaborating across teams (dev/QA/product)
- Need to verify user experience

### Use [TDD](./tdd.md) When

- Implementing complex internal algorithms
- Building MCP server logic
- Creating utility functions
- Refactoring existing code safely
- Need to ensure code correctness

## Integrated Workflow

```text
1. ATDD: Define user story and run Example Mapping workshop
         Yellow Card -> User story
         Blue Cards -> Acceptance criteria
         Green Cards -> Concrete examples
         Red Cards -> Questions to resolve
         [See ATDD Workflow ->](./atdd.md#workflow)

2. BDD:  Convert Green Cards to Gherkin scenarios
         Add appropriate tags (@cli, @error, @success, etc.)
         One scenario per Green Card
         [See BDD Workflow ->](./bdd.md#workflow)

3. TDD:  Implement with unit tests for internal logic
         Reference feature name in test comments
         [See TDD Workflow ->](./tdd.md#workflow)

4. BDD:  Verify scenarios pass (acceptance tests)

5. ATDD: Validate acceptance criteria met (Blue Cards checked)

6. Trace: Verify feature name links properly
```

## Complete Example

**Feature File** (`requirements/cli/init_project.feature`):

```gherkin
@cli @flag @critical
Feature: Initialize a new project

  # ATDD Layer: Business context
  As a developer
  I want to initialize a CLI project with a single command
  So that I can quickly start development with proper structure

  Acceptance Criteria:
  - [ ] Creates project directory structure
  - [ ] Generates valid configuration file
  - [ ] Exits with clear success/error messages
  - [ ] Handles existing projects gracefully

  # BDD Layer: User-facing behavior
  @success
  Scenario: Initialize in current directory
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And the file should contain valid YAML
    And the command should exit with code 0

  @error
  Scenario: Initialize in non-empty directory
    Given I am in a directory containing files
    When I run "cc init"
    Then the command should fail
    And stderr should contain "Directory must be empty"
```

**Unit Test** (pseudocode - see [TDD guide](./tdd.md) for language-specific examples):

```
COMMENT: Feature: init_project

TEST init_in_empty_directory:
    ARRANGE:
        temp_dir = create_temporary_directory()

    ACT:
        result = init_project(temp_dir)

    ASSERT:
        file_exists(temp_dir + "/cc.yaml") == true
        result.exit_code == 0
```

## Detailed Guides

- **[ATDD Guide](./atdd.md)** - Learn how to write effective acceptance criteria and user stories
- **[BDD Guide](./bdd.md)** - Master Gherkin scenarios and tagging strategies
- **[TDD Guide](./tdd.md)** - Understand unit testing and the Red-Green-Refactor cycle

## Traceability Model

### From Feature to Tests

Search codebase for `Feature: init_project` (in comments) to find all related unit tests.

### From Tests to Features

Look at test comment to find feature file: `requirements/cli/init_project.feature`

### Coverage Verification

All feature files should have at least one unit test referencing them.

### Module Organization

Feature files are organized by module (cli, vscode, docs, mcp) for easy navigation.

## Contract Integration

This testing strategy aligns with versioned contracts:

- `contracts/testing/0.1.0/specifications.yml` - Defines tdd/bdd/atdd classifications
- `contracts/testing/0.1.0/taxonomy.yml` - Defines test levels (l0-l4, horizontal-e2e)

## Migration from Legacy Files

Many existing `.feature` files may contain only BDD scenarios. See the migration strategies in:

- [ATDD Migration Guide](./atdd.md#migration)
- [BDD Migration Guide](./bdd.md#migration)

---

**Next Steps**: Read the detailed guides for [ATDD](./atdd.md), [BDD](./bdd.md), and [TDD](./tdd.md) to master each testing layer.
