# Create Feature Spec

Create acceptance.spec and behavior.feature files for a new feature.

---

## Prerequisites

- [Gauge installed and configured](./setup-gauge.md)
- [Godog installed and configured](./setup-godog.md)
- Example Mapping workshop completed (see [Run Example Mapping](./run-example-mapping.md))
- Workshop cards (Yellow, Blue, Green) available

---

## Overview

This guide walks through creating both test specification files for a feature:

1. **acceptance.spec** (Gauge/ATDD) - Business requirements
2. **behavior.feature** (Godog/BDD) - Executable scenarios

---

## Step 1: Create Feature Directory

Determine the feature module and name, then create the directory:

**Format**: `specs/<module>/<feature_name>/`

**Example**:

```bash
mkdir -p specs/cli/init_project
cd specs/cli/init_project
```

**Module names**: `cli`, `vscode`, `docs`, `mcp`

---

## Step 2: Create acceptance.spec (ATDD)

### Determine Feature ID

**Format**: `<module>_<feature_name>`

**Example**: `cli_init_project`

### Create File

Create `specs/<module>/<feature_name>/acceptance.spec`:

```bash
touch acceptance.spec
```

### Write File Structure

Use your Yellow Card (user story) and Blue Cards (acceptance criteria) from the Example Mapping workshop.

**Template**:

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

* [Criterion 1 from Blue Card 1]
* [Criterion 2 from Blue Card 2]
* [Criterion 3 from Blue Card 3]

## Acceptance Tests

### AC1: [Criterion 1]
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: <tags>

* [Gauge step 1]
* [Gauge step 2]
* [Gauge step 3]
* [Verification step]

### AC2: [Criterion 2]
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: <tags>

* [Gauge step 1]
* [Gauge step 2]
* [Verification step]
```

### Complete Example

**File**: `specs/cli/init_project/acceptance.spec`

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
* Command completes in under 2 seconds
* Works on Linux, macOS, and Windows

## Acceptance Tests

### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Verify "cc.yaml" file exists
* Verify "src/" directory exists
* Verify "tests/" directory exists
* Verify "docs/" directory exists
* Verify command exit code is "0"

### AC2: Generates valid configuration file
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Read "cc.yaml" file contents
* Verify YAML is valid
* Verify YAML contains key "name"
* Verify YAML contains key "version"

### AC3: Command completes in under 2 seconds
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: performance

* Create empty test directory
* Start performance timer
* Run "cc init" command
* Stop performance timer
* Assert execution time is less than "2" seconds

### AC4: Works on Linux, macOS, and Windows
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: cross-platform

* Detect current operating system
* Create empty test directory
* Run "cc init" command
* Verify paths use OS-specific separators
* Verify command succeeds
```

---

## Step 3: Create behavior.feature (BDD)

### Use Same Feature ID

Ensure Feature ID matches the one in acceptance.spec: `cli_init_project`

### Create File

Create `specs/<module>/<feature_name>/behavior.feature`:

```bash
touch behavior.feature
```

### Write File Structure

Use your Green Cards from the Example Mapping workshop to create scenarios.

**Template**:

```gherkin
# Feature ID: <module>_<feature_name>
# Acceptance Spec: acceptance.spec
# Module: <Module>

@<module> @<priority> @<feature_name>
Feature: [Feature Name]

  Background:
    Given [common precondition for all scenarios]

  # Green Card 1a
  @success @ac1
  Scenario: [Happy path scenario name]
    Given [precondition]
    When [action]
    Then [observable outcome]
    And [additional verification]

  # Green Card 1b
  @error @ac1
  Scenario: [Error scenario name]
    Given [precondition]
    When [invalid action]
    Then [error behavior]
    And [error message verification]

  # Green Card 2a
  @success @ac2
  Scenario: [Another scenario]
    Given [precondition]
    When [action]
    Then [outcome]
```

### Complete Example

**File**: `specs/cli/init_project/behavior.feature`

```gherkin
# Feature ID: cli_init_project
# Acceptance Spec: acceptance.spec
# Module: CLI

@cli @critical @init_project
Feature: Initialize project command behavior

  Background:
    Given I am in a clean test environment

  # Green Card 1a: Empty folder → init → creates dirs
  @success @ac1
  Scenario: Initialize in empty directory creates structure
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And a directory named "src/" should exist
    And a directory named "tests/" should exist
    And a directory named "docs/" should exist
    And the command should exit with code 0

  # Green Card 1b: Existing project → init → error
  @error @ac1
  Scenario: Initialize in existing project shows error
    Given I am in a directory with "cc.yaml"
    When I run "cc init"
    Then the command should fail
    And stderr should contain "already initialized"
    And the command should exit with code 1

  # Green Card 2a: New project → creates cc.yaml with defaults
  @success @ac2
  Scenario: Initialize creates configuration with defaults
    Given I am in an empty folder
    When I run "cc init"
    Then a file named "cc.yaml" should be created
    And the file should contain valid YAML
    And the file should contain "version: 1.0.0"
    And the command should exit with code 0

  # Green Card 2b: With --name flag → contains custom name
  @flag @success @ac2
  Scenario: Initialize with custom name flag
    Given I am in an empty folder
    When I run "cc init --name my-project"
    Then a file named "cc.yaml" should be created
    And the file should contain "name: my-project"
    And the command should exit with code 0

  # Green Card 3a: Standard project → measure time → <2s
  @success @ac3 @PV
  Scenario: Initialize completes within performance threshold
    Given I am in an empty folder
    When I run "cc init"
    Then the command should complete within 2 seconds
    And the command should exit with code 0

  # Green Card 4a: Linux/macOS/Windows → init → succeeds
  @success @ac4
  Scenario: Initialize works on current operating system
    Given I am in an empty folder
    When I run "cc init"
    Then the command should succeed
    And the paths should use OS-specific separators
```

---

## Step 4: Validate Files

### Check Feature ID Consistency

Ensure Feature ID is identical in both files:

```bash
# Should show the same Feature ID in both files
grep "Feature ID" acceptance.spec
grep "Feature ID" behavior.feature
```

### Check Acceptance Criteria Tags

Verify @ac tags in behavior.feature match acceptance criteria:

```bash
# Should show @ac1, @ac2, @ac3, @ac4
grep "@ac" behavior.feature
```

### Check File Links

Verify cross-references:

```bash
# acceptance.spec should link to behavior.feature
grep "behavior.feature" acceptance.spec

# behavior.feature should link to acceptance.spec
grep "acceptance.spec" behavior.feature
```

---

## Step 5: Implement Gauge Steps

Create `acceptance_test.go` to implement Gauge steps.

**File**: `specs/cli/init_project/acceptance_test.go`

```go
// Feature: cli_init_project
// Type: ATDD (Gauge)
package init_project_test

import (
    "os"
    "path/filepath"
    "time"

    "github.com/getgauge-contrib/gauge-go/gauge"
)

var (
    testDir       string
    exitCode      int
    startTime     time.Time
    executionTime time.Duration
)

func init() {
    // Register Gauge steps
    gauge.Step("Create empty test directory", createEmptyTestDirectory)
    gauge.Step("Run <command> command", runCommand)
    gauge.Step("Verify <file> file exists", verifyFileExists)
    gauge.Step("Verify <dir> directory exists", verifyDirectoryExists)
    gauge.Step("Verify command exit code is <code>", verifyExitCode)
    gauge.Step("Start performance timer", startPerformanceTimer)
    gauge.Step("Stop performance timer", stopPerformanceTimer)
    gauge.Step("Assert execution time is less than <seconds> seconds", assertExecutionTime)
}

func createEmptyTestDirectory() {
    var err error
    testDir, err = os.MkdirTemp("", "gauge-test-*")
    if err != nil {
        gauge.GetScenarioStore()["error"] = err.Error()
    }
}

func runCommand(command string) {
    // Implementation: execute command in testDir
    // Store exitCode
}

func verifyFileExists(file string) {
    path := filepath.Join(testDir, file)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        gauge.GetScenarioStore()["error"] = "File does not exist: " + file
    }
}

func verifyDirectoryExists(dir string) {
    path := filepath.Join(testDir, dir)
    info, err := os.Stat(path)
    if os.IsNotExist(err) || !info.IsDir() {
        gauge.GetScenarioStore()["error"] = "Directory does not exist: " + dir
    }
}

func verifyExitCode(expectedCode string) {
    // Implementation: check exitCode
}

func startPerformanceTimer() {
    startTime = time.Now()
}

func stopPerformanceTimer() {
    executionTime = time.Since(startTime)
}

func assertExecutionTime(seconds string) {
    // Implementation: check executionTime < threshold
}
```

---

## Step 6: Implement Godog Steps

Create `step_definitions_test.go` to implement Godog steps.

**File**: `specs/cli/init_project/step_definitions_test.go`

```go
// Feature: cli_init_project
// Type: BDD (Godog)
package init_project_test

import (
    "context"
    "testing"
    "github.com/cucumber/godog"
)

type testContext struct {
    workDir string
    stdout  string
    stderr  string
    exitCode int
}

func iAmInACleanTestEnvironment(ctx context.Context) (context.Context, error) {
    // Implementation
    return ctx, nil
}

func iAmInAnEmptyFolder(ctx context.Context) (context.Context, error) {
    // Implementation
    return ctx, nil
}

func iAmInADirectoryWith(ctx context.Context, filename string) (context.Context, error) {
    // Implementation
    return ctx, nil
}

func iRun(ctx context.Context, command string) (context.Context, error) {
    // Implementation
    return ctx, nil
}

func aFileNamedShouldBeCreated(ctx context.Context, filename string) error {
    // Implementation
    return nil
}

func aDirectoryNamedShouldExist(ctx context.Context, dirname string) error {
    // Implementation
    return nil
}

func theCommandShouldExitWithCode(ctx context.Context, code int) error {
    // Implementation
    return nil
}

func theCommandShouldFail(ctx context.Context) error {
    // Implementation
    return nil
}

func stderrShouldContain(ctx context.Context, text string) error {
    // Implementation
    return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^I am in a clean test environment$`, iAmInACleanTestEnvironment)
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I am in a directory with "([^"]*)"$`, iAmInADirectoryWith)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    ctx.Step(`^a file named "([^"]*)" should be created$`, aFileNamedShouldBeCreated)
    ctx.Step(`^a directory named "([^"]*)" should exist$`, aDirectoryNamedShouldExist)
    ctx.Step(`^the command should exit with code (\d+)$`, theCommandShouldExitWithCode)
    ctx.Step(`^the command should fail$`, theCommandShouldFail)
    ctx.Step(`^stderr should contain "([^"]*)"$`, stderrShouldContain)
}

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
```

---

## Step 7: Run Tests

### Run Gauge Tests

```bash
gauge run specs/cli/init_project/
```

### Run Godog Tests

```bash
godog specs/cli/init_project/behavior.feature
```

### Run via Go Test

```bash
cd specs/cli/init_project
go test -v
cd ../../..
```

---

## Step 8: Track Questions (Optional)

If you had Red Cards from Example Mapping, create an issues tracker:

**File**: `specs/cli/init_project/issues.md`

```markdown
# Open Questions

## RED-1: What if cc.yaml already exists?

**Status**: Open
**Raised**: 2025-11-03
**Decision needed by**: Product Owner
**Resolution**: TBD

## RED-2: Should we support a --force flag?

**Status**: Open
**Raised**: 2025-11-03
**Decision needed by**: Development Team
**Resolution**: TBD
```

---

## Checklist

Before considering the feature spec complete:

- [ ] Feature directory created
- [ ] acceptance.spec file created with user story and acceptance criteria
- [ ] behavior.feature file created with scenarios from Green Cards
- [ ] Feature ID is identical in both files
- [ ] @ac tags in behavior.feature match acceptance criteria
- [ ] Files link to each other correctly
- [ ] acceptance_test.go created with Gauge step implementations
- [ ] step_definitions_test.go created with Godog step definitions
- [ ] Both test files run without errors
- [ ] Red Cards tracked in issues.md (if any)

---

## Next Steps

- ✅ Feature specification files created
- **Next**: Implement the feature using [TDD](../../reference/testing/tdd-format.md)
- **Then**: [Run Tests](./run-tests.md) to validate

---

## Related Documentation

- [ATDD Format](../../reference/testing/atdd-format.md) - Specification format
- [BDD Format](../../reference/testing/bdd-format.md) - Gherkin syntax
- [Run Example Mapping](./run-example-mapping.md) - Workshop guide
- [Three-Layer Approach](../../explanation/testing/three-layer-approach.md) - Understanding the approach
