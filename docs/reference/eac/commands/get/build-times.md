# Get build-times

<!-- book:cmd get build-times -->

Returns build timing metrics parsed from UoW manifest files in `out/build/`. Provides per-module durations, pass/fail status, and aggregated statistics by module type.

## Usage

```bash
eac get build-times [flags]
```

Requires a prior `eac build` run to populate timing data.

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

## Examples

```bash
eac get build-times
eac get build-times --as-json
```

## See Also

- [show build-times](../show/build-times.md) - Formatted table
- [get test-timings](./test-timings.md)
