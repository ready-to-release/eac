# ATDD Format Reference

Quick reference for acceptance.spec file structure and Gauge syntax.

---

## File Location

```text
requirements/<module>/<feature_name>/acceptance.spec
```

---

## Template Structure

```markdown
# [Feature Name]

> **Feature ID**: <module>_<feature_name>
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: <Module>
> **Tags**: <tags>

## User Story

* As a [user role]
* I want [capability]
* So that [business value]

## Acceptance Criteria

* [Measurable criterion 1]
* [Measurable criterion 2]
* [Measurable criterion 3]
* [Measurable criterion 4]

## Acceptance Tests

### AC1: [Criterion 1 description]
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, <module>

* [Gauge step 1]
* [Gauge step 2]
* [Gauge step 3]
* [Verification step]

### AC2: [Criterion 2 description]
**Validated by**: behavior.feature -> @ac2 scenarios

* [Gauge step 1]
* [Gauge step 2]
* [Verification step]
```

---

## Component Breakdown

### 1. Metadata Header

**Format**:

```markdown
> **Feature ID**: <module>_<feature_name>
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: <Module>
> **Tags**: <tags>
```

**Fields**:

- **Feature ID**: Unique identifier (e.g., `cli_init_project`)
- **BDD Scenarios**: Link to behavior.feature file
- **Module**: Module name (e.g., CLI, VSCode, Docs)
- **Tags**: Space-separated tags (e.g., `cli, critical`)

**Example**:

```markdown
> **Feature ID**: cli_init_project
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: CLI
> **Tags**: cli, critical
```

### 2. User Story

**Format**: Bullet list with three components

```markdown
## User Story

* As a [user role]
* I want [capability]
* So that [business value]
```

**Components**:

- **As a**: User persona or stakeholder
- **I want**: Desired functionality
- **So that**: Business benefit or outcome

**Example**:

```markdown
## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure
```

### 3. Acceptance Criteria

**Format**: Bullet list with measurable outcomes (2-6 items)

```markdown
## Acceptance Criteria

* [Measurable criterion 1]
* [Measurable criterion 2]
```

**Guidelines**:

- Keep to 2-6 items (more means feature too large)
- Each must be **measurable** (pass/fail, not subjective)
- Include functional AND non-functional requirements
- Cover cross-platform, performance, security, usability

**Example**:

```markdown
## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file
* Command completes in under 2 seconds
* Works on Linux, macOS, and Windows
* Exits with clear success/error messages
```

### 4. Acceptance Tests

**Format**: Gauge scenarios with executable steps

```markdown
### AC1: [Criterion description]
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: <tag1>, <tag2>

* [Gauge step 1]
* [Gauge step 2]
* [Verification step]
```

**Components**:

- **Heading**: AC number and description
- **Validated by**: Link to BDD scenarios (e.g., `@ac1`)
- **Tags**: Optional Gauge tags
- **Steps**: Executable Gauge steps (bullet points)

**Example**:

```markdown
### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Verify "cc.yaml" file exists
* Verify "src/" directory exists
* Verify "tests/" directory exists
```

---

## Gauge Step Syntax

### Plain Steps

```markdown
* Step description without parameters
```

### Parameterized Steps

**Static parameters** (quoted strings):

```markdown
* Run "cc init" command
* Verify "config.yaml" file exists
```

**Dynamic parameters** (angle brackets):

```markdown
* Run <command> command
* Verify <filename> file exists
* Assert execution time is less than <seconds> seconds
```

### Tables

**Inline tables**:

```markdown
* Verify files exist
    | File         | Exists |
    |--------------|--------|
    | cc.yaml      | true   |
    | src/main.go  | true   |
```

### Concepts

Reusable step groups:

**File**: `requirements/concepts/setup.cpt`

```markdown
# Setup test environment
* Create empty test directory
* Set environment variables
* Initialize test database
```

**Usage**:

```markdown
* Setup test environment
* Run tests
```

---

## Feature ID Linkage

**Purpose**: Traceability across all test layers

**Linkage Pattern**:

```text
Feature ID: cli_init_project

Used in:
- acceptance.spec → > **Feature ID**: cli_init_project
- behavior.feature → # Feature ID: cli_init_project
- acceptance_test.go → // Feature: cli_init_project
- step_definitions_test.go → // Feature: cli_init_project
- Unit tests → // Feature: cli_init_project
```

---

## Related Documentation

- [ATDD Concepts](../../explanation/testing/atdd-concepts.md) - Understanding ATDD
- [Create Feature Spec](../../how-to-guides/testing/create-feature-spec.md) - Step-by-step guide
- [Gauge Commands](./gauge-commands.md) - Command reference
- [BDD Format](./bdd-format.md) - Behavior scenarios format
