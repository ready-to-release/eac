# Show Build Times

<!-- book:cmd show build-times -->

Displays build timing analysis parsed from build logs, including overall statistics, breakdown by module type, and the slowest individual builds.

## Usage

```bash
eac show build-times
```

Requires a prior `eac build` run. Reads timing data from the build output directory (default: `out/build`).

## Output Sections

The report includes three tables:

1. **Overall Summary** - Key-value table with metrics: Total Builds, Passed Builds, Failed Builds, Total Duration, Average Duration.
2. **Summary by Type** - Table with columns: Type, Builds, Passed, Failed, Total (s), Avg (s). Sorted by total duration (slowest type first).
3. **Top 20 Slowest Builds** - Table with columns: #, Duration (s), Status, Module, Type. Ranked by individual build duration.

## Examples

```bash
# Show build timing analysis after a build
eac build
eac show build-times
```

## See Also

- [get build-times](../get/build-times.md) - JSON output
- [show build-summary](./build-summary.md)
- [build](../build/build.md)
