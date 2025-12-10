# Run Tests for Module

{{ page_breadcrumb() }}

## What You'll Accomplish

Execute tests for a module and view results with timing information and coverage.

## Prerequisites

### Required Knowledge
**New to testing?** Learn these concepts first:
- [Understanding Test Suites](../../../../tutorials/getting-started/understanding-test-suites.md) - Test levels and when to use each
- [Building and Testing Changes](../../../../tutorials/core-workflows/building-and-testing.md) - Efficient test workflows

### Required Setup
- Module has tests defined
- Module is built (or will be built automatically)
- Test dependencies available

## Steps

### 1. Run Tests

```bash
r2r eac test src-auth
```

**What happens**: Executes all tests for src-auth module, shows pass/fail results

### 2. View Test Summary

```bash
r2r eac show test-summary src-auth
```

**What happens**: Displays formatted summary with pass/fail counts, timing, and coverage

### 3. Check Test Timings

```bash
r2r eac show test-timings src-auth
```

**What happens**: Shows which tests are slowest, helps identify bottlenecks

## Run Specific Tests

```bash
# Run only unit tests
r2r eac test src-auth --tags unit

# Run with verbose output
r2r eac test src-auth --verbose

# Run and stop on first failure
r2r eac test src-auth --fail-fast
```

## Example Scenario

You've added login functionality and want to verify tests pass:

```bash
# Run tests
r2r eac test src-auth

# Output:
# Running tests for src-auth...
# === RUN   TestLogin
# --- PASS: TestLogin (0.05s)
# === RUN   TestLoginInvalidPassword
# --- PASS: TestLoginInvalidPassword (0.02s)
# PASS
# coverage: 85.2% of statements

# View detailed summary
r2r eac show test-summary src-auth

# Check if any tests are slow
r2r eac show test-timings src-auth
```

## Common Issues

| Problem | Solution |
|---------|----------|
| Tests fail | Run `test debug` to see failure details |
| Timeout | Increase timeout or check for infinite loops |
| Coverage too low | Add tests for uncovered code |

## Next Steps

- [Debug Test Failures](./debug-test-failures.md) → Fix failing tests
- [Run Test Suites](./run-test-suites.md) → Run integration tests

## Related Commands

- [`test`](../../../../reference/commands/test/test.md) - Full command reference
- [`show test-summary`](../../../../reference/commands/show/test-summary.md) - View summary
- [`test debug`](../../../../reference/commands/test/debug.md) - Debug failures

{{ diataxis_footer() }}
