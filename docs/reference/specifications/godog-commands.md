# Godog Commands Reference

Quick reference for running Godog tests with go test.

---

## Overview

**Godog** is integrated with `go test` for seamless testing within the Go toolchain.

**Specifications**: `specs/<module>/<feature>/specification.feature`
**Implementations**: `src/<module>/tests/steps_test.go`

**Recommended Approach**: Use `go test` instead of the deprecated `godog` CLI for full Go tooling support.

---

## Basic Commands

### Run All Tests from Module

```bash
# Navigate to test directory
cd src/<module>/tests

# Run all tests
go test -v

# Example
cd src/commands/tests
go test -v
```

### Run All Tests from Project Root

```bash
# Run all test packages
go test -v ./src/...

# Run specific module's tests
go test -v ./src/commands/tests
go test -v ./src/cli/tests
```

### Run Specific Test Function

```bash
# Run specific test function
cd src/<module>/tests
go test -v -run TestFeatures

# From root
go test -v -run TestFeatures ./src/commands/tests
```

### Run with Coverage

```bash
# Run with coverage report
cd src/<module>/tests
go test -v -cover

# Generate coverage profile
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Tag-Based Filtering

**Note**: Tag filtering with `go test` is configured in your `godog_test.go` file using `godog.Options.Tags`.

### Configure Tag Filtering in Test Code

Edit `src/<module>/tests/godog_test.go`:

```go
func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../../../specs/module/"},
            Tags:     "@success", // Filter by tag
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}
```

### Common Tag Filters

```go
// Single tag
Tags: "@critical"
Tags: "@success"
Tags: "@error"

// Multiple tags (AND)
Tags: "@critical && @success"
Tags: "@cli && @success"

// Multiple tags (OR)
Tags: "@cli || @vscode"

// NOT operator
Tags: "~@wip"  // Exclude work-in-progress

// Complex expressions
Tags: "@success && ~@slow"
```

### Verification Tag Examples (Regulatory/Audit)

```go
// Installation Verification only
Tags: "@IV"

// Performance Verification only
Tags: "@PV"

// Operational Verification (exclude IV and PV)
Tags: "~@IV && ~@PV"
```

---

## Acceptance Criteria Filtering

### Filter by Acceptance Criterion

```bash
# Run scenarios for acceptance criterion 1
godog --tags="@ac1" specs/**/behavior.feature

# Run scenarios for acceptance criterion 2
godog --tags="@ac2" specs/**/behavior.feature

# Run scenarios for multiple criteria
godog --tags="@ac1 || @ac2" specs/**/behavior.feature
```

---

## Report Generation

### Pretty Format (Human-Readable)

```bash
# Default pretty format
godog --format=pretty specs/**/behavior.feature

# Pretty format is the default if no format specified
godog specs/**/behavior.feature
```

### JUnit XML Format

```bash
# Generate JUnit XML report
godog --format=junit:test-results/godog.xml specs/**/behavior.feature
```

### Cucumber JSON Format

```bash
# Generate Cucumber JSON report
godog --format=cucumber:test-results/godog.json specs/**/behavior.feature
```

### Multiple Formats Simultaneously

```bash
# Generate pretty console output + JUnit XML
godog --format=pretty --format=junit:test-results/godog.xml specs/**/behavior.feature

# Generate pretty + JUnit + Cucumber JSON
godog --format=pretty \
      --format=junit:test-results/godog.xml \
      --format=cucumber:test-results/godog.json \
      specs/**/behavior.feature
```

### Generate Reports by Verification Type

```bash
# Installation Verification report
godog --tags="@IV" --format=junit:test-results/iv-godog.xml specs/**/behavior.feature

# Performance Verification report
godog --tags="@PV" --format=junit:test-results/pv-godog.xml specs/**/behavior.feature

# Operational Verification report (default scenarios)
godog --tags="~@IV && ~@PV" --format=junit:test-results/ov-godog.xml specs/**/behavior.feature
```

---

## Scenario Execution Options

### Stop on Failure

```bash
# Stop immediately on first failure
godog --stop-on-failure specs/**/behavior.feature
```

### Strict Mode

```bash
# Fail on undefined or pending steps
godog --strict specs/**/behavior.feature
```

### Random Execution Order

```bash
# Run scenarios in random order
godog --random specs/**/behavior.feature

# Run with specific random seed (for reproducibility)
godog --random --seed=12345 specs/**/behavior.feature
```

---

## Verbosity and Output

### Verbose Output

```bash
# Show step definitions source location
godog --verbose specs/**/behavior.feature
```

### No Colors

```bash
# Disable colored output
godog --no-colors specs/**/behavior.feature
```

---

## Common Use Cases

### Run All Critical Tests

```bash
godog --tags="@critical" specs/**/behavior.feature
```

### Run Tests for Specific Module

```bash
godog specs/cli/**/behavior.feature
```

### Generate JUnit Report

```bash
godog --format=junit:test-results/godog.xml specs/**/behavior.feature
```

### Run Success Scenarios Only

```bash
godog --tags="@success" specs/**/behavior.feature
```

### Run Error Scenarios Only

```bash
godog --tags="@error" specs/**/behavior.feature
```

### Run Tests Excluding WIP

```bash
godog --tags="~@wip" specs/**/behavior.feature
```

### Generate Regulatory Reports (IV/PV/OV)

```bash
# Installation Verification
godog --tags="@IV" \
      --format=junit:test-results/iv-godog.xml \
      specs/**/behavior.feature

# Performance Verification
godog --tags="@PV" \
      --format=junit:test-results/pv-godog.xml \
      specs/**/behavior.feature

# Operational Verification (default)
godog --tags="~@IV && ~@PV" \
      --format=junit:test-results/ov-godog.xml \
      specs/**/behavior.feature
```

---

## Report Locations

After running with `--format=junit` or `--format=cucumber`, reports are generated in:

```text
test-results/
├── godog.xml                # JUnit XML report
├── godog.json               # Cucumber JSON report
├── iv-godog.xml             # Installation Verification report
├── pv-godog.xml             # Performance Verification report
└── ov-godog.xml             # Operational Verification report
```

---

## Godog Configuration

### Using godog.yaml

**File**: `godog.yaml` (project root)

```yaml
default:
  paths:
    - specs/**/behavior.feature
  format: pretty,junit:test-results/godog.xml
  tags: ~@wip
  strict: true
  stop-on-failure: false
```

**Running with configuration**:

```bash
# Uses godog.yaml by default
godog

# Use specific configuration
godog --config=godog.yaml
```

---

## Integration with Go Tests

### Run Godog via Go Test

```bash
# Run as Go test
go test -v ./specs/cli/init_project/

# Run with coverage
go test -v -coverprofile=coverage.out ./specs/cli/init_project/
```

**Test file**: `specs/<module>/<feature>/step_definitions_test.go`

```go
package init_project_test

import (
    "testing"
    "github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"behavior.feature"},
            Tags:     "~@wip",
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}
```

---

## Related Documentation

- [Gherkin Format](./gherkin-format.md) - Specification syntax reference
- [ATDD and BDD with Gherkin](../../explanation/specifications/atdd-bdd-with-gherkin.md) - Concepts and workflow
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - How ATDD/BDD/TDD work together
- [Testing How-to Guides](../../how-to-guides/specifications/index.md) - Task-oriented guides
