# Run Tests

Execute tests at all three layers (ATDD, BDD, TDD).

---

## Overview

This guide covers running tests for:

- **ATDD/BDD** - Acceptance criteria (Rules) and behavior scenarios with Godog (unified)
- **TDD** - Unit tests with Go test

---

## Prerequisites

- [Godog configured](./setup-godog.md)
- Test files created (see [Create Specifications](./create-specifications.md))
- Step definitions implemented in `src/<module>/tests/`

---

## Quick Reference

```bash
# Run BDD/ATDD tests (from src/)
cd src/<module>/tests
go test -v

# Run all tests from project root
go test -v ./src/...

# Run specific module's tests
go test -v ./src/commands/tests
go test -v ./src/cli/tests

# Run with coverage
cd src/<module>/tests
go test -v -cover

# Run TDD unit tests
go test ./src/<module>
```

---

## ATDD/BDD: Run Godog Tests

**Note**: Godog executes both ATDD (Rule blocks) and BDD (Scenario blocks) from the same specification files.

### Run All Tests for a Module

```bash
cd src/<module>/tests
go test -v
```

**Example**:

```bash
cd src/commands/tests
go test -v
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
gauge run specs/cli/
gauge run specs/vscode/
```

### Run Specific Feature

```bash
gauge run specs/cli/init_project/
```

### Run with Tags

```bash
# Run only critical tests
gauge run --tags "critical" specs/

# Run only performance tests
gauge run --tags "performance" specs/

# Run critical CLI tests (AND)
gauge run --tags "cli & critical" specs/

# Exclude WIP tests (NOT)
gauge run --tags "!wip" specs/
```

### Generate HTML Report

```bash
gauge run --html-report specs/
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
gauge run -p=4 specs/

# Use all available cores
gauge run -p specs/
```

### Validate Without Running

```bash
# Check specs are valid
gauge validate specs/

# Validate specific feature
gauge validate specs/cli/init_project/
```

---

## BDD: Run Godog Tests

### Run All Behavior Tests

```bash
godog specs/**/behavior.feature
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
godog specs/cli/**/behavior.feature
godog specs/vscode/**/behavior.feature
```

### Run Specific Feature

```bash
godog specs/cli/init_project/behavior.feature
```

### Run with Tags

```bash
# Run only success scenarios
godog --tags="@success" specs/**/behavior.feature

# Run only error scenarios
godog --tags="@error" specs/**/behavior.feature

# Run critical tests (AND)
godog --tags="@success && @cli" specs/**/behavior.feature

# Exclude WIP tests (NOT)
godog --tags="~@wip" specs/**/behavior.feature

# Run specific acceptance criterion
godog --tags="@ac1" specs/**/behavior.feature
```

### Run by Verification Type

```bash
# Installation Verification only
godog --tags="@IV" specs/**/behavior.feature

# Performance Verification only
godog --tags="@PV" specs/**/behavior.feature

# Operational Verification only (default scenarios)
godog --tags="~@IV && ~@PV" specs/**/behavior.feature
```

### Generate Reports

#### JUnit XML (for CI/CD)

```bash
godog --format=junit:test-results/godog.xml specs/**/behavior.feature
```

#### Cucumber JSON

```bash
godog --format=cucumber:test-results/godog.json specs/**/behavior.feature
```

#### Multiple Formats

```bash
godog --format=pretty \
      --format=junit:test-results/godog.xml \
      --format=cucumber:test-results/godog.json \
      specs/**/behavior.feature
```

#### Separate Reports by Verification Type

```bash
# Installation Verification report
godog --tags="@IV" --format=junit:test-results/iv-godog.xml specs/**/behavior.feature

# Performance Verification report
godog --tags="@PV" --format=junit:test-results/pv-godog.xml specs/**/behavior.feature

# Operational Verification report
godog --tags="~@IV && ~@PV" --format=junit:test-results/ov-godog.xml specs/**/behavior.feature
```

### Run via Go Test

```bash
# Run as Go test
cd specs/cli/init_project
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
gauge run specs/

echo "Running BDD behavior tests..."
godog specs/**/behavior.feature

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
gauge run --html-report specs/
echo

echo "=== BDD: Godog Behavior Tests ==="
godog --format=pretty \
      --format=junit:test-results/godog.xml \
      specs/**/behavior.feature
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
        run: gauge run --html-report specs/

      - name: Run BDD Tests (Godog)
        run: |
          godog --format=pretty \
                --format=junit:test-results/godog.xml \
                specs/**/behavior.feature

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

## Related Documentation

- [Godog Commands](../../reference/specifications/godog-commands.md) - Full Godog reference
- [TDD Format](../../reference/specifications/tdd-format.md) - Unit test patterns
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - Understanding the approach
