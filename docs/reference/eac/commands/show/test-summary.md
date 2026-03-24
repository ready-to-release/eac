# show test-summary

<!-- book:cmd show test-summary -->

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

## See Also

- [test](../test/test.md) - Run tests
- [show test-timings](./test-timings.md) - Performance analysis
- [test debug](../test/debug.md) - Debug failures
