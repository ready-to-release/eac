# Best Practices

{{ page_breadcrumb() }}

> **Testing patterns and conventions for Go/Godog**

Follow these best practices to write clear, maintainable, and reliable tests.

---

## Test Naming

### Convention

**Pattern**: `Test<Function>_<Scenario>_<ExpectedResult>`

### Good Examples

- `TestParseConfig_WithValidYAML_ShouldSucceed`
- `TestCreateUser_WithExistingEmail_ShouldReturnError`
- `TestCalculateTotal_WithDiscount_ReturnsDiscountedAmount`
- `TestValidateInput_WhenEmpty_ReturnsFalse`

### Bad Examples

- `TestParse` (too vague)
- `TestParseConfigSuccess` (doesn't describe scenario)
- `Test1`, `Test2` (meaningless)

---

## Table-Driven Tests

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

---

## Test Isolation

### Use Subtests for Isolation

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

### Use `t.TempDir()` for Filesystem Tests

```go
func TestCreateFile(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    path := filepath.Join(tmpDir, "test.txt")

    // Test filesystem operations
}
```

### Clean Up Resources

```go
func TestDatabaseOperation(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close() // Always clean up

    // Test database operations
}
```

---

## Test Organization

### Group Related Tests in Same File

```
config_test.go          # All config-related tests
config_parse_test.go    # Config parsing tests specifically
config_validate_test.go # Config validation tests
```

### Use Build Tags for Special Tests

```go
//go:build integration
// +build integration

package tests

// Integration tests only run with: go test -tags=integration
```

---

## Arrange-Act-Assert Pattern

Structure tests clearly with AAA pattern:

```go
func TestCreateConfig_InEmptyDirectory_ShouldSucceed(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "r2r.yaml")
    config := DefaultConfig()

    // Act
    err := CreateConfig(configPath, config)

    // Assert
    if err != nil {
        t.Fatalf("CreateConfig failed: %v", err)
    }
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Errorf("config file not created")
    }
}
```

---

## Error Messages

### Provide Context in Errors

**Bad**:
```go
if err != nil {
    t.Fatal("failed")
}
```

**Good**:
```go
if err != nil {
    t.Fatalf("CreateConfig failed: %v", err)
}
```

### Show Expected vs Actual

**Bad**:
```go
if got != want {
    t.Error("values don't match")
}
```

**Good**:
```go
if got != want {
    t.Errorf("want %v, got %v", want, got)
}
```

---

## Test Helpers

### Extract Common Setup

```go
func setupTestDirectory(t *testing.T) string {
    t.Helper() // Mark as helper for better error reporting
    tmpDir := t.TempDir()
    if err := os.Chdir(tmpDir); err != nil {
        t.Fatalf("failed to change directory: %v", err)
    }
    return tmpDir
}

func TestCreateFile(t *testing.T) {
    dir := setupTestDirectory(t)
    // Test implementation
}
```

### Use `t.Helper()` in Helper Functions

```go
func assertFileExists(t *testing.T, path string) {
    t.Helper() // Stack traces show calling test, not this function

    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Errorf("file %s does not exist", path)
    }
}
```

---

## Parallel Tests

### Enable Parallel Execution

```go
func TestParseConfig(t *testing.T) {
    t.Parallel() // Run in parallel with other parallel tests

    tests := []struct {
        name string
        // ...
    }{
        // test cases
    }

    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel

            // Test implementation
        })
    }
}
```

### Be Careful with Shared State

**Bad**:
```go
var globalState string // Shared across parallel tests - race condition!

func TestFunction(t *testing.T) {
    t.Parallel()
    globalState = "test" // RACE!
}
```

**Good**:
```go
func TestFunction(t *testing.T) {
    t.Parallel()
    localState := "test" // Local to this test
}
```

---

## Skipping Tests

### Skip with Good Reason

```go
func TestDatabaseOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Test implementation
}
```

### Skip Platform-Specific Tests

```go
func TestWindowsSpecific(t *testing.T) {
    if runtime.GOOS != "windows" {
        t.Skip("Windows-only test")
    }

    // Test implementation
}
```

---

## Coverage

### Run Coverage Analysis

```bash
# Generate coverage report
go test -cover ./...

# Generate detailed coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage by function
go tool cover -func=coverage.out
```

### Focus on Meaningful Coverage

Don't chase 100% coverage. Focus on:
- Business logic
- Error handling paths
- Edge cases
- Complex functions

Skip coverage for:
- Trivial getters/setters
- Auto-generated code
- Test helpers

---

## Benchmarking

### Write Benchmarks for Performance-Critical Code

```go
func BenchmarkParseConfig(b *testing.B) {
    input := []byte("key: value\nname: test")

    b.ResetTimer() // Reset timer after setup

    for i := 0; i < b.N; i++ {
        ParseConfig(input)
    }
}
```

### Run Benchmarks

```bash
# Run all benchmarks
go test -bench=.

# Run specific benchmark
go test -bench=BenchmarkParseConfig

# With memory allocation stats
go test -bench=. -benchmem

# Compare benchmarks
go test -bench=. -benchmem > old.txt
# Make changes
go test -bench=. -benchmem > new.txt
benchcmp old.txt new.txt
```

---

## Common Pitfalls

### Don't Use `t.FailNow()` in Goroutines

**Bad**:
```go
func TestConcurrent(t *testing.T) {
    go func() {
        t.FailNow() // WRONG - must be called from test goroutine
    }()
}
```

**Good**:
```go
func TestConcurrent(t *testing.T) {
    errCh := make(chan error, 1)
    go func() {
        errCh <- doSomething()
    }()

    if err := <-errCh; err != nil {
        t.Fatalf("goroutine failed: %v", err)
    }
}
```

### Capture Range Variables in Loops

**Bad**:
```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // tt is captured incorrectly - all subtests see last value
    })
}
```

**Good**:
```go
for _, tt := range tests {
    tt := tt // Capture range variable
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // tt is correctly captured
    })
}
```

### Don't Ignore Cleanup Errors

**Bad**:
```go
defer os.RemoveAll(tmpDir) // Ignores error
```

**Good**:
```go
defer func() {
    if err := os.RemoveAll(tmpDir); err != nil {
        t.Logf("cleanup failed: %v", err)
    }
}()
```

---

## Related Documentation

- [File Organization](./file-organization.md) - Test file structure
- [Test Levels](./test-levels.md) - Build tags and isolation
- [Step Definitions](./step-definitions.md) - Godog patterns

{{ diataxis_footer() }}
