# Run Test Suites

{{ page_breadcrumb() }}

## What You'll Accomplish

Execute specific test suites (unit, integration, acceptance) for targeted testing.

## Prerequisites

- Test suites configured in repository
- Tests tagged appropriately

## Steps

### 1. List Available Suites

```bash
r2r eac test list-suites
```

**What happens**: Shows all configured test suites with descriptions

### 2. View Suite Details

```bash
r2r eac show suite integration
```

**What happens**: Displays what tests are in the suite, which modules

### 3. Run Specific Suite

```bash
r2r eac test suite integration
```

**What happens**: Executes all tests in the integration suite

### 4. View Suite Results

```bash
r2r eac show test-summary --suite integration
```

**What happens**: Shows pass/fail summary for suite execution

## Suite Types

Common test suites:

- **unit** - Fast unit tests only
- **integration** - Integration tests (may require services)
- **acceptance** - End-to-end BDD tests
- **smoke** - Quick smoke tests
- **regression** - Full regression suite

## Example Scenario

Running integration tests before merge:

```bash
# See what suites exist
r2r eac test list-suites
# Available suites:
# - unit: Fast unit tests
# - integration: Integration tests with database
# - acceptance: End-to-end tests
# - smoke: Quick health checks

# Check what's in integration suite
r2r eac show suite integration
# Suite: integration
# Modules: src-auth, src-api, src-db
# Tests: 45 tests
# Estimated time: 2m 30s

# Run integration tests
r2r eac test suite integration
# Running integration suite...
# ✓ 43 passed, ✗ 2 failed

# View results
r2r eac show test-summary --suite integration
```

## CI Pipeline Example

```bash
# Quick feedback
r2r eac test suite unit

# If unit passes, run integration
r2r eac test suite integration

# If all pass, run acceptance
r2r eac test suite acceptance
```

## Common Issues

| Problem | Solution |
|---------|----------|
| Suite not found | Run `test list-suites` to see available |
| Tests timeout | Suite may need more resources |
| Flaky tests | Check integration test stability |

## Next Steps

- [Debug Test Failures](./debug-test-failures.md) → Fix failures

## Related Commands

- [`test suite`](../../../../reference/commands/test/suite.md) - Run test suite
- [`test list-suites`](../../../../reference/commands/test/list-suites.md) - List suites
- [`show suite`](../../../../reference/commands/show/suite.md) - Suite details

{{ diataxis_footer() }}
