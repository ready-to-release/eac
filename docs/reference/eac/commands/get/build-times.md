# get build-times

<!-- book:cmd get build-times -->

## Output Structure

```yaml
total_builds: 12
passed_builds: 11
failed_builds: 1
total_duration_seconds: 145.2
avg_duration_seconds: 12.1
build_output_dir: out/build
timings:
  - module: clie
    duration_seconds: 32.5
    status: PASS
    type: go
by_type:
  go:
    total_builds: 8
    total_duration_seconds: 120.0
    avg_duration_seconds: 15.0
    modules: [...]
```

## See Also

- [show build-times](../show/build-times.md) - Formatted table
- [get test-timings](./test-timings.md)
