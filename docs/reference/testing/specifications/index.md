# Testing Strategy Overview

This project uses a **layered testing approach** that separates concerns across three complementary testing methodologies:

- **[ATDD (Acceptance Test-Driven Development)](./atdd.md)** - Business requirements and customer value
- **[BDD (Behavior-Driven Development)](./bdd.md)** - User-facing behavior specifications
- **[TDD (Test-Driven Development)](./tdd.md)** - Implementation correctness and code quality

## Quick Reference

| Layer             | Focus               | Format                  | Tool         | Location                   | Who                  |
|-------------------|---------------------|-------------------------|--------------|----------------------------|----------------------|
| [ATDD](./atdd.md) | Business value      | Gauge specs (markdown)  | Gauge        | `acceptance.spec` files    | Product Owner + Team |
| [BDD](./bdd.md)   | Observable behavior | Gherkin scenarios       | Godog        | `behavior.feature` files   | QA + Developers      |
| [TDD](./tdd.md)   | Code correctness    | Unit tests              | Go test/etc  | Test files                 | Developers           |

---

## Core Principle

**Separate files for separate concerns:**

- **acceptance.spec** (Gauge) - Contains business requirements (ATDD)
- **behavior.feature** (Godog) - Contains executable scenarios (BDD)
- **Unit tests** - Contains implementation validation (TDD)

Traceability is maintained through **Feature ID**:

```text
Feature ID: cli_init_project

Used in:
- acceptance.spec (Gauge ATDD)
- behavior.feature (Godog BDD)
- acceptance_test.go (Gauge step implementations)
- step_definitions_test.go (Godog step definitions)
- Unit tests (// Feature: cli_init_project)
```

## Testing Tools

| Layer | Tool | File Format | Purpose |
|-------|------|-------------|---------|
| ATDD | [Gauge](https://gauge.org/) | `.spec` (markdown) | Execute acceptance criteria validation |
| BDD | [Godog](https://github.com/cucumber/godog) | `.feature` (Gherkin) | Execute behavioral scenarios |
| TDD | Go test / pytest / etc | Unit test files | Verify implementation correctness |

**Installation**:

```bash
# Gauge
go install github.com/getgauge/gauge@latest
gauge install go

# Godog
go get github.com/cucumber/godog/cmd/godog@latest
```

---

## Example Mapping Workflow

**[Example Mapping](https://cucumber.io/blog/bdd/example-mapping-introduction/)** is a collaborative workshop technique that produces colored cards which map to our test files.

### Card to File Mapping

| Card Color | Contains | → | Output File | Tool |
|-----------|----------|---|-------------|------|
| 🟡 **Yellow** (1) | User Story | → | `acceptance.spec` | Gauge |
| 🔵 **Blue** (2-6) | Rules/Criteria | → | `acceptance.spec` | Gauge |
| 🟢 **Green** (2-4 per Blue) | Concrete Examples | → | `behavior.feature` | Godog |
| 🔴 **Red** (0-N) | Questions | → | `issues.md` | Tracking |

### Workshop to Execution Flow

```text
1. Example Mapping Workshop (15-25 min)
   +-- Yellow Card -> User story
   +-- Blue Cards -> Acceptance criteria
   +-- Green Cards -> Concrete examples
   +-- Red Cards -> Questions/blockers

2. Convert Cards to Files
   +-- Yellow + Blue -> acceptance.spec (Gauge)
   +-- Green -> behavior.feature (Godog)
   +-- Red -> issues.md (track separately)

3. Implement Test Steps
   +-- acceptance_test.go (Gauge steps)
   +-- step_definitions_test.go (Godog steps)

4. Execute Tests
   +-- gauge run requirements/<module>/<feature>/
   +-- godog run requirements/<module>/<feature>/behavior.feature

5. Implement Feature Code (TDD)
   +-- Write unit tests with Feature ID comment
   +-- Implement to pass tests

6. Verify All Layers
   +-- Gauge tests pass (ATDD)
   +-- Godog tests pass (BDD)
   +-- Unit tests pass (TDD)
```

**See**: [ATDD Guide - Example Mapping Workshop](./atdd.md#example-mapping-workshop-collaborative-discovery) for complete workshop details.

---

## Project Structure

```text
requirements/
+-- cli/                                    # CLI module
|   +-- init_project/                       # Feature directory
|       +-- acceptance.spec                 # Gauge (ATDD)
|       +-- behavior.feature                # Godog (BDD)
|       +-- acceptance_test.go              # Gauge step implementations
|       +-- step_definitions_test.go        # Godog step definitions
|       +-- issues.md                       # Red cards (optional)
|   +-- deploy_module/
|       +-- acceptance.spec
|       +-- behavior.feature
|       +-- acceptance_test.go
|       +-- step_definitions_test.go
+-- vscode/                                 # VS Code extension module
|   +-- commit_button/
|       +-- acceptance.spec
|       +-- behavior.feature
|       +-- acceptance_test.go
|       +-- step_definitions_test.go
+-- docs/                                   # Documentation module
+-- mcp/                                    # MCP server module

src/**/                                     # Implementation code
tests/**/                                   # TDD unit tests with Feature ID
contracts/testing/0.1.0/
+-- specifications.yml                      # Contract definitions
+-- taxonomy.yml                            # Test level classifications
```

---

## Decision Tree for AI Assistants

### 1. User asks to CREATE a new feature specification

**Action**: Generate both acceptance.spec and behavior.feature

**Step 1**: Create feature directory

```bash
mkdir -p requirements/<module>/<feature_name>
```

**Step 2**: Create acceptance.spec (Gauge)

**File**: `requirements/<module>/<feature_name>/acceptance.spec`

```markdown
# [Feature Name]

> **Feature ID**: <module>_<feature_name>
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: <Module>
> **Tags**: <tags>

## User Story

* As a [role]
* I want [capability]
* So that [business value]

## Acceptance Criteria

* [Measurable criterion 1]
* [Measurable criterion 2]
* [Measurable criterion 3]

## Acceptance Tests

### AC1: [Criterion 1]
**Validated by**: behavior.feature -> @ac1 scenarios

* [Gauge step 1]
* [Gauge step 2]
* [Gauge step 3]
```

**Step 3**: Create behavior.feature (Godog)

**File**: `requirements/<module>/<feature_name>/behavior.feature`

```gherkin
# Feature ID: <module>_<feature_name>
# Acceptance Spec: acceptance.spec
# Module: <Module>

@<module> @critical @<feature_name>
Feature: [Feature Name]

  @success @ac1
  Scenario: [Happy path scenario name]
    Given [precondition]
    When [action]
    Then [observable outcome]

  @error @ac1
  Scenario: [Error scenario name]
    Given [precondition]
    When [invalid action]
    Then [error behavior]
```

**Step 4**: Create step implementation files

- `acceptance_test.go` - Gauge step implementations
- `step_definitions_test.go` - Godog step definitions

### 2. User asks to UPDATE an existing feature

**Action**:

- **If updating business requirements**: Edit `acceptance.spec`
- **If updating behavior/scenarios**: Edit `behavior.feature`
- **Always**: Maintain Feature ID consistency across both files
- **Always**: Update linkage tags (@ac1, @ac2, etc.)

### 3. User asks to VALIDATE feature files

**Action**:

- Check both `acceptance.spec` and `behavior.feature` exist
- Verify Feature ID is identical in both files
- Ensure tags are consistent
- Confirm acceptance criteria are measurable
- Verify scenarios use Given/When/Then format
- Check @ac tags link scenarios to criteria

### 4. User asks to IMPLEMENT a feature

**Workflow**:

1. Read both `acceptance.spec` and `behavior.feature`
2. Extract acceptance criteria (ATDD) from acceptance.spec
3. Extract scenarios (BDD) from behavior.feature
4. Note the Feature ID for traceability
5. Create Gauge step implementations in `acceptance_test.go`
6. Create Godog step definitions in `step_definitions_test.go`
7. Write unit tests with Feature ID comment (e.g., `// Feature: cli_init_project`)
8. Implement feature code to pass all tests
9. Run Gauge: `gauge run requirements/<module>/<feature>/`
10. Run Godog: `godog run requirements/<module>/<feature>/behavior.feature`
11. Run unit tests: `go test ./...`
12. Verify all tests pass

## Layer Selection Guide

| User Request                 | File to Use                          | Why                                              |
|------------------------------|--------------------------------------|--------------------------------------------------|
| "Add a feature for X"        | Create both [acceptance.spec](./atdd.md) and [behavior.feature](./bdd.md) | Need both business value and behavior |
| "Write acceptance criteria"  | [acceptance.spec](./atdd.md) only    | Business stakeholder perspective                 |
| "Write a scenario for Y"     | [behavior.feature](./bdd.md) only    | User explicitly wants behavioral spec            |
| "What tests do I need?"      | Check unit tests ([TDD](./tdd.md))   | Implementation testing is in code                |
| "Run acceptance tests"       | `gauge run` on acceptance.spec       | Execute ATDD validation                          |
| "Run behavior tests"         | `godog run` on behavior.feature      | Execute BDD scenarios                            |

## When to Use Each Layer

### Use [ATDD](./atdd.md) (acceptance.spec) When

- Defining new features with business stakeholders
- Running Example Mapping workshops (Yellow + Blue cards)
- Creating user stories for sprint planning
- Establishing definition of "done"
- Features require customer sign-off
- Need to validate business value with Gauge tests

### Use [BDD](./bdd.md) (behavior.feature) When

- Converting Green Cards from Example Mapping
- Specifying CLI command behavior
- Documenting user-facing interactions
- Creating executable Godog scenarios
- Collaborating across teams (dev/QA/product)
- Need to verify user experience

### Use [TDD](./tdd.md) (unit tests) When

- Implementing complex internal algorithms
- Building MCP server logic
- Creating utility functions
- Refactoring existing code safely
- Need to ensure code correctness
- Testing internal implementation details

---

## Complete Example

### Example Mapping Workshop Output

```text
[YELLOW] As a developer, I want to initialize a CLI project with one command,
         so that I can quickly start development

[BLUE-1] Creates project directory structure
  [GREEN-1a] Empty folder -> init -> creates src/, tests/, docs/
  [GREEN-1b] Existing project -> init -> error "already initialized"

[BLUE-2] Generates valid configuration file
  [GREEN-2a] New project -> init -> creates cc.yaml with defaults
  [GREEN-2b] With --name flag -> cc.yaml contains custom name

[RED-1] What if cc.yaml already exists?
```

### Converted to Files

#### acceptance.spec (Gauge)

**File**: `requirements/cli/init_project/acceptance.spec`

```markdown
# Initialize Project

> **Feature ID**: cli_init_project
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: CLI
> **Tags**: cli, critical

## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure

## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file
* Handles errors gracefully
* Command completes in under 2 seconds

## Acceptance Tests

### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

* Create empty test directory
* Run "cc init" command
* Verify "cc.yaml" file exists
* Verify "src/" directory exists
* Verify "tests/" directory exists
* Verify "docs/" directory exists

### AC2: Generates valid configuration file
**Validated by**: behavior.feature -> @ac2 scenarios

* Create empty test directory
* Run "cc init" command
* Read "cc.yaml" file contents
* Verify YAML is valid
* Verify default values are present
* Verify file has correct permissions
```

#### behavior.feature (Godog)

**File**: `requirements/cli/init_project/behavior.feature`

```gherkin
# Feature ID: cli_init_project
# Acceptance Spec: acceptance.spec
# Module: CLI

@cli @critical @init_project
Feature: Initialize project command behavior

  Background:
    Given I am in a clean test environment

  # Green Card 1a: Empty folder -> init -> creates dirs
  @success @ac1
  Scenario: Initialize in empty directory creates structure
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And a directory named "src/" should exist
    And a directory named "tests/" should exist
    And a directory named "docs/" should exist
    And the command should exit with code 0

  # Green Card 1b: Existing project -> init -> error
  @error @ac1
  Scenario: Initialize in existing project shows error
    Given I am in a directory with "cc.yaml"
    When I run "cc init"
    Then the command should fail
    And stderr should contain "already initialized"
    And the command should exit with code 1

  # Green Card 2a: New project -> creates cc.yaml with defaults
  @success @ac2
  Scenario: Initialize creates configuration with defaults
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And the file should contain valid YAML
    And the file should contain "version: 1.0.0"
    And the command should exit with code 0

  # Green Card 2b: With --name flag -> contains custom name
  @flag @success @ac2
  Scenario: Initialize with custom name flag
    Given I am in an empty folder
    When I run "cc init --name my-project"
    Then a file named "cc.yaml" should be created
    And the file should contain "name: my-project"
    And the command should exit with code 0
```

#### Unit Test (TDD)

**File**: `src/cli/init_test.go` (example)

```go
// Feature: cli_init_project
package cli_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestInitProject_CreatesConfigFile(t *testing.T) {
    // Arrange
    tempDir := createTempDirectory(t)
    defer cleanup(tempDir)

    // Act
    err := InitProject(tempDir)

    // Assert
    assert.NoError(t, err)
    assert.FileExists(t, filepath.Join(tempDir, "cc.yaml"))
}
```

---

## Test Execution

### Run All Tests

```bash
# Run ATDD acceptance tests (Gauge)
gauge run requirements/

# Run BDD behavior tests (Godog)
godog requirements/**/behavior.feature

# Run TDD unit tests
go test ./...
```

### Run Specific Feature

```bash
# Run acceptance tests for one feature
gauge run requirements/cli/init_project/

# Run behavior tests for one feature
godog requirements/cli/init_project/behavior.feature

# Run unit tests for one package
go test ./src/cli/...
```

### Run by Tags

```bash
# Run critical features only (Gauge)
gauge run --tags "critical" requirements/

# Run critical scenarios only (Godog)
godog --tags="@critical" requirements/**/behavior.feature

# Run success scenarios only (Godog)
godog --tags="@success" requirements/**/behavior.feature
```

---

## Traceability Model

### Feature ID Linkage

```text
Feature ID: cli_init_project

Links all files:
  +-- acceptance.spec
  |     (Business requirements)
  |
  +-- behavior.feature
  |     (Executable scenarios)
  |
  +-- acceptance_test.go
  |     (Gauge step implementations)
  |
  +-- step_definitions_test.go
  |     (Godog step definitions)
  |
  +-- Unit tests
        (Implementation validation)
        // Feature: cli_init_project
```

### From Acceptance Criteria to Scenarios

```text
acceptance.spec:
  AC1: Creates project directory structure
    |
    +-- @ac1 tag
          |
          v
behavior.feature:
  @success @ac1
  Scenario: Initialize in empty directory...
```

### Finding Related Files

```bash
# Find all files for a feature
ls -la requirements/cli/init_project/

# Find by Feature ID in code
grep -r "Feature: cli_init_project" .
grep -r "Feature ID: cli_init_project" requirements/

# Find by tags
gauge run --tags "init_project" requirements/
godog --tags="@init_project" requirements/
```

---

## Detailed Guides

- **[ATDD Guide](./atdd.md)** - Gauge specs, acceptance criteria, and business validation
- **[BDD Guide](./bdd.md)** - Godog scenarios, Gherkin syntax, and behavior testing
- **[TDD Guide](./tdd.md)** - Unit testing patterns across multiple languages

## Contract Integration

This testing strategy aligns with versioned contracts:

- `contracts/testing/0.1.0/specifications.yml` - Defines tdd/bdd/atdd classifications
- `contracts/testing/0.1.0/taxonomy.yml` - Defines test levels (l0-l4, horizontal-e2e)

---

**Next Steps**: Read the detailed guides for [ATDD](./atdd.md), [BDD](./bdd.md), and [TDD](./tdd.md) to master each testing layer.
