# Go Implementation Guide

> **Implementation-specific guide for Go/Godog BDD and testing**

This guide provides Go-specific implementation details for the BDD and testing approach described in [Three-Layer Testing](../../explanation/specifications/three-layer-approach.md).

---

## Overview

This project uses:

- **Go** for production code and unit tests
- **Godog** for executing Gherkin specifications
- **Go test framework** for unit and integration tests
- **Build tags** for test level isolation

---

## File Organization

### Specification Files

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

### Step Definitions

**Location**: `src/<module>/tests/steps_test.go`

**Purpose**: Glue code connecting Gherkin scenarios to Go functions

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

### Unit Test Files

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

### Directory Structure

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

## Test Levels with Go Build Tags

Go build tags control which tests run based on isolation level.

### L0 Tests - Fully Isolated

**Build Tag**: `//go:build L0`

**Purpose**: Ultra-fast tests with zero I/O, no network, no filesystem

**Execution**: `go test -tags=L0 ./...`

**Example**:

```go
//go:build L0

package core

import "testing"

func TestParsePath_WithValidInput_ShouldSucceed(t *testing.T) {
    result := ParsePath("/foo/bar")

    if result != "/foo/bar" {
        t.Errorf("expected /foo/bar, got %s", result)
    }
}

func TestAdd_TwoNumbers_ReturnsSum(t *testing.T) {
    result := Add(2, 3)

    if result != 5 {
        t.Errorf("expected 5, got %d", result)
    }
}
```

**Characteristics**:
- Runs in microseconds
- No external dependencies
- Pure functions only
- Parallel-safe

### L1 Tests - Unit Tests (Default)

**Build Tag**: None (default)

**Purpose**: Unit tests with minimal dependencies (temp files, simple mocks)

**Execution**: `go test ./...`

**Example**:

```go
package core

import "testing"

func TestParseConfig_WithValidYAML_ShouldSucceed(t *testing.T) {
    input := []byte("key: value\nname: test")

    result, err := ParseConfig(input)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.Key != "value" {
        t.Errorf("expected value, got %s", result.Key)
    }
}

func TestCreateTempFile_WithContent_ShouldWriteFile(t *testing.T) {
    tmpDir := t.TempDir() // Uses filesystem
    path := filepath.Join(tmpDir, "test.txt")

    err := CreateFile(path, []byte("content"))

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

**Characteristics**:
- Default test level (no build tag needed)
- Can use `t.TempDir()` for filesystem
- Simple mocks allowed
- Fast execution (milliseconds)

### L2 Tests - Integration Tests

**Build Tag**: `//go:build L2`

**Purpose**: Integration tests with emulated/mocked external dependencies

**Execution**: `go test -tags=L2 ./...`

**Example**:

```go
//go:build L2

package integration

import (
    "testing"
    "github.com/testcontainers/testcontainers-go"
)

func TestDatabase_Connect_ShouldSucceed(t *testing.T) {
    // Start test container
    ctx := context.Background()
    req := testcontainers.ContainerRequest{
        Image: "postgres:14",
        ExposedPorts: []string{"5432/tcp"},
    }
    container, err := testcontainers.GenericContainer(ctx, req)
    if err != nil {
        t.Fatal(err)
    }
    defer container.Terminate(ctx)

    // Test database connection
    db, err := ConnectDB(container.GetConnectionString())
    if err != nil {
        t.Fatalf("connection failed: %v", err)
    }
    defer db.Close()
}
```

**Characteristics**:
- Uses test containers, mocked APIs
- Emulated dependencies
- Slower (seconds per test)
- Isolated from production

### L3 Tests - Pre-Production (Godog)

**Tag**: `@L3` in Gherkin scenarios OR auto-inferred from `@iv`/`@pv`

**Purpose**: Testing in production-like test environment (PLTE)

**Execution**: `godog run --tags=@L3` or `godog run --tags=@iv`

**Example**:

```gherkin
@L3 @iv
Scenario: Deploy application to PLTE
  Given a production-like test environment
  When I deploy version 1.2.3
  Then the deployment should succeed
  And health checks should pass
```

**Characteristics**:
- Real infrastructure (test environment)
- Full integration testing
- Deployment verification
- Slower (minutes)

### L4 Tests - Production Verification (Godog)

**Tag**: `@L4` in Gherkin OR auto-inferred from `@piv`/`@ppv`

**Purpose**: Smoke tests and monitoring in production

**Execution**: `godog run --tags=@L4` or `godog run --tags=@piv`

**Example**:

```gherkin
@L4 @piv
Scenario: Production health check
  Given the production environment
  When I check system health
  Then all services should be running
  And response time should be under 200ms
```

**Characteristics**:
- Runs against production
- Read-only operations
- Continuous monitoring
- Slowest (minutes to hours)

---

## Build Tag to Gherkin Tag Mapping

| Go Build Tag | Gherkin Tag | Auto-inferred from | Execution |
|--------------|-------------|-------------------|-----------|
| `//go:build L0` | `@L0` | N/A | `go test -tags=L0` |
| (none) | `@L1` | `@ov` (default) | `go test` |
| `//go:build L2` | `@L2` | (explicit only) | `go test -tags=L2` |
| N/A (Godog) | `@L3` | `@iv`, `@pv` | `godog --tags=@L3` |
| N/A (Godog) | `@L4` | `@piv`, `@ppv` | `godog --tags=@L4` |

**Note**: Godog scenarios don't use Go build tags. Test level is determined by Gherkin tags.

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

## Canon TDD Workflow in Go

### Step 1: List Behavioral Variants

```go
// Feature: cli_init-project
// Function: CreateConfig

// Behavioral variants to test:
// 1. Create config in empty directory (success)
// 2. Create config with custom values (success)
// 3. Create config when file exists (error)
// 4. Create config in read-only directory (error)
// 5. Create config with invalid YAML (error)
```

### Step 2-5: Test, Pass, Refactor, Repeat

```go
package core

import (
    "testing"
    "os"
    "path/filepath"
)

// Test 1: Create config in empty directory
func TestCreateConfig_InEmptyDirectory_ShouldSucceed(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "r2r.yaml")

    // Act
    err := CreateConfig(configPath, DefaultConfig())

    // Assert
    if err != nil {
        t.Fatalf("CreateConfig failed: %v", err)
    }

    // Verify file exists
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Errorf("config file not created")
    }
}

// Test 2: Create config with custom values
func TestCreateConfig_WithCustomValues_ShouldPersist(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "r2r.yaml")
    config := Config{Name: "custom", Version: "1.0"}

    err := CreateConfig(configPath, config)

    if err != nil {
        t.Fatalf("CreateConfig failed: %v", err)
    }

    // Read and verify
    loaded, _ := LoadConfig(configPath)
    if loaded.Name != "custom" {
        t.Errorf("expected custom, got %s", loaded.Name)
    }
}

// Test 3: File exists - should fail
func TestCreateConfig_WhenFileExists_ShouldFail(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "r2r.yaml")

    // Create existing file
    os.WriteFile(configPath, []byte("existing"), 0644)

    err := CreateConfig(configPath, DefaultConfig())

    if err == nil {
        t.Fatal("Expected error when file exists")
    }
    if !errors.Is(err, ErrFileExists) {
        t.Errorf("expected ErrFileExists, got %v", err)
    }
}

// Test 4: Read-only directory - should fail
func TestCreateConfig_InReadOnlyDirectory_ShouldFail(t *testing.T) {
    if os.Getuid() == 0 {
        t.Skip("Cannot test read-only as root")
    }

    tmpDir := t.TempDir()
    os.Chmod(tmpDir, 0555) // Read-only
    defer os.Chmod(tmpDir, 0755)
    configPath := filepath.Join(tmpDir, "r2r.yaml")

    err := CreateConfig(configPath, DefaultConfig())

    if err == nil {
        t.Fatal("Expected error with read-only directory")
    }
}
```

---

## Best Practices

### Test Naming

**Convention**: `Test<Function>_<Scenario>_<ExpectedResult>`

**Good Examples**:
- `TestParseConfig_WithValidYAML_ShouldSucceed`
- `TestCreateUser_WithExistingEmail_ShouldReturnError`
- `TestCalculateTotal_WithDiscount_ReturnsDiscountedAmount`
- `TestValidateInput_WhenEmpty_ReturnsFalse`

**Bad Examples**:
- `TestParse` (too vague)
- `TestParseConfigSuccess` (doesn't describe scenario)
- `Test1`, `Test2` (meaningless)

### Table-Driven Tests

Use table-driven tests for multiple variants of same behavior:

```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        want    Config
        wantErr bool
    }{
        {
            name:    "valid YAML",
            input:   []byte("key: value\nname: test"),
            want:    Config{Key: "value", Name: "test"},
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   []byte(""),
            want:    Config{},
            wantErr: true,
        },
        {
            name:    "invalid YAML",
            input:   []byte("{invalid}"),
            want:    Config{},
            wantErr: true,
        },
        {
            name:    "missing required field",
            input:   []byte("key: value"),
            want:    Config{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseConfig(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr %v, got error: %v", tt.wantErr, err)
            }
            if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("want %+v, got %+v", tt.want, got)
            }
        })
    }
}
```

### Step Definition Patterns

**Use regex for flexibility**:

```go
// Matches both "I run" and "I execute"
ctx.Step(`^I (?:run|execute) "([^"]*)"$`, iRun)

// Matches optional negation
ctx.Step(`^the command should( not)? succeed$`, theCommandShouldSucceed)
```

**Keep steps reusable**:

```go
// Good - reusable across scenarios
ctx.Step(`^I run "([^"]*)"$`, iRun)
ctx.Step(`^a file named "([^"]*)" should exist$`, aFileNamedShouldExist)

// Bad - too specific
ctx.Step(`^I run the init command$`, iRunInitCommand)
ctx.Step(`^the r2r\.yaml file should exist$`, theR2RYamlShouldExist)
```

**Avoid overly generic steps**:

```go
// Bad - Too generic
ctx.Step(`^I do something$`, iDoSomething)
ctx.Step(`^it should work$`, itShouldWork)

// Good - Specific and clear
ctx.Step(`^I initialize a new project$`, iInitializeNewProject)
ctx.Step(`^the project structure should be created$`, theProjectStructureShouldBeCreated)
```

### Test Isolation

**Use subtests for isolation**:

```go
func TestProjectInit(t *testing.T) {
    t.Run("creates directory structure", func(t *testing.T) {
        // Each subtest is isolated
    })

    t.Run("generates config file", func(t *testing.T) {
        // Independent from previous subtest
    })
}
```

**Use `t.TempDir()` for filesystem tests**:

```go
func TestCreateFile(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    path := filepath.Join(tmpDir, "test.txt")

    // Test filesystem operations
}
```

**Clean up resources**:

```go
func TestDatabaseOperation(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close() // Always clean up

    // Test database operations
}
```

### Test Organization

**Group related tests in same file**:

```
config_test.go          # All config-related tests
config_parse_test.go    # Config parsing tests specifically
config_validate_test.go # Config validation tests
```

**Use build tags for special tests**:

```go
//go:build integration
// +build integration

package tests

// Integration tests only run with: go test -tags=integration
```

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

### Context Management

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

## Related Documentation

### Conceptual Understanding
- [Three-Layer Testing Approach](../../explanation/specifications/three-layer-approach.md) - Conceptual overview
- [Working with Specifications](../../explanation/specifications/working-with-specifications.md) - BDD fundamentals
- [Tag Reference](../../explanation/specifications/tag-reference.md) - Tag taxonomy concepts

### Practical Guides
- [BDD Development Workflow](../../how-to-guides/eac/specifications/bdd-development-workflow.md) - Step-by-step workflow
- [Canon TDD Workflow](../specifications/canon-tdd-workflow.md) - Kent Beck's TDD approach

### Organizational
- [Gherkin File Organization](../../explanation/specifications/gherkin-concepts.md) - Specification structure
- [Example Mapping](../../explanation/specifications/example-mapping.md) - Requirements discovery

---

## Quick Reference

### Essential Commands

```bash
# Unit tests
go test ./...                    # L0 + L1
go test -tags=L0 ./...          # L0 only
go test -tags=L2 ./...          # L2 integration

# BDD scenarios
godog run                        # All scenarios
godog run --tags=@ov            # Operational verification
godog run --tags=@cli           # CLI features only

# Coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### File Templates

**Unit Test**:
```go
package mypackage

import "testing"

func TestFunction_Scenario_ExpectedResult(t *testing.T) {
    // Arrange

    // Act

    // Assert
}
```

**Step Definition**:
```go
package tests

import "github.com/cucumber/godog"

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^step pattern$`, stepImplementation)
}

func stepImplementation() error {
    // Implementation
    return nil
}
```
