# Godog Commands Reference

Quick reference for running Godog BDD tests.

---

## Overview

**Godog** executes BDD scenarios defined in `behavior.feature` files using Gherkin (Given/When/Then) syntax.

**Test Location**: `requirements/<module>/<feature>/behavior.feature`

---

## Basic Commands

### Run All Tests

```bash
# Run all behavior tests
godog requirements/**/behavior.feature
```

### Run Specific Module

```bash
# Run tests for specific module
godog requirements/cli/**/behavior.feature
godog requirements/vscode/**/behavior.feature
godog requirements/docs/**/behavior.feature
```

### Run Specific Feature

```bash
# Run tests for specific feature
godog requirements/cli/init_project/behavior.feature
godog requirements/vscode/commit_message/behavior.feature
```

### Run Specific Feature File (Split Features)

```bash
# Run specific sub-feature file
godog requirements/cli/module_detection/automation_module_detection.feature
godog requirements/cli/module_detection/source_module_detection.feature

# Run all sub-features for a feature
godog requirements/cli/module_detection/*.feature
```

---

## Tag-Based Filtering

### Filter by Single Tag

```bash
# Run only critical tests
godog --tags="@critical" requirements/**/behavior.feature

# Run only success scenarios
godog --tags="@success" requirements/**/behavior.feature

# Run only error scenarios
godog --tags="@error" requirements/**/behavior.feature
```

### Filter by Multiple Tags (AND)

```bash
# Run critical CLI tests
godog --tags="@success && @cli" requirements/**/behavior.feature

# Run critical success scenarios
godog --tags="@critical && @success" requirements/**/behavior.feature
```

### Filter by Multiple Tags (OR)

```bash
# Run critical OR error tests
godog --tags="@critical || @error" requirements/**/behavior.feature
```

### Exclude Tags (NOT)

```bash
# Run all except WIP tests
godog --tags="~@wip" requirements/**/behavior.feature

# Run all except integration tests
godog --tags="~@integration" requirements/**/behavior.feature
```

---

## Verification Tags (Regulatory/Audit)

### Installation Verification (IV)

```bash
# Run only installation verification scenarios
godog --tags="@IV" requirements/**/behavior.feature

# Generate separate IV report
godog --tags="@IV" --format=junit:test-results/iv-godog.xml requirements/**/behavior.feature
```

### Performance Verification (PV)

```bash
# Run only performance verification scenarios
godog --tags="@PV" requirements/**/behavior.feature

# Generate separate PV report
godog --tags="@PV" --format=junit:test-results/pv-godog.xml requirements/**/behavior.feature
```

### Operational Verification (OV - default)

```bash
# Run only operational verification scenarios (exclude IV and PV)
godog --tags="~@IV && ~@PV" requirements/**/behavior.feature

# Generate separate OV report
godog --tags="~@IV && ~@PV" --format=junit:test-results/ov-godog.xml requirements/**/behavior.feature
```

---

## Acceptance Criteria Filtering

### Filter by Acceptance Criterion

```bash
# Run scenarios for acceptance criterion 1
godog --tags="@ac1" requirements/**/behavior.feature

# Run scenarios for acceptance criterion 2
godog --tags="@ac2" requirements/**/behavior.feature

# Run scenarios for multiple criteria
godog --tags="@ac1 || @ac2" requirements/**/behavior.feature
```

---

## Report Generation

### Pretty Format (Human-Readable)

```bash
# Default pretty format
godog --format=pretty requirements/**/behavior.feature

# Pretty format is the default if no format specified
godog requirements/**/behavior.feature
```

### JUnit XML Format

```bash
# Generate JUnit XML report
godog --format=junit:test-results/godog.xml requirements/**/behavior.feature
```

### Cucumber JSON Format

```bash
# Generate Cucumber JSON report
godog --format=cucumber:test-results/godog.json requirements/**/behavior.feature
```

### Multiple Formats Simultaneously

```bash
# Generate pretty console output + JUnit XML
godog --format=pretty --format=junit:test-results/godog.xml requirements/**/behavior.feature

# Generate pretty + JUnit + Cucumber JSON
godog --format=pretty \
      --format=junit:test-results/godog.xml \
      --format=cucumber:test-results/godog.json \
      requirements/**/behavior.feature
```

### Generate Reports by Verification Type

```bash
# Installation Verification report
godog --tags="@IV" --format=junit:test-results/iv-godog.xml requirements/**/behavior.feature

# Performance Verification report
godog --tags="@PV" --format=junit:test-results/pv-godog.xml requirements/**/behavior.feature

# Operational Verification report (default scenarios)
godog --tags="~@IV && ~@PV" --format=junit:test-results/ov-godog.xml requirements/**/behavior.feature
```

---

## Scenario Execution Options

### Stop on Failure

```bash
# Stop immediately on first failure
godog --stop-on-failure requirements/**/behavior.feature
```

### Strict Mode

```bash
# Fail on undefined or pending steps
godog --strict requirements/**/behavior.feature
```

### Random Execution Order

```bash
# Run scenarios in random order
godog --random requirements/**/behavior.feature

# Run with specific random seed (for reproducibility)
godog --random --seed=12345 requirements/**/behavior.feature
```

---

## Verbosity and Output

### Verbose Output

```bash
# Show step definitions source location
godog --verbose requirements/**/behavior.feature
```

### No Colors

```bash
# Disable colored output
godog --no-colors requirements/**/behavior.feature
```

---

## Common Use Cases

### Run All Critical Tests

```bash
godog --tags="@critical" requirements/**/behavior.feature
```

### Run Tests for Specific Module

```bash
godog requirements/cli/**/behavior.feature
```

### Generate JUnit Report

```bash
godog --format=junit:test-results/godog.xml requirements/**/behavior.feature
```

### Run Success Scenarios Only

```bash
godog --tags="@success" requirements/**/behavior.feature
```

### Run Error Scenarios Only

```bash
godog --tags="@error" requirements/**/behavior.feature
```

### Run Tests Excluding WIP

```bash
godog --tags="~@wip" requirements/**/behavior.feature
```

### Generate Regulatory Reports (IV/PV/OV)

```bash
# Installation Verification
godog --tags="@IV" \
      --format=junit:test-results/iv-godog.xml \
      requirements/**/behavior.feature

# Performance Verification
godog --tags="@PV" \
      --format=junit:test-results/pv-godog.xml \
      requirements/**/behavior.feature

# Operational Verification (default)
godog --tags="~@IV && ~@PV" \
      --format=junit:test-results/ov-godog.xml \
      requirements/**/behavior.feature
```

---

## Example Output

### Successful Scenario Run

```text
Feature: Initialize project command behavior

  Scenario: Initialize in empty directory creates structure       # requirements/cli/init_project/behavior.feature:12
    Given I am in an empty folder                                 # step_definitions_test.go:15
    When I run "cc init"                                          # step_definitions_test.go:20
    Then a file named "cc.yaml" should be created                 # step_definitions_test.go:25
    And a directory named "src/" should exist                     # step_definitions_test.go:30
    And the command should exit with code 0                       # step_definitions_test.go:35

  Scenario: Initialize in existing project shows error            # requirements/cli/init_project/behavior.feature:19
    Given I am in a directory with "cc.yaml"                      # step_definitions_test.go:40
    When I run "cc init"                                          # step_definitions_test.go:20
    Then the command should fail                                  # step_definitions_test.go:45
    And stderr should contain "already initialized"               # step_definitions_test.go:50

2 scenarios (2 passed)
10 steps (10 passed)
125.456µs
```

### Failed Scenario Run

```text
Feature: Initialize project command behavior

  Scenario: Initialize in empty directory creates structure       # requirements/cli/init_project/behavior.feature:12
    Given I am in an empty folder                                 # step_definitions_test.go:15
    When I run "cc init"                                          # step_definitions_test.go:20
    Then a file named "cc.yaml" should be created                 # step_definitions_test.go:25
      Error: file "cc.yaml" does not exist
      step_definitions_test.go:28
    And a directory named "src/" should exist                     # step_definitions_test.go:30 - skipped
    And the command should exit with code 0                       # step_definitions_test.go:35 - skipped

1 scenarios (1 failed)
5 steps (2 passed, 1 failed, 2 skipped)
89.123µs

--- Failed steps:

  Scenario: Initialize in empty directory creates structure # requirements/cli/init_project/behavior.feature:12
    Then a file named "cc.yaml" should be created # requirements/cli/init_project/behavior.feature:15
      Error: file "cc.yaml" does not exist
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
    - requirements/**/behavior.feature
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
go test -v ./requirements/cli/init_project/

# Run with coverage
go test -v -coverprofile=coverage.out ./requirements/cli/init_project/
```

**Test file**: `requirements/<module>/<feature>/step_definitions_test.go`

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

- [BDD Format](./bdd-format.md) - Behavior feature format
- [ATDD Format](./atdd-format.md) - Acceptance spec format
- [Gauge Commands](./gauge-commands.md) - ATDD test commands
- [BDD Concepts](../../explanation/testing/bdd-concepts.md) - Understanding BDD
- [Testing How-to Guides](../../how-to-guides/testing/index.md) - Task-oriented guides
