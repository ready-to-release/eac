# Gauge Commands Reference

Quick reference for running Gauge acceptance tests.

---

## Overview

**Gauge** executes acceptance tests defined in `acceptance.spec` files using Markdown-based specifications.

**Test Location**: `specs/<module>/<feature>/acceptance.spec`

---

## Basic Commands

### Run All Tests

```bash
# Run all acceptance tests
gauge run specs/
```

### Run Specific Module

```bash
# Run tests for specific module
gauge run specs/cli/
gauge run specs/vscode/
gauge run specs/docs/
```

### Run Specific Feature

```bash
# Run tests for specific feature
gauge run specs/cli/init_project/
gauge run specs/vscode/commit_message/
```

---

## Tag-Based Filtering

### Filter by Single Tag

```bash
# Run only critical tests
gauge run --tags "critical" specs/

# Run only performance tests
gauge run --tags "performance" specs/

# Run only cross-platform tests
gauge run --tags "cross-platform" specs/
```

### Filter by Multiple Tags (AND)

```bash
# Run critical CLI tests
gauge run --tags "cli & critical" specs/

# Run critical performance tests
gauge run --tags "critical & performance" specs/
```

### Filter by Multiple Tags (OR)

```bash
# Run critical OR performance tests
gauge run --tags "critical | performance" specs/
```

### Exclude Tags (NOT)

```bash
# Run all except WIP tests
gauge run --tags "!wip" specs/

# Run all except integration tests
gauge run --tags "!integration" specs/
```

---

## Report Generation

### HTML Report

```bash
# Generate HTML report
gauge run --html-report specs/

# HTML report location: test-results/gauge/html-report/index.html
```

### Multiple Report Formats

```bash
# Generate HTML and XML reports
gauge run --html-report --xml-report specs/
```

---

## Parallel Execution

### Run Tests in Parallel

```bash
# Run with 4 parallel streams
gauge run -p=4 specs/

# Run with maximum available cores
gauge run -p specs/
```

**Note**: Parallel execution requires thread-safe step implementations.

---

## Verbosity and Output

### Verbose Output

```bash
# Show detailed step execution
gauge run --verbose specs/
```

### Simple Console Output

```bash
# Minimal console output
gauge run --simple-console specs/
```

---

## Validation and Debugging

### Validate Specifications

```bash
# Validate specs without running
gauge validate specs/

# Validate specific feature
gauge validate specs/cli/init_project/
```

### List Specifications

```bash
# List all scenarios
gauge list specs/

# List scenarios with tags
gauge list --tags "critical" specs/
```

---

## Environment Configuration

### Set Environment

```bash
# Run with specific environment
gauge run --env staging specs/

# Environment config location: env/staging/default.properties
```

### Override Properties

```bash
# Override specific property
gauge run -e "test_url=http://localhost:8080" specs/
```

---

## Common Use Cases

### Run Critical Tests Only

```bash
gauge run --tags "critical" specs/
```

### Run Tests for Specific Module

```bash
gauge run specs/cli/
```

### Generate Report and Run in Parallel

```bash
gauge run -p=4 --html-report specs/
```

### Run Tests with Verbose Output

```bash
gauge run --verbose specs/cli/init_project/
```

### Validate Before Running

```bash
# Validate all specs
gauge validate specs/

# If valid, run tests
gauge run specs/
```

---

## Example Output

### Successful Test Run

```text
# Initialize Project

## Acceptance Tests

  ### AC1: Creates project directory structure     ✓

  ### AC2: Generates valid configuration file     ✓

  ### AC3: Command completes in under 2 seconds   ✓


Successfully generated html-report to => test-results/gauge/html-report/index.html
Specifications: 1 executed     1 passed     0 failed     0 skipped
Scenarios:      3 executed     3 passed     0 failed     0 skipped

Total time taken: 1.234s
```

### Failed Test Run

```text
# Initialize Project

## Acceptance Tests

  ### AC1: Creates project directory structure     ✓

  ### AC2: Generates valid configuration file     ✗

    Step: Verify YAML contains key "version"
    Error: YAML does not contain key "version"
    File: specs/cli/init_project/acceptance_test.go:123


Specifications: 1 executed     0 passed     1 failed     0 skipped
Scenarios:      3 executed     2 passed     1 failed     0 skipped

Total time taken: 1.234s
```

---

## Report Locations

After running with `--html-report`, reports are generated in:

```text
test-results/gauge/
├── html-report/
│   └── index.html          # HTML report
└── xml-report/
    └── result.xml           # JUnit XML report
```

---

## Gauge Configuration

### Default Configuration

**File**: `.gauge/gauge.properties` (project root)

```properties
# Default settings
screenshot_on_failure = true
enable_multithreading = false
gauge_reports_dir = test-results/gauge
```

### Environment-Specific Configuration

**File**: `env/<environment>/default.properties`

```properties
# Staging environment
test_url = https://staging.example.com
test_timeout = 30
```

---

## Related Documentation

- [ATDD Format](./atdd-format.md) - Acceptance spec format
- [ATDD Concepts](../../explanation/testing/atdd-concepts.md) - Understanding ATDD
- [Godog Commands](./godog-commands.md) - BDD test commands
- [Testing How-to Guides](../../how-to-guides/testing/index.md) - Task-oriented guides
