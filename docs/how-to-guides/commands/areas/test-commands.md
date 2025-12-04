# Test Commands

Command reference for EAC's testing system.

## Quick Reference

| Command | Description |
|---------|-------------|
| `test` | Test one or more modules by moniker |
| `test suite` | Run tests for a specific test suite |
| `test list-suites` | List all available test suites |
| `test debug` | Parse test results and list all failures |

---

## test

Test one or more modules by moniker.

### Synopsis

```bash
r2r eac test [module1] [module2] ... [options]
```

### Description

Runs tests for modules based on their type. Automatically dispatches to the correct test runner (Go test, godog, etc.).

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `module...` | No | Module monikers to test (defaults to all) |

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--as-cucumber` | | bool | true | Output in Cucumber JSON format |
| `--as-junit` | | bool | false | Output in JUnit XML format |
| `--suite` | `-s` | string | `commit` | Filter tests by suite |
| `--coverage` | | bool | false | Generate coverage reports |

### Examples

```bash
# Test all modules
r2r eac test

# Test single module (verbose output)
r2r eac test eac-commands

# Test multiple modules (parallel)
r2r eac test r2r-cli eac-core

# Test with JUnit output
r2r eac test eac-core --as-junit

# Test specific suite
r2r eac test eac-commands --suite integration

# Test with coverage
r2r eac test eac-core --coverage

# Combine flags
r2r eac test eac-core --coverage --as-junit
```

### Output (Single Module - Verbose)

```text
Testing module: eac-commands (type: go-commands)
Module root: go/eac/commands
Test suite: commit

=== RUN   TestBuildCommand
--- PASS: TestBuildCommand (0.25s)
=== RUN   TestValidateCommand
--- PASS: TestValidateCommand (0.18s)
=== RUN   TestShowModules
--- PASS: TestShowModules (0.12s)

PASS
ok      github.com/ready-to-release/eac/go/eac/commands    0.55s
```

### Output (Multiple Modules - Summary)

```text
Testing 3 modules in parallel...

  ✓ eac-core (2.3s) - 15 tests passed
  ✓ eac-commands (3.1s) - 23 tests passed
  ✓ r2r-cli (4.5s) - 8 tests passed

✅ All 3 modules passed (46 tests)
Total time: 4.5s
```

### Output (With Failures)

```text
Testing module: src-auth

=== RUN   TestLoginHandler
    login_test.go:45: Expected status 200, got 500
--- FAIL: TestLoginHandler (0.15s)
=== RUN   TestTokenValidation
--- PASS: TestTokenValidation (0.08s)

FAIL
exit status 1

✗ 1 of 2 tests failed
```

### Supported Module Types

| Type | Test Runner | Output |
|------|-------------|--------|
| `go-cli` | `go test` | Go test output |
| `go-commands` | `go test` | Go test output |
| `go-library` | `go test` | Go test output |
| `go-mcp` | `go test` | Go test output |
| `go-tests` | godog | Cucumber JSON |
| `specifications` | godog | Cucumber JSON |

### Exit Codes

| Code | Description |
|------|-------------|
| 0 | All tests passed |
| 1 | One or more tests failed |
| 2 | Module not found |
| 3 | Test runner error |

---

## test suite

Run tests for a specific test suite.

### Synopsis

```bash
r2r eac test suite <name> [options]
```

### Description

Runs all tests matching the specified suite's tag configuration.

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Suite name (commit, integration, e2e, smoke) |

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--as-cucumber` | | bool | true | Cucumber JSON output |
| `--as-junit` | | bool | false | JUnit XML output |
| `--parallel` | `-p` | bool | suite default | Run tests in parallel |

### Examples

```bash
# Run commit suite
r2r eac test suite commit

# Run integration suite
r2r eac test suite integration

# Run e2e suite with JUnit output
r2r eac test suite e2e --as-junit

# Run smoke tests
r2r eac test suite smoke
```

### Output

```text
Running test suite: integration
Tags: @integration
Timeout: 900s

Discovering tests...
  Found 12 scenarios across 4 features

Running tests...
  ✓ Feature: API Authentication (3 scenarios)
  ✓ Feature: User Management (4 scenarios)
  ✓ Feature: Data Export (2 scenarios)
  ✓ Feature: Notifications (3 scenarios)

✅ Suite passed: 12 scenarios in 45.2s
```

### Exit Codes

| Code | Description |
|------|-------------|
| 0 | Suite passed |
| 1 | Suite failed |
| 2 | Suite not found |
| 3 | Timeout exceeded |

---

## test list-suites

List all available test suites.

### Synopsis

```bash
r2r eac test list-suites
```

### Description

Displays all configured test suites with their descriptions and test counts.

### Examples

```bash
r2r eac test list-suites
```

### Output

```text
Available Test Suites
═══════════════════════════════════════════════════════

│ Suite       │ Tests │ Timeout │ Description                    │
├─────────────┼───────┼─────────┼────────────────────────────────┤
│ commit      │ 45    │ 5m      │ Pre-commit validation tests    │
│ integration │ 23    │ 15m     │ Integration tests              │
│ e2e         │ 12    │ 30m     │ End-to-end tests               │
│ smoke       │ 8     │ 1m      │ Quick health checks            │

Run a suite with: r2r eac test suite <name>
```

### Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error reading configuration |

---

## test debug

Parse test results and list all failures.

### Synopsis

```bash
r2r eac test debug [options]
```

### Description

Analyzes test output files to identify and display all test failures with detailed information.

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--verbose` | `-v` | bool | false | Show detailed failure information |

### Examples

```bash
# List all failures
r2r eac test debug

# Verbose output
r2r eac test debug --verbose
```

### Output

```text
Test Failures Summary
═══════════════════════════════════════════════════════

Module: src-auth
─────────────────────────────────────────────────────
  ✗ TestLoginHandler
    File: src/auth/login_test.go:45
    Error: Expected status 200, got 500

  ✗ TestTokenValidation
    File: src/auth/token_test.go:78
    Error: Token validation failed: invalid signature

Module: eac-core
─────────────────────────────────────────────────────
  ✗ TestDatabaseConnection
    File: go/eac/core/db_test.go:23
    Error: Connection timeout after 30s

Summary:
  Total failures: 3
  Modules affected: 2

Fix the failures and re-run: r2r eac test
```

### Verbose Output

```text
Test Failures Summary (Verbose)
═══════════════════════════════════════════════════════

Module: src-auth
─────────────────────────────────────────────────────

✗ TestLoginHandler
  File: src/auth/login_test.go:45
  Error: Expected status 200, got 500

  Stack trace:
    login_test.go:45: handler.ServeHTTP(rr, req)
    login_test.go:46: assert.Equal(t, 200, rr.Code)

  Test code:
    func TestLoginHandler(t *testing.T) {
        req := httptest.NewRequest("POST", "/login", body)
        rr := httptest.NewRecorder()
        handler.ServeHTTP(rr, req)
        assert.Equal(t, 200, rr.Code)  // <- Failure here
    }
```

### Exit Codes

| Code | Description |
|------|-------------|
| 0 | No failures found |
| 1 | Failures found |
| 2 | No test results to analyze |

---

## Common Workflows

### Local Development

```bash
# Quick test cycle
r2r eac build src-auth
r2r eac test src-auth

# Debug failures
r2r eac test debug --verbose
```

### TDD Workflow

```bash
# 1. Create spec
r2r eac create spec "User can reset password"

# 2. Run tests (expect failure)
r2r eac test src-auth

# 3. Implement
# ... code ...

# 4. Run tests (expect pass)
r2r eac test src-auth

# 5. Verify coverage
r2r eac test src-auth --coverage
```

### CI/CD Pipeline

```bash
# Build and test with JUnit output
r2r eac build
r2r eac test --as-junit > test-results.xml

# Upload to CI system
# ... CI-specific upload ...
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
r2r eac test suite commit || exit 1
echo "✅ Tests passed"
```

---

## Coverage Workflows

### Generate Coverage

```bash
# Single module
r2r eac test eac-core --coverage

# All modules
r2r eac test --coverage
```

### View Coverage

```bash
# HTML report
open out/reports/coverage/eac-core.html

# Console summary
go tool cover -func=out/reports/coverage/eac-core.out
```

### Coverage in CI

```bash
# Generate and upload
r2r eac test --coverage
codecov -f out/reports/coverage/*.out
```

---

## Integration Patterns

### GitHub Actions

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5

      - name: Run tests
        run: r2r eac test --as-junit > test-results.xml

      - name: Publish results
        uses: dorny/test-reporter@v1
        if: always()
        with:
          name: Tests
          path: test-results.xml
          reporter: java-junit
```

### Makefile

```makefile
.PHONY: test test-coverage test-suite

test:
  r2r eac test

test-coverage:
  r2r eac test --coverage

test-suite:
  r2r eac test suite $(SUITE)
```

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Tests not found | Check module type in contract |
| Timeout | Increase timeout in suite config |
| No coverage | Add `--coverage` flag |
| Suite not found | Check `test-suites.yml` |
| Parse error | Check test output format |

---

## Related Documentation

- [Test Overview](test-overview.md) - Testing concepts
- [Test Configuration](test-configuration.md) - Configuration options
- [Build Commands](build-commands.md) - Build before testing
