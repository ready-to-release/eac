# Run Tests

Execute tests at all three layers (ATDD, BDD, TDD).

---

## Overview

This guide covers running tests for:

- **ATDD** - Acceptance tests with Gauge
- **BDD** - Behavior tests with Godog
- **TDD** - Unit tests with Go test

---

## Prerequisites

- [Gauge installed](./setup-gauge.md)
- [Godog installed](./setup-godog.md)
- Test files created (see [Create Feature Spec](./create-feature-spec.md))

---

## Quick Reference

```bash
# Run all tests
gauge run requirements/
godog requirements/**/behavior.feature
go test ./...

# Run specific feature
gauge run requirements/cli/init_project/
godog requirements/cli/init_project/behavior.feature
go test ./src/cli/...

# Run with tags
gauge run --tags "critical" requirements/
godog --tags="@critical" requirements/**/behavior.feature

# Generate reports
gauge run --html-report requirements/
godog --format=junit:test-results/godog.xml requirements/**/behavior.feature
go test -coverprofile=coverage.out ./...
```

---

## ATDD: Run Gauge Tests

### Run All Acceptance Tests

```bash
gauge run requirements/
```

**Output**:

```text
# Initialize Project

## Acceptance Tests

  ### AC1: Creates project directory structure     ✓
  ### AC2: Generates valid configuration file     ✓

Specifications: 1 executed     1 passed     0 failed     0 skipped
Scenarios:      2 executed     2 passed     0 failed     0 skipped
```

### Run Specific Module

```bash
gauge run requirements/cli/
gauge run requirements/vscode/
```

### Run Specific Feature

```bash
gauge run requirements/cli/init_project/
```

### Run with Tags

```bash
# Run only critical tests
gauge run --tags "critical" requirements/

# Run only performance tests
gauge run --tags "performance" requirements/

# Run critical CLI tests (AND)
gauge run --tags "cli & critical" requirements/

# Exclude WIP tests (NOT)
gauge run --tags "!wip" requirements/
```

### Generate HTML Report

```bash
gauge run --html-report requirements/
```

**Report location**: `test-results/gauge/html-report/index.html`

**Open report**:

```bash
# macOS
open test-results/gauge/html-report/index.html

# Linux
xdg-open test-results/gauge/html-report/index.html

# Windows
start test-results/gauge/html-report/index.html
```

### Run in Parallel

```bash
# Run with 4 parallel streams
gauge run -p=4 requirements/

# Use all available cores
gauge run -p requirements/
```

### Validate Without Running

```bash
# Check specs are valid
gauge validate requirements/

# Validate specific feature
gauge validate requirements/cli/init_project/
```

---

## BDD: Run Godog Tests

### Run All Behavior Tests

```bash
godog requirements/**/behavior.feature
```

**Output**:

```text
Feature: Initialize project command behavior

  Scenario: Initialize in empty directory creates structure
    Given I am in an empty folder
    When I run "cc init"
    Then directories should exist                            ✓

2 scenarios (2 passed)
6 steps (6 passed)
```

### Run Specific Module

```bash
godog requirements/cli/**/behavior.feature
godog requirements/vscode/**/behavior.feature
```

### Run Specific Feature

```bash
godog requirements/cli/init_project/behavior.feature
```

### Run with Tags

```bash
# Run only success scenarios
godog --tags="@success" requirements/**/behavior.feature

# Run only error scenarios
godog --tags="@error" requirements/**/behavior.feature

# Run critical tests (AND)
godog --tags="@success && @cli" requirements/**/behavior.feature

# Exclude WIP tests (NOT)
godog --tags="~@wip" requirements/**/behavior.feature

# Run specific acceptance criterion
godog --tags="@ac1" requirements/**/behavior.feature
```

### Run by Verification Type

```bash
# Installation Verification only
godog --tags="@IV" requirements/**/behavior.feature

# Performance Verification only
godog --tags="@PV" requirements/**/behavior.feature

# Operational Verification only (default scenarios)
godog --tags="~@IV && ~@PV" requirements/**/behavior.feature
```

### Generate Reports

#### JUnit XML (for CI/CD)

```bash
godog --format=junit:test-results/godog.xml requirements/**/behavior.feature
```

#### Cucumber JSON

```bash
godog --format=cucumber:test-results/godog.json requirements/**/behavior.feature
```

#### Multiple Formats

```bash
godog --format=pretty \
      --format=junit:test-results/godog.xml \
      --format=cucumber:test-results/godog.json \
      requirements/**/behavior.feature
```

#### Separate Reports by Verification Type

```bash
# Installation Verification report
godog --tags="@IV" --format=junit:test-results/iv-godog.xml requirements/**/behavior.feature

# Performance Verification report
godog --tags="@PV" --format=junit:test-results/pv-godog.xml requirements/**/behavior.feature

# Operational Verification report
godog --tags="~@IV && ~@PV" --format=junit:test-results/ov-godog.xml requirements/**/behavior.feature
```

### Run via Go Test

```bash
# Run as Go test
cd requirements/cli/init_project
go test -v
cd ../../..
```

---

## TDD: Run Unit Tests

### Run All Unit Tests

```bash
go test ./...
```

**Output**:

```text
ok      github.com/simply-cli/cli/src/cli    0.123s
ok      github.com/simply-cli/cli/src/mcp    0.089s
```

### Run with Verbose Output

```bash
go test -v ./...
```

### Run Specific Package

```bash
go test ./src/cli/...
go test ./src/mcp/...
```

### Run Specific Test

```bash
# By test name
go test -run TestInitProject ./src/cli/...

# By regex pattern
go test -run "^TestInit" ./src/cli/...
```

### Run with Coverage

```bash
# Show coverage percentage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage by function
go tool cover -func=coverage.out
```

**Example coverage output**:

```text
ok      github.com/simply-cli/cli/src/cli    0.123s  coverage: 85.7% of statements
ok      github.com/simply-cli/cli/src/mcp    0.089s  coverage: 92.3% of statements
```

### Run with Race Detection

```bash
go test -race ./...
```

### Run Tests by Feature ID

```bash
# Find all tests for a feature
grep -r "Feature: init_project" src/ tests/

# Run specific package containing the feature
go test ./src/cli/init/...
```

---

## Run All Tests (All Layers)

### Sequential Execution

```bash
#!/bin/bash
# Run all tests in sequence

echo "Running ATDD acceptance tests..."
gauge run requirements/

echo "Running BDD behavior tests..."
godog requirements/**/behavior.feature

echo "Running TDD unit tests..."
go test ./...

echo "All tests completed!"
```

### Create Test Script

**File**: `scripts/run-all-tests.sh`

```bash
#!/bin/bash
set -e

echo "=== Running All Tests ==="
echo

echo "=== ATDD: Gauge Acceptance Tests ==="
gauge run --html-report requirements/
echo

echo "=== BDD: Godog Behavior Tests ==="
godog --format=pretty \
      --format=junit:test-results/godog.xml \
      requirements/**/behavior.feature
echo

echo "=== TDD: Go Unit Tests ==="
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
echo

echo "=== Test Summary ==="
echo "Gauge report: test-results/gauge/html-report/index.html"
echo "Godog report: test-results/godog.xml"
echo "Coverage report: coverage.out"
echo
echo "✅ All tests completed!"
```

**Make executable**:

```bash
chmod +x scripts/run-all-tests.sh
```

**Run**:

```bash
./scripts/run-all-tests.sh
```

---

## CI/CD Integration

### GitHub Actions Example

**File**: `.github/workflows/test.yml`

```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Gauge
        run: |
          curl -SsL https://downloads.gauge.org/stable | sh
          gauge install go

      - name: Install Godog
        run: go install github.com/cucumber/godog/cmd/godog@latest

      - name: Run ATDD Tests (Gauge)
        run: gauge run --html-report requirements/

      - name: Run BDD Tests (Godog)
        run: |
          godog --format=pretty \
                --format=junit:test-results/godog.xml \
                requirements/**/behavior.feature

      - name: Run TDD Tests (Go)
        run: go test -v -coverprofile=coverage.out ./...

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: |
            test-results/
            coverage.out
```

---

## Troubleshooting

### Gauge: Command not found

**Problem**: `gauge: command not found`

**Solution**:

```bash
# Check installation
which gauge

# Reinstall if needed
curl -SsL https://downloads.gauge.org/stable | sh
gauge install go
```

### Godog: No step definitions found

**Problem**: Steps show as undefined

**Solution**:

- Verify step definitions are registered in `InitializeScenario`
- Check regex patterns match step text exactly
- Run with `--verbose` to see matching details:

```bash
godog --verbose requirements/cli/init_project/behavior.feature
```

### Go Test: No tests to run

**Problem**: `no test files`

**Solution**:

- Ensure test files end with `_test.go`
- Ensure test functions start with `Test`
- Check you're in the right directory

### Tests Failing After Refactor

**Problem**: BDD/ATDD tests fail after code refactor

**Solution**:

- This is expected - tests verify behavior
- If behavior changed, update scenarios
- If behavior same, fix implementation to pass tests

---

## Best Practices

### During Development

Run tests frequently:

```bash
# After each small change
go test ./src/cli/...

# Before committing
gauge run requirements/cli/feature/
godog requirements/cli/feature/behavior.feature
go test ./...
```

### Before Commit

```bash
# Run all tests
./scripts/run-all-tests.sh

# Check coverage
go test -cover ./...
```

### In CI/CD

- Run all three test layers
- Generate and archive reports
- Fail build if any tests fail
- Track coverage trends

---

## Next Steps

- ✅ Tests are running successfully
- **Continuous**: Run tests during development
- **Before release**: Ensure all tests pass
- **Monitor**: Track test coverage over time

---

## Related Documentation

- [Gauge Commands](../../reference/testing/gauge-commands.md) - Full Gauge reference
- [Godog Commands](../../reference/testing/godog-commands.md) - Full Godog reference
- [TDD Format](../../reference/testing/tdd-format.md) - Unit test patterns
- [Three-Layer Approach](../../explanation/testing/three-layer-approach.md) - Understanding the approach
