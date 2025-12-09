<!-- EDITOR
# Editor: how-to-guides/commands/areas/test-configuration.md

## Soul

Configuration reference for testing system including test suite definitions, tag contracts with pattern validation, output formats, coverage settings, and parallel execution control.

## Sections

1. Test Suite Configuration
   - Suite Definition
   - Suite Properties
2. Test Tag Configuration
   - Tag Contract
   - Pattern Tags
3. Output Format Configuration
   - Cucumber JSON (Default)
   - JUnit XML
   - Output Directory Structure
4. Coverage Configuration
   - Coverage Flags
   - Coverage Output
   - Coverage Thresholds
5. Module Type Test Configuration
   - Go Modules
   - Specification Modules
6. Environment Variables
7. CI/CD Configuration
   - GitHub Actions
   - Coverage Integration
8. Parallel Test Execution
   - Default Behavior
   - Controlling Parallelism
9. Test Filtering
   - By Suite
   - By Tag (in specs)
10. Timeout Configuration
    - Suite-level Timeout
    - Module-level Timeout
11. Troubleshooting
12. Related Documentation
-->

# Test Configuration

This guide covers configuration options for EAC's testing system, including test suites, output formats, and coverage settings.

## Test Suite Configuration

### Suite Definition

```yaml
# .r2r/eac/test-suites.yml
suites:
  - name: commit
    description: Pre-commit validation tests
    tags:
      - "@commit"
      - "@unit"
    timeout: 300
    parallel: true

  - name: integration
    description: Integration tests
    tags:
      - "@integration"
    timeout: 900
    parallel: true

  - name: e2e
    description: End-to-end tests
    tags:
      - "@e2e"
    timeout: 1800
    parallel: false

  - name: smoke
    description: Quick health checks
    tags:
      - "@smoke"
    timeout: 60
    parallel: true
```

### Suite Properties

| Property      | Type     | Description                |
| ------------- | -------- | -------------------------- |
| `name`        | string   | Suite identifier           |
| `description` | string   | Human-readable description |
| `tags`        | string[] | Gherkin tags to include    |
| `timeout`     | int      | Timeout in seconds         |
| `parallel`    | bool     | Run tests in parallel      |

## Test Tag Configuration

### Tag Contract

```yaml
# .r2r/eac/testing-tags.yml
tags:
  - tag: "@commit"
    description: "Tests that run on every commit"

  - tag: "@integration"
    description: "Integration tests requiring external services"

  - tag: "@e2e"
    description: "End-to-end tests"

  - tag: "@smoke"
    description: "Quick health check tests"

  - tag: "@slow"
    description: "Tests that take longer than 5 seconds"

skip_reasons:
  - code: "flaky"
    description: "Test is flaky and needs investigation"

  - code: "wip"
    description: "Work in progress"

  - code: "ci-only"
    description: "Only runs in CI environment"
```

### Pattern Tags

| Pattern          | Description                | Validation                        |
| ---------------- | -------------------------- | --------------------------------- |
| `@skip:<reason>` | Skip with reason code      | Against `skip_reasons`            |
| `@deps:<name>`   | Requires system dependency | Against `system-dependencies.yml` |
| `@env:<moniker>` | Requires environment       | Against `environments.yml`        |
| `@depm:<module>` | Requires module            | Against `modules.yml`             |

## Output Format Configuration

### Cucumber JSON (Default)

```json
{
  "description": "Module test results",
  "elements": [
    {
      "id": "feature-1",
      "name": "Authentication",
      "type": "scenario",
      "steps": [
        {
          "keyword": "Given",
          "name": "user exists",
          "result": {"status": "passed"}
        }
      ]
    }
  ]
}
```

### JUnit XML

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="Module Tests">
  <testsuite name="eac-commands" tests="10" failures="0" errors="0" time="2.5">
    <testcase name="TestExample" classname="commands" time="0.25"/>
  </testsuite>
</testsuites>
```

### Output Directory Structure

```text
out/
└── reports/
    ├── test-results.json       # Cucumber JSON
    ├── test-results.xml        # JUnit XML (if --as-junit)
    └── coverage/
        ├── <module>.out        # Go coverage profile
        └── <module>.html       # HTML coverage report
```

## Coverage Configuration

### Coverage Flags

| Flag             | Description                    |
| ---------------- | ------------------------------ |
| `--coverage`     | Enable coverage collection     |
| `--coverprofile` | Custom profile path (advanced) |

### Coverage Output

```bash
# Generate coverage
r2r eac test eac-core --coverage

# Output files:
# out/reports/coverage/eac-core.out   - Coverage profile
# out/reports/coverage/eac-core.html  - HTML report
```

### Coverage Thresholds

```bash
# Check coverage meets threshold
COVERAGE=$(go tool cover -func=out/reports/coverage/eac-core.out | \
  grep total | awk '{print $3}' | sed 's/%//')

if (( $(echo "$COVERAGE < 80" | bc -l) )); then
  echo "Coverage $COVERAGE% is below threshold 80%"
  exit 1
fi
```

## Module Type Test Configuration

### Go Modules

```yaml
modules:
  - moniker: eac-core
    type: go
    test:
      timeout: 300
      coverage: true
      tags:
        - "!integration"  # Exclude integration tests by default
```

### Specification Modules

```yaml
modules:
  - moniker: specs-auth
    type: specifications
    files:
      root: specs/auth
      patterns:
        - "**/*.feature"
    test:
      runner: godog
      format: cucumber
```

## Environment Variables

| Variable          | Description           | Default        |
| ----------------- | --------------------- | -------------- |
| `TEST_TIMEOUT`    | Default test timeout  | 300s           |
| `TEST_PARALLEL`   | Enable parallel tests | true           |
| `TEST_COVERAGE`   | Enable coverage       | false          |
| `TEST_OUTPUT_DIR` | Output directory      | `out/reports/` |

## CI/CD Configuration

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
        with:
          go-version: '1.21'

      - name: Run tests
        run: r2r eac test --as-junit > test-results.xml

      - name: Publish results
        uses: dorny/test-reporter@v1
        if: always()
        with:
          name: Test Results
          path: test-results.xml
          reporter: java-junit

      - name: Upload coverage
        if: success()
        run: |
          r2r eac test --coverage
          codecov -f out/reports/coverage/*.out
```

### Coverage Integration

```yaml
- name: Coverage check
  run: |
    r2r eac test --coverage
    COVERAGE=$(go tool cover -func=out/reports/coverage/eac-core.out | \
      grep total | awk '{print $3}' | sed 's/%//')
    echo "Coverage: $COVERAGE%"
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "::error::Coverage below threshold"
      exit 1
    fi
```

## Parallel Test Execution

### Default Behavior

- Single module: Sequential, verbose output
- Multiple modules: Parallel execution

### Controlling Parallelism

```bash
# Parallel (default for multiple)
r2r eac test eac-commands eac-core

# Sequential (for debugging)
for module in eac-commands eac-core; do
  r2r eac test $module
done
```

## Test Filtering

### By Suite

```bash
r2r eac test suite commit
r2r eac test suite integration
```

### By Tag (in specs)

```gherkin
@commit @auth
Feature: User Authentication

  @smoke
  Scenario: Valid login
    Given a valid user
    When they login
    Then they are authenticated
```

## Timeout Configuration

### Suite-level Timeout

```yaml
suites:
  - name: e2e
    timeout: 1800  # 30 minutes
```

### Module-level Timeout

```yaml
modules:
  - moniker: slow-tests
    test:
      timeout: 600  # 10 minutes
```

## Troubleshooting

| Issue             | Cause              | Solution                         |
| ----------------- | ------------------ | -------------------------------- |
| Tests timeout     | Long-running tests | Increase timeout in suite config |
| No coverage       | Missing flag       | Add `--coverage` flag            |
| Tag not found     | Undefined tag      | Add to `testing-tags.yml`        |
| Parallel failures | Race conditions    | Set `parallel: false`            |

## Related Documentation

- [Test Overview](test-overview.md) - Testing concepts
- [Test Commands](test-commands.md) - Command reference
- [Specifications Configuration](specifications-configuration.md) - BDD test configuration
