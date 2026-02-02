# get test-results

<!-- book:cmd get test-results -->

## Output Structure

Returns structured data including:

- `modules_tested` - Number of modules with test results
- `last_run` - Timestamp of most recent test run
- `total_tests` - Total test count
- `total_passed` - Total passed count
- `total_failed` - Total failed count
- `tests[]` - Individual test results with status, tags, duration
- `spec_coverage[]` - Godog specification coverage
- `control_summary[]` - Aggregated control coverage
- `module_stats[]` - Per-module statistics
- `summary_by_type[]` - Aggregations by test type
- `summary_by_suite[]` - Aggregations by suite

## Tags Processing

Tags are returned without the @ prefix:

- `["L1", "ov", "control:ai-2"]` instead of `["@L1", "@ov", "@control:ai-2"]`

## Prerequisites

Test manifests must exist from previous test runs:

```bash
# Run tests first to generate manifests
eac test <module>

# Then get results
eac get test-results
```

## See Also

- [show test-results](../show/test-results.md) - Human-readable display
- [get tests](./tests.md) - Test definitions
- [test](../test/test.md) - Run tests
