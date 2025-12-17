# Test Levels with Go Build Tags
> **Build tags and test isolation (L0-L4)**

Go build tags control which tests run based on isolation level.

---

## L0 Tests - Fully Isolated

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

---

## L1 Tests - Unit Tests (Default)

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

---

## L2 Tests - Integration Tests

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

---

## L3 Tests - Pre-Production (Godog)

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

---

## L4 Tests - Production Verification (Godog)

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

## Related Documentation

- [File Organization](./file-organization.md) - Directory structure and file naming
- [Testing Taxonomy](../../taxonomy/) - Complete tag reference
- [Three-Layer Approach](../../concepts/three-layer-approach.md) - Conceptual overview
