# Testing System

The testing system in EAC provides unified test execution across all module types with multiple output formats and test suite management.

## What is the Testing System?

EAC's testing system enables you to:

- **Test any module type** with a single command
- **Run test suites** for organized test execution
- **Generate reports** in Cucumber JSON or JUnit XML
- **Collect coverage** for Go modules
- **Debug failures** with detailed analysis

The system automatically detects module types and invokes appropriate test tooling.

## When to Use Test Commands

Use test commands when you need:

| Scenario              | Command             |
| --------------------- | ------------------- |
| Test a single module  | `test <module>`     |
| Test multiple modules | `test <m1> <m2>`    |
| Test all modules      | `test`              |
| Run a test suite      | `test-suite <name>` |
| List available suites | `test-list-suites`  |
| Debug test failures   | `test-debug`        |

### Common Use Cases

- **Local development** - Verify changes before commit
- **CI/CD pipelines** - Automated quality gates
- **TDD workflow** - Write specs, run tests, implement
- **Coverage reporting** - Track test coverage

## Key Concepts

### Test Output Formats

| Format        | Flag                      | Use Case                |
| ------------- | ------------------------- | ----------------------- |
| Cucumber JSON | `--as-cucumber` (default) | BDD tooling integration |
| JUnit XML     | `--as-junit`              | CI/CD systems           |
| Standard      | (verbose)                 | Local development       |

### Test Suites

Test suites group tests by purpose:

| Suite         | Purpose               | Typical Run Time |
| ------------- | --------------------- | ---------------- |
| `commit`      | Pre-commit validation | < 5 min          |
| `integration` | Integration tests     | 5-15 min         |
| `e2e`         | End-to-end tests      | 15-30 min        |
| `smoke`       | Quick health checks   | < 1 min          |

### Coverage Reports

Coverage is generated for Go modules:

```text
out/reports/coverage/
├── <module>.out      # Coverage profile
└── <module>.html     # HTML report
```

### Module Type Dispatch

| Module Type      | Test Runner | Output         |
| ---------------- | ----------- | -------------- |
| `go-cli`         | `go test`   | Go test output |
| `go-commands`    | `go test`   | Go test output |
| `go-library`     | `go test`   | Go test output |
| `go-tests`       | godog       | Cucumber JSON  |
| `specifications` | godog       | Cucumber JSON  |

## Workflow Overview

### Local Development

```bash
# Test single module (verbose)
r2r eac test eac-commands

# Test with coverage
r2r eac test eac-core --coverage

# Debug failures
r2r eac test-debug
```

### TDD Workflow

```bash
# 1. Write spec
r2r eac create-spec "User authentication"

# 2. Run tests (expect failure)
r2r eac test src-auth

# 3. Implement code
# ... edit src/auth/*.go ...

# 4. Run tests (expect pass)
r2r eac test src-auth

# 5. Check coverage
r2r eac test src-auth --coverage
```

### CI/CD Pipeline

```bash
# Run all tests with JUnit output
r2r eac test --as-junit > test-results.xml

# Run specific suite
r2r eac test-suite integration --as-junit
```

## Test Suite System

### Suite Configuration

Test suites are defined in contracts:

```yaml
# .r2r/eac/test-suites.yml
suites:
  - name: commit
    description: Pre-commit validation tests
    tags:
      - "@commit"
      - "@unit"
    timeout: 300

  - name: integration
    description: Integration tests
    tags:
      - "@integration"
    timeout: 900
```

### Suite Execution

```bash
# List available suites
r2r eac test-list-suites

# Run specific suite
r2r eac test-suite integration

# Run suite with JUnit output
r2r eac test-suite e2e --as-junit
```

## Coverage Collection

### Enabling Coverage

```bash
# Single module coverage
r2r eac test eac-core --coverage

# Multiple modules
r2r eac test r2r-cli eac-core --coverage

# All modules
r2r eac test --coverage
```

### Coverage Output

```bash
# View HTML report
open out/reports/coverage/eac-core.html

# View summary
go tool cover -func=out/reports/coverage/eac-core.out

# Check threshold
COVERAGE=$(go tool cover -func=out/reports/coverage/eac-core.out | grep total | awk '{print $3}')
```

## Debugging Test Failures

### Using test-debug

```bash
# Run tests
r2r eac test src-auth

# If failures, analyze
r2r eac test-debug
```

### Debug Output

```text
Test Failures Summary
=====================

Module: src-auth
  - TestLoginHandler: Expected status 200, got 500
    File: src/auth/login_test.go:45

  - TestTokenValidation: Token validation failed
    File: src/auth/token_test.go:78

Module: eac-core
  - TestDatabaseConnection: Connection timeout
    File: go/eac/core/db_test.go:23

Total: 3 failures across 2 modules
```

## Integration Points

### With Build

```bash
# Build then test
r2r eac build eac-core
r2r eac test eac-core
```

### With Validation

```bash
# Full CI workflow
r2r eac build && r2r eac test && r2r eac validate
```

### With Specifications

```bash
# Generate spec, test, implement
r2r eac create-spec "Feature description"
r2r eac test src-feature  # Fails
# ... implement ...
r2r eac test src-feature  # Passes
```

## Best Practices

### Do's

- **Test after build** - Ensure code compiles first
- **Use test suites** - Organize by purpose
- **Track coverage** - Monitor trends over time
- **Use JUnit in CI** - Better integration

### Don'ts

- **Don't skip failing tests** - Fix or mark as skip
- **Don't ignore coverage** - Aim for meaningful coverage
- **Don't test without building** - May get stale results

## Troubleshooting

| Problem           | Solution                           |
| ----------------- | ---------------------------------- |
| Tests not found   | Check module type in contract      |
| Timeout           | Increase timeout or optimize tests |
| Coverage missing  | Use `--coverage` flag              |
| JUnit parse error | Check XML output format            |

## Next Steps

- [Test Configuration](test-configuration.md) - Configure test settings
- [Test Commands](test-commands.md) - Full command reference

## Related Areas

- [Build](build-overview.md) - Build before testing
- [Specifications](specifications-overview.md) - BDD specs for tests
- [Validate](validate-overview.md) - Validate test tags
