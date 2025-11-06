# Gherkin Format Reference

Quick reference for specification.feature file structure and Gherkin syntax with Rule blocks.

---

## Overview

This project uses Gherkin's Rule syntax to maintain ATDD/BDD separation within a single file.

---

## Architectural Context

```text
specs/<module>/<feature>/specification.feature    ← This reference documents THIS
    (WHAT the system should do)

src/<module>/tests/steps_test.go                  ← Separate documentation for this
    (HOW to verify the specifications)
```

**Separation of Concerns**:

- **Specifications** in `specs/`: Business-readable Gherkin describing WHAT
- **Implementations** in `src/`: Technical Go code describing HOW

---

## File Location

```text
specs/<module>/<feature_name>/specification.feature
```

**Examples**:

- `specs/cli/init-project/specification.feature`
- `specs/src-commands/ai-commit-generation/specification.feature`

---

## Template Structure

```gherkin
@<module> @critical @<feature-name>
Feature: [Feature Name]

  As a [role]
  I want [capability]
  So that [business value]

  Background:
    Given [common precondition for all scenarios]
    And [common setup]

  Rule: [Acceptance Criterion 1 - Business Rule]

    @success @ac1 @IV
    Scenario: [Installation/setup scenario]
      Given [precondition]
      When [installation action]
      Then [installation verified]
      And [configuration verified]

    @success @ac1
    Scenario: [Happy path scenario - OV by default]
      Given [precondition]
      When [action]
      Then [observable outcome]
      And [verification]

    @error @ac1
    Scenario: [Error scenario]
      Given [precondition]
      When [invalid action]
      Then [error behavior]
      And [error message verification]

  Rule: [Acceptance Criterion 2]

    @success @ac2 @PV
    Scenario: [Performance scenario]
      Given [precondition]
      When [action]
      Then [outcome within time threshold]
      And [resource usage within limits]

    @success @ac2 @risk1
    Scenario: [Risk control scenario]
      Given [security precondition]
      When [authenticated action]
      Then [access granted]
```

---

## Component Breakdown

### Metadata Header (Comments)

**Purpose**: Provides traceability and module context

**Format**:

```gherkin
# Feature ID: <module>_<feature-name>
# Module: <Module>
```

**Example**:

```gherkin
# Feature ID: cli_init-project
# Module: CLI
```

**Note**: Use underscores between module and feature, dashes within feature names.

---

### Feature Tags

**Purpose**: Categorize features for filtering and reporting

**Placement**: Feature line (applies to all scenarios)

**Format**:

```gherkin
@<module> @<priority> @<feature-name>
Feature: [Feature Name]
```

**Common Tags**:

- **Module**: `@cli`, `@vscode`, `@mcp`, `@docs`, `@src-commands`
- **Priority**: `@critical`, `@high`, `@medium`, `@low`
- **Type**: `@integration`, `@io`, `@git`, `@ai`

**Example**:

```gherkin
@cli @critical @init
Feature: cli_init-project
```

---

### Feature Description (User Story)

**Purpose**: Provides context and business value

**Format**:

```gherkin
Feature: [Feature Name]

  As a [role]
  I want [capability]
  So that [business value]
```

**Example**:

```gherkin
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project with one command
  So that I can quickly start development
```

**Best Practices**:

- Keep feature names aligned with Feature ID
- Use clear, stakeholder-friendly language
- Focus on user value, not implementation

---

### Background

**Purpose**: Define common setup shared by all scenarios in the feature

**When to use**:

- Preconditions needed by multiple scenarios
- Common test data setup
- Shared environment configuration

**Format**:

```gherkin
Background:
  Given [common precondition]
  And [additional setup]
```

**Example**:

```gherkin
Background:
  Given the repository has module contracts defined
  And I am in the src/commands directory
```

**When NOT to use**:

- Setup specific to one scenario (put in that scenario instead)
- Complex setup that makes scenarios hard to understand

---

### Rule Blocks (ATDD Layer)

**Purpose**: Define acceptance criteria in business terms

**Representation**: Each Rule represents one acceptance criterion from Example Mapping (Blue Card)

**Format**:

```gherkin
Rule: [Measurable business rule or acceptance criterion]

  [Scenarios that validate this rule...]
```

**Examples**:

```gherkin
Rule: All modules must be shown with details

Rule: Module types must be shown with counts

Rule: Contract loading errors must be handled gracefully
```

**Best Practices**:

- Each Rule should be measurable
- Rules should reflect business requirements, not implementation
- Aim for 2-6 Rules per feature
- If >6 Rules, consider splitting the feature

**Connection to Example Mapping**:

- Blue Card (Example Mapping) → Rule block (Gherkin)

---

### Scenario Blocks (BDD Layer)

**Purpose**: Executable examples of behavior

**Location**: Nested under Rule blocks

**Format**:

```gherkin
Rule: [Acceptance Criterion]

  @success @ac1
  Scenario: [Clear description of happy path]
    Given [precondition]
    When [action]
    Then [observable outcome]
    And [verification]
```

**Scenario Structure** (Given/When/Then):

| Keyword | Purpose | Example |
|---------|---------|---------|
| **Given** | Set up preconditions | `Given I am in an empty folder` |
| **When** | Perform action | `When I run "r2r init"` |
| **Then** | Assert outcome | `Then a file named "r2r.yaml" should be created` |
| **And** | Continue previous keyword | `And the command should exit with code 0` |
| **But** | Negative continuation | `But the file should not be empty` |

**Best Practices**:

- One scenario per specific behavior
- Use concrete examples, not abstract descriptions
- Include both success and error cases
- Keep scenarios focused and readable

**Connection to Example Mapping**:

- Green Card (Example Mapping) → Scenario block (Gherkin)

---

### Scenario Tags

**Purpose**: Link scenarios to acceptance criteria and verification types

**Placement**: Scenario line

**Tag Categories**:

#### 1. Outcome Tags (Required)

- `@success` - Happy path scenarios
- `@error` - Error/failure scenarios

#### 2. Acceptance Criteria Links (Required)

- `@ac1`, `@ac2`, `@ac3`, etc.
- Links scenario to Rule (acceptance criterion)

#### 3. Verification Type Tags (Optional)

- `@IV` - Installation Verification (setup, initial config)
- `@PV` - Performance Verification (timing, resource usage)
- No tag = `@OV` (Operational Verification - default)

#### 4. Risk Control Tags (Optional)

- `@risk1`, `@risk2`, etc.
- Links to risk control requirements

**Example**:

```gherkin
Rule: User authentication must be required

  @success @ac1 @risk1
  Scenario: Valid credentials grant access
    Given I have valid credentials
    When I run "simply login --user admin"
    Then I should be authenticated
```

**Tag Combinations**:

```gherkin
@success @ac1          # Standard operational scenario
@success @ac1 @IV      # Installation verification
@success @ac2 @PV      # Performance verification
@error @ac1            # Error case
@success @ac3 @risk5   # Success case implementing risk control #5
```

---

## Verification Tags in Detail

### @IV - Installation Verification

**Purpose**: Verify installation, setup, and initial configuration

**When to use**:

- First-time setup scenarios
- Installation processes
- Initial configuration
- Environment preparation

**Example**:

```gherkin
@success @ac1 @IV
Scenario: Install creates directory structure
  Given no project exists
  When I run "r2r install"
  Then directories "src/", "tests/", "docs/" are created
```

### @PV - Performance Verification

**Purpose**: Verify performance characteristics and resource usage

**When to use**:

- Response time requirements
- Throughput requirements
- Resource consumption limits
- Scalability tests

**Example**:

```gherkin
@success @ac2 @PV
Scenario: Command completes within time limit
  Given a project with 100 files
  When I run "r2r build"
  Then the command completes in under 5 seconds
```

### @OV - Operational Verification (Default)

**Purpose**: Verify standard operational behavior

**When to use**:

- Regular feature behavior
- Business logic validation
- Standard use cases

**Note**: This is the default - no tag needed

**Example**:

```gherkin
@success @ac1
Scenario: Create new project
  Given I am in an empty directory
  When I run "r2r init my-project"
  Then project "my-project" is created
```

---

## Risk Control Tags

**Purpose**: Link scenarios to risk control requirements

**Format**: `@risk<ID>` where ID is the risk control number

**Example**:

```gherkin
# Risk control definition (in specs/risk-controls/)
@risk1
Scenario: RC-001 - User authentication required
  Given a system with protected resources
  Then all user access MUST be authenticated

# User scenario implementing the control
@success @ac1 @risk1
Scenario: Valid credentials grant access
  Given I have valid credentials
  When I run "simply login"
  Then I should be authenticated
```

**See**: [How to Link Risk Controls](../../how-to-guides/specifications/link-risk-controls.md)

---

## Best Practices

### Feature Organization

✅ **Do**:

- Keep related scenarios together under their Rule
- Use Background for common setup
- Maintain 10-20 scenarios per file
- Split large features into multiple files

❌ **Don't**:

- Mix unrelated scenarios in one feature
- Duplicate setup across scenarios (use Background)
- Create features with >30 scenarios
- Put all scenarios at root (nest under Rules)

### Rule Block Guidelines

✅ **Do**:

- One Rule per acceptance criterion
- Make Rules measurable
- Use business language
- Aim for 2-6 Rules per feature

❌ **Don't**:

- Create Rules for implementation details
- Use technical jargon in Rules
- Have >6 Rules (split feature instead)
- Skip Rules (defeats ATDD purpose)

### Scenario Writing

✅ **Do**:

- Use concrete examples
- Focus on observable behavior
- Include both success and error cases
- Keep scenarios independent

❌ **Don't**:

- Use abstract descriptions
- Test implementation details
- Create dependent scenarios
- Make scenarios too long (>10 steps)

### Naming Conventions

**Feature IDs**: `module_feature-name`

- Examples: `cli_init-project`, `src-commands_module-inspection`

**Feature Names**: Same as Feature ID

- Example: `Feature: cli_init-project`

**Directory Names**: Use dashes

- Examples: `init-project`, `module-inspection`

**File Names**: Always `specification.feature`

---

## Step Definition Implementation

**Location**: `src/<module>/tests/steps_test.go`

**Example Structure**:

```go
// Feature: cli_init-project
// Step definitions for specs/cli/init-project/specification.feature
package tests

import (
    "github.com/cucumber/godog"
)

func InitializeScenario(sc *godog.ScenarioContext) {
    // Register steps that appear in the specification
    sc.Step(`^I am in an empty directory$`, iAmInAnEmptyDirectory)
    sc.Step(`^I run "([^"]*)"$`, iRun)
    sc.Step(`^directory "([^"]*)" should exist$`, directoryShouldExist)
}

func iAmInAnEmptyDirectory() error {
    // Implementation
    return nil
}

func iRun(command string) error {
    // Implementation
    return nil
}

func directoryShouldExist(path string) error {
    // Implementation
    return nil
}
```

**See**: [Godog Commands](./godog-commands.md) for running tests

---

## Related Documentation

- [ATDD and BDD with Gherkin](../../explanation/specifications/atdd-bdd-with-gherkin.md) - Concepts and workflow
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - How ATDD/BDD/TDD work together
- [Godog Commands](./godog-commands.md) - Running tests
- [Verification Tags](./verification-tags.md) - Detailed verification tag guide
- [Create Feature Spec](../../how-to-guides/specifications/create-specifications.md) - Step-by-step guide
