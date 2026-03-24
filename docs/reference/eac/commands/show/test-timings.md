# show test-timings

<!-- book:cmd show test-timings -->

## Output Sections

1. **Overall Summary** - Key-value table with metrics: Total Tests, Passed Tests, Failed Tests, Test Duration, Average Duration. If wall-clock time is available, also shows Setup/Overhead and Wall-Clock Time.
2. **Summary by Module** - Table with columns: Module, Tests, Passed, Failed, Total (s), Avg (s). Sorted by total duration (slowest module first).
3. **Top 20 Slowest Tests** - Table with columns: #, Duration (s), Status, Module, Scenario. Ranked by individual test duration.

## See Also

- [get test-timings](../get/test-timings.md) - JSON output
- [show test-summary](./test-summary.md)
- [test](../test/test.md)
