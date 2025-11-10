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

**Phases**: Arrange (setup) → Act (execute) → Assert (verify)

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

## Feature Name Linkage

**Purpose**: Traceability between specifications and unit tests

**Format**: `[module-name_feature-name]` using kebab-case (e.g., `cli_init-project`, `src-commands_design-command`)

**In unit tests**:

```go
// TestInitProject validates project initialization logic
// Feature: cli_init-project
func TestInitProject(t *testing.T) {
    // Test implementation
}
```

**Traceability**:

- Specification: `specs/cli/init-project/specification.feature` (Feature: cli_init-project)
- Unit test: `src/cli/init-project/init_test.go` with `// Feature: cli_init-project` comment
- Step definitions: `src/cli/tests/steps_test.go` with `// Feature: cli_init-project` comment

**Note**: The Feature name provides traceability across all layers without requiring separate ID comments in the specification files.

**Find all tests for a feature**:

```bash
grep -r "Feature: cli_init-project" src/
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

| Assertion | Pattern |
|-----------|---------|
| **Equality** | `if result != expected { t.Errorf("got %v, want %v", result, expected) }` |
| **String contains** | `if !strings.Contains(result, sub) { t.Errorf("result = %q, want substring %q", result, sub) }` |
| **Boolean** | `if result != true { t.Errorf("got %v, want true", result) }` |
| **Not nil** | `if result == nil { t.Error("expected non-nil result") }` |
| **Is nil** | `if result != nil { t.Errorf("expected nil, got %v", result) }` |
| **Slice length** | `if len(result) != expected { t.Errorf("got length %d, want %d", len(result), expected) }` |

**Fatal vs Error**:

- `t.Fatal()` / `t.Fatalf()` - Stops test immediately (use for setup failures)
- `t.Error()` / `t.Errorf()` - Continues test (use for assertion failures)

---

## Test Helpers

**Setup/Teardown**:

```go
func TestFeature(t *testing.T) {
    cleanup := setupTest(t)
    defer cleanup()
    // Test implementation
}

func setupTest(t *testing.T) func() {
    tmpDir := t.TempDir()
    return func() { /* teardown if needed */ }
}
```

**Helper Functions** (use `t.Helper()` for better error reporting):

```go
func assertFileExists(t *testing.T, path string) {
    t.Helper()
    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Errorf("File %s does not exist", path)
    }
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
ok      github.com/ready-to-release/eac/src/cli    0.123s  coverage: 85.7% of statements
ok      github.com/ready-to-release/eac/src/mcp    0.089s  coverage: 92.3% of statements
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

## Related Documentation

- [Gherkin Format](./gherkin-format.md) - Specification syntax
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - How ATDD/BDD/TDD work together
- [Specifications How-to Guides](../../how-to-guides/specifications/index.md) - Task-oriented guides
