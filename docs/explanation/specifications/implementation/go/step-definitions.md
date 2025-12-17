# Step Definitions
> **Writing and organizing Godog step definitions**

Learn how to create effective step definitions that connect Gherkin scenarios to Go code.

---

## Godog Setup

### Test Suite Setup

```go
package tests

import (
    "os"
    "testing"

    "github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"specs"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    // Register before hooks
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        // Setup before each scenario
        return ctx, nil
    })

    // Register steps
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    ctx.Step(`^the command should succeed$`, theCommandShouldSucceed)

    // Register after hooks
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        // Cleanup after each scenario
        return ctx, nil
    })
}
```

---

## Context Management

Use a context structure to maintain state across steps:

```go
type TestContext struct {
    workingDir  string
    lastCommand *exec.Cmd
    lastOutput  string
    lastError   error
}

var testContext *TestContext

func iAmInAnEmptyFolder() error {
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        return err
    }
    testContext = &TestContext{workingDir: tmpDir}
    return os.Chdir(tmpDir)
}

func iRun(command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = testContext.workingDir
    output, err := cmd.CombinedOutput()
    testContext.lastOutput = string(output)
    testContext.lastError = err
    return nil
}
```

---

## Step Definition Patterns

### Use Regex for Flexibility

```go
// Matches both "I run" and "I execute"
ctx.Step(`^I (?:run|execute) "([^"]*)"$`, iRun)

// Matches optional negation
ctx.Step(`^the command should( not)? succeed$`, theCommandShouldSucceed)
```

### Keep Steps Reusable

**Good - reusable across scenarios**:
```go
ctx.Step(`^I run "([^"]*)"$`, iRun)
ctx.Step(`^a file named "([^"]*)" should exist$`, aFileNamedShouldExist)
```

**Bad - too specific**:
```go
ctx.Step(`^I run the init command$`, iRunInitCommand)
ctx.Step(`^the r2r\.yaml file should exist$`, theR2RYamlShouldExist)
```

### Avoid Overly Generic Steps

**Bad - Too generic**:
```go
ctx.Step(`^I do something$`, iDoSomething)
ctx.Step(`^it should work$`, itShouldWork)
```

**Good - Specific and clear**:
```go
ctx.Step(`^I initialize a new project$`, iInitializeNewProject)
ctx.Step(`^the project structure should be created$`, theProjectStructureShouldBeCreated)
```

---

## Common Step Patterns

### Given Steps - Setup

```go
func iAmInAnEmptyFolder() error {
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        return err
    }
    testContext.workingDir = tmpDir
    return os.Chdir(tmpDir)
}

func aFileNamedExists(filename, content string) error {
    path := filepath.Join(testContext.workingDir, filename)
    return os.WriteFile(path, []byte(content), 0644)
}
```

### When Steps - Actions

```go
func iRun(command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = testContext.workingDir
    output, err := cmd.CombinedOutput()
    testContext.lastOutput = string(output)
    testContext.lastError = err
    return nil
}

func iDeleteTheFile(filename string) error {
    path := filepath.Join(testContext.workingDir, filename)
    return os.Remove(path)
}
```

### Then Steps - Assertions

```go
func theCommandShouldSucceed() error {
    if testContext.lastError != nil {
        return fmt.Errorf("command failed: %v\nOutput: %s",
            testContext.lastError, testContext.lastOutput)
    }
    return nil
}

func aFileNamedShouldExist(filename string) error {
    path := filepath.Join(testContext.workingDir, filename)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Errorf("file %s does not exist", filename)
    }
    return nil
}

func theOutputShouldContain(expected string) error {
    if !strings.Contains(testContext.lastOutput, expected) {
        return fmt.Errorf("output does not contain '%s'\nActual output: %s",
            expected, testContext.lastOutput)
    }
    return nil
}
```

---

## Handling Tables

Godog supports data tables in steps:

```gherkin
Given the following configuration:
  | key     | value      |
  | name    | my-project |
  | version | 1.0.0      |
```

```go
func theFollowingConfiguration(table *godog.Table) error {
    config := make(map[string]string)

    // Skip header row, iterate data rows
    for _, row := range table.Rows[1:] {
        key := row.Cells[0].Value
        value := row.Cells[1].Value
        config[key] = value
    }

    testContext.config = config
    return nil
}
```

---

## Error Handling

Always return descriptive errors:

```go
func aFileNamedShouldExist(filename string) error {
    path := filepath.Join(testContext.workingDir, filename)

    if _, err := os.Stat(path); os.IsNotExist(err) {
        // Show what was expected vs actual
        files, _ := os.ReadDir(testContext.workingDir)
        fileList := make([]string, len(files))
        for i, f := range files {
            fileList[i] = f.Name()
        }

        return fmt.Errorf("file %s does not exist\nAvailable files: %v",
            filename, fileList)
    }

    return nil
}
```

---

## File Templates

### Step Definition File Template

```go
package tests

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/cucumber/godog"
)

type TestContext struct {
    workingDir  string
    lastOutput  string
    lastError   error
}

var testContext *TestContext

func InitializeScenario(ctx *godog.ScenarioContext) {
    // Setup
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        testContext = &TestContext{}
        return ctx, nil
    })

    // Given steps
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^a file named "([^"]*)" with content "([^"]*)"$`, aFileNamedWithContent)

    // When steps
    ctx.Step(`^I run "([^"]*)"$`, iRun)

    // Then steps
    ctx.Step(`^the command should succeed$`, theCommandShouldSucceed)
    ctx.Step(`^a file named "([^"]*)" should exist$`, aFileNamedShouldExist)
    ctx.Step(`^the output should contain "([^"]*)"$`, theOutputShouldContain)

    // Cleanup
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        if testContext.workingDir != "" {
            os.RemoveAll(testContext.workingDir)
        }
        return ctx, nil
    })
}

// Step implementations

func iAmInAnEmptyFolder() error {
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        return err
    }
    testContext.workingDir = tmpDir
    return os.Chdir(tmpDir)
}

func aFileNamedWithContent(filename, content string) error {
    path := filepath.Join(testContext.workingDir, filename)
    return os.WriteFile(path, []byte(content), 0644)
}

func iRun(command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = testContext.workingDir
    output, err := cmd.CombinedOutput()
    testContext.lastOutput = string(output)
    testContext.lastError = err
    return nil
}

func theCommandShouldSucceed() error {
    if testContext.lastError != nil {
        return fmt.Errorf("command failed: %v\nOutput: %s",
            testContext.lastError, testContext.lastOutput)
    }
    return nil
}

func aFileNamedShouldExist(filename string) error {
    path := filepath.Join(testContext.workingDir, filename)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Errorf("file %s does not exist", filename)
    }
    return nil
}

func theOutputShouldContain(expected string) error {
    if !strings.Contains(testContext.lastOutput, expected) {
        return fmt.Errorf("output does not contain '%s'\nActual: %s",
            expected, testContext.lastOutput)
    }
    return nil
}
```

---

## Complete Working Example

This example shows how all the pieces fit together: feature file, step definitions, test runner, and execution.

### Project Structure

```text
myproject/
├── specs/
│   └── cli/
│       └── init-project/
│           └── specification.feature    # Gherkin scenarios
├── src/
│   └── cli/
│       └── tests/
│           ├── features_test.go        # Test runner
│           └── steps_test.go           # Step definitions
└── go.mod
```

### 1. Feature File

`specs/cli/init-project/specification.feature`:

```gherkin
@L2 @ov
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project with one command
  So that I can quickly start development

  Rule: Creates project directory structure

    @ov
    Scenario: Initialize in empty directory
      Given I am in an empty folder
      When I run "r2r init my-project"
      Then the command should succeed
      And a file named "my-project/r2r.yaml" should exist
      And a directory named "my-project/specs/" should exist

    @ov
    Scenario: Initialize in existing project fails
      Given I am in a directory with a file named "r2r.yaml"
      When I run "r2r init"
      Then the command should fail
      And the output should contain "already initialized"
```

### 2. Test Runner

`src/cli/tests/features_test.go`:

```go
//go:build L2

package tests

import (
    "os"
    "testing"

    "github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../../../specs/cli/init-project/"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}
```

### 3. Step Definitions

`src/cli/tests/steps_test.go`:

```go
//go:build L2

package tests

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/cucumber/godog"
)

// TestContext maintains state between steps
type TestContext struct {
    workingDir string
    lastOutput string
    lastError  error
    lastExitCode int
}

var testCtx *TestContext

// InitializeScenario registers step definitions and hooks
func InitializeScenario(ctx *godog.ScenarioContext) {
    // Before hook - setup before each scenario
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        testCtx = &TestContext{}
        return ctx, nil
    })

    // Given steps - setup preconditions
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I am in a directory with a file named "([^"]*)"$`, iAmInADirectoryWithFile)

    // When steps - actions
    ctx.Step(`^I run "([^"]*)"$`, iRun)

    // Then steps - assertions
    ctx.Step(`^the command should succeed$`, theCommandShouldSucceed)
    ctx.Step(`^the command should fail$`, theCommandShouldFail)
    ctx.Step(`^a file named "([^"]*)" should exist$`, aFileNamedShouldExist)
    ctx.Step(`^a directory named "([^"]*)" should exist$`, aDirectoryNamedShouldExist)
    ctx.Step(`^the output should contain "([^"]*)"$`, theOutputShouldContain)

    // After hook - cleanup after each scenario
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        if testCtx.workingDir != "" {
            os.RemoveAll(testCtx.workingDir)
        }
        return ctx, nil
    })
}

// Given step implementations

func iAmInAnEmptyFolder() error {
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        return fmt.Errorf("failed to create temp dir: %w", err)
    }
    testCtx.workingDir = tmpDir
    return os.Chdir(tmpDir)
}

func iAmInADirectoryWithFile(filename string) error {
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        return fmt.Errorf("failed to create temp dir: %w", err)
    }
    testCtx.workingDir = tmpDir

    // Create the file
    path := filepath.Join(tmpDir, filename)
    if err := os.WriteFile(path, []byte(""), 0644); err != nil {
        return fmt.Errorf("failed to create file %s: %w", filename, err)
    }

    return os.Chdir(tmpDir)
}

// When step implementations

func iRun(command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = testCtx.workingDir
    output, err := cmd.CombinedOutput()

    testCtx.lastOutput = string(output)
    testCtx.lastError = err

    if exitErr, ok := err.(*exec.ExitError); ok {
        testCtx.lastExitCode = exitErr.ExitCode()
    } else if err != nil {
        testCtx.lastExitCode = 1
    } else {
        testCtx.lastExitCode = 0
    }

    return nil // Don't return error here - let Then steps verify
}

// Then step implementations

func theCommandShouldSucceed() error {
    if testCtx.lastExitCode != 0 {
        return fmt.Errorf("expected command to succeed but got exit code %d\nOutput: %s",
            testCtx.lastExitCode, testCtx.lastOutput)
    }
    return nil
}

func theCommandShouldFail() error {
    if testCtx.lastExitCode == 0 {
        return fmt.Errorf("expected command to fail but it succeeded\nOutput: %s",
            testCtx.lastOutput)
    }
    return nil
}

func aFileNamedShouldExist(filename string) error {
    path := filepath.Join(testCtx.workingDir, filename)

    if _, err := os.Stat(path); os.IsNotExist(err) {
        // List available files to help debug
        files, _ := filepath.Glob(filepath.Join(testCtx.workingDir, "*"))
        return fmt.Errorf("file %s does not exist\nAvailable files: %v", filename, files)
    }

    return nil
}

func aDirectoryNamedShouldExist(dirname string) error {
    path := filepath.Join(testCtx.workingDir, dirname)

    info, err := os.Stat(path)
    if os.IsNotExist(err) {
        return fmt.Errorf("directory %s does not exist", dirname)
    }
    if err != nil {
        return fmt.Errorf("error checking directory %s: %w", dirname, err)
    }
    if !info.IsDir() {
        return fmt.Errorf("%s exists but is not a directory", dirname)
    }

    return nil
}

func theOutputShouldContain(expected string) error {
    if !strings.Contains(testCtx.lastOutput, expected) {
        return fmt.Errorf("expected output to contain '%s'\nActual output: %s",
            expected, testCtx.lastOutput)
    }
    return nil
}
```

### 4. Execution

**Run all tests**:

```bash
# Run all L2 integration tests
go test -tags=L2 ./src/cli/tests/

# Run with verbose output
go test -tags=L2 -v ./src/cli/tests/

# Run specific scenario by name
godog run --tags=@ov specs/cli/init-project/
```

**Expected output**:

```text
Feature: cli_init-project

  Rule: Creates project directory structure

    Scenario: Initialize in empty directory
      Given I am in an empty folder
      When I run "r2r init my-project"
      Then the command should succeed
      And a file named "my-project/r2r.yaml" should exist
      And a directory named "my-project/specs/" should exist

    Scenario: Initialize in existing project fails
      Given I am in a directory with a file named "r2r.yaml"
      When I run "r2r init"
      Then the command should fail
      And the output should contain "already initialized"

2 scenarios (2 passed)
10 steps (10 passed)
```

### Key Integration Points

1. **Build Tags**: `//go:build L2` ensures tests run only when explicitly requested
2. **Path Configuration**: `Paths: []string{"../../../specs/cli/init-project/"}` points to feature files
3. **Context Management**: `TestContext` maintains state between Given/When/Then steps
4. **Hooks**: `Before` and `After` hooks handle setup and cleanup
5. **Step Registration**: `InitializeScenario` connects Gherkin steps to Go functions
6. **Error Handling**: Descriptive errors show expected vs actual for debugging

### Debugging Tips

**When scenarios fail**:

```bash
# Run with verbose output to see each step
go test -tags=L2 -v ./src/cli/tests/

# Run single scenario
godog run --tags="@ov and @cli" specs/cli/init-project/

# Check step definitions are registered
godog run --format=progress specs/cli/init-project/
```

**Common issues**:

- **Undefined step**: Add step definition to `InitializeScenario`
- **Wrong exit code**: Check `iRun` captures exit codes correctly
- **File not found**: Verify `workingDir` is set and `Chdir` succeeded
- **Context lost**: Ensure `testCtx` is initialized in `Before` hook

---

## Related Documentation

- [File Organization](./file-organization.md) - Where to place step definition files
- [Test Levels](./test-levels.md) - Understanding test isolation
- [Best Practices](./best-practices.md) - Testing patterns and conventions
