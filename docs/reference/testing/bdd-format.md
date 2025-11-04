# BDD Format Reference

Quick reference for behavior.feature file structure and Gherkin syntax.

---

## File Location

```text
specs/<module>/<feature_name>/behavior.feature
```

---

## Template Structure

```gherkin
# Feature ID: <module>_<feature_name>
# Acceptance Spec: acceptance.spec
# Module: <Module>

@<module> @critical @<feature_name>
Feature: [Feature Name]

  Background:
    Given [common precondition for all scenarios]

  @success @ac1 @IV
  Scenario: [Installation/setup scenario name]
    Given [precondition]
    When [installation/setup action]
    Then [installation verified]
    And [configuration verified]

  @success @ac1
  Scenario: [Happy path scenario name - OV by default]
    Given [precondition]
    When [action]
    Then [observable outcome]
    And [additional verification]

  @error @ac1
  Scenario: [Error scenario name - OV by default]
    Given [precondition]
    When [invalid action]
    Then [error behavior]
    And [error message verification]

  @success @ac2 @PV
  Scenario: [Performance scenario name]
    Given [precondition]
    When [action]
    Then [outcome within time threshold]
    And [resource usage within limits]
```

---

## Component Breakdown

### Metadata Header (Comments)

**Purpose**: Links this feature to acceptance.spec and provides traceability

**Format**:

```gherkin
# Feature ID: <module>_<feature_name>
# Acceptance Spec: acceptance.spec
# Module: <Module>
```

**Example**:

```gherkin
# Feature ID: cli_init_project
# Acceptance Spec: acceptance.spec
# Module: CLI
```

### Feature Tags

**Purpose**: Categorize features by testing level and priority

**Placement**: Feature line (applies to all scenarios)

**Format**:

```gherkin
@<module> @<priority> @<feature_name>
Feature: [Feature Name]
```

**Example**:

```gherkin
@cli @critical @init_project
Feature: Initialize project command behavior
```

**Common feature tags**: `@cli`, `@vscode`, `@io`, `@integration`, `@critical`

### Feature Line

**Format**: `Feature: [Descriptive feature name]`

**Guidelines**:

- Clear, concise description of the feature
- Focus on user-facing behavior, not implementation
- Usually matches the feature directory name

**Example**:

```gherkin
Feature: Initialize project command behavior
```

### Background (Optional)

**Purpose**: Common setup for all scenarios in the feature

**Format**:

```gherkin
Background:
  Given [common precondition]
  And [additional setup]
```

**Example**:

```gherkin
Background:
  Given I am in a clean test environment
  And all previous test artifacts are removed
```

**When to use**:

- All scenarios share the same initial setup
- Reduces repetition across scenarios
- Improves readability

**When NOT to use**:

- Setup is specific to individual scenarios
- Background would be longer than scenarios

### Scenarios

**Purpose**: Describe one specific user interaction

**Format**:

```gherkin
@<tags>
Scenario: [Descriptive scenario name]
  Given [precondition]
  When [action]
  Then [outcome]
  And [additional verification]
```

**Example**:

```gherkin
@success @ac1
Scenario: Initialize in empty directory
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the command should exit with code 0
```

### Scenario Tags

**Purpose**: Categorize individual scenarios by type

**Placement**: Scenario line (applies to that scenario only)

**Format**:

```gherkin
@<type> @<ac_link> @<verification_tag>
Scenario: [Name]
```

**Common scenario tags**: `@success`, `@error`, `@flag`, `@ac1`, `@ac2`, `@IV`, `@PV`

#### Risk Control Tags

Link scenarios to risk control requirements defined in `specs/risk-controls/`.

**Format**: `@risk<ID>` (e.g., `@risk1`, `@risk2`, `@risk10`)

**Purpose**:

- Create traceability between test scenarios and risk controls
- Support audit requirements
- Enable risk-based testing prioritization
- Generate compliance reports

**Placement**:

- **Feature-level**: When entire feature implements a risk control
- **Scenario-level**: When specific scenarios validate different controls

**How it works**:

1. **Risk controls are Gherkin scenarios** in `specs/risk-controls/` that define what the control requires
2. **User scenarios are tagged** with `@risk<ID>` to link to risk control definitions
3. **Traceability** is created through tag matching

**Example**:

Risk control definition:

```gherkin
# specs/risk-controls/authentication-controls.feature

@risk1
Scenario: RC-001 - User authentication required
  Given a system with protected resources
  Then all user access MUST be authenticated
```

User scenario implementation:

```gherkin
# specs/cli/user-authentication/behavior.feature

@success @ac1 @risk1
Scenario: Valid credentials grant access
  Given I have valid credentials
  When I run "simply login"
  Then I should be authenticated
```

**Naming Convention for Risk Control Scenarios**:

Risk control scenarios in `specs/risk-controls/` follow this pattern:

```gherkin
@risk<ID>
Scenario: RC-<ID> - <Short description>
  Given <context>
  Then <requirement> MUST <condition>
  And <requirement> MUST <condition>
```

**Example**:

```gherkin
@risk1
Scenario: RC-001 - User authentication required
  Given a system with protected resources
  Then all user access MUST be authenticated
  And authentication MUST occur before granting access
  And failed attempts MUST be logged
```

**Key points**:

- Tag: `@risk<ID>` (e.g., @risk1, @risk5, @risk10)
- Scenario name: `RC-<ID> - <Description>` (e.g., RC-001 - User authentication required)
- Use "MUST" for mandatory requirements
- Keep scenarios focused on one control

**Querying Risk Tags**:

```bash
# Find all scenarios for a specific risk control
grep -r "@risk1" specs/

# Find all features with any risk tag
grep -r "@risk:" specs/ | grep "Feature:"

# Count scenarios per risk control
grep -r "@risk" specs/ | grep -oP '@risk\K[0-9]+' | sort | uniq -c
```

**See**: [How to Link Risk Controls](../../how-to-guides/testing/link-risk-controls.md) for detailed guide.

---

## Gherkin Syntax

### Given/When/Then Structure

| Keyword | Purpose | Example |
|---------|---------|---------|
| **Given** | Set up preconditions | `Given I am in an empty folder` |
| **When** | Perform action | `When I run "cc init"` |
| **Then** | Verify outcome | `Then a file named "cc.yaml" should be created` |
| **And** | Continue previous step | `And the file should contain valid YAML` |
| **But** | Negative continuation | `But no error message should appear` |

### Writing Effective Steps

**Given** (Preconditions):

- Describe the initial state
- Set up test context
- Use present tense: "I am in...", "the file exists..."

**Example**:

```gherkin
Given I am in an empty folder
Given the file "config.yaml" exists
Given the environment variable "DEBUG" is set to "true"
```

**When** (Action):

- Describe the user action
- Usually a single action per scenario
- Quote CLI commands: `When I run "cc init --force"`

**Example**:

```gherkin
When I run "cc init"
When I run "cc deploy staging"
When I delete the file "config.yaml"
```

**Then** (Outcome):

- Verify observable results
- Check files, output, exit codes
- Use "should" language: "should be created", "should contain"

**Example**:

```gherkin
Then a file named "cc.yaml" should be created
Then the command should exit with code 0
Then stdout should contain "Success"
Then stderr should contain "Error: config not found"
```

### Scenario Outline (Data Tables)

Use **Scenario Outline** to test the same behavior with different inputs.

**Format**:

```gherkin
Scenario Outline: [Name with placeholders]
  Given [step with <placeholder>]
  When [step with <placeholder>]
  Then [step with <placeholder>]

  Examples:
    | placeholder1 | placeholder2 | placeholder3 |
    | value1       | value2       | value3       |
    | value4       | value5       | value6       |
```

**Example**:

```gherkin
@success @ac1
Scenario Outline: Detect automation modules
  When I determine module for "<path>"
  Then the detected module is "<module>"

  Examples:
    | path                                      | module              |
    | automation/cli/deploy/script.sh           | cli-deploy          |
    | automation/container/registry/config.yml  | container-registry  |
    | automation/docs/build/Makefile            | docs-build          |
```

**When to use**:

- Testing multiple variations of the same scenario
- Repetitive scenarios with only data differences
- Boundary value testing

**Benefits**:

- Reduces duplication
- Improves maintainability
- Makes data-driven testing explicit

---

## Feature ID Linkage

**Purpose**: Traceability across all test layers

**Linkage Pattern**:

```text
Feature ID: cli_init_project

Used in:
- acceptance.spec → > **Feature ID**: cli_init_project
- behavior.feature → # Feature ID: cli_init_project
- step_definitions_test.go → // Feature: cli_init_project
- Unit tests → // Feature: cli_init_project
```

**In behavior.feature**:

```gherkin
# Feature ID: cli_init_project
# Acceptance Spec: acceptance.spec
# Module: CLI

@cli @critical @init_project
Feature: Initialize project command behavior
```

---

## Tagging Taxonomy

### Feature-Level Tags

Apply at **Feature level** (affects all scenarios):

| Tag | Description | Usage |
|-----|-------------|-------|
| `@cli` | CLI-level interaction | Use for all CLI commands |
| `@vscode` | VS Code extension | Use for extension features |
| `@flag` | Involves command flags or arguments | Use when flags present |
| `@io` | Involves filesystem or network I/O | Use for file/network operations |
| `@integration` | Interacts with external systems | Use for Docker, APIs, databases |

**Example**:

```gherkin
@cli @critical @init_project
Feature: Initialize project command behavior
```

### Scenario-Level Tags

Apply at **Scenario level** (scenario-specific):

| Tag | Description | Usage |
|-----|-------------|-------|
| `@success` | Normal successful operation | Happy path scenarios |
| `@error` | Negative or invalid input scenario | Error handling scenarios |
| `@critical` | Business-critical functionality | Important acceptance criteria |
| `@wip` | Work in progress | Exclude from CI runs |
| `@ac1`, `@ac2`, etc. | Links to acceptance criterion | Maps to acceptance.spec |

**Example**:

```gherkin
@success @ac1
Scenario: Initialize in empty directory
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
```

### Verification Tags (Regulatory/Audit)

Apply at **Scenario level** for traceability in implementation reports:

| Tag | Description | Purpose |
|-----|-------------|---------|
| `@IV` | Installation Verification | Scenarios that verify installation, setup, deployment, configuration, and version checks |
| `@PV` | Performance Verification | Scenarios that verify performance requirements, response times, resource usage, and throughput |
| (none) | Operational Verification (OV) | Default - functional scenarios not tagged with @IV or @PV |

**When to Use Verification Tags**:

| Scenario Type | Tag | Examples |
|---------------|-----|----------|
| Installation, deployment, configuration, version checks, baseline setup | `@IV` | Install CLI, verify installation paths, check version numbers, validate environment configuration |
| Functional behavior, business logic, error handling, data processing | (none) | Run commands, process data, handle errors, validate outputs - these are OV by default |
| Performance requirements, response times, resource limits, throughput | `@PV` | Command completes in <2s, handles 1000+ items, memory usage under threshold, concurrent operations |

**Example**:

```gherkin
@cli @critical
Feature: CLI Installation and Performance

  @success @ac1 @IV
  Scenario: Install CLI on clean system
    Given a clean test environment
    When I run the installation script
    Then the CLI should be installed at "/usr/local/bin/cc"
    And the version should be "1.0.0"

  @success @ac2
  Scenario: Run basic help command
    Given the CLI is installed
    When I run "cc --help"
    Then help text should be displayed
    And the command should exit with code 0

  @success @ac3 @PV
  Scenario: Status command responds within performance threshold
    Given the CLI is installed
    And the project is initialized
    When I run "cc status"
    Then the command should complete within 2 seconds
    And the command should exit with code 0
```

### Tag Inheritance

**Feature-level tags apply to all scenarios**:

```gherkin
@cli @critical
Feature: Deploy application

  @success @ac1
  Scenario: Deploy to staging
    # Effective tags: @cli @critical @success @ac1
    Given the application is built
    When I run "cc deploy staging"
    Then the deployment should succeed

  @io @error @ac2
  Scenario: Deploy with missing config
    # Effective tags: @cli @critical @io @error @ac2
    Given the config file is missing
    When I run "cc deploy staging"
    Then the command should fail
    And stderr should contain "config not found"
```

---

## Acceptance Criteria Linking

**Purpose**: Map scenarios to acceptance criteria in acceptance.spec

**Pattern**: Use `@ac1`, `@ac2`, `@ac3`, etc. tags

**In acceptance.spec**:

```markdown
### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

### AC2: Generates valid configuration file
**Validated by**: behavior.feature -> @ac2 scenarios
```

**In behavior.feature**:

```gherkin
@success @ac1
Scenario: Initialize in empty directory creates structure
  Given I am in an empty folder
  When I run "cc init"
  Then directories "src/", "tests/", "docs/" should exist

@success @ac2
Scenario: Initialize creates valid configuration file
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the file should contain valid YAML
```

---

## Related Documentation

- [ATDD Format](./atdd-format.md) - Acceptance spec format
- [TDD Format](./tdd-format.md) - Unit test format
- [Godog Commands](./godog-commands.md) - Command reference
- [BDD Concepts](../../explanation/testing/bdd-concepts.md) - Understanding BDD
- [Create Feature Spec](../../how-to-guides/testing/create-feature-spec.md) - Step-by-step guide
