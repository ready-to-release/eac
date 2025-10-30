# BDD: Behavior-Driven Development

**[<- Back to Testing Overview](./index.md)**

## What is BDD?

**Behavior-Driven Development (BDD)** is a specification technique that describes user-facing behavior through concrete examples. In this project, BDD scenarios define **observable CLI behaviors** using Gherkin syntax (Given/When/Then).

### Key Characteristics

| Aspect | Description |
|--------|-------------|
| **Who** | Developers and testers (informed by stakeholders) |
| **When** | During feature specification |
| **Format** | Gherkin scenarios (Given/When/Then) |
| **Location** | Feature files in `requirements/<module>/` |
| **Purpose** | Specify observable CLI behaviors from user perspective |

## BDD in Feature Files

BDD scenarios appear **below the ATDD layer** in `.feature` files, providing executable specifications of user-facing behavior.

### Template Structure

```gherkin
@cli @critical
Feature: [Feature Name]

  # ATDD Layer (above - see ATDD guide)
  As a [role]
  I want [capability]
  So that [value]

  Acceptance Criteria:
  - [ ] [Criterion 1]

  # BDD Layer: User-facing behavior
  @success
  Scenario: [Happy path scenario name]
    Given [precondition]
    When [action]
    Then [observable outcome]
    And [additional verification]

  @error
  Scenario: [Error scenario name]
    Given [precondition]
    When [invalid action]
    Then [error behavior]
    And [error message verification]
```

### Component Breakdown

#### 1. Feature Tags

**Purpose**: Categorize features by testing level and priority

**Placement**: Feature line (applies to all scenarios)

**Example**:

```gherkin
@cli @flag @critical
Feature: Initialize a new project
```

**Common tags**: `@cli`, `@flag`, `@io`, `@integration`, `@critical`

#### 2. Scenarios

**Purpose**: Describe one specific user interaction

**Format**: Given/When/Then steps

**Example**:

```gherkin
@success
Scenario: Initialize in current directory
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the file should contain valid YAML
  And the command should exit with code 0
```

#### 3. Scenario Tags

**Purpose**: Categorize individual scenarios by type

**Placement**: Scenario line (applies to that scenario only)

**Example**: `@success`, `@error`, `@io`, `@flag`

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

**When** (Action):

- Describe the user action
- Usually a single action per scenario
- Quote CLI commands: `When I run "cc init --force"`

**Then** (Outcome):

- Verify observable results
- Check files, output, exit codes
- Use "should" language: "should be created", "should contain"

## Tagging Taxonomy

### Testing Level Tags

Apply at **Feature level** (affects all scenarios):

| Tag | Description | Usage |
|-----|-------------|-------|
| `@cli` | CLI-level interaction | Use for all CLI commands |
| `@flag` | Involves command flags or arguments | Use when flags present |
| `@io` | Involves filesystem or network I/O | Use for file/network operations |
| `@integration` | Interacts with external systems | Use for Docker, APIs, databases |

### Scenario Type Tags

Apply at **Scenario level** (scenario-specific):

| Tag | Description | Usage |
|-----|-------------|-------|
| `@success` | Normal successful operation | Happy path scenarios |
| `@error` | Negative or invalid input scenario | Error handling scenarios |
| `@critical` | Business-critical functionality | ATDD priority features |

### Tagging Rules

1. **Feature-level tags**: Include testing level tags (`@cli`, `@io`, etc.)
2. **Scenario-level tags**: Include scenario type tags (`@success`, `@error`)
3. **Inheritance**: Scenario inherits feature tags (combined)
4. **Multiple tags**: Can apply multiple tags per feature/scenario

### Tagging Examples

```gherkin
@cli @critical
Feature: Deploy application

  @success
  Scenario: Deploy to staging
    # Effective tags: @cli @critical @success
    Given the application is built
    When I run "cc deploy staging"
    Then the deployment should succeed

  @io @error
  Scenario: Deploy with missing config
    # Effective tags: @cli @critical @io @error
    Given the config file is missing
    When I run "cc deploy staging"
    Then the command should fail
    And stderr should contain "config not found"
```

## File Organization

### Module-Based Structure

Feature files are organized by module:

```text
requirements/
 cli/                      # CLI command features
    init_project.feature
    deploy_module.feature
    run_tests.feature
 vscode/                   # VS Code extension features
    commit_button.feature
    status_bar.feature
 docs/                     # Documentation features
    build_docs.feature
 mcp/                      # MCP server features
     server_startup.feature
```

### File Naming Conventions

**Format**: `<feature-description>.feature`

**Guidelines**:

- Use snake_case: `init_project.feature`
- Be descriptive: `generate_commit_message.feature` (not `commit.feature`)
- Reflect feature purpose: `handle_config_errors.feature`

**Good examples**:

- `requirements/cli/init_project.feature`
- `requirements/vscode/commit_button.feature`
- `requirements/docs/build_docs.feature`

**Bad examples**:

- `requirements/feature1.feature` (not descriptive)
- `requirements/test.feature` (too generic)
- `requirements/cli/Init-Project.feature` (wrong case)

## From Example Mapping to Scenarios

After the ATDD team completes an Example Mapping workshop (see [ATDD Guide](./atdd.md#example-mapping-workshop-collaborative-discovery)), you'll have a collection of **Green Cards** representing concrete examples. These cards become your BDD scenarios.

### Green Card to Gherkin Conversion

Each Green Card describes a specific example of a rule in action. Convert these to Gherkin scenarios by:

1. **Extract the context** -> GIVEN
2. **Identify the action** -> WHEN
3. **Determine the outcome** -> THEN

### Conversion Guidelines

**Card Structure**: `[Context] -> [Action] -> [Result]`

**Gherkin Structure**:

```gherkin
Scenario: [Descriptive name based on card]
  Given [Context]
  When [Action]
  Then [Result]
```

### Example Conversions

#### Example 1: Simple Conversion

**Green Card**:

```text
Empty folder -> run init -> creates directories
```

**Becomes**:

```gherkin
@success
Scenario: Initialize in empty folder
  Given I am in an empty folder
  When I run "cc init"
  Then directories "src/", "tests/", "docs/" should exist
  And the command should exit with code 0
```

#### Example 2: Error Case

**Green Card**:

```text
Existing project -> run init -> error "already initialized"
```

**Becomes**:

```gherkin
@error
Scenario: Initialize in existing project
  Given I am in a directory with "cc.yaml"
  When I run "cc init"
  Then the command should fail
  And stderr should contain "already initialized"
  And the command should exit with code 1
```

#### Example 3: With Flags

**Green Card**:

```text
With --name flag -> cc.yaml contains custom name
```

**Becomes**:

```gherkin
@flag @success
Scenario: Initialize with custom project name
  Given I am in an empty folder
  When I run "cc init --name my-project"
  Then a file named "cc.yaml" should be created
  And the file should contain "name: my-project"
  And the command should exit with code 0
```

### Handling Multiple Green Cards per Rule

A single Blue Card (rule/acceptance criterion) often has multiple Green Cards (examples). Create one scenario per Green Card.

**Blue Card**: Generates valid configuration file

**Green Cards**:

1. New project -> init -> creates cc.yaml with defaults
2. With --name flag -> cc.yaml contains custom name
3. With --verbose flag -> shows configuration being written

**Result**: 3 separate scenarios, all testing configuration generation

### Tips for Effective Conversion

- **Add exit code verification**: Green Cards may not specify exit codes, but BDD scenarios must
- **Specify stdout vs stderr**: Make output assertions explicit
- **Add scenario tags**: Use @success, @error, @flag, @io based on behavior
- **Make assertions observable**: Convert "creates directories" to specific directory names
- **Keep 1:1 mapping**: One Green Card = One Scenario (avoid combining)

### Relationship to ATDD Layer

Green Cards validate Blue Cards (acceptance criteria):

```text
ATDD Layer (Blue Card):
  - [ ] Creates project directory structure

BDD Layer (Green Cards):
  Scenario: Initialize creates required directories
  Scenario: Initialize in existing project shows error
```

Each Blue Card should have 2-4 Green Cards testing different aspects of that rule.

## Workflow

### BDD Development Process

```text
BDD Workflow

1. Read ATDD Context
    Review user story (As a/I want/So that) - YELLOW CARD
    Study acceptance criteria - BLUE CARDS
    Understand business value

2. Gather Green Cards from Example Mapping
    Retrieve Green Cards from workshop (see ATDD guide)
    Each Green Card is a concrete example
    Format: [Context] -> [Action] -> [Result]
    Expect 2-4 Green Cards per Blue Card (rule)

3. Convert Each Green Card to Scenario
    Extract context -> GIVEN step
    Identify action -> WHEN step
    Determine outcome -> THEN step
    Create one scenario per Green Card

4. Enhance Given/When/Then Steps
    Given: Set up preconditions clearly
    When: Describe user action (quote CLI commands)
    Then: Verify observable outcome
    Add: Exit code verification
    Add: stdout/stderr assertions
    Keep steps concise (max 2 lines)

5. Add Appropriate Tags
    Feature-level: @cli, @io, @integration
    Scenario-level: @success, @error, @flag
    Priority: @critical if business-critical

6. Review Scenario Completeness
    Does it verify acceptance criteria (Blue Cards)?
    Is it executable/testable?
    Are steps clear and unambiguous?
    Are all edge cases covered?
    Each Green Card has matching scenario?

7. Save to Module Folder
    Choose correct module (cli, vscode, docs, mcp)
    Use descriptive file name
    Commit to requirements/ directory

8. Proceed to TDD
    Reference feature file name in unit tests
    See: TDD Workflow ->
```

### Prerequisites

Before starting BDD:

- ATDD layer exists (user story + acceptance criteria)
- Example Mapping workshop completed (Green Cards available)
- Feature file name is chosen
- Module folder is identified

### Outputs

After completing BDD:

- One scenario per Green Card (2-4 scenarios total)
- Scenarios written in Gherkin (Given/When/Then)
- Scenarios tagged appropriately (@success, @error, etc.)
- All acceptance criteria (Blue Cards) have corresponding scenarios
- Scenarios are executable specifications
- File saved in `requirements/<module>/`

## Style Rules

### Do

- **Write from user's perspective**: "When I run...", "Then I see..."
- **Use present tense**: "I am in...", "the file exists..."
- **Quote CLI commands**: `When I run "cc init --force"`
- **Keep steps concise**: Max 2 lines per step
- **Test one behavior per scenario**: Single purpose
- **Include 2-4 scenarios per feature**: Happy path + error cases
- **Verify exit codes**: `And the command should exit with code 0`
- **Check error output**: `And stderr should contain "error message"`
- **Use "should" language**: "should be created", "should contain"

### L Don't

- **Reference internal functions**: `When CreateProject() is called`
- **Use implementation details**: Class names, method calls
- **Make scenarios too long**: >8 steps is usually too complex
- **Write ambiguous assertions**: "Then it should work"
- **Skip error scenarios**: Always test failure cases
- **Forget exit codes**: CLI tools must specify exit behavior
- **Mix stdout/stderr**: Be explicit about which stream

## Examples

### Example 1: CLI Initialization Feature

**File**: `requirements/cli/init_project.feature`

```gherkin
@cli @flag @critical
Feature: Initialize a new project

  # ATDD Layer
  As a developer
  I want to initialize a CLI project with a single command
  So that I can quickly start development with proper structure

  Acceptance Criteria:
  - [ ] Creates project directory structure
  - [ ] Generates valid configuration file
  - [ ] Handles existing projects gracefully

  # BDD Layer
  @success
  Scenario: Initialize in current directory
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And the file should contain valid YAML
    And a directory named "src/" should exist
    And the command should exit with code 0
    And stdout should contain "Project initialized successfully"

  @flag @success
  Scenario: Initialize with custom project name
    Given I am in an empty folder
    When I run "cc init --name my-project"
    Then a file named "cc.yaml" should be created
    And the file should contain "name: my-project"
    And the command should exit with code 0

  @io @error
  Scenario: Initialize in non-empty directory
    Given I am in a directory containing files
    When I run "cc init"
    Then the command should fail
    And stderr should contain "Directory must be empty"
    And no configuration file should be created
    And the command should exit with code 1

  @flag @error
  Scenario: Initialize with invalid flag
    Given I am in an empty folder
    When I run "cc init --invalid-flag"
    Then the command should fail
    And stderr should contain "unknown flag: --invalid-flag"
    And the command should exit with code 1
```

**Why this works**:

- Four scenarios cover happy path and error cases
- Tags clearly indicate scenario types (@success, @error)
- Steps verify files, output, and exit codes
- Error scenarios specify stderr (not stdout)
- Each scenario tests one specific behavior

### Example 2: VS Code Extension Feature

**File**: `requirements/vscode/commit_button.feature`

```gherkin
@vscode @integration @critical
Feature: Generate commit messages via button

  # ATDD Layer
  As a developer using VS Code
  I want to generate semantic commit messages by clicking a button
  So that I can create consistent, well-formatted commits

  Acceptance Criteria:
  - [ ] Button appears in VS Code Source Control panel
  - [ ] Generated message follows semantic commit format
  - [ ] Message reflects actual code changes accurately

  # BDD Layer
  @success
  Scenario: Generate commit message for single file change
    Given I have modified "src/index.ts"
    And the file has added a new function "calculateTotal"
    When I click the "Generate Commit Message" button
    Then the commit message should start with "feat(src):"
    And the commit message should contain "add calculateTotal function"
    And the message should appear in the commit input field

  @success
  Scenario: Generate commit message for bug fix
    Given I have modified "src/validator.ts"
    And the file has fixed a null pointer error
    When I click the "Generate Commit Message" button
    Then the commit message should start with "fix(validator):"
    And the commit message should describe the bug fix
    And the message should follow semantic commit format

  @error
  Scenario: Handle no changes staged
    Given I have no staged changes
    When I click the "Generate Commit Message" button
    Then an error notification should appear
    And the notification should say "No changes staged"
    And the commit input field should remain empty
```

**Why this works**:

- Uses `@vscode` tag to indicate VS Code context
- Scenarios describe UI interactions (clicking button)
- Verifies semantic commit format (domain-specific requirement)
- Error scenario handles edge case (no changes)

### Example 3: Documentation Build Feature

**File**: `requirements/docs/build_docs.feature`

```gherkin
@cli @io @critical
Feature: Build documentation site

  # ATDD Layer
  As a documentation maintainer
  I want to build a static documentation site with one command
  So that I can deploy updated docs quickly

  Acceptance Criteria:
  - [ ] Builds complete in under 30 seconds
  - [ ] Generates valid HTML output
  - [ ] Works without internet connection

  # BDD Layer
  @success
  Scenario: Build docs successfully
    Given Docker is running
    And documentation source files exist in "docs/"
    When I run "cc build-docs"
    Then static HTML files should be created in "dist/"
    And the files should contain valid HTML
    And the command should complete in under 30 seconds
    And the command should exit with code 0
    And stdout should contain "Build completed: X files generated"

  @io @error
  Scenario: Build fails when Docker is not running
    Given Docker is not running
    When I run "cc build-docs"
    Then the command should fail
    And stderr should contain "Docker is not running"
    And no output files should be created
    And the command should exit with code 1

  @io @error
  Scenario: Build fails with missing source files
    Given Docker is running
    And the "docs/" directory is empty
    When I run "cc build-docs"
    Then the command should fail
    And stderr should contain "No documentation source files found"
    And the command should exit with code 1
```

**Why this works**:

- Performance criterion (30s) is tested in scenario
- Uses `@io` tag for filesystem operations
- Docker dependency is explicit in Given steps
- Error scenarios cover infrastructure and data issues

### Example 4: Error Handling Feature

**File**: `requirements/cli/handle_config_errors.feature`

```gherkin
@cli @io @error
Feature: Handle configuration file errors gracefully

  # ATDD Layer
  As a CLI user
  I want clear error messages when configuration is invalid
  So that I can quickly fix issues without debugging

  Acceptance Criteria:
  - [ ] Identifies specific line number with syntax error
  - [ ] Suggests valid syntax or correction
  - [ ] Exits with non-zero code

  # BDD Layer
  @error
  Scenario: Handle YAML syntax error
    Given a file "cc.yaml" exists with invalid YAML on line 5
    When I run "cc validate"
    Then the command should fail
    And stderr should contain "Syntax error at line 5"
    And stderr should contain "Expected key:value format"
    And the command should exit with code 1

  @error
  Scenario: Handle missing required field
    Given a file "cc.yaml" exists without a "name" field
    When I run "cc validate"
    Then the command should fail
    And stderr should contain "Missing required field: name"
    And stderr should contain an example of valid syntax
    And the command should exit with code 1

  @error
  Scenario: Handle invalid field type
    Given a file "cc.yaml" has "version" set to a string instead of number
    When I run "cc validate"
    Then the command should fail
    And stderr should contain "Invalid type for 'version'"
    And stderr should contain "Expected: number, got: string"
    And the command should exit with code 1
```

**Why this works**:

- Entire feature focuses on error handling (tagged `@error`)
- Each scenario tests different error type
- Verifies helpful error messages (line numbers, suggestions)
- All scenarios verify exit codes

## Common Patterns

### CLI Command Execution

```gherkin
Scenario: Run command with flags
  Given I am in the project directory
  When I run "cc deploy --env production --verbose"
  Then the command should succeed
  And stdout should contain deployment progress
  And the command should exit with code 0
```

### File System Operations

```gherkin
Scenario: Create files and directories
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And a directory named "src/" should exist
  And the file should contain valid YAML
```

### Error Output Validation

```gherkin
Scenario: Handle invalid input
  Given I am in the project directory
  When I run "cc deploy --env invalid"
  Then the command should fail
  And stderr should contain "Invalid environment: invalid"
  And stderr should contain "Valid options: development, staging, production"
  And the command should exit with code 1
```

### Performance Validation

```gherkin
Scenario: Complete operation within time limit
  Given I have a project with 100 modules
  When I run "cc build"
  Then the command should complete in under 5 seconds
  And the command should exit with code 0
```

### Multi-Step Workflows

```gherkin
Scenario: Initialize and configure project
  Given I am in an empty folder
  When I run "cc init"
  And I run "cc config set name my-project"
  And I run "cc config set version 1.0.0"
  Then the file "cc.yaml" should contain "name: my-project"
  And the file "cc.yaml" should contain "version: 1.0.0"
```

## Validation Checklist

Use this checklist when reviewing BDD scenarios:

### Feature-Level

- [ ] Feature has appropriate tags (@cli, @io, @integration, @critical)
- [ ] Feature name is descriptive and clear
- [ ] File is saved in correct module folder
- [ ] File name follows naming convention (snake_case)
- [ ] ATDD layer exists above BDD scenarios

### Scenario-Level

- [ ] Scenario has appropriate tags (@success, @error)
- [ ] Scenario name clearly describes the behavior
- [ ] Uses Given/When/Then structure
- [ ] Steps are from user perspective (not implementation)
- [ ] Each step is concise (max 2 lines)
- [ ] Verifies observable outcomes (files, output, exit codes)
- [ ] Error scenarios check stderr (not stdout)
- [ ] All scenarios include exit code verification
- [ ] Scenarios are executable/testable

### Coverage

- [ ] At least one @success scenario (happy path)
- [ ] At least one @error scenario (failure case)
- [ ] All acceptance criteria have corresponding scenarios
- [ ] Edge cases are covered
- [ ] Total scenarios: 2-4 per feature (not too many)

## Migration

### Enhancing Legacy BDD Files

Some existing `.feature` files may need enhancement.

#### Before (Minimal BDD)

```gherkin
@cli
Feature: Initialize project

  Scenario: Run init command
    When I run "cc init"
    Then it should work
```

#### After (Complete BDD)

```gherkin
@cli @flag @critical
Feature: Initialize a new project

  # ATDD Layer: Added business context
  As a developer
  I want to initialize a CLI project with a single command
  So that I can quickly start development

  Acceptance Criteria:
  - [ ] Creates project directory structure
  - [ ] Generates valid configuration file

  # BDD Layer: Enhanced scenarios
  @success
  Scenario: Initialize in current directory
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And the file should contain valid YAML
    And a directory named "src/" should exist
    And the command should exit with code 0
    And stdout should contain "Project initialized successfully"

  @error
  Scenario: Initialize in non-empty directory
    Given I am in a directory containing files
    When I run "cc init"
    Then the command should fail
    And stderr should contain "Directory must be empty"
    And no configuration file should be created
    And the command should exit with code 1
```

**Improvements**:

- Added ATDD layer with business context
- Made assertions specific (not "it should work")
- Added Given steps (preconditions)
- Verified files, output, exit codes
- Added error scenario
- Added appropriate tags

## Integration with ATDD and TDD

### From ATDD to BDD

Each ATDD acceptance criterion should have corresponding BDD scenarios.

**ATDD Criterion**:

```text
- [ ] Creates project directory structure
```

**BDD Scenarios**:

```gherkin
@success
Scenario: Initialize creates required directories
  Given I am in an empty folder
  When I run "cc init"
  Then a directory named "src/" should exist
  And a directory named "tests/" should exist
  And a directory named "docs/" should exist
```

### From BDD to TDD

BDD scenarios inform unit test implementation.

**BDD Scenario**:

```gherkin
Scenario: Initialize in current directory
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
```

**TDD Unit Tests** (see [TDD guide](./tdd.md) for language-specific examples):

```text
# Feature: init_project
test_init_creates_config_file() {
    tmp_dir = create_temp_dir()
    init_project(tmp_dir)

    config_path = join_path(tmp_dir, "cc.yaml")
    assert file_exists(config_path), "Config file was not created"
}
```

## Related Resources

- **[ATDD Guide](./atdd.md)** - Define business value and acceptance criteria
- **[TDD Guide](./tdd.md)** - Implement features with unit tests
- **[Testing Overview](./index.md)** - Understand the complete testing strategy

---

**Next**: Learn how to implement scenarios with [TDD unit tests](./tdd.md).
