# Show test-summary

<!-- book:cmd show test-summary -->

Generate a test summary for one or all modules, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Reads test results from UoW manifests and shows pass/fail/skip counts, individual test results grouped by unit tests and BDD features, and per-module breakdowns.

## Usage

```
eac show test-summary [module] [suite] [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module name (optional -- omit for all modules) |
| `suite` | Test suite name (optional -- omit for all suites) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--status` | string | Test status override (`success` or `failure`) |
| `--run-id` | string | GitHub Actions run ID for linking to workflow |

## Output Sections

**All-modules mode** (no arguments):
- **Header**: overall pass/fail with aggregate counts (modules, tests, passed, failed, skipped)
- **Per-module section**: each module gets its own heading with counts and test tables
  - **Unit Tests**: sorted by name with status and duration
  - **Features**: grouped by feature name, each with scenario count and pass/fail breakdown

**Single-module mode**:
- **Header**: module name with test emoji and suite indicator
- **Status**: pass/fail message
- **Test Metrics** (success) or **Diagnostics** (failure): last 100 lines of test log
- **Test Configuration** (collapsible): suite and module details

## Examples

```bash
# Summary for all modules with test results
eac show test-summary

# Summary for a specific module
eac show test-summary core

# Summary for a specific module and suite
eac show test-summary core L0

# Redirect to GitHub Actions step summary
eac show test-summary core >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [test](../test/test.md) - Run tests
- [show test-timings](./test-timings.md) - Performance analysis
- [test debug](../test/debug.md) - Debug failures
