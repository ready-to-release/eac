# BDD Concepts

Understanding Behavior-Driven Development with Godog.

---

## What is BDD?

**Behavior-Driven Development (BDD)** is a specification technique that describes user-facing behavior through concrete examples. BDD focuses on **observable behavior** - what users can see and interact with, not internal implementation details.

### Core Purpose

BDD answers the question: **"How does the system behave from the user's perspective?"**

By writing scenarios in natural language (Given/When/Then), teams create:

- Shared understanding of expected behavior
- Executable specifications that become automated tests
- Living documentation that stays synchronized with code

---

## Why Use BDD?

### Common Language

**Problem**: Developers, testers, and product owners speak different languages
**Solution**: BDD uses Gherkin (Given/When/Then), which is readable by all stakeholders

### Focus on Behavior

**Problem**: Tests focus on implementation details that change frequently
**Solution**: BDD tests describe **what the system does**, not **how it does it**

**Example**:

**Bad** (implementation-focused):

```gherkin
When the ConfigParser reads the YAML file
And the YAML is deserialized into a Config struct
Then the struct fields are populated
```

**Good** (behavior-focused):

```gherkin
When I run "cc init"
Then a file named "cc.yaml" should be created
And the file should contain valid YAML
```

### Living Documentation

**Problem**: Documentation becomes outdated
**Solution**: BDD scenarios are executable - if they pass, the documentation is accurate

### Concrete Examples

**Problem**: Abstract requirements are ambiguous
**Solution**: BDD uses concrete examples that show exactly what should happen

---

## BDD with Godog

This project uses **[Godog](https://github.com/cucumber/godog)**, the official Cucumber implementation for Go.

### Why Godog?

- **Gherkin syntax**: Industry-standard Given/When/Then format
- **Executable**: Scenarios become automated tests
- **Go native**: Integrates seamlessly with Go projects
- **Rich features**: Tables, scenario outlines, hooks, tags
- **Clear reporting**: Multiple output formats (pretty, junit, cucumber)

### Godog Scenarios

Godog scenarios are written in Gherkin syntax in `.feature` files:

```gherkin
# Feature ID: cli_init_project
# Acceptance Spec: acceptance.spec
# Module: CLI

@cli @critical @init_project
Feature: Initialize project command behavior

  @success @ac1
  Scenario: Initialize in empty directory
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And a directory named "src/" should exist
    And the command should exit with code 0

  @error @ac1
  Scenario: Initialize in existing project shows error
    Given I am in a directory with "cc.yaml"
    When I run "cc init"
    Then the command should fail
    And stderr should contain "already initialized"
```

**Key Point**: These scenarios are **executable** - they map to Go step definitions that validate the behavior.

---

## The Gherkin Language

### Given/When/Then Structure

BDD scenarios follow a **Given/When/Then** pattern that mirrors how we naturally describe behavior:

| Keyword | Purpose | Question | Example |
|---------|---------|----------|---------|
| **Given** | Preconditions | "What's the starting state?" | `Given I am in an empty folder` |
| **When** | Action | "What action do I take?" | `When I run "cc init"` |
| **Then** | Outcome | "What should I observe?" | `Then a file should be created` |
| **And** | Continuation | "What else?" | `And the command should succeed` |
| **But** | Negative continuation | "What shouldn't happen?" | `But no error should appear` |

### Writing Effective Steps

#### Given (Preconditions)

**Purpose**: Describe the initial state before the action

**Characteristics**:

- Set up context
- Use present tense ("I am...", "the file exists...")
- Focus on relevant state only

**Examples**:

```gherkin
Given I am in an empty folder
Given the file "config.yaml" exists
Given the environment variable "DEBUG" is set to "true"
Given I am in a directory with "package.json"
```

#### When (Action)

**Purpose**: Describe the user action or event

**Characteristics**:

- Usually one action per scenario
- Describes what the user does
- Quote CLI commands for clarity

**Examples**:

```gherkin
When I run "cc init"
When I run "cc deploy staging"
When I delete the file "config.yaml"
When I press Ctrl+C
```

#### Then (Outcome)

**Purpose**: Verify observable results

**Characteristics**:

- Check what the user can see
- Verify files, output, exit codes
- Use "should" language

**Examples**:

```gherkin
Then a file named "cc.yaml" should be created
Then the command should exit with code 0
Then stdout should contain "Success"
Then stderr should contain "Error: config not found"
Then the deployment log should be created
```

---

## Observable Behavior vs Implementation

### The Key Principle

**BDD describes WHAT the system does (observable behavior)**
**Not HOW it does it (implementation details)**

### Examples

#### ❌ Bad: Implementation Details

```gherkin
Scenario: Parse configuration file
  Given a YAML file with valid syntax exists
  When the ConfigParser.Parse() method is called
  And the YAML is deserialized into a Config struct
  And the struct is validated using ValidateConfig()
  Then the Config object should have non-nil fields
```

**Problems**:

- Mentions internal classes (ConfigParser)
- Mentions methods (Parse(), ValidateConfig())
- Tests implementation, not behavior
- Changes when code is refactored

#### ✅ Good: Observable Behavior

```gherkin
Scenario: Initialize with valid configuration
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the file should contain valid YAML
  And the file should contain "version: 1.0.0"
```

**Benefits**:

- Describes user-facing behavior
- Independent of implementation
- Tests what users care about
- Stable during refactoring

---

## From Example Mapping to Scenarios

BDD scenarios are derived from **Green Cards** created during Example Mapping workshops (see [ATDD Concepts](./atdd-concepts.md#example-mapping-workshop)).

### Green Card Format

Each Green Card describes a concrete example:

**Format**: [Context] → [Action] → [Result]

**Example**:

```text
Empty folder → run init → creates src/, tests/, docs/
```

### Converting to Gherkin

Extract the three parts:

| Green Card Part | → | Gherkin Keyword |
|-----------------|---|-----------------|
| **Context** (Empty folder) | → | `Given I am in an empty folder` |
| **Action** (run init) | → | `When I run "cc init"` |
| **Result** (creates dirs) | → | `Then directories should exist` |

**Full conversion**:

```text
Green Card:
  Empty folder → init → creates src/, tests/, docs/

Gherkin Scenario:
  @success @ac1
  Scenario: Initialize in empty directory creates structure
    Given I am in an empty folder
    When I run "cc init"
    Then a directory named "src/" should exist
    And a directory named "tests/" should exist
    And a directory named "docs/" should exist
```

### Multiple Green Cards → Multiple Scenarios

Each Green Card becomes one scenario:

```text
[BLUE-1] Creates project directory structure
  [GREEN-1a] Empty folder → init → creates dirs
  [GREEN-1b] Existing project → init → error
```

**Becomes**:

```gherkin
# Green Card 1a
@success @ac1
Scenario: Initialize in empty directory creates structure
  Given I am in an empty folder
  When I run "cc init"
  Then directories should exist

# Green Card 1b
@error @ac1
Scenario: Initialize in existing project shows error
  Given I am in a directory with "cc.yaml"
  When I run "cc init"
  Then the command should fail
  And stderr should contain "already initialized"
```

---

## Scenario Tags

Tags categorize and filter scenarios.

### Feature-Level Tags

Applied at the **Feature** level, inherited by all scenarios:

```gherkin
@cli @critical @init_project
Feature: Initialize project command behavior
  # All scenarios inherit @cli @critical @init_project
```

**Common feature tags**:

- `@cli` - CLI-level interaction
- `@vscode` - VS Code extension
- `@io` - Involves file/network I/O
- `@integration` - Interacts with external systems
- `@critical` - Business-critical functionality

### Scenario-Level Tags

Applied to individual scenarios:

```gherkin
@success @ac1
Scenario: Initialize in empty directory
  # Effective tags: @cli @critical @init_project @success @ac1
```

**Common scenario tags**:

- `@success` - Happy path (normal operation)
- `@error` - Error handling scenario
- `@flag` - Involves command flags
- `@ac1`, `@ac2` - Links to acceptance criteria

### Acceptance Criteria Linking

**Purpose**: Map scenarios to acceptance criteria in acceptance.spec

**Pattern**: Use `@ac1`, `@ac2`, `@ac3`, etc.

**In acceptance.spec**:

```markdown
### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios
```

**In behavior.feature**:

```gherkin
@success @ac1
Scenario: Initialize in empty directory creates structure
  # This scenario validates AC1
```

**Benefits**:

- Traceability from requirements to tests
- Easy to find scenarios for a specific criterion
- Ensures all criteria have test coverage

---

## Scenario Outlines (Data-Driven Tests)

When testing the same behavior with different inputs, use **Scenario Outline** to avoid repetition.

### Without Scenario Outline (Repetitive)

```gherkin
Scenario: Detect CLI module
  When I determine module for "automation/cli/deploy/script.sh"
  Then the detected module is "cli-deploy"

Scenario: Detect container module
  When I determine module for "automation/container/registry/config.yml"
  Then the detected module is "container-registry"

Scenario: Detect docs module
  When I determine module for "automation/docs/build/Makefile"
  Then the detected module is "docs-build"
```

### With Scenario Outline (Concise)

```gherkin
Scenario Outline: Detect automation modules
  When I determine module for "<path>"
  Then the detected module is "<module>"

  Examples:
    | path                                      | module              |
    | automation/cli/deploy/script.sh           | cli-deploy          |
    | automation/container/registry/config.yml  | container-registry  |
    | automation/docs/build/Makefile            | docs-build          |
```

**Benefits**:

- Reduces duplication
- Easier to add new test cases
- Data is separated from logic
- Intent is clearer

---

## BDD Best Practices

### 1. One Scenario Per Example

Each Green Card = one scenario. Don't combine multiple examples into one scenario.

### 2. Use Background for Common Setup

If all scenarios start with the same steps, use `Background`:

```gherkin
Feature: Deploy application

  Background:
    Given the application is built
    And the deployment environment is configured

  @success @ac1
  Scenario: Deploy to staging
    When I run "cc deploy staging"
    Then the deployment should succeed

  @success @ac2
  Scenario: Deploy to production
    When I run "cc deploy production"
    Then the deployment should succeed
```

### 3. Keep Scenarios Focused

Each scenario should test **one thing**. If a scenario has many Then statements, it might be testing too much.

**Bad** (too broad):

```gherkin
Scenario: Initialize project
  When I run "cc init"
  Then a file should be created
  And directories should exist
  And the config should be valid
  And permissions should be correct
  And the version should be set
  And the timestamp should be recent
```

**Good** (focused):

```gherkin
Scenario: Initialize creates config file
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the file should contain valid YAML

Scenario: Initialize creates directory structure
  When I run "cc init"
  Then directories "src/", "tests/", "docs/" should exist
```

### 4. Use Declarative Language

Focus on **what** should happen, not **how** to do it.

**Bad** (imperative/procedural):

```gherkin
Given I open the terminal
And I navigate to the project directory
And I type "cc init"
And I press Enter
```

**Good** (declarative):

```gherkin
Given I am in the project directory
When I run "cc init"
```

---

## BDD vs ATDD vs TDD

### BDD (behavior.feature)

- **Focus**: Observable user behavior
- **Language**: Gherkin (Given/When/Then)
- **Stakeholder**: QA and Developers
- **Question**: "Does it behave as expected?"
- **Example**: "When I run 'cc init', a file is created"

### ATDD (acceptance.spec)

- **Focus**: Business requirements
- **Language**: Natural language (Markdown)
- **Stakeholder**: Product Owner
- **Question**: "Are we building the right thing?"
- **Example**: "Creates project directory structure"

### TDD (unit tests)

- **Focus**: Implementation correctness
- **Language**: Go test code
- **Stakeholder**: Developers
- **Question**: "Does the code work correctly?"
- **Example**: "TestCreateConfig validates file creation logic"

**All three work together** - see [Three-Layer Approach](./three-layer-approach.md).

---

## Common Questions

### Q: Who writes the behavior.feature file?

**A**: Initially, the **developer** or **QA** writes it after the Example Mapping workshop, based on the Green Cards. The team reviews together before implementation.

### Q: How detailed should scenarios be?

**A**: Detailed enough to be **unambiguous**, but not so detailed that they specify implementation. Focus on observable behavior.

### Q: Should scenarios test happy path only?

**A**: No. Include both:

- **@success** scenarios (happy path)
- **@error** scenarios (error handling)

Both are essential for complete behavior specification.

### Q: How many scenarios per feature?

**A**: **10-20 scenarios** is ideal. If you have more than 30, consider splitting into multiple `.feature` files.

### Q: What if behavior changes during development?

**A**: Update the scenarios. BDD scenarios are **living documentation** - they should always reflect current behavior.

---

## Benefits of BDD

### 1. Shared Understanding

Developers, QA, and product owners all understand the scenarios. No translation needed.

### 2. Confidence in Refactoring

Because scenarios test behavior (not implementation), you can refactor code without changing scenarios.

### 3. Regression Protection

Scenarios become automated tests that catch regressions.

### 4. Documentation That Never Lies

If scenarios pass, the documentation is accurate. If behavior changes, scenarios fail until updated.

### 5. Faster Onboarding

New team members read scenarios to understand how the system behaves.

---

## Related Documentation

- [Three-Layer Testing Approach](./three-layer-approach.md) - How ATDD/BDD/TDD work together
- [ATDD Concepts](./atdd-concepts.md) - Understanding ATDD with Gauge
- [BDD Format Reference](../../reference/testing/bdd-format.md) - Gherkin syntax
- [Godog Commands](../../reference/testing/godog-commands.md) - Running BDD tests
- [Create Feature Spec](../../how-to-guides/testing/create-feature-spec.md) - Step-by-step guide
