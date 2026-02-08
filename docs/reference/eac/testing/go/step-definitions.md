# Step Definitions

Writing and organizing Godog step definitions.

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
        return ctx, nil
    })

    // Register steps
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    ctx.Step(`^the command should succeed$`, theCommandShouldSucceed)

    // Register after hooks
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
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
ctx.Step(`^the clie\.yaml file should exist$`, theCLIEYamlShouldExist)
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

## File Template

### Complete Step Definition File

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

    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        if testContext.workingDir != "" {
            os.RemoveAll(testContext.workingDir)
        }
        return ctx, nil
    })
}

// Step implementations...
```

---

## Related Documentation

- [File Organization](./file-organization.md) - Where to place step definition files
- [Test Levels](./test-levels.md) - Understanding test isolation
- [Best Practices](./best-practices.md) - Testing patterns and conventions
