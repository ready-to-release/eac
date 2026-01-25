# Best Practices

Testing patterns and conventions for Go/Godog.

---

## Test Naming

### Convention

**Pattern**: `Test<Function>_<Scenario>_<ExpectedResult>`

| Good                                                      | Bad                                    |
| --------------------------------------------------------- | -------------------------------------- |
| `TestParseConfig_WithValidYAML_ShouldSucceed`             | `TestParse` (too vague)                |
| `TestCreateUser_WithExistingEmail_ShouldReturnError`      | `TestParseConfigSuccess` (no scenario) |
| `TestCalculateTotal_WithDiscount_ReturnsDiscountedAmount` | `Test1`, `Test2` (meaningless)         |

---

## Table-Driven Tests

Use table-driven tests for multiple variants:

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

### Use Subtests

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

### Use `t.TempDir()`

```go
func TestCreateFile(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    path := filepath.Join(tmpDir, "test.txt")
    // Test filesystem operations
}
```

---

## Arrange-Act-Assert Pattern

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

### Use `t.Helper()`

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

```go
func TestParseConfig(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name string
    }{
        // test cases
    }

    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // Test implementation
        })
    }
}
```

---

## Coverage

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

### Focus Areas

**Do test**: Business logic, error handling paths, edge cases, complex functions

**Skip testing**: Trivial getters/setters, auto-generated code, test helpers

---

## Benchmarking

```go
func BenchmarkParseConfig(b *testing.B) {
    input := []byte("key: value\nname: test")
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ParseConfig(input)
    }
}
```

```bash
# Run all benchmarks
go test -bench=.

# With memory allocation stats
go test -bench=. -benchmem
```

---

## Common Pitfalls

### Capture Range Variables

**Bad**:

```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // tt is captured incorrectly
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

### Don't Use `t.FailNow()` in Goroutines

**Bad**:

```go
go func() {
    t.FailNow() // WRONG
}()
```

**Good**:

```go
errCh := make(chan error, 1)
go func() {
    errCh <- doSomething()
}()

if err := <-errCh; err != nil {
    t.Fatalf("goroutine failed: %v", err)
}
```

---

## Related Documentation

- [File Organization](./file-organization.md) - Test file structure
- [Test Levels](./test-levels.md) - Build tags and isolation
- [Step Definitions](./step-definitions.md) - Godog patterns
