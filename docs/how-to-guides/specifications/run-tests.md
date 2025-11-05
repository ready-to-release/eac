# Run Tests

Execute tests at all three layers (ATDD, BDD, TDD).

---

## Overview

This guide covers running tests for:

- **ATDD/BDD** - Acceptance criteria (Rules) and behavior scenarios with Godog (unified in `specification.feature`)
- **TDD** - Unit tests with Go test

---

## Prerequisites

- Go installed (version 1.21+)
- Test files created (see [Create Specifications](./create-specifications.md))
- Step definitions implemented in `src/<module>/tests/steps_test.go`

---

## Quick Reference

```bash
# Run ATDD/BDD tests for a module (from project root)
go test -v ./src/<module>/tests

# Run all ATDD/BDD tests
go test -v ./src/*/tests

# Run with coverage
go test -v -cover ./src/<module>/tests

# Run TDD unit tests
go test ./src/<module>

# Run all tests (ATDD/BDD + TDD)
go test -v ./src/...
```

---

## ATDD/BDD: Run Godog Tests

**Note**: Godog executes both ATDD (Rule blocks) and BDD (Scenario blocks) from the same `specification.feature` files.

### Understanding the Test Structure

**Specifications** (what to test):

- Located in: `specs/<module>/<feature>/specification.feature`
- Contains: Feature, Rules (ATDD), and Scenarios (BDD)

**Test Implementations** (how to test):

- Located in: `src/<module>/tests/steps_test.go`
- Contains: Step definitions that implement the scenarios

**Tests run from**: `src/<module>/tests` directory using `go test`

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
Feature: AI Commit Message Generation
  As a developer
  I want to generate commit messages from staged changes
  So that I can maintain consistent commit message quality

  Rule: Generated commit messages must follow semantic commit format

    Scenario: Generate commit message from staged changes         # specs/src-commands/ai-commit-generation/specification.feature:15
      Given I have staged changes in the repository
      When I run the commit generation command
      Then a commit message should be generated
      And the message should follow semantic commit format

2 scenarios (2 passed)
8 steps (8 passed)
```

### Run Tests from Project Root

```bash
# Run specific module's tests
go test -v ./src/commands/tests
go test -v ./src/cli/tests

# Run all ATDD/BDD tests
go test -v ./src/*/tests
```

### Run with Tags

Tags in the `specification.feature` files can filter which scenarios run:

```bash
# Run only success scenarios
cd src/<module>/tests
go test -v -godog.tags="@success"

# Run only error scenarios
go test -v -godog.tags="@error"

# Run critical tests
go test -v -godog.tags="@critical"

# Run specific acceptance criterion scenarios
go test -v -godog.tags="@ac1"

# Run by module tag
go test -v -godog.tags="@cli"

# Combine tags with AND
go test -v -godog.tags="@success && @cli"

# Exclude WIP tests (NOT)
go test -v -godog.tags="~@wip"
```

### Run by Verification Type

```bash
# Installation Verification only
go test -v -godog.tags="@IV"

# Performance Verification only
go test -v -godog.tags="@PV"

# Operational Verification only (default, no tags)
go test -v -godog.tags="~@IV && ~@PV"
```

### Run with Coverage

```bash
cd src/<module>/tests
go test -v -cover

# Generate coverage profile
go test -v -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Generate Reports

#### JUnit XML (for CI/CD)

```bash
cd src/<module>/tests
go test -v -godog.format=junit > test-results/godog.xml
```

#### Multiple Output Formats

```bash
cd src/<module>/tests
go test -v -godog.format=pretty -godog.format=junit:test-results/godog.xml
```

#### Separate Reports by Verification Type

```bash
cd src/<module>/tests

# Installation Verification report
go test -v -godog.tags="@IV" -godog.format=junit:test-results/iv-godog.xml

# Performance Verification report
go test -v -godog.tags="@PV" -godog.format=junit:test-results/pv-godog.xml

# Operational Verification report
go test -v -godog.tags="~@IV && ~@PV" -godog.format=junit:test-results/ov-godog.xml
```

### Run Specific Feature

While specifications live in `specs/`, tests run from `src/`:

```bash
# Run tests for a specific module (which reads its specs automatically)
cd src/commands/tests
go test -v

# The test runner automatically finds specs at:
# specs/src-commands/*/specification.feature
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
# Find all tests for a feature by Feature ID
grep -r "Feature: cli_init-project" src/

# Run the tests for that module
go test -v ./src/cli/tests
```

---

## Run All Tests (All Layers)

### Sequential Execution

```bash
#!/bin/bash
# Run all tests in sequence

echo "Running ATDD/BDD tests (Godog via go test)..."
go test -v ./src/*/tests

echo "Running TDD unit tests..."
go test -v ./src/...

echo "All tests completed!"
```

### Create Test Script

**File**: `scripts/run-all-tests.sh`

```bash
#!/bin/bash
set -e

echo "=== Running All Tests ==="
echo

echo "=== ATDD/BDD: Godog Tests (via go test) ==="
go test -v ./src/*/tests
echo

echo "=== TDD: Go Unit Tests ==="
go test -v -coverprofile=coverage.out ./src/...
go tool cover -func=coverage.out
echo

echo "=== Test Summary ==="
echo "Coverage report: coverage.out"
echo "View coverage: go tool cover -html=coverage.out"
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

      - name: Run ATDD/BDD Tests (Godog via go test)
        run: go test -v ./src/*/tests

      - name: Run TDD Tests (Go)
        run: go test -v -coverprofile=coverage.out ./src/...

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

      - name: Generate Test Reports
        if: always()
        run: |
          # Generate JUnit reports for ATDD/BDD tests
          go test -v -godog.format=junit:test-results/atdd-bdd.xml ./src/*/tests

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

- [Create Specifications](./create-specifications.md) - How to write specification files
- [Gherkin Format](../../reference/specifications/gherkin-format.md) - Feature file syntax
- [TDD Format](../../reference/specifications/tdd-format.md) - Unit test patterns
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - Understanding ATDD/BDD/TDD
