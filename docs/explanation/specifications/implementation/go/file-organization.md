# File Organization

{{ page_breadcrumb() }}

> **Directory structure and file naming conventions for Go/Godog projects**

Learn how to organize specification files, step definitions, and unit tests in Go projects.

---

## Specification Files

**Location**: `specs/<module>/<feature>/specification.feature`

**Format**: Gherkin (`.feature` files)

**Tool**: Godog test runner

**Example**:

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
      When I run "r2r init"
      Then a file named "r2r.yaml" should be created
```

---

## Step Definitions

**Location**: `src/<module>/tests/steps_test.go`

**Purpose**: Step definitions connecting Gherkin scenarios to Go functions

**Naming Convention**: `steps_test.go` for Godog step definitions

**Example**:

```go
package tests

import (
    "github.com/cucumber/godog"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    ctx.Step(`^a file named "([^"]*)" should be created$`, aFileNamedShouldBeCreated)
    ctx.Step(`^the command should succeed$`, theCommandShouldSucceed)
}

func iAmInAnEmptyFolder() error {
    // Setup empty directory
    tmpDir := os.TempDir()
    os.Chdir(tmpDir)
    return nil
}

func iRun(command string) error {
    // Execute command
    cmd := exec.Command("sh", "-c", command)
    output, err := cmd.CombinedOutput()
    testContext.lastOutput = string(output)
    testContext.lastError = err
    return nil
}

func aFileNamedShouldBeCreated(filename string) error {
    // Assert file exists
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return fmt.Errorf("file %s does not exist", filename)
    }
    return nil
}
```

---

## Unit Test Files

**Location**: `src/<module>/*_test.go`

**Naming Convention**: `*_test.go` suffix (Go convention)

**Framework**: Go `testing` package

**Naming Pattern**: `Test<Function>_<Scenario>_<ExpectedResult>`

**Example**:

```go
package core

import (
    "testing"
    "path/filepath"
)

func TestCreateConfig_InEmptyDirectory_ShouldSucceed(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "r2r.yaml")

    err := CreateConfig(configPath)

    if err != nil {
        t.Fatalf("CreateConfig failed: %v", err)
    }

    // Verify file exists
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Errorf("config file not created")
    }
}

func TestCreateConfig_WhenFileExists_ShouldFail(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "r2r.yaml")

    // Create existing file
    os.WriteFile(configPath, []byte("existing"), 0644)

    err := CreateConfig(configPath)

    if err == nil {
        t.Fatal("Expected error when file exists")
    }
}
```

---

## Directory Structure

```
project/
├── specs/
│   └── <module>/
│       └── <feature>/
│           └── specification.feature    ← Gherkin specs
└── src/
    └── <module>/
        ├── *.go                          ← Production code
        ├── *_test.go                     ← L1 unit tests (default)
        ├── *_l0_test.go                  ← L0 tests (optional naming)
        └── tests/
            ├── steps_test.go             ← Godog step definitions
            └── *_integration_test.go     ← L2 integration tests
```

---

## File Naming Conventions

### Production Code

**Pattern**: `<name>.go`

**Examples**:
- `config.go` - Configuration handling
- `parser.go` - Parsing logic
- `validator.go` - Validation logic

### Unit Tests

**Pattern**: `<name>_test.go`

**Examples**:
- `config_test.go` - Tests for config.go
- `parser_test.go` - Tests for parser.go
- `validator_test.go` - Tests for validator.go

### L0 Tests (Optional)

**Pattern**: `<name>_l0_test.go`

**Purpose**: Clearly identify ultra-fast tests

**Examples**:
- `parser_l0_test.go` - L0 tests for parser
- `validator_l0_test.go` - L0 tests for validator

### Integration Tests

**Pattern**: `<name>_integration_test.go`

**Purpose**: Clearly identify integration tests

**Build Tag**: `//go:build L2`

**Examples**:
- `database_integration_test.go`
- `api_integration_test.go`

### Step Definitions

**Pattern**: `steps_test.go`

**Purpose**: Godog step definitions

**Location**: `tests/` subdirectory

---

## Test Execution Commands

### Unit Tests

```bash
# Run all unit tests (L0 + L1)
go test ./...

# Run L0 tests only (fastest)
go test -tags=L0 ./...

# Run L0 tests with verbose output
go test -tags=L0 -v ./...

# Run specific package
go test ./src/module/core/...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Run L2 integration tests
go test -tags=L2 ./...

# Run L2 with verbose output
go test -tags=L2 -v ./...
```

### BDD Scenarios (Godog)

```bash
# Run all scenarios
godog run

# Run specific test suite
godog run --tags=@ov          # Operational verification
godog run --tags=@iv          # Installation verification
godog run --tags=@L3          # Pre-production tests

# Run scenarios for specific feature
godog run specs/cli/init-project/

# Run with formatting
godog run --format=pretty
godog run --format=progress

# Run parallel
godog run --concurrency=4
```

### Combined Test Suites

```bash
# Full test suite (sequential)
go test -tags=L0 ./...
go test ./...
go test -tags=L2 ./...
godog run

# Fast feedback loop (L0 + L1 only)
go test ./...
```

---

## Related Documentation

- [Test Levels](./test-levels.md) - Build tags and test isolation (L0-L4)
- [Step Definitions](./step-definitions.md) - Writing Godog step definitions
- [Best Practices](./best-practices.md) - Go testing best practices

{{ diataxis_footer() }}
