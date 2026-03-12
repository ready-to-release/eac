# Show Test Timings

<!-- book:cmd show test-timings -->

Displays test timing analysis parsed from test logs, including overall statistics, breakdown by module, and the slowest individual test scenarios.

## Usage

```bash
eac show test-timings
```

Requires a prior `eac test` run. Reads timing data from the test output directory (default: `out/test`).

## Output Sections

1. **Overall Summary** - Key-value table with metrics: Total Tests, Passed Tests, Failed Tests, Test Duration, Average Duration. If wall-clock time is available, also shows Setup/Overhead and Wall-Clock Time.
2. **Summary by Module** - Table with columns: Module, Tests, Passed, Failed, Total (s), Avg (s). Sorted by total duration (slowest module first).
3. **Top 20 Slowest Tests** - Table with columns: #, Duration (s), Status, Module, Scenario. Ranked by individual test duration.

## Examples

```bash
# Show test timing analysis after running tests
eac test
eac show test-timings
```

## See Also

- [get test-timings](../get/test-timings.md) - JSON output
- [show test-summary](./test-summary.md)
- [test](../test/test.md)
