# show build-times

<!-- book:cmd show build-times -->

## Output Sections

The report includes three tables:

1. **Overall Summary** - Key-value table with metrics: Total Builds, Passed Builds, Failed Builds, Total Duration, Average Duration.
2. **Summary by Type** - Table with columns: Type, Builds, Passed, Failed, Total (s), Avg (s). Sorted by total duration (slowest type first).
3. **Top 20 Slowest Builds** - Table with columns: #, Duration (s), Status, Module, Type. Ranked by individual build duration.

## See Also

- [get build-times](../get/build-times.md) - JSON output
- [show build-summary](./build-summary.md)
- [build](../build/build.md)
