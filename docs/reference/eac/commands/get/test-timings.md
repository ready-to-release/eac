# Get test-timings

<!-- book:cmd get test-timings -->

Returns test timing metrics parsed from `out/test/` log files. Provides per-scenario durations, pass/fail status, and aggregated statistics by module.

## Usage

```bash
eac get test-timings [flags]
```

Requires a prior `eac test` run to populate timing data.

## Output Structure

```yaml
total_tests: 45
passed_tests: 43
failed_tests: 2
total_duration_seconds: 92.5
avg_duration_seconds: 2.05
test_output_dir: out/test
timings:
  - module: core
    package: config
    scenario: Load_default_configuration
    duration_seconds: 1.2
    status: PASS
by_module:
  core:
    module: core
    total_tests: 20
    passed_tests: 20
    failed_tests: 0
    total_duration_seconds: 45.0
    avg_duration_seconds: 2.25
    scenarios: [...]
```

Parses both JSON test events (from `go test -json`) and plain text logs as fallback.

## Examples

```bash
eac get test-timings
eac get test-timings --as-json
```

## See Also

- [show test-timings](../show/test-timings.md) - Formatted table
- [get build-times](./build-times.md)
