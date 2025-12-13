# Test Commands

{{ page_breadcrumb() }}

## Overview

Test commands execute and manage tests across the repository. They provide parallel test execution, suite-based organization, and comprehensive failure diagnostics for quality assurance.

**Key Characteristics**:

- Parallel test execution by default
- Suite-based test organization
- Module-aware test discovery
- Detailed failure diagnostics
- Integration with BDD/Gherkin specs

**When to use**: When running tests during development, in CI/CD pipelines, or debugging test failures.

## All Test Commands

| Command                                    | Purpose                    | Use Case                              |
| ------------------------------------------ | -------------------------- | ------------------------------------- |
| [test](../test/test.md)                    | Run tests for modules      | Execute tests for one or more modules |
| [test suite](../test/suite.md)             | Run specific test suite    | Execute tests organized by suite      |
| [test list-suites](../test/list-suites.md) | List available suites      | Discover available test suites        |
| [test debug](../test/debug.md)             | Parse and analyze failures | Debug failed tests                    |

## Test Organization

### Module-Based Testing

Tests are organized by module, with each module containing its own test files:

```text
go/src/auth/
├── auth.go
├── auth_test.go
└── specs/
    └── authentication.feature
```

### Suite-Based Testing

Tests can also be grouped into logical suites that span multiple modules:

```yaml
# Test suite configuration
suites:
  - name: integration
    modules: [src-api, src-auth, src-database]
  - name: unit
    modules: [eac-core, eac-commands]
```

## Common Workflows

### Running Tests During Development

```bash
# Test all modules
r2r eac test

# Test specific module
r2r eac test src-auth

# Test multiple modules
r2r eac test src-auth src-api
```

### Running Test Suites

```bash
# List available suites
r2r eac test list-suites

# Run specific suite
r2r eac test suite integration

# Run suite with verbose output
r2r eac test suite integration --verbose
```

### Debugging Test Failures

```bash
# Run tests and capture failures
r2r eac test src-auth

# Parse failures for debugging
r2r eac test debug

# Show detailed failure information
r2r eac show test-summary src-auth
```

### CI/CD Integration

```bash
# Run tests in CI
r2r eac test --all --parallel

# Run specific suite for stage
r2r eac test suite smoke

# Check test results
r2r eac show test-timings
```

## Test Execution Modes

### Sequential Execution

Run tests one at a time:

```bash
r2r eac test src-auth --sequential
```

**Use when**:

- Debugging test failures
- Tests have resource contention
- Need predictable execution order

### Parallel Execution (Default)

Run tests concurrently:

```bash
r2r eac test src-auth src-api
# Runs in parallel by default
```

**Use when**:

- Running in CI/CD
- Tests are independent
- Need fast feedback

## Test Types

### Unit Tests

Go standard testing framework:

```go
func TestAuthentication(t *testing.T) {
    // Test implementation
}
```

**Run with**: `r2r eac test <module>`

### BDD/Gherkin Tests

Godog framework for behavior-driven development:

```gherkin
Feature: User Authentication
  Scenario: Valid login
    Given a registered user
    When they provide valid credentials
    Then authentication succeeds
```

**Run with**: `r2r eac test <module>` (automatically detected)

### Integration Tests

Tests spanning multiple modules:

```bash
# Run integration suite
r2r eac test suite integration
```

## Test Discovery

### Automatic Discovery

Tests are automatically discovered based on:

- `*_test.go` files for Go tests
- `*.feature` files for Gherkin specs
- Module contracts defining test locations

### Manual Suite Configuration

Define custom test suites:

```yaml
# .eac/contracts/test-suites.yml
suites:
  - name: smoke
    description: Quick smoke tests
    modules: [src-api]
    tags: ["@smoke"]

  - name: regression
    description: Full regression suite
    modules: [src-auth, src-api, src-database]
    tags: ["@regression"]
```

## Test Output

### Default Output

```bash
r2r eac test src-auth

# Output:
Testing module: src-auth
Running 15 tests...
✓ TestValidateUser (0.12s)
✓ TestAuthentication (0.34s)
✗ TestAuthorization (0.21s)

15 tests, 14 passed, 1 failed
Total time: 2.3s
```

### Verbose Output

```bash
r2r eac test src-auth --verbose

# Shows:
# - Individual test output
# - Step-by-step execution for BDD tests
# - Stack traces for failures
# - Resource usage
```

### JSON Output

```bash
r2r eac get tests | jq '.'

# Structured test metadata
{
  "tests": [
    {
      "module": "src-auth",
      "file": "auth_test.go",
      "tests": ["TestValidateUser", "TestAuthentication"]
    }
  ]
}
```

## Test Performance

### Analyzing Test Times

```bash
# Show test timing data
r2r eac show test-timings

# Get slowest tests
r2r eac get test-timings | jq '[.tests[]] | sort_by(.duration) | reverse | .[0:10]'
```

### Optimizing Test Execution

```bash
# Run only fast tests
r2r eac test suite unit

# Skip slow integration tests
r2r eac test --exclude-tags @slow

# Parallel execution for speed
r2r eac test --all --parallel
```

## Common Patterns

### Pre-commit Testing

```bash
# Test affected modules
CHANGED=$(r2r eac get changed-modules | jq -r '.changed_modules[]')
r2r eac test $CHANGED
```

### CI Pipeline Testing

```bash
# Full test suite with coverage
r2r eac test --all --parallel --coverage

# Smoke tests for quick feedback
r2r eac test suite smoke

# Regression tests before release
r2r eac test suite regression
```

### Test-Driven Development

```bash
# Watch and re-run tests
while true; do
  r2r eac test src-auth
  sleep 2
done
```

## Test Tagging

### Gherkin Tags

```gherkin
@smoke @api
Feature: API Authentication

  @positive
  Scenario: Valid credentials
    Given a registered user
    When they authenticate
    Then access is granted

  @negative @slow
  Scenario: Invalid credentials
    Given invalid credentials
    When they authenticate
    Then access is denied
```

### Running Tagged Tests

```bash
# Run smoke tests only
r2r eac test --tags @smoke

# Exclude slow tests
r2r eac test --exclude-tags @slow

# Multiple tags (AND)
r2r eac test --tags @api,@positive
```

## Failure Handling

### Immediate Failure

Stop on first failure:

```bash
r2r eac test --fail-fast
```

### Collect All Failures

Continue running all tests:

```bash
r2r eac test --continue-on-error
```

### Debug Failures

```bash
# Parse test output for failures
r2r eac test debug

# Show only failures
r2r eac test src-auth | grep "✗"

# Get detailed failure report
r2r eac show test-summary src-auth
```

## Test Coverage

### Generate Coverage

```bash
# Run with coverage
r2r eac test src-auth --coverage

# View coverage report
go tool cover -html=coverage.out
```

### Coverage Thresholds

```bash
# Enforce minimum coverage
r2r eac test --coverage --coverage-threshold 80
```

## Integration with Other Commands

### Build and Test

```bash
# Build before testing
r2r eac build src-auth && r2r eac test src-auth
```

### Test in Clean Environment

```bash
# Create workspace for testing
r2r eac work create test/feature-x
cd ../work/test-feature-x

# Run tests in isolation
r2r eac test
```

### Validate Before Commit

```bash
# Validate changes
r2r eac validate

# Run tests
r2r eac test

# Commit if passing
r2r eac work commit --all
```

## Best Practices

### Test Frequently

```bash
# Test during development
# After each significant change
r2r eac test
```

### Use Appropriate Granularity

```bash
# Development: Test single module
r2r eac test src-auth

# Pre-commit: Test changed modules
r2r eac test $(r2r eac get changed-modules | jq -r '.changed_modules[]')

# CI: Test all
r2r eac test --all
```

### Organize with Suites

```bash
# Fast feedback with unit tests
r2r eac test suite unit

# Comprehensive validation with full suite
r2r eac test suite regression
```

### Monitor Performance

```bash
# Track test times
r2r eac show test-timings

# Identify slow tests
r2r eac get test-timings | jq '[.tests[]] | sort_by(.duration) | reverse'
```

## Common Issues

### Tests Fail Locally but Pass in CI

**Problem**: Environment differences

**Solution**: Use consistent test environment

```bash
# Run tests in Docker (if configured)
docker run -v $(pwd):/app test-image r2r eac test

# Or use workspace isolation
r2r eac work create test/investigation
cd ../work/test-investigation
r2r eac test
```

### Flaky Tests

**Problem**: Tests pass/fail intermittently

**Solution**: Run multiple times and analyze

```bash
# Run test multiple times
for i in {1..10}; do
  r2r eac test src-auth
done

# Identify flaky tests
r2r eac test debug
```

### Slow Test Suite

**Problem**: Tests take too long

**Solution**: Analyze and optimize

```bash
# Find slowest tests
r2r eac show test-timings

# Run only fast tests during development
r2r eac test suite unit

# Use parallel execution
r2r eac test --all --parallel
```

### Test Discovery Issues

**Problem**: Tests not being found

**Solution**: Verify module contracts

```bash
# Check module configuration
r2r eac show modules

# Validate contracts
r2r eac validate
```

## Command Details

### test

Execute tests for one or more modules:

```bash
# Single module
r2r eac test src-auth

# Multiple modules
r2r eac test src-auth src-api

# All modules
r2r eac test --all

# With options
r2r eac test src-auth --verbose --coverage
```

### test suite

Run tests organized in a suite:

```bash
# List available suites
r2r eac test list-suites

# Run specific suite
r2r eac test suite integration

# Run with tags
r2r eac test suite regression --tags @critical
```

### test list-suites

Display all configured test suites:

```bash
r2r eac test list-suites

# Output:
# Available test suites:
# - unit: Unit tests (fast)
# - integration: Integration tests
# - smoke: Quick smoke tests
# - regression: Full regression suite
```

### test debug

Parse test output and show failures:

```bash
# After running tests
r2r eac test src-auth
r2r eac test debug

# Shows:
# - Failed test names
# - Error messages
# - File locations
# - Stack traces
```

## See Also

- [build](../other/build.md) - Build modules before testing
- [validate specs](../validate/specs.md) - Validate Gherkin specifications
- [show tests](../show/tests.md) - List all tests
- [show test-summary](../show/test-summary.md) - Test execution summary
- [show test-timings](../show/test-timings.md) - Test performance analysis
- [get tests](../get/tests.md) - Test metadata as JSON

{{ diataxis_footer() }}
