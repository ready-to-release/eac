# TDD Format Reference

Quick reference for unit test structure and Go test conventions.

---

## File Location

```text
src/<module>/<feature>/
├── <feature>.go          # Implementation
└── <feature>_test.go     # Tests
```

---

## Test File Pattern

| Language | Test File Pattern | Example |
|----------|-------------------|---------|
| **Go** | `*_test.go` | `config_test.go` |

---

## Test Function Format

### Go Test Function

```go
package <package_name>

import "testing"

// TestFunctionName describes what is being tested
// Feature: <feature_name>
func TestFunctionName(t *testing.T) {
    // Arrange: Set up test data

    // Act: Execute functionality

    // Assert: Verify outcomes
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**Components**:

- **Package**: Same as implementation package
- **Test name**: Must start with `Test`
- **Comment**: Feature traceability (required)
- **Parameter**: `t *testing.T` for standard tests
- **Structure**: Arrange-Act-Assert pattern

---

## Arrange-Act-Assert Pattern

### Structure

```go
func TestExample(t *testing.T) {
    // Arrange: Set up test data and preconditions
    input := "test input"
    expected := "expected result"

    // Act: Execute the functionality being tested
    result := FunctionUnderTest(input)

    // Assert: Verify the outcomes
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Components

| Phase | Purpose | Example |
|-------|---------|---------|
| **Arrange** | Set up test data and preconditions | Create test inputs, mock dependencies, set up temp directories |
| **Act** | Execute the functionality | Call the function/method being tested |
| **Assert** | Verify outcomes | Check return values, error states, side effects |

---

## Test Naming Conventions

### Function Names

**Pattern**: `Test<FunctionName>_<Scenario>`

**Examples**:

```go
TestCreateConfig           // Basic happy path
TestCreateConfig_FileExists // Specific error case
TestValidateConfig_EmptyName // Another error case
```

**Guidelines**:

- Start with `Test`
- Include function/method name
- Add underscore + scenario for specific cases
- Use descriptive scenario names
- CamelCase formatting

---

## Feature ID Linkage

**Purpose**: Traceability across all test layers

**Pattern**:

```text
Feature ID: cli_init_project

Used in:
- acceptance.spec → > **Feature ID**: cli_init_project
- behavior.feature → # Feature ID: cli_init_project
- step_definitions_test.go → // Feature: cli_init_project
- Unit tests → // Feature: cli_init_project
```

**In unit tests**:

```go
// TestInitProject validates project initialization logic
// Feature: init_project
func TestInitProject(t *testing.T) {
    // Test implementation
}
```

**Location**:

- Feature directory: `requirements/cli/init_project/`
- ATDD spec: `acceptance.spec`
- BDD scenarios: `behavior.feature`
- Unit test comment: `Feature: init_project`

**Searching**:

```bash
# Find all tests for a feature
grep -r "Feature: init_project" src/
```

---

## Table-Driven Tests

Use table-driven tests for multiple test cases with similar structure.

### Basic Pattern

```go
// Feature: <feature_name>
func TestFunction_Scenario(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        expected  string
        wantError bool
    }{
        {
            name:      "valid input",
            input:     "test",
            expected:  "result",
            wantError: false,
        },
        {
            name:      "empty input",
            input:     "",
            expected:  "",
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Act
            result, err := FunctionUnderTest(tt.input)

            // Assert
            if (err != nil) != tt.wantError {
                t.Errorf("error = %v, wantError %v", err, tt.wantError)
            }

            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### When to Use

- Testing multiple variations of the same scenario
- Boundary value testing
- Testing different input combinations
- Parameterized tests with similar structure

---

## Error Handling Tests

### Testing Error Cases

```go
// Feature: <feature_name>
func TestFunction_ErrorCase(t *testing.T) {
    // Arrange
    invalidInput := ""

    // Act
    _, err := FunctionUnderTest(invalidInput)

    // Assert
    if err == nil {
        t.Fatal("Expected error, got nil")
    }

    if !strings.Contains(err.Error(), "expected message") {
        t.Errorf("error = %q, want substring %q", err.Error(), "expected message")
    }
}
```

### Error Assertion Patterns

**Check error exists**:

```go
if err == nil {
    t.Fatal("Expected error, got nil")
}
```

**Check error message**:

```go
if !strings.Contains(err.Error(), "expected message") {
    t.Errorf("error = %q, want substring %q", err.Error(), "expected message")
}
```

**Check specific error type**:

```go
if !errors.Is(err, ErrExpectedError) {
    t.Errorf("error = %v, want %v", err, ErrExpectedError)
}
```

---

## File and Directory Testing

### Using Temp Directories

```go
// Feature: <feature_name>
func TestFileOperation(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir() // Automatically cleaned up
    filePath := filepath.Join(tmpDir, "test.txt")

    // Act
    err := CreateFile(filePath, "content")

    // Assert
    if err != nil {
        t.Fatalf("CreateFile failed: %v", err)
    }

    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        t.Error("File was not created")
    }

    // Verify file content
    content, err := os.ReadFile(filePath)
    if err != nil {
        t.Fatalf("Failed to read file: %v", err)
    }

    if string(content) != "content" {
        t.Errorf("got %q, want %q", string(content), "content")
    }
}
```

**Best Practices**:

- Use `t.TempDir()` for automatic cleanup
- Test file creation, modification, deletion
- Verify file permissions if relevant
- Check file content, not just existence

---

## Common Assertions

### Equality

```go
if result != expected {
    t.Errorf("got %v, want %v", result, expected)
}
```

### String Contains

```go
if !strings.Contains(result, substring) {
    t.Errorf("result = %q, want substring %q", result, substring)
}
```

### Boolean

```go
if result != true {
    t.Errorf("got %v, want true", result)
}
```

### Nil/Not Nil

```go
if result == nil {
    t.Error("expected non-nil result")
}

if result != nil {
    t.Errorf("expected nil, got %v", result)
}
```

### Slice/Array Length

```go
if len(result) != expected {
    t.Errorf("got length %d, want %d", len(result), expected)
}
```

### Fatal vs Error

**Use `t.Fatal()`** when test cannot continue:

```go
if err != nil {
    t.Fatalf("Setup failed: %v", err) // Stops test immediately
}
```

**Use `t.Error()`** when test can continue:

```go
if result != expected {
    t.Errorf("got %v, want %v", result, expected) // Continues to next assertion
}
```

---

## Test Helpers

### Setup and Teardown

```go
// Feature: <feature_name>
func TestFeature(t *testing.T) {
    // Setup
    cleanup := setupTest(t)
    defer cleanup() // Teardown

    // Test implementation
}

func setupTest(t *testing.T) func() {
    // Setup code
    tmpDir := t.TempDir()

    // Return cleanup function
    return func() {
        // Teardown code (if needed beyond t.TempDir())
    }
}
```

### Helper Functions

```go
// assertFileExists is a test helper that checks file existence
func assertFileExists(t *testing.T, path string) {
    t.Helper() // Marks this as a helper function for better error reporting

    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Errorf("File %s does not exist", path)
    }
}

// Usage in tests
func TestExample(t *testing.T) {
    // ...
    assertFileExists(t, "test.txt")
}
```

---

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...

# Run specific package
go test ./src/cli/...

# Run specific test
go test -run TestInitProject ./src/cli/...

# Run tests with race detection
go test -race ./...
```

### Coverage Commands

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage by function
go tool cover -func=coverage.out

# Coverage for specific package
go test -cover ./src/cli/...
```

### Example Output

```text
ok      github.com/simply-cli/cli/src/cli    0.123s  coverage: 85.7% of statements
ok      github.com/simply-cli/cli/src/mcp    0.089s  coverage: 92.3% of statements
```

---

## Style Guidelines

### ✅ Do

- **Add feature comment** to every test: `// Feature: feature_name`
- **Use descriptive test names**: Clear purpose
- **Write tests first** (Red-Green-Refactor)
- **Test one thing per test**: Single responsibility
- **Use table/parameterized tests** for multiple cases
- **Test edge cases**: Empty input, nil/null values, errors
- **Use setup/teardown** appropriately
- **Use temp directories** for file operations
- **Verify error messages**: Not just error existence
- **Fail fast**: Stop test when precondition fails

### ❌ Don't

- **Forget feature comment**: Always link to feature file
- **Skip Red phase**: Always see test fail first
- **Write tests after code**: TDD means tests first
- **Test implementation details**: Test behavior, not internals
- **Ignore test failures**: Fix or update immediately
- **Use external dependencies**: Mock or stub external services
- **Write flaky tests**: Tests must be deterministic
- **Skip refactor phase**: Clean code is critical

---

## Coverage Guidelines

**Target**: >80% coverage

**Focus**:

- Cover critical paths first
- Test error handling
- Test edge cases
- Test boundary conditions

**Exclude from coverage**:

- Main functions
- Generated code
- Test helpers

---

## Examples

### Example 1: Simple Function Test

**BDD Scenario** (from `requirements/cli/init_project/behavior.feature`):

```gherkin
Scenario: Initialize creates config file
  When I run "cc init"
  Then a file named "cc.yaml" should be created
  And the file should contain valid YAML
```

**Go Implementation**:

```go
package config

import (
    "os"
    "path/filepath"
    "testing"
)

// TestCreateConfig validates config file creation
// Feature: init_project
func TestCreateConfig(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "cc.yaml")

    // Act
    err := CreateConfig(configPath)

    // Assert
    if err != nil {
        t.Fatalf("CreateConfig failed: %v", err)
    }

    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Error("Config file was not created")
    }
}
```

### Example 2: Table-Driven Test

```go
// Feature: handle_config_errors
func TestValidateConfig_InvalidTypes(t *testing.T) {
    tests := []struct {
        name      string
        config    Config
        wantError string
    }{
        {
            name:      "version as string",
            config:    Config{Version: "1.0.0"},
            wantError: "Invalid type for 'version'",
        },
        {
            name:      "valid config",
            config:    Config{Name: "test", Version: 1.0},
            wantError: "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateConfig(tt.config)

            if tt.wantError == "" && err != nil {
                t.Errorf("unexpected error: %v", err)
            }

            if tt.wantError != "" {
                if err == nil {
                    t.Error("expected error, got nil")
                } else if !strings.Contains(err.Error(), tt.wantError) {
                    t.Errorf("error = %q, want substring %q", err.Error(), tt.wantError)
                }
            }
        })
    }
}
```

### Example 3: Error Handling

```go
// Feature: init_project
func TestInitProject_NonEmptyDirectory(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    os.WriteFile(filepath.Join(tmpDir, "existing.txt"), []byte("content"), 0644)

    // Act
    err := InitProject(tmpDir)

    // Assert
    if err == nil {
        t.Fatal("Expected error for non-empty directory")
    }

    if !strings.Contains(err.Error(), "Directory must be empty") {
        t.Errorf("Error = %q, want substring %q", err.Error(), "Directory must be empty")
    }
}
```

---

## Related Documentation

- [ATDD Format](./atdd-format.md) - Acceptance spec format
- [BDD Format](./bdd-format.md) - Behavior scenarios format
- [TDD Concepts](../../explanation/testing/tdd-concepts.md) - Understanding TDD
- [Testing How-to Guides](../../how-to-guides/testing/index.md) - Task-oriented guides
