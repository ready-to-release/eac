# BDD: Behavior-Driven Development

**[<- Back to Testing Overview](./index.md)**

## What is BDD?

**Behavior-Driven Development (BDD)** is a specification technique that describes user-facing behavior through concrete examples. In this project, BDD scenarios define **observable CLI behaviors** using Gherkin syntax (Given/When/Then).

This project uses **[Godog](https://github.com/cucumber/godog)** to execute BDD scenarios in `behavior.feature` files.

### Key Characteristics

| Aspect | Description |
|--------|-------------|
| **Who** | Developers and testers (informed by stakeholders) |
| **When** | During feature specification (after ATDD workshop) |
| **Format** | Gherkin scenarios (Given/When/Then) |
| **Location** | `behavior.feature` files in `requirements/<module>/<feature>/` |
| **Purpose** | Specify and execute observable CLI behaviors |
| **Tool** | [Godog](https://github.com/cucumber/godog) - Cucumber for Go |

## BDD with Godog

Godog is the official Cucumber implementation for Go, allowing us to write and execute Gherkin scenarios.

### Why Godog for BDD?

- **Gherkin syntax** - Industry standard Given/When/Then format
- **Executable** - Scenarios become automated tests
- **Go native** - Integrates seamlessly with Go projects
- **Rich features** - Tables, scenario outlines, hooks, tags
- **Clear reporting** - Multiple output formats (pretty, junit, cucumber)

## behavior.feature File Structure

BDD content appears in `behavior.feature` files within each feature directory, **separate from** `acceptance.spec` (ATDD).

### Template Structure

**File**: `requirements/<module>/<feature_name>/behavior.feature`

```gherkin
# Feature ID: <module>_<feature_name>
# Acceptance Spec: acceptance.spec
# Module: <Module>

@<module> @critical @<feature_name>
Feature: [Feature Name]

  Background:
    Given [common precondition for all scenarios]

  @success @ac1
  Scenario: [Happy path scenario name]
    Given [precondition]
    When [action]
    Then [observable outcome]
    And [additional verification]

  @error @ac1
  Scenario: [Error scenario name]
    Given [precondition]
    When [invalid action]
    Then [error behavior]
    And [error message verification]
```

### Component Breakdown

#### 1. Metadata Header (Comments)

**Purpose**: Links this feature to acceptance.spec and provides traceability

**Example**:

```gherkin
# Feature ID: cli_init_project
# Acceptance Spec: acceptance.spec
# Module: CLI
```

#### 2. Feature Tags

**Purpose**: Categorize features by testing level and priority

**Placement**: Feature line (applies to all scenarios)

**Example**:

```gherkin
@cli @critical @init_project
Feature: Initialize project command behavior
```

**Common feature tags**: `@cli`, `@vscode`, `@io`, `@integration`, `@critical`

#### 3. Background (Optional)

**Purpose**: Common setup for all scenarios in the feature

**Example**:

```gherkin
Background:
  Given I am in a clean test environment
  And all previous test artifacts are removed
```

#### 4. Scenarios

**Purpose**: Describe one specific user interaction

**Format**: Given/When/Then steps

**Example**:

```gherkin
@success @ac1
Scenario: Initialize in empty directory
  Given I am in an empty folder
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the command should exit with code 0
```

#### 5. Scenario Tags

**Purpose**: Categorize individual scenarios by type

**Placement**: Scenario line (applies to that scenario only)

**Common scenario tags**: `@success`, `@error`, `@flag`, `@io`

## Godog Installation and Setup

### Installation

```bash
# Install Godog
go get github.com/cucumber/godog/cmd/godog@latest

# Verify installation
godog version
```

### Project Setup

Create Godog configuration:

**File**: `godog.yaml` (project root)

```yaml
default:
  paths:
    - requirements/**/behavior.feature
  format: pretty,junit:test-results/godog.xml,html:test-results/godog.html
  tags: ~@wip
  strict: true
  stop-on-failure: false
```

### Step Definitions

Godog steps are implemented in Go test files within each feature directory.

**File**: `requirements/<module>/<feature>/step_definitions_test.go`

```go
// Feature: cli_init_project
// Type: BDD (Godog)
package init_project_test

import (
    "context"
    "github.com/cucumber/godog"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
    // Register step definitions
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    ctx.Step(`^a file named "([^"]*)" should be created$`, aFileNamedShouldBeCreated)
    ctx.Step(`^the command should exit with code (\d+)$`, theCommandShouldExitWithCode)
    ctx.Step(`^stderr should contain "([^"]*)"$`, stderrShouldContain)
}

func iAmInAnEmptyFolder() error {
    // Implementation
    return nil
}

func iRun(command string) error {
    // Implementation: execute command
    return nil
}

func aFileNamedShouldBeCreated(filename string) error {
    // Implementation: verify file exists
    return nil
}

func theCommandShouldExitWithCode(code int) error {
    // Implementation: verify exit code
    return nil
}

func stderrShouldContain(message string) error {
    // Implementation: verify stderr
    return nil
}
```

## From Example Mapping to Scenarios

After the ATDD team completes an Example Mapping workshop (see [ATDD Guide](./atdd.md#example-mapping-workshop-collaborative-discovery)), you'll have **Green Cards** representing concrete examples. These become your Godog scenarios.

### Green Card to Gherkin Conversion

Each Green Card describes a specific example of a rule in action. Convert these to Gherkin scenarios by:

1. **Extract the context** → GIVEN
2. **Identify the action** → WHEN
3. **Determine the outcome** → THEN

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
```
Empty folder -> run init -> creates directories
```

**Becomes**:

```gherkin
@success @ac1
Scenario: Initialize in empty directory creates structure
  Given I am in an empty folder
  When I run "cc init"
  Then directories "src/", "tests/", "docs/" should exist
  And the command should exit with code 0
```

#### Example 2: Error Case

**Green Card**:
```
Existing project -> run init -> error "already initialized"
```

**Becomes**:

```gherkin
@error @ac1
Scenario: Initialize in existing project shows error
  Given I am in a directory with "cc.yaml"
  When I run "cc init"
  Then the command should fail
  And stderr should contain "already initialized"
  And the command should exit with code 1
```

#### Example 3: With Flags

**Green Card**:
```
With --name flag -> cc.yaml contains custom name
```

**Becomes**:

```gherkin
@flag @success @ac2
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

**Result**: 3 separate scenarios, all testing configuration generation, all tagged `@ac2`

### Tips for Effective Conversion

- **Add exit code verification**: Green Cards may not specify exit codes, but BDD scenarios must
- **Specify stdout vs stderr**: Make output assertions explicit
- **Add scenario tags**: Use @success, @error, @flag, @io based on behavior
- **Make assertions observable**: Convert "creates directories" to specific directory names
- **Keep 1:1 mapping**: One Green Card = One Scenario (avoid combining)

### Relationship to ATDD Layer

Green Cards validate Blue Cards (acceptance criteria):

```text
ATDD Layer (acceptance.spec):
  AC1: Creates project directory structure

BDD Layer (behavior.feature):
  @ac1 Scenario: Initialize creates required directories
  @ac1 Scenario: Initialize in existing project shows error
```

Each Blue Card should have 2-4 Green Cards testing different aspects of that rule.

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

## Workflow

### BDD Development Process with Godog

```text
BDD Workflow (8 steps)

1. Read ATDD Context
   +-- Review acceptance.spec file
   +-- Study user story (Yellow Card)
   +-- Study acceptance criteria (Blue Cards)
   +-- Understand business value

2. Gather Green Cards from Example Mapping
   +-- Retrieve Green Cards from workshop
   +-- Each Green Card is a concrete example
   +-- Format: [Context] -> [Action] -> [Result]
   +-- Expect 2-4 Green Cards per Blue Card (AC)

3. Convert Each Green Card to Scenario
   +-- Extract context -> GIVEN step
   +-- Identify action -> WHEN step
   +-- Determine outcome -> THEN step
   +-- Create one scenario per Green Card
   +-- Tag with @ac1, @ac2, etc. to link to acceptance criteria

4. Enhance Given/When/Then Steps
   +-- Given: Set up preconditions clearly
   +-- When: Describe user action (quote CLI commands)
   +-- Then: Verify observable outcome
   +-- Add: Exit code verification
   +-- Add: stdout/stderr assertions
   +-- Keep steps concise (max 2 lines)

5. Add Appropriate Tags
   +-- Feature-level: @cli, @io, @integration
   +-- Scenario-level: @success, @error, @flag
   +-- Link to AC: @ac1, @ac2, @ac3, etc.
   +-- Priority: @critical if business-critical

6. Implement Godog Step Definitions
   +-- Create step_definitions_test.go
   +-- Implement step functions
   +-- Wire up to actual CLI code
   +-- Run godog to verify
   +-- Iterate until all scenarios pass

7. Review Scenario Completeness
   +-- Does it verify acceptance criteria (Blue Cards)?
   +-- Is it executable/testable?
   +-- Are steps clear and unambiguous?
   +-- Are all edge cases covered?
   +-- Each Green Card has matching scenario?
   +-- Exit codes specified?

8. Save and Execute
   +-- Save behavior.feature in feature directory
   +-- Run: godog requirements/<module>/<feature>/behavior.feature
   +-- Fix failures
   +-- Generate reports
   +-- Proceed to TDD (unit tests)
```

### Prerequisites

Before starting BDD:

- acceptance.spec exists (ATDD layer complete)
- Example Mapping workshop completed (Green Cards available)
- Feature directory created
- Feature ID defined
- Godog installed and configured

### Outputs

After completing BDD:

- One scenario per Green Card (2-4 scenarios per AC)
- Scenarios written in Gherkin (Given/When/Then)
- Scenarios tagged appropriately (@success, @error, @ac1, etc.)
- All acceptance criteria (Blue Cards) have corresponding scenarios
- behavior.feature saved in feature directory
- step_definitions_test.go with Godog implementations
- All scenarios executable with `godog run`

## Running Godog Tests

### Execute BDD Scenarios

```bash
# Run all behavior tests
godog requirements/**/behavior.feature

# Run tests for specific module
godog requirements/cli/**/behavior.feature

# Run tests for specific feature
godog requirements/cli/init_project/behavior.feature

# Run with tags
godog --tags="@critical" requirements/**/behavior.feature
godog --tags="@success && @cli" requirements/**/behavior.feature
godog --tags="~@wip" requirements/**/behavior.feature  # Exclude @wip

# Generate reports
godog --format=pretty --format=junit:test-results/godog.xml requirements/**/behavior.feature
```

### Godog Output

```text
Feature: Initialize project command behavior

  Scenario: Initialize in empty directory creates structure       # requirements/cli/init_project/behavior.feature:12
    Given I am in an empty folder                                 # step_definitions_test.go:15
    When I run "cc init"                                          # step_definitions_test.go:20
    Then a file named "cc.yaml" should be created                 # step_definitions_test.go:25
    And a directory named "src/" should exist                     # step_definitions_test.go:30
    And the command should exit with code 0                       # step_definitions_test.go:35

  Scenario: Initialize in existing project shows error            # requirements/cli/init_project/behavior.feature:19
    Given I am in a directory with "cc.yaml"                      # step_definitions_test.go:40
    When I run "cc init"                                          # step_definitions_test.go:20
    Then the command should fail                                  # step_definitions_test.go:45
    And stderr should contain "already initialized"               # step_definitions_test.go:50

2 scenarios (2 passed)
10 steps (10 passed)
125.456µs
```

## Complete Example

### Workshop Output (Green Cards)

```text
[BLUE-1] Creates project directory structure
  [GREEN-1a] Empty folder -> init -> creates src/, tests/, docs/
  [GREEN-1b] Existing project -> init -> error "already initialized"

[BLUE-2] Generates valid configuration file
  [GREEN-2a] New project -> init -> creates cc.yaml with defaults
  [GREEN-2b] With --name flag -> cc.yaml contains custom name
```

### behavior.feature (Godog)

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
    And stdout should contain "Project initialized successfully"

  # Green Card 1b: Existing project -> init -> error
  @error @ac1
  Scenario: Initialize in existing project shows error
    Given I am in a directory with "cc.yaml"
    When I run "cc init"
    Then the command should fail
    And stderr should contain "already initialized"
    And stderr should contain the path to the current directory
    And no new files should be created
    And the command should exit with code 1

  # Green Card 2a: New project -> creates cc.yaml with defaults
  @success @ac2
  Scenario: Initialize creates configuration with defaults
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And the file should contain valid YAML
    And the file should contain "name"
    And the file should contain "version: 1.0.0"
    And the command should exit with code 0

  # Green Card 2b: With --name flag -> contains custom name
  @flag @success @ac2
  Scenario: Initialize with custom project name
    Given I am in an empty folder
    When I run "cc init --name my-project"
    Then a file named "cc.yaml" should be created
    And the file should contain "name: my-project"
    And the command should exit with code 0
```

### step_definitions_test.go (Godog)

**File**: `requirements/cli/init_project/step_definitions_test.go`

```go
// Feature: cli_init_project
// Type: BDD (Godog)
package init_project_test

import (
    "context"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/cucumber/godog"
)

// Test context to store state between steps
type testContext struct {
    workDir     string
    stdout      string
    stderr      string
    exitCode    int
    lastError   error
}

var ctx *testContext

func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"behavior.feature"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}

func InitializeScenario(sc *godog.ScenarioContext) {
    // Initialize context for each scenario
    sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        ctx = &testContext{}
        return ctx, nil
    })

    // Register steps
    sc.Step(`^I am in a clean test environment$`, iAmInACleanTestEnvironment)
    sc.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    sc.Step(`^I am in a directory with "([^"]*)"$`, iAmInADirectoryWith)
    sc.Step(`^I run "([^"]*)"$`, iRun)
    sc.Step(`^a file named "([^"]*)" should be created$`, aFileNamedShouldBeCreated)
    sc.Step(`^a directory named "([^"]*)" should exist$`, aDirectoryNamedShouldExist)
    sc.Step(`^the command should exit with code (\d+)$`, theCommandShouldExitWithCode)
    sc.Step(`^the command should fail$`, theCommandShouldFail)
    sc.Step(`^stdout should contain "([^"]*)"$`, stdoutShouldContain)
    sc.Step(`^stderr should contain "([^"]*)"$`, stderrShouldContain)
    sc.Step(`^the file should contain "([^"]*)"$`, theFileShouldContain)
    sc.Step(`^the file should contain valid YAML$`, theFileShouldContainValidYAML)
    sc.Step(`^no new files should be created$`, noNewFilesShouldBeCreated)

    // Cleanup after scenario
    sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        if ctx.workDir != "" {
            os.RemoveAll(ctx.workDir)
        }
        return ctx, nil
    })
}

func iAmInACleanTestEnvironment() error {
    ctx = &testContext{}
    return nil
}

func iAmInAnEmptyFolder() error {
    var err error
    ctx.workDir, err = os.MkdirTemp("", "godog-test-*")
    return err
}

func iAmInADirectoryWith(filename string) error {
    var err error
    ctx.workDir, err = os.MkdirTemp("", "godog-test-*")
    if err != nil {
        return err
    }

    // Create the specified file
    filePath := filepath.Join(ctx.workDir, filename)
    return os.WriteFile(filePath, []byte("existing content"), 0644)
}

func iRun(command string) error {
    // Parse command and execute
    // Store stdout, stderr, exitCode in ctx
    // Implementation would call actual CLI code
    return nil
}

func aFileNamedShouldBeCreated(filename string) error {
    filePath := filepath.Join(ctx.workDir, filename)
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return fmt.Errorf("file %s does not exist", filename)
    }
    return nil
}

func aDirectoryNamedShouldExist(dirname string) error {
    dirPath := filepath.Join(ctx.workDir, dirname)
    info, err := os.Stat(dirPath)
    if os.IsNotExist(err) {
        return fmt.Errorf("directory %s does not exist", dirname)
    }
    if !info.IsDir() {
        return fmt.Errorf("%s is not a directory", dirname)
    }
    return nil
}

func theCommandShouldExitWithCode(expectedCode int) error {
    if ctx.exitCode != expectedCode {
        return fmt.Errorf("expected exit code %d, got %d", expectedCode, ctx.exitCode)
    }
    return nil
}

func theCommandShouldFail() error {
    if ctx.exitCode == 0 {
        return fmt.Errorf("expected command to fail, but it succeeded")
    }
    return nil
}

func stdoutShouldContain(expected string) error {
    if !strings.Contains(ctx.stdout, expected) {
        return fmt.Errorf("stdout does not contain %q", expected)
    }
    return nil
}

func stderrShouldContain(expected string) error {
    if !strings.Contains(ctx.stderr, expected) {
        return fmt.Errorf("stderr does not contain %q, got: %s", expected, ctx.stderr)
    }
    return nil
}

func theFileShouldContain(content string) error {
    // Implementation: read last referenced file and check content
    return nil
}

func theFileShouldContainValidYAML() error {
    // Implementation: read last referenced file and validate YAML
    return nil
}

func noNewFilesShouldBeCreated() error {
    // Implementation: verify no files were created in workDir
    return nil
}
```

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

### Scenario-Level Tags

Apply at **Scenario level** (scenario-specific):

| Tag | Description | Usage |
|-----|-------------|-------|
| `@success` | Normal successful operation | Happy path scenarios |
| `@error` | Negative or invalid input scenario | Error handling scenarios |
| `@critical` | Business-critical functionality | Important acceptance criteria |
| `@wip` | Work in progress | Exclude from CI runs |
| `@ac1`, `@ac2`, etc. | Links to acceptance criterion | Maps to acceptance.spec |

### Tagging Examples

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

## Related Resources

- **[ATDD Guide](./atdd.md)** - Define business value and acceptance criteria with Gauge
- **[TDD Guide](./tdd.md)** - Implement features with unit tests
- **[Testing Overview](./index.md)** - Complete testing strategy
- **[Godog Documentation](https://github.com/cucumber/godog)** - Official Godog docs

---

**Next**: Implement scenarios with [TDD unit tests](./tdd.md).
