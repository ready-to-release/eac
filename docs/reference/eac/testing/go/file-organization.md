# File Organization

Directory structure and file naming conventions for Go/Godog projects.

---

## Directory Structure

```text
project/
├── specs/
│   └── <module>/
│       └── <feature>/
│           └── specification.feature    # Gherkin specs
└── src/
    └── <module>/
        ├── *.go                          # Production code
        ├── *_test.go                     # L1 unit tests (default)
        ├── *_l0_test.go                  # L0 tests (optional naming)
        └── tests/
            ├── steps_test.go             # Godog step definitions
            └── *_integration_test.go     # L2 integration tests
```

---

## Specification Files

**Location**: `specs/<module>/<feature>/specification.feature`

**Format**: Gherkin (`.feature` files)

**Tool**: Godog test runner

---

## Step Definitions

**Location**: `src/<module>/tests/steps_test.go`

**Purpose**: Step definitions connecting Gherkin scenarios to Go functions

**Naming Convention**: `steps_test.go` for Godog step definitions

---

## File Naming Conventions

### Production Code

| Pattern     | Examples                                 |
| ----------- | ---------------------------------------- |
| `<name>.go` | `config.go`, `parser.go`, `validator.go` |

### Unit Tests

| Pattern          | Examples                           |
| ---------------- | ---------------------------------- |
| `<name>_test.go` | `config_test.go`, `parser_test.go` |

### L0 Tests (Optional)

| Pattern             | Purpose                           | Examples            |
| ------------------- | --------------------------------- | ------------------- |
| `<name>_l0_test.go` | Clearly identify ultra-fast tests | `parser_l0_test.go` |

### Integration Tests

| Pattern                      | Build Tag       | Examples                       |
| ---------------------------- | --------------- | ------------------------------ |
| `<name>_integration_test.go` | `//go:build L2` | `database_integration_test.go` |

### Step Definitions

| Pattern         | Location              | Purpose                |
| --------------- | --------------------- | ---------------------- |
| `steps_test.go` | `tests/` subdirectory | Godog step definitions |

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
