# Test Commands

## Overview

Test commands execute and manage tests across the repository.

They provide parallel test execution, suite-based organization, and comprehensive failure diagnostics for quality assurance.

**Key Characteristics**:

- Parallel test execution by default
- Suite-based test organization
- Module-aware test discovery
- Detailed failure diagnostics
- Integration with BDD/Gherkin specs

**When to use**: When running tests during development, in CI/CD pipelines, or debugging test failures.

## All Test Commands

<!-- book:category-commands test -->

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
eac test

# Test specific module
eac test src-auth

# Test multiple modules
eac test src-auth src-api
```

### Running Test Suites

```bash
# List available suites
eac test list-suites

# Run specific suite
eac test suite integration

# Run suite with verbose output
eac test suite integration --verbose
```

### Debugging Test Failures

```bash
# Run tests and capture failures
eac test src-auth

# Parse failures for debugging
eac test debug

# Show detailed failure information
eac show test-summary src-auth
```

### CI/CD Integration

```bash
# Run tests in CI (unit + integration for PR, add acceptance for main)
eac test --suite unit+integration

# Run specific suite for stage
eac test --suite acceptance

# Check test results
eac show test-timings
```

## Test Execution Modes

### Sequential Execution

Run tests one at a time:

```bash
eac test src-auth --sequential
```

**Use when**:

- Debugging test failures
- Tests have resource contention
- Need predictable execution order

### Parallel Execution (Default)

Run tests concurrently:

```bash
eac test src-auth src-api
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

**Run with**: `eac test <module>`

### BDD/Gherkin Tests

Godog framework for behavior-driven development:

```gherkin
Feature: User Authentication
  Scenario: Valid login
    Given a registered user
    When they provide valid credentials
    Then authentication succeeds
```

**Run with**: `eac test <module>` (automatically detected)

### Integration Tests

Tests spanning multiple modules:

```bash
# Run integration suite
eac test suite integration
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
eac test src-auth

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
eac test src-auth --verbose

# Shows:
# - Individual test output
# - Step-by-step execution for BDD tests
# - Stack traces for failures
# - Resource usage
```

### JSON Output

```bash
eac get tests | jq '.'

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
eac show test-timings

# Get slowest tests
eac get test-timings | jq '[.tests[]] | sort_by(.duration) | reverse | .[0:10]'
```

### Optimizing Test Execution

```bash
# Run only fast tests
eac test suite unit

# Skip slow integration tests
eac test --exclude-tags @slow

# All suites (parallel is default)
eac test --suite unit+integration+acceptance
```

## Common Patterns

### Pre-commit Testing

```bash
# Test affected modules
CHANGED=$(eac get changed-modules | jq -r '.changed_modules[]')
eac test $CHANGED
```

### CI Pipeline Testing

```bash
# Full test suite with coverage
eac test --suite unit+integration+acceptance --coverage

# Acceptance tests for thorough testing
eac test --suite acceptance

# Integration tests with coverage
eac test --suite integration --coverage
```

### Test-Driven Development

```bash
# Watch and re-run tests
while true; do
  eac test src-auth
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
eac test --tags @smoke

# Exclude slow tests
eac test --exclude-tags @slow

# Multiple tags (AND)
eac test --tags @api,@positive
```

## Failure Handling

### Immediate Failure

Stop on first failure:

```bash
eac test --fail-fast
```

### Collect All Failures

Continue running all tests:

```bash
eac test --continue-on-error
```

### Debug Failures

```bash
# Parse test output for failures
eac test debug

# Show only failures
eac test src-auth | grep "✗"

# Get detailed failure report
eac show test-summary src-auth
```

## Test Coverage

### Generate Coverage

```bash
# Run with coverage
eac test src-auth --coverage

# View coverage report
go tool cover -html=coverage.out
```

### Coverage Thresholds

```bash
# Enforce minimum coverage
eac test --coverage --coverage-threshold 80
```

## Integration with Other Commands

### Build and Test

```bash
# Build before testing
eac build src-auth && eac test src-auth
```

### Test in Clean Environment

```bash
# Create workspace for testing
eac work create test/feature-x
cd ../work/test-feature-x

# Run tests in isolation
eac test
```

### Validate Before Commit

```bash
# Validate changes
eac validate

# Run tests
eac test

# Commit if passing
eac work commit --all
```

## Best Practices

### Test Frequently

```bash
# Test during development
# After each significant change
eac test
```

### Use Appropriate Granularity

```bash
# Development: Test single module
eac test src-auth

# Pre-commit: Test changed modules
eac test $(eac get changed-modules | jq -r '.changed_modules[]')

# CI: Test all modules with all suites
eac test --suite unit+integration+acceptance
```

### Organize with Suites

```bash
# Fast feedback with unit tests
eac test --suite unit

# Comprehensive validation with all suites
eac test --suite unit+integration+acceptance
```

### Monitor Performance

```bash
# Track test times
eac show test-timings

# Identify slow tests
eac get test-timings | jq '[.tests[]] | sort_by(.duration) | reverse'
```

## Common Issues

### Tests Fail Locally but Pass in CI

**Problem**: Environment differences

**Solution**: Use consistent test environment

```bash
# Run tests in Docker (if configured)
docker run -v $(pwd):/app test-image eac test

# Or use workspace isolation
eac work create test/investigation
cd ../work/test-investigation
eac test
```

### Flaky Tests

**Problem**: Tests pass/fail intermittently

**Solution**: Run multiple times and analyze

```bash
# Run test multiple times
for i in {1..10}; do
  eac test src-auth
done

# Identify flaky tests
eac test debug
```

### Slow Test Suite

**Problem**: Tests take too long

**Solution**: Analyze and optimize

```bash
# Find slowest tests
eac show test-timings

# Run only fast tests during development
eac test --suite unit

# Run all suites (parallel is default)
eac test --suite unit+integration+acceptance
```

### Test Discovery Issues

**Problem**: Tests not being found

**Solution**: Verify module contracts

```bash
# Check module configuration
eac show modules

# Validate contracts
eac validate
```

## Command Details

### test

Execute tests for one or more modules:

```bash
# Single module
eac test src-auth

# Multiple modules
eac test src-auth src-api

# All modules (no args = all modules)
eac test

# With specific suites and coverage
eac test src-auth --suite unit+integration --coverage
```

### test suite

Run tests organized in a suite:

```bash
# List available suites
eac test list-suites

# Run specific suite
eac test suite integration

# Run with tags
eac test suite regression --tags @critical
```

### test list-suites

Display all configured test suites:

```bash
eac test list-suites

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
eac test src-auth
eac test debug

# Shows:
# - Failed test names
# - Error messages
# - File locations
# - Stack traces
```

## See Also

- [build](../build/build.md) - Build modules before testing
- [validate specs](../validate/specs.md) - Validate Gherkin specifications
- [show tests](../show/tests.md) - List all tests
- [show test-summary](../show/test-summary.md) - Test execution summary
- [show test-timings](../show/test-timings.md) - Test performance analysis
- [get tests](../get/tests.md) - Test metadata as JSON
