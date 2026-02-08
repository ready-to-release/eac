# Go/Godog Overview

Introduction to Go/Godog BDD testing setup and installation.

---

## Stack

This project uses:

- **Go** for production code and unit tests
- **Godog** for executing Gherkin specifications
- **Go test framework** for unit and integration tests
- **Build tags** for test level isolation

---

## Why Godog?

**Godog** is the Go implementation of Cucumber, enabling BDD with Gherkin syntax.

| Feature               | Description                                                    |
| --------------------- | -------------------------------------------------------------- |
| Native Go Integration | Works seamlessly with `go test` framework                      |
| Gherkin Support       | Full support for Feature, Rule, Scenario, Examples             |
| Table Tests           | Natural mapping from Gherkin examples to Go table-driven tests |
| Parallel Execution    | Built-in support for concurrent scenario execution             |
| Output Formats        | Multiple formats (pretty, progress, JSON, JUnit)               |

---

## Quick Start

### 1. Install Godog

```bash
go get github.com/cucumber/godog/cmd/godog@latest
```

### 2. Create a Feature File

**Location**: `specs/<module>/<feature>/specification.feature`

```gherkin
@cli @critical
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project
  So that I can quickly start development

  Rule: Creates project directory structure

    @ov
    Scenario: Initialize in empty directory
      Given I am in an empty folder
      When I run "clie init"
      Then a file named "clie.yaml" should be created
```

### 3. Implement Step Definitions

**Location**: `src/<module>/tests/steps_test.go`

```go
package tests

import (
    "github.com/cucumber/godog"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    ctx.Step(`^a file named "([^"]*)" should be created$`, aFileNamedShouldBeCreated)
}

func iAmInAnEmptyFolder() error {
    tmpDir := os.TempDir()
    os.Chdir(tmpDir)
    return nil
}

func iRun(command string) error {
    cmd := exec.Command("sh", "-c", command)
    output, err := cmd.CombinedOutput()
    testContext.lastOutput = string(output)
    testContext.lastError = err
    return nil
}

func aFileNamedShouldBeCreated(filename string) error {
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return fmt.Errorf("file %s does not exist", filename)
    }
    return nil
}
```

### 4. Run Tests

```bash
# Run all scenarios
godog run

# Run with pretty output
godog run --format=pretty

# Run specific scenarios
godog run --tags=@ov
```

---

## Related Documentation

- [File Organization](./file-organization.md) - Directory structure
- [Test Levels](./test-levels.md) - Build tags (L0-L4)
- [Step Definitions](./step-definitions.md) - Step definition patterns
- [Best Practices](./best-practices.md) - Go testing conventions
