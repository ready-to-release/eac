# Run Test Suites

## What You'll Accomplish

Execute specific test suites (commit, integration, acceptance, production-verification) for targeted testing.

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

Standard test suites (defined in `.r2r/eac/test-suites.yml`):

- **unit** - L0-L1 fast unit tests (pre-commit, <5 min)
- **integration** - L2 Docker-based emulated tests (<15 min)
- **acceptance** - L3 production-like tests in PLTE (1-2 hours)
- **production-verification** - L4 production smoke tests (continuous)

## Example Scenario

Running integration tests before merge:

```bash
# See what suites exist
r2r eac test list-suites
# Available suites:
# - unit: L0-L1 fast unit tests
# - integration: L2 Docker-based tests
# - acceptance: L3 production-like tests
# - production-verification: L4 production smoke tests

# Check what's in integration suite
r2r eac show suite integration
# Suite: integration
# Tags: @L2 (excludes @L0, @L1, @L3, @L4)
# Tests: 45 tests
# Estimated time: 10m

# Run integration tests
r2r eac test suite integration
# Running integration suite...
# ✓ 43 passed, ✗ 2 failed

# View results
r2r eac show test-summary --suite integration
```

## Running Multiple Suites at Once

Use composite suites with `+` to run multiple suites in a single pass:

```bash
# Run all suites (single init, single summary)
r2r eac test --suite unit+integration+acceptance

# Run unit and integration for a specific module
r2r eac test src-auth --suite unit+integration
```

**What happens**:

- Single initialization phase
- Tests output to module folder (`out/test/<module>/`)
- Single summary showing results from all suites

This is useful for local development when you want comprehensive test coverage without running separate commands.

## CI Pipeline Example

```bash
# PR workflow: quick feedback (unit + integration)
r2r eac test --suite unit+integration

# Main branch: comprehensive testing
r2r eac test --suite unit+integration+acceptance
```

<!-- markdownlint-disable MD046 -->

!!! note "CI vs Local"

    CI pipelines typically run unit+integration for PRs and add acceptance for main branch commits.
    Use composite suites for convenience.

<!-- markdownlint-enable MD046 -->

## Common Issues

| Problem         | Solution                                |
| --------------- | --------------------------------------- |
| Suite not found | Run `test list-suites` to see available |
| Tests timeout   | Suite may need more resources           |
| Flaky tests     | Check integration test stability        |

## Next Steps

- [Debug Test Failures](./debug-test-failures.md) → Fix failures

## Related Commands

- [`test suite`](../../../../reference/eac/commands/test/suite.md) - Run test suite
- [`test list-suites`](../../../../reference/eac/commands/test/list-suites.md) - List suites
- [`show suite`](../../../../reference/eac/commands/show/suite.md) - Suite details
