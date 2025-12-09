# Debug Test Failures

{{ page_breadcrumb() }}

## What You'll Accomplish

Identify and understand test failures with detailed output and timing information.

## Prerequisites

- Tests have been run
- Some tests are failing

## Steps

### 1. List All Failures

```bash
r2r eac test debug
```

**What happens**: Parses test results and lists all failures with details

### 2. Check Test Timings

```bash
r2r eac show test-timings src-auth
```

**What happens**: Shows which tests are slowest, may indicate issues

### 3. Run Specific Failing Test

```bash
r2r eac test src-auth --verbose --fail-fast
```

**What happens**:

- Verbose output shows detailed test execution
- Stops at first failure for focused debugging

### 4. Run Single Test

```bash
# Run specific test by name
go test -v ./src/auth -run TestLoginFails
```

**What happens**: Runs only the failing test for quick iteration

## Example Scenario

Your tests are failing after auth changes:

```bash
# See all failures
r2r eac test debug

# Output:
# Failed Tests:
# ✗ TestLogin (src-auth)
#   Error: expected status 200, got 401
#   File: src/auth/login_test.go:45
#
# ✗ TestRefreshToken (src-auth)
#   Error: token expired
#   File: src/auth/token_test.go:23

# Check if timeout issue
r2r eac show test-timings src-auth
# TestRefreshToken: 31.2s (timeout?)

# Run single test with verbose
go test -v ./src/auth -run TestLogin
# Shows detailed output for debugging

# Fix issue and retest
r2r eac test src-auth --fail-fast
# ✓ All tests pass
```

## Common Issues

| Problem | Solution |
|---------|----------|
| Flaky tests | Check for timing/race conditions |
| Timeouts | Increase timeout or optimize test |
| Setup failures | Verify test fixtures and data |

## Next Steps

- [Run Tests for Module](./run-tests-for-module.md) → Verify fixes

## Related Commands

- [`test debug`](../../../reference/commands/test/debug.md) - List failures
- [`show test-timings`](../../../reference/commands/show/test-timings.md) - Test performance
- [`test`](../../../reference/commands/test/test.md) - Run tests

{{ diataxis_footer() }}
