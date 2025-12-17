# Test

<!-- book:cmd test -->

## Language Support

The test command supports multiple test frameworks via language-specific runners:

- **Go** - `gotest` (unit tests), `godog` (BDD/Gherkin)
- **TypeScript** - `mocha` (unit tests), `cucumber-js` (BDD/Gherkin)

Tests are discovered by file patterns (`*_test.go`, `*.test.ts`, `*.feature`) and executed by the appropriate runner based on module type and test type. See [Language Support](../language-support.md) for details.

## Test Suites

Use composite suites with `+` to run multiple suites in a single pass. Test results are output to the module's test folder:

- `out/test/<module>/` - All test results for the module

This provides a single initialization and summary with organized output per module.

## See Also

- [test suite](./suite.md) - Run test suites
- [test debug](./debug.md) - Debug failures
- [show tests](../show/tests.md)
- [test Commands](../categories/test.md)
- [Run Test Suites](../../../how-to-guides/eac/commands/build-test-validate/run-test-suites.md) - How-to guide
