# test

<!-- book:cmd test -->

## Language Support

The test command supports multiple test frameworks:

- **Go** - `gotest` (unit tests), `godog` (BDD/Gherkin)
- **TypeScript** - `mocha` (unit tests), `cucumber-js` (BDD/Gherkin)

Tests are discovered by file patterns (`*_test.go`, `*.test.ts`, `*.feature`).

## Test Suites

Use composite suites with `+` to run multiple suites in a single pass:

```bash
eac test eac-commands --suite unit+integration
```

Test results output to: `out/test/<module>/`

## See Also

- [test suite](./suite.md) - Run test suites
- [test debug](./debug.md) - Debug failures
- [show tests](../show/tests.md) - View test assertions
