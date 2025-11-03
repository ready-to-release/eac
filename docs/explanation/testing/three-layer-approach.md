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

## Why Three Separate Layers?

### Separation of Concerns

Each layer focuses on a specific aspect of quality:

| Layer | Answers | Stakeholders | Perspective |
|-------|---------|--------------|-------------|
| **ATDD** | "What business value does this deliver?" | Product Owner, Business Stakeholders | Business perspective |
| **BDD** | "How does the user interact with this?" | QA, Developers, Product Owner | User perspective |
| **TDD** | "Does the code work correctly?" | Developers | Implementation perspective |

### Different Questions, Different Tools

**ATDD asks**: "Are we building the right thing?"

- Focus: Business requirements, acceptance criteria
- Tool: Gauge with Markdown specifications
- Output: Validated business value

**BDD asks**: "Does it behave as expected?"

- Focus: Observable user-facing behavior
- Tool: Godog with Gherkin scenarios
- Output: Executable living documentation

**TDD asks**: "Is the implementation correct?"

- Focus: Code correctness, internal logic
- Tool: Go test framework
- Output: Verified implementation quality

---

## The Three Layers

### Layer 1: ATDD (Acceptance Test-Driven Development)

**Purpose**: Define and validate business requirements

**Format**: Markdown specifications in `acceptance.spec` files

**Tool**: [Gauge](https://gauge.org/)

**Who writes it**: Product Owner with Development Team

**Content**:

- User stories (business value)
- Acceptance criteria (definition of "done")
- Measurable success metrics
- Business-level test steps

**Example**:

```markdown
# Initialize Project

> **Feature ID**: cli_init_project

## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure

## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file
* Command completes in under 2 seconds
```

**Key Characteristics**:

- Written before development starts
- Drives feature scope and definition of "done"
- Validated through Example Mapping workshops
- Executable with Gauge

### Layer 2: BDD (Behavior-Driven Development)

**Purpose**: Specify and verify user-facing behavior

**Format**: Gherkin scenarios in `behavior.feature` files

**Tool**: [Godog](https://github.com/cucumber/godog)

**Who writes it**: QA and Developers, informed by stakeholders

**Content**:

- Given/When/Then scenarios
- Observable CLI behaviors
- User interaction patterns
- Expected outputs and error messages

**Example**:

```gherkin
@success @ac1
Scenario: Initialize in empty directory
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the command should exit with code 0
```

**Key Characteristics**:

- Describes observable behavior, not implementation
- Written in natural language (Gherkin)
- Maps directly to acceptance criteria via @ac tags
- Executable with Godog

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
    configPath := filepath.Join(tmpDir, "cc.yaml")

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
3. ATDD Layer (acceptance.spec)
   ├─ Yellow Card → User Story
   └─ Blue Cards → Acceptance Criteria
   ↓
4. BDD Layer (behavior.feature)
   └─ Green Cards → Scenarios
   ↓
5. TDD Layer (unit tests)
   └─ Implementation testing
   ↓
6. Implementation Code
```

### Card to File Mapping

| Card Color | Contains | → | Output File | Tool |
|-----------|----------|---|-------------|------|
| 🟡 **Yellow** (1) | User Story | → | `acceptance.spec` | Gauge |
| 🔵 **Blue** (2-6) | Rules/Criteria | → | `acceptance.spec` | Gauge |
| 🟢 **Green** (2-4 per Blue) | Concrete Examples | → | `behavior.feature` | Godog |
| 🔴 **Red** (0-N) | Questions | → | `issues.md` | Tracking |

### Traceability Across Layers

All layers are linked through **Feature ID**:

```text
Feature ID: cli_init_project

Links all files:
  ├─ acceptance.spec
  │    > **Feature ID**: cli_init_project
  │
  ├─ behavior.feature
  │    # Feature ID: cli_init_project
  │
  ├─ acceptance_test.go
  │    // Feature: cli_init_project
  │
  ├─ step_definitions_test.go
  │    // Feature: cli_init_project
  │
  └─ Unit tests
       // Feature: cli_init_project
```

### Acceptance Criteria to Scenario Linkage

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

---

## When to Use Each Layer

### Use ATDD (acceptance.spec) When

✅ Defining new features with business stakeholders
✅ Running Example Mapping workshops
✅ Creating user stories for sprint planning
✅ Establishing definition of "done"
✅ Features require customer sign-off
✅ Need to validate business value

**Example Question**: "What does 'initialize project' mean for the business?"

### Use BDD (behavior.feature) When

✅ Converting concrete examples from workshops
✅ Specifying CLI command behavior
✅ Documenting user-facing interactions
✅ Creating executable scenarios
✅ Collaborating across teams (dev/QA/product)
✅ Need to verify user experience

**Example Question**: "What happens when a user runs 'cc init' in an empty folder?"

### Use TDD (unit tests) When

✅ Implementing complex internal algorithms
✅ Building server-side logic
✅ Creating utility functions
✅ Refactoring existing code safely
✅ Testing internal implementation details
✅ Ensuring code correctness

**Example Question**: "Does the config parser handle invalid YAML correctly?"

---

## Complete Example Workflow

### Step 1: Example Mapping Workshop (15-25 minutes)

**Output**:

```text
[YELLOW] As a developer, I want to initialize a CLI project with one command,
         so that I can quickly start development

[BLUE-1] Creates project directory structure
  [GREEN-1a] Empty folder → init → creates src/, tests/, docs/
  [GREEN-1b] Existing project → init → error "already initialized"

[BLUE-2] Generates valid configuration file
  [GREEN-2a] New project → init → creates cc.yaml with defaults
  [GREEN-2b] With --name flag → cc.yaml contains custom name

[RED-1] What if cc.yaml already exists? (→ issues.md)
```

### Step 2: Create acceptance.spec (ATDD)

**File**: `requirements/cli/init_project/acceptance.spec`

```markdown
# Initialize Project

> **Feature ID**: cli_init_project
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: CLI

## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure

## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file

## Acceptance Tests

### AC1: Creates project directory structure
**Validated by**: behavior.feature → @ac1 scenarios

* Create empty test directory
* Run "cc init" command
* Verify "src/" directory exists
* Verify "tests/" directory exists
* Verify command exit code is "0"

### AC2: Generates valid configuration file
**Validated by**: behavior.feature → @ac2 scenarios

* Create empty test directory
* Run "cc init" command
* Read "cc.yaml" file contents
* Verify YAML is valid
```

### Step 3: Create behavior.feature (BDD)

**File**: `requirements/cli/init_project/behavior.feature`

```gherkin
# Feature ID: cli_init_project
# Acceptance Spec: acceptance.spec
# Module: CLI

@cli @critical @init_project
Feature: Initialize project command behavior

  # Green Card 1a
  @success @ac1
  Scenario: Initialize in empty directory creates structure
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And a directory named "src/" should exist
    And the command should exit with code 0

  # Green Card 1b
  @error @ac1
  Scenario: Initialize in existing project shows error
    Given I am in a directory with "cc.yaml"
    When I run "cc init"
    Then the command should fail
    And stderr should contain "already initialized"

  # Green Card 2a
  @success @ac2
  Scenario: Initialize creates configuration with defaults
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And the file should contain valid YAML
    And the file should contain "version: 1.0.0"

  # Green Card 2b
  @flag @success @ac2
  Scenario: Initialize with custom name flag
    Given I am in an empty folder
    When I run "cc init --name my-project"
    Then the file should contain "name: my-project"
```

### Step 4: Implement Gauge and Godog Steps

**File**: `requirements/cli/init_project/acceptance_test.go` (Gauge)

```go
// Feature: cli_init_project
package init_project_test

import (
    "github.com/getgauge-contrib/gauge-go/gauge"
)

func init() {
    gauge.Step("Run <command> command", runCommand)
    gauge.Step("Verify <dir> directory exists", verifyDirectoryExists)
    // ... more step implementations
}
```

**File**: `requirements/cli/init_project/step_definitions_test.go` (Godog)

```go
// Feature: cli_init_project
package init_project_test

import (
    "github.com/cucumber/godog"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    // ... more step definitions
}
```

### Step 5: Write Unit Tests (TDD)

**File**: `src/cli/init_test.go`

```go
// Feature: cli_init_project
package cli_test

import (
    "testing"
)

func TestInitProject_CreatesConfigFile(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()

    // Act
    err := InitProject(tmpDir)

    // Assert
    if err != nil {
        t.Fatalf("InitProject failed: %v", err)
    }

    configPath := filepath.Join(tmpDir, "cc.yaml")
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Error("Config file was not created")
    }
}
```

### Step 6: Implement Feature Code

```go
package cli

func InitProject(dir string) error {
    // Implementation guided by all three test layers
    return createProjectStructure(dir)
}
```

### Step 7: Verify All Layers Pass

```bash
# Run ATDD acceptance tests
gauge run requirements/cli/init_project/

# Run BDD behavior tests
godog requirements/cli/init_project/behavior.feature

# Run TDD unit tests
go test ./src/cli/...
```

---

## Benefits of the Three-Layer Approach

### Clear Separation of Concerns

- **Business logic** lives in ATDD layer (acceptance.spec)
- **User behavior** lives in BDD layer (behavior.feature)
- **Implementation details** live in TDD layer (unit tests)

### Improved Communication

- **ATDD**: Stakeholders understand business value
- **BDD**: QA and developers share common language
- **TDD**: Developers have implementation confidence

### Better Traceability

Every feature can be traced from:

1. Business requirement (ATDD)
2. User scenario (BDD)
3. Implementation (TDD)

### Living Documentation

All three layers serve as documentation:

- **ATDD**: Documents business requirements
- **BDD**: Documents user-facing behavior
- **TDD**: Documents code behavior

### Reduced Ambiguity

- **ATDD**: "Creates project structure" is measurable
- **BDD**: Concrete examples show exactly what happens
- **TDD**: Tests verify implementation works

---

## Common Questions

### Q: Why not just use BDD scenarios?

**A**: BDD scenarios focus on **observable behavior**. ATDD provides:

- Business context (user story)
- Measurable acceptance criteria
- Stakeholder buy-in
- High-level test steps

### Q: Why not just use unit tests?

**A**: Unit tests focus on **implementation details**. They don't capture:

- Business requirements
- User-facing behavior
- End-to-end workflows
- Stakeholder language

### Q: Is this too much overhead?

**A**: Each layer has a specific purpose:

- Skip ATDD if feature has no business stakeholder
- Skip BDD if feature has no user-facing behavior
- Never skip TDD (code must be tested)

For most features, all three layers add value and reduce risk.

### Q: How do I keep the layers in sync?

**A**: Use **Feature ID** to link all files:

```bash
# Find all files for a feature
grep -r "Feature: cli_init_project" .
grep -r "Feature ID: cli_init_project" requirements/
```

---

## Related Documentation

- [ATDD Concepts](./atdd-concepts.md) - Understanding ATDD with Gauge
- [BDD Concepts](./bdd-concepts.md) - Understanding BDD with Godog
- [ATDD Format Reference](../../reference/testing/atdd-format.md) - Specification format
- [BDD Format Reference](../../reference/testing/bdd-format.md) - Scenario format
- [TDD Format Reference](../../reference/testing/tdd-format.md) - Unit test format
